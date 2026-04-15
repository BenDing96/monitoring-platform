# monitoring-app

Go backend services for monitoring-platform.

## Services

| Binary | Port | Description |
|---|---|---|
| `collector` | `:4318` | OTLP/HTTP ingress — receives spans, stamps project, writes to ClickHouse |
| `ingestor` | `:8081` | Queue consumer (stub — Pub/Sub consumer wired in phase 4) |
| `api` | `:8080` | REST read API — runs list, span tree |
| `worker` | — | Eval + alert jobs (phase 4) |

## Local development

```bash
# From repo root: start kind cluster
make kind-up

# Build images + deploy full stack to kind
make deploy-dev

# Or use Tilt for live reload
make tilt

# Run tests
make app-test

# Build binaries only (no Docker)
make app-build
```

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `COLLECTOR_ADDR` | `:4318` | Collector listen address |
| `API_ADDR` | `:8080` | API listen address |
| `INGESTOR_ADDR` | `:8081` | Ingestor listen address |
| `DEFAULT_PROJECT_ID` | `default` | Project ID stamped on spans (phase 3: derived from API key) |
| `CLICKHOUSE_ADDR` | `clickhouse:9000` | ClickHouse address |
| `CLICKHOUSE_DB` | `monitoring` | Database name |
| `CLICKHOUSE_USER` | `default` | Username |
| `CLICKHOUSE_PASSWORD` | — | Password |
| `CLICKHOUSE_TLS` | `false` | Enable TLS |

## Package layout

```
cmd/           one main package per binary
internal/
  domain/      core types: Span, Run, StatusCode
  httpx/       shared HTTP helpers (health handler)
  ingest/      SpanSink interface
  otelconv/    OTLP proto → domain mapping
  pricing/     model cost table (USD per 1M tokens)
  storage/
    clickhouse/ schema, writer, reader
```

## Adding a new service

1. `cmd/<name>/main.go` — minimal main with graceful shutdown and `/healthz`.
2. `deploy/docker/<name>.Dockerfile` — distroless, nonroot.
3. `deploy/helm/charts/<name>/` — Chart.yaml, values.yaml, templates/.
4. Add `<name>` to `SERVICES` in `Makefile` and `services` list in `Tiltfile`.
5. Add `<name>` as a dependency in `deploy/helm/charts/monitoring-app-dev/Chart.yaml`.
6. Add `helm_release.<name>` in `infrastructure/terraform/envs/dev/20-app/main.tf`.
