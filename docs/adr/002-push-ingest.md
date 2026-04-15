# ADR-002: Push-mode OTLP ingest over pull

**Status:** Accepted  
**Date:** 2026-04-14

## Context

We need to decide how agent runtimes send observability data to the platform: push (agents send
to us) or pull (we scrape agents).

## Decision

Push-only via OTLP (HTTP `:4318` and gRPC `:4317`). No pull/scrape endpoint.

## Reasoning

**Pull** (Prometheus-style) requires the collector to reach into the agent's network. This is
impractical for:
- External / cloud-hosted agents — would require VPC peering, reverse tunnels, or public
  endpoints on the agent side.
- Serverless / ephemeral workloads — there is nothing persistent to scrape.
- Multi-tenancy — we would need credentials per customer environment to authenticate inbound
  scrapes.

**Push** is the natural model for a SaaS-style observability platform:
- Agents only need outbound HTTPS to our endpoint — works from any network.
- Authentication happens at the collector edge on every request (credential → `project_id`).
- `project_id` is stamped from the credential, not the payload — clients cannot spoof tenancy.
- Backpressure is explicit: `429 Too Many Requests` + `Retry-After`.
- OTel SDKs (Go, Python, JS, Java, .NET, Rust) all support OTLP push natively — zero custom
  instrumentation required beyond pointing the exporter at our endpoint.

**Custom protocol vs OTLP:** We use standard OTLP rather than a proprietary format. This means
any OTel-instrumented agent (LangChain, LlamaIndex, generic Python with `opentelemetry-sdk`)
works out of the box. We follow the `gen_ai.*` semantic conventions for LLM-specific attributes.

## Consequences

- The collector is the sole public ingress point — TLS terminates there.
- API key validation (phase 3) goes entirely in the collector. No auth logic in the ingestor
  or api services.
- OTLP/JSON (text) is also valid OTLP/HTTP — we may add support as a convenience but it is not
  required for any known SDK.
