package domain

import "time"

// StatusCode mirrors OTel span status.
type StatusCode uint8

const (
	StatusUnset StatusCode = 0
	StatusOK    StatusCode = 1
	StatusError StatusCode = 2
)

// Span is the canonical representation of a single OTel span after ingest.
type Span struct {
	ProjectID    string
	TraceID      string
	SpanID       string
	ParentSpanID string
	Name         string

	StartTime time.Time
	EndTime   time.Time

	StatusCode    StatusCode
	StatusMessage string

	// LLM-specific (gen_ai.* semconv)
	Model        string
	InputTokens  uint32
	OutputTokens uint32
	CostUSD      float64

	// All other attributes serialised as JSON.
	Attributes string

	IngestTime time.Time
}

// DurationMS returns the span duration in milliseconds.
func (s Span) DurationMS() float64 {
	return float64(s.EndTime.Sub(s.StartTime).Microseconds()) / 1000.0
}
