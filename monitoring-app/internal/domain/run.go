package domain

import "time"

// Run represents a complete agent trace (root span + all children aggregated).
type Run struct {
	ProjectID string
	TraceID   string
	Name      string

	StartTime time.Time
	EndTime   time.Time

	StatusCode StatusCode

	TotalInputTokens  uint32
	TotalOutputTokens uint32
	TotalCostUSD      float64
	SpanCount         uint32

	// Root span attributes serialised as JSON.
	Attributes string

	IngestTime time.Time
}

// DurationMS returns the run duration in milliseconds.
func (r Run) DurationMS() float64 {
	return float64(r.EndTime.Sub(r.StartTime).Microseconds()) / 1000.0
}
