# Architecture

## Overview

`monitoring-platform` is an observability backend for LLM agents. It ingests OpenTelemetry traces
pushed by agent runtimes (Claude Agent SDK, LangChain, any OTel-instrumented code), maps them to
an LLM-specific domain model (runs, spans, token usage, cost), stores them in ClickHouse, and
exposes a read API consumed by the React dashboard.

## System diagram

```
                        ┌────────────────────────────────┐
                        │   Cloud IAM / API Gateway      │
                        │  (GCP IAP — phase 3+)          │
                        └───────────────┬────────────────┘
                                        │ OTLP/HTTP :4318  OTLP/gRPC :4317
                                        ▼
┌──────────────┐                ┌──────────────────┐
│  Agent SDKs  │ ─────────────▶ │    collector     │  stateless, HPA on CPU
│  (OTel any)  │                │  · receive OTLP  │
└──────────────┘                │  · stamp project │
                                │  · validate/redact│
                                │  · batch → sink  │
                                └────────┬─────────┘
                                         │  SpanSink interface
                                         │  phase 0–2: direct CH write
                                         │  phase 3+:  Kafka/Pub-Sub
                                         ▼
                                ┌──────────────────┐
                                │    ingestor      │  stateless worker
                                │  · domain map    │
                                │  · cost calc     │
                                │  · payload split │
                                └────────┬─────────┘
                                         │
              ┌──────────────────────────┼──────────────────────┐
              ▼                          ▼                       ▼
    ┌──────────────────┐      ┌──────────────────┐    ┌──────────────────┐
    │   ClickHouse     │      │    Postgres      │    │   GCS (S3)       │
    │  spans, runs     │      │  projects, evals │    │  prompt/         │
    │  (time-series)   │      │  alerts, keys    │    │  completion      │
    └────────┬─────────┘      └────────┬─────────┘    │  payloads        │
             │                         │              └──────────────────┘
             └──────────┬──────────────┘
                        ▼
              ┌──────────────────┐        ┌──────────────────┐
              │      api         │ ◀────▶ │  eval worker     │
              │  REST + gRPC     │        │  (phase 4)       │
              └────────┬─────────┘        └──────────────────┘
                       │
                       ▼
              ┌──────────────────┐
              │ monitoring-      │
              │ console (UI)     │
              └──────────────────┘

  Shared: Redis (rate limits, caches) · Workload Identity (no key files)
```

## Components

### collector

- **Role:** Public-facing OTLP ingress. The only component exposed outside the cluster.
- **Protocols:** OTLP/HTTP (`:4318`), OTLP/gRPC (`:4317`, phase 3).
- **Responsibilities:** receive spans, stamp `project_id` from the authenticated credential
  (never from the payload — this is the multi-tenancy trust boundary), validate, size-limit,
  optionally redact PII, batch, and forward to the `SpanSink`.
- **Scaling:** stateless; HPA on CPU; multiple replicas behind an NLB/GKE Gateway.
- **Phase 0–2:** `SpanSink` writes directly to ClickHouse.
- **Phase 3+:** `SpanSink` produces to Kafka/Pub-Sub; `ingestor` consumes.

### ingestor

- **Role:** Async span → storage worker. Decoupled from the write path by the `SpanSink` interface.
- **Responsibilities:** map OTel spans to domain types, calculate cost from the model pricing table,
  split large payloads to GCS, write spans + upsert run aggregates to ClickHouse.
- **Phase 2:** stub with `/healthz` — queue consumer wired in phase 3.

### api

- **Role:** Read-side service. Serves the dashboard and any external consumers.
- **Endpoints (current):**
  - `GET /healthz`
  - `GET /v1/runs` — list runs for a project, newest first
  - `GET /v1/runs/:traceID` — spans for a single trace
- **Scaling:** stateless; HPA on CPU + p95 latency.

### worker *(phase 4)*

- **Role:** Async jobs — eval scoring (LLM-as-judge, heuristics), alert rule evaluation, cost
  roll-ups.
- **Scaling:** KEDA on queue depth; separate node pool from latency-sensitive services.

### monitoring-console

- **Stack:** Vite + React 18 + TypeScript + Tailwind CSS + TanStack Query + React Router.
- **API client:** typed fetch wrapper against `api`'s REST endpoints.
- **Deployment:** static build → nginx:alpine pod; `/v1/*` proxied to `api`.

## Data model

### Span

```
project_id      String        ← stamped at ingest from credential
trace_id        String        ← 128-bit hex (from OTel)
span_id         String        ← 64-bit hex
parent_span_id  String        ← empty = root span
name            String
start_time      DateTime64(9)
end_time        DateTime64(9)
status_code     UInt8         ← 0 Unset, 1 OK, 2 Error
model           String        ← gen_ai.request.model
input_tokens    UInt32        ← gen_ai.usage.input_tokens
output_tokens   UInt32        ← gen_ai.usage.output_tokens
cost_usd        Float64       ← calculated at ingest from pricing table
attributes      String        ← all other attrs as JSON
```

### Run

An aggregated view of a complete trace — one row per `trace_id`, built from its spans.

```
project_id            String
trace_id              String
name                  String        ← root span name
start_time            DateTime64(9) ← min(span.start_time)
end_time              DateTime64(9) ← max(span.end_time)
status_code           UInt8         ← root span status
total_input_tokens    UInt32
total_output_tokens   UInt32
total_cost_usd        Float64
span_count            UInt32
```

