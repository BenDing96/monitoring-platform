-- spans: one row per OTel span
CREATE TABLE IF NOT EXISTS spans
(
    project_id      String,
    trace_id        String,
    span_id         String,
    parent_span_id  String,
    name            String,
    start_time      DateTime64(9, 'UTC'),
    end_time        DateTime64(9, 'UTC'),
    duration_ms     Float64,
    status_code     UInt8,
    status_message  String,
    model           String,
    input_tokens    UInt32,
    output_tokens   UInt32,
    cost_usd        Float64,
    attributes      String,  -- JSON blob
    ingest_time     DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, start_time, trace_id, span_id)
TTL start_time + INTERVAL 90 DAY;

-- runs: one row per trace (root-span aggregation), updated on each ingest via ReplacingMergeTree
CREATE TABLE IF NOT EXISTS runs
(
    project_id           String,
    trace_id             String,
    name                 String,
    start_time           DateTime64(9, 'UTC'),
    end_time             DateTime64(9, 'UTC'),
    duration_ms          Float64,
    status_code          UInt8,
    total_input_tokens   UInt32,
    total_output_tokens  UInt32,
    total_cost_usd       Float64,
    span_count           UInt32,
    attributes           String,  -- root span attributes as JSON
    ingest_time          DateTime DEFAULT now()
)
ENGINE = ReplacingMergeTree(ingest_time)
PARTITION BY toYYYYMM(start_time)
ORDER BY (project_id, start_time, trace_id)
TTL start_time + INTERVAL 90 DAY;
