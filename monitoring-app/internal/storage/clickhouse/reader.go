package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"monitoring-platform/internal/domain"
)

// Reader queries runs and spans from ClickHouse.
type Reader struct {
	conn driver.Conn
}

// NewReader returns a Reader backed by the given connection.
func NewReader(conn driver.Conn) *Reader {
	return &Reader{conn: conn}
}

// RunFilter constrains a ListRuns query.
type RunFilter struct {
	ProjectID string
	Since     time.Time
	Until     time.Time
	Limit     int
}

// ListRuns returns runs matching the filter, newest first.
func (r *Reader) ListRuns(ctx context.Context, f RunFilter) ([]domain.Run, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Since.IsZero() {
		f.Since = time.Now().UTC().Add(-24 * time.Hour)
	}
	if f.Until.IsZero() {
		f.Until = time.Now().UTC()
	}

	const q = `
		SELECT
			project_id, trace_id, name,
			start_time, end_time, duration_ms,
			status_code,
			total_input_tokens, total_output_tokens, total_cost_usd,
			span_count, attributes, ingest_time
		FROM runs FINAL
		WHERE project_id = @project_id
		  AND start_time >= @since
		  AND start_time < @until
		ORDER BY start_time DESC
		LIMIT @limit
	`
	rows, err := r.conn.Query(ctx, q,
		clickhouseNamed("project_id", f.ProjectID),
		clickhouseNamed("since", f.Since),
		clickhouseNamed("until", f.Until),
		clickhouseNamed("limit", f.Limit),
	)
	if err != nil {
		return nil, fmt.Errorf("list runs query: %w", err)
	}
	defer rows.Close()

	var out []domain.Run
	for rows.Next() {
		var run domain.Run
		var durMS float64
		if err := rows.Scan(
			&run.ProjectID, &run.TraceID, &run.Name,
			&run.StartTime, &run.EndTime, &durMS,
			&run.StatusCode,
			&run.TotalInputTokens, &run.TotalOutputTokens, &run.TotalCostUSD,
			&run.SpanCount, &run.Attributes, &run.IngestTime,
		); err != nil {
			return nil, fmt.Errorf("scan run: %w", err)
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

// SpansByTrace returns all spans for a given trace ID.
func (r *Reader) SpansByTrace(ctx context.Context, projectID, traceID string) ([]domain.Span, error) {
	const q = `
		SELECT
			project_id, trace_id, span_id, parent_span_id, name,
			start_time, end_time, duration_ms,
			status_code, status_message,
			model, input_tokens, output_tokens, cost_usd,
			attributes, ingest_time
		FROM spans
		WHERE project_id = @project_id
		  AND trace_id = @trace_id
		ORDER BY start_time ASC
	`
	rows, err := r.conn.Query(ctx, q,
		clickhouseNamed("project_id", projectID),
		clickhouseNamed("trace_id", traceID),
	)
	if err != nil {
		return nil, fmt.Errorf("spans by trace query: %w", err)
	}
	defer rows.Close()

	var out []domain.Span
	for rows.Next() {
		var s domain.Span
		var durMS float64
		if err := rows.Scan(
			&s.ProjectID, &s.TraceID, &s.SpanID, &s.ParentSpanID, &s.Name,
			&s.StartTime, &s.EndTime, &durMS,
			&s.StatusCode, &s.StatusMessage,
			&s.Model, &s.InputTokens, &s.OutputTokens, &s.CostUSD,
			&s.Attributes, &s.IngestTime,
		); err != nil {
			return nil, fmt.Errorf("scan span: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// clickhouseNamed is a helper to build a named parameter.
func clickhouseNamed(name string, value any) driver.NamedValue {
	return driver.NamedValue{Name: name, Value: value}
}
