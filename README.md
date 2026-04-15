# monitoring-platform

An open observability platform for LLM agents. Collects traces, token usage, cost, and quality
evaluations from AI workloads via OpenTelemetry push, stores them in ClickHouse, and surfaces
them through a React dashboard.

## Repository layout

```
monitoring-platform/
├── monitoring-app/        Go services (collector, ingestor, api, worker)
├── monitoring-console/    React + TypeScript dashboard
├── infrastructure/        Terraform (GCP) + Helm cluster addons
└── docs/                  Architecture, ADRs, runbooks
```

## Documentation

| Document | Description |
|---|---|
| [docs/architecture.md](docs/architecture.md) | System architecture, components, data flow |
| [docs/adr/001-clickhouse.md](docs/adr/001-clickhouse.md) | Why ClickHouse over Cassandra/Prometheus |
| [docs/adr/002-push-ingest.md](docs/adr/002-push-ingest.md) | Why push-mode OTLP over pull |
| [docs/adr/003-monorepo.md](docs/adr/003-monorepo.md) | Why a monorepo |

## Quick start (local)

**Prerequisites:** Go 1.25, Node 22, Docker, [kind](https://kind.sigs.k8s.io), [kubectl](https://kubernetes.io/docs/tasks/tools), [helm](https://helm.sh/docs/intro/install), [tilt](https://docs.tilt.dev/install.html) (optional)

```bash
# 1. Start a local Kubernetes cluster + registry
make kind-up

# 2. Start the full stack with live reload
make tilt
# or without tilt:
make deploy-dev

# Collector OTLP/HTTP  →  localhost:4318
# API                  →  localhost:8080
# Console (dev server) →  see below
```

**Run the console dev server** (proxies /v1/* to localhost:8080):

```bash
cd monitoring-console
npm install
npm run dev       # http://localhost:5173
```

**Run backend tests:**

```bash
make app-test
```

## Sending a test trace

```bash
# Send a minimal OTLP trace to the local collector
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @- <<'EOF'
# (use an OTel SDK or otelcol to generate a real protobuf payload)
EOF
```

Or instrument your agent with any OTel SDK pointing at `http://localhost:4318`.

## CI / deployment

Each GitHub Actions workflow is path-scoped — only the relevant pipeline runs on a given PR:

| Workflow | Trigger path | What it does |
|---|---|---|
| `app-ci.yaml` | `monitoring-app/**` | `go test` → build images → push to Artifact Registry → bump tag in `terraform.tfvars` |
| `console-ci.yaml` | `monitoring-console/**` | `npm run build` → build image → push → bump tag |
| `terraform.yaml` | `infrastructure/**` | `tf plan` on PR (posts diff as comment), `tf apply` on merge (manual approval gate) |

See [docs/architecture.md](docs/architecture.md) for the full deploy flow.

## Sub-project READMEs

- [monitoring-app/README.md](monitoring-app/README.md) — Go services dev guide
- [monitoring-console/README.md](monitoring-console/README.md) — UI dev guide
- [infrastructure/README.md](infrastructure/README.md) — Terraform bootstrap guide
