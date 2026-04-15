# ADR-001: Use ClickHouse for span and run storage

**Status:** Accepted  
**Date:** 2026-04-14

## Context

We need a time-series store for high-cardinality span data. Candidates considered:
ClickHouse, Cassandra, Prometheus, Elasticsearch.

## Decision

Use ClickHouse with `MergeTree` / `ReplacingMergeTree` engines.

## Reasoning

**Prometheus** is purpose-built for metrics (counters, gauges) — not traces. High-cardinality
labels (trace IDs, span IDs) violate its data model and it has no JOIN or GROUP BY support.
Ruled out immediately for span storage.

**Cassandra** excels at high-throughput writes with predictable point-lookup latency, but
aggregation queries (`GROUP BY model, SUM(cost_usd)`) require a separate query engine
(Spark, Trino) on top. Our read patterns are analytical — per-project cost roll-ups, p95
latency by run name, token usage over time — which Cassandra cannot serve efficiently alone.

**ClickHouse** natively supports all our read patterns with vectorised SQL execution:

```sql
SELECT model, sum(cost_usd), quantile(0.95)(duration_ms)
FROM spans
WHERE project_id = 'x' AND start_time > now() - INTERVAL 7 DAY
GROUP BY model
```

This runs in milliseconds over millions of rows. Key properties:
- Columnar layout — scanning `sum(cost_usd)` touches only the cost column.
- `PARTITION BY toYYYYMM(start_time)` — time-range queries skip irrelevant partitions entirely.
- `ReplacingMergeTree` — handles streaming run aggregation (spans arrive in batches).
- Native TTL — `TTL start_time + INTERVAL 90 DAY` with partition-level expiry (no row-by-row
  overhead).
- Precedent: Langfuse, PostHog, SigNoz, Plausible all converged on ClickHouse for the same
  observability-at-scale use case.

## Consequences

- ClickHouse must be available before spans can be written. The collector degrades gracefully
  (logs warning, drops spans) if CH is unreachable.
- `FINAL` modifier required on `runs` table reads to force `ReplacingMergeTree` deduplication.
- Large prompt/completion payloads are stored in GCS (not CH rows) to keep CH column widths
  narrow.
