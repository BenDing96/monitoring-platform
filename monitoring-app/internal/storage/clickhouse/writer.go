package clickhouse

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"monitoring-platform/internal/domain"
)

// Writer writes domain spans and runs to ClickHouse.
type Writer struct {
	conn driver.Conn
}

// NewWriter returns a Writer backed by the given connection.
func NewWriter(conn driver.Conn) *Writer {
	return &Writer{conn: conn}
}

// WriteSpans inserts a batch of spans and upserts a run row per trace.
func (w *Writer) WriteSpans(ctx context.Context, spans []domain.Span) error {
	if err := w.insertSpans(ctx, spans); err != nil {
		return err
	}
	runs := aggregateRuns(spans)
	return w.upsertRuns(ctx, runs)
}

func (w *Writer) insertSpans(ctx context.Context, spans []domain.Span) error {
	batch, err := w.conn.PrepareBatch(ctx, "INSERT INTO spans")
	if err != nil {
		return fmt.Errorf("prepare spans batch: %w", err)
	}
	for _, s := range spans {
		if err := batch.Append(
			s.ProjectID,
			s.TraceID,
			s.SpanID,
			s.ParentSpanID,
			s.Name,
			s.StartTime,
			s.EndTime,
			s.DurationMS(),
			uint8(s.StatusCode),
			s.StatusMessage,
			s.Model,
			s.InputTokens,
			s.OutputTokens,
			s.CostUSD,
			s.Attributes,
			s.IngestTime,
		); err != nil {
			return fmt.Errorf("append span: %w", err)
		}
	}
	return batch.Send()
}

func (w *Writer) upsertRuns(ctx context.Context, runs []domain.Run) error {
	batch, err := w.conn.PrepareBatch(ctx, "INSERT INTO runs")
	if err != nil {
		return fmt.Errorf("prepare runs batch: %w", err)
	}
	for _, r := range runs {
		if err := batch.Append(
			r.ProjectID,
			r.TraceID,
			r.Name,
			r.StartTime,
			r.EndTime,
			r.DurationMS(),
			uint8(r.StatusCode),
			r.TotalInputTokens,
			r.TotalOutputTokens,
			r.TotalCostUSD,
			r.SpanCount,
			r.Attributes,
			r.IngestTime,
		); err != nil {
			return fmt.Errorf("append run: %w", err)
		}
	}
	return batch.Send()
}

// aggregateRuns builds one Run per trace from the given spans.
// The root span (no parent) provides name, start, status, and attributes.
func aggregateRuns(spans []domain.Span) []domain.Run {
	type agg struct {
		run        domain.Run
		minStart   time.Time
		maxEnd     time.Time
		seenSpanID map[string]struct{}
	}
	m := make(map[string]*agg)

	for _, s := range spans {
		a, ok := m[s.TraceID]
		if !ok {
			a = &agg{
				run: domain.Run{
					ProjectID:  s.ProjectID,
					TraceID:    s.TraceID,
					IngestTime: s.IngestTime,
				},
				seenSpanID: make(map[string]struct{}),
			}
			m[s.TraceID] = a
		}

		if _, dup := a.seenSpanID[s.SpanID]; dup {
			continue
		}
		a.seenSpanID[s.SpanID] = struct{}{}
		a.run.SpanCount++
		a.run.TotalInputTokens += s.InputTokens
		a.run.TotalOutputTokens += s.OutputTokens
		a.run.TotalCostUSD += s.CostUSD

		// root span: empty parent span ID
		if s.ParentSpanID == "" || isZeroHex(s.ParentSpanID) {
			a.run.Name = s.Name
			a.run.StatusCode = s.StatusCode
			a.run.Attributes = s.Attributes
		}

		if a.minStart.IsZero() || s.StartTime.Before(a.minStart) {
			a.minStart = s.StartTime
		}
		if s.EndTime.After(a.maxEnd) {
			a.maxEnd = s.EndTime
		}
	}

	out := make([]domain.Run, 0, len(m))
	for _, a := range m {
		a.run.StartTime = a.minStart
		a.run.EndTime = a.maxEnd
		out = append(out, a.run)
	}
	return out
}

// isZeroHex returns true for all-zero hex strings (empty parent span ID in OTLP).
func isZeroHex(s string) bool {
	for _, c := range s {
		if c != '0' {
			return false
		}
	}
	return true
}