Stored in `ReplacingMergeTree(ingest_time)` so re-ingesting spans for the same trace
updates the aggregates.

## OTel semantic conventions

All LLM-specific attributes follow the
[OpenTelemetry GenAI semantic conventions](https://opentelemetry.io/docs/specs/semconv/gen-ai/)
(`gen_ai.*` namespace). Key attributes:

| Attribute | Description |
|---|---|
| `gen_ai.system` | e.g. `anthropic`, `openai` |
| `gen_ai.request.model` | model identifier |
| `gen_ai.usage.input_tokens` | prompt tokens consumed |
| `gen_ai.usage.output_tokens` | completion tokens generated |
| `gen_ai.operation.name` | `chat`, `text_completion`, etc. |

## Storage

| Store | Engine | Used for | Retention |
|---|---|---|---|
| ClickHouse | `MergeTree` / `ReplacingMergeTree` | spans, runs | 90-day TTL + tiered S3 |
| Postgres | CloudSQL Postgres 16 | projects, eval defs, alert rules, API keys | indefinite |
| GCS | Standard → Nearline → Coldline | prompt/completion payloads | lifecycle rules |
| Redis | Memorystore | rate limits, API key cache, hot aggregates | ephemeral |

### ClickHouse schema design choices

- `PARTITION BY toYYYYMM(start_time)` — enables fast partition pruning on time range queries
  and efficient TTL expiry (drop whole partitions, not row-by-row).
- `ORDER BY (project_id, start_time, trace_id, span_id)` — primary key optimised for the most
  common access patterns: per-project time-range scans and per-trace lookups.
- `ReplacingMergeTree` for runs — handles the "streaming aggregation" pattern where spans for
  the same trace arrive in multiple batches. `FINAL` in read queries forces deduplication.

## Infrastructure (GCP)

| Concern | Service |
|---|---|
| Kubernetes | GKE Autopilot (private nodes, Workload Identity) |
| Postgres | Cloud SQL for PostgreSQL 16 (private IP, automated backups) |
| ClickHouse | Self-hosted on GKE via Altinity operator (phase 3+) or ClickHouse Cloud |
| Blob storage | GCS (payloads, TF state, backups) |
| Cache | Memorystore for Redis |
| Queue (phase 3) | Cloud Pub/Sub or MSK-compatible Kafka |
| Secrets | Secret Manager + External Secrets Operator |
| Registry | Artifact Registry (Docker images + OCI Helm charts) |
| Ingress | GKE Gateway API + Google-managed certs |
| Identity (phase 3) | IAP (console) + Cloud Endpoints (collector authN) |
| IaC | Terraform 1.10+, state in GCS |

### Terraform state layers

State is split into four independent layers per environment to limit blast radius:

```
infrastructure/terraform/envs/{dev,staging,prod}/
├── 00-network/     VPC, subnets, Cloud NAT
├── 10-platform/    GKE, CloudSQL, GCS, Artifact Registry, IAM
├── 20-app/         helm_release per service (image tags bumped by CI)
└── 30-observability/ alert policies, log sinks, dashboards
```

App deploys only touch `20-app/` — fast plans, isolated blast radius.

## Deployment flow

```
Developer merges PR to monitoring-app/main
  │
  ├─ CI: go test ./... + go vet
  ├─ CI: docker build collector:$SHA → push to Artifact Registry
  ├─ CI: bump collector_tag = "$SHA" in infrastructure/terraform/envs/dev/20-app/terraform.tfvars
  │
  └─ Triggers terraform.yaml (path: infrastructure/**)
       ├─ tf plan on 20-app layer → posts diff as PR comment
       ├─ Reviewer approves
       └─ tf apply → helm_release.collector updated → GKE rolling deploy
```

Console deploys follow the same pattern (`console_tag`).

Infrastructure changes (VPC, cluster, CloudSQL) go through the same `terraform.yaml` workflow
but touch `00-network` or `10-platform` layers, which require separate reviewer approval and
carry higher blast radius.

## Authentication & authorisation *(deferred to phase 3)*

- **Ingest (collector):** API keys (`mpk_live_<32-byte-random>`) stored as bcrypt hashes in
  Postgres; cached in Redis with 60s TTL. Validated at the collector edge. `project_id` is
  derived from the key — never trusted from the payload.
- **Dashboard (console):** GCP IAP for internal access; OIDC (Google/Okta) for external.
- **Machine access (CI, SDK):** Workload Identity — no long-lived JSON key files anywhere.

## Phased roadmap

| Phase | Status | Deliverable |
|---|---|---|
| 0 | ✅ Done | Repo layout, kind + Helm dev loop, `/healthz` on all services |
| 1 | ✅ Done | OTLP/HTTP receiver, ClickHouse schema, domain types, pricing table, REST read API |
| 2 | ✅ Done | `SpanSink` interface, ingestor stub, ClickHouse enabled in umbrella chart, React console scaffold |
| 3 | ✅ Done | Terraform GCP modules (GKE Autopilot, CloudSQL, GCS, Artifact Registry, IAM), CI pipelines, monorepo |
| 4 | Planned | Eval engine (LLM-as-judge, heuristics), alert worker, queue (Pub-Sub) decoupling |
| 5 | Planned | API key authN/Z, OIDC for console, multi-tenancy isolation, staging + prod envs |
| 6 | Planned | Multi-region, second cloud (AWS), per-tenant node pools |
