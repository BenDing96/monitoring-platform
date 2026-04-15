package ingest

import (
	"context"

	"monitoring-platform/internal/domain"
)

// SpanSink is the write-side abstraction for span storage.
// Phase 0–1: backed by a direct ClickHouse writer.
// Phase 3+:  backed by a Kafka producer; ingestor binary consumes and writes to CH.
type SpanSink interface {
	WriteSpans(ctx context.Context, spans []domain.Span) error
}
