package otelconv

import (
	"encoding/hex"
	"encoding/json"
	"time"

	coltracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"

	"monitoring-platform/internal/domain"
	"monitoring-platform/internal/pricing"
)

// SpansFromRequest converts an OTLP ExportTraceServiceRequest into domain spans.
// projectID is stamped from the authenticated API key — never from the payload.
func SpansFromRequest(req *coltracev1.ExportTraceServiceRequest, projectID string) []domain.Span {
	now := time.Now().UTC()
	var out []domain.Span

	for _, rs := range req.GetResourceSpans() {
		for _, ss := range rs.GetScopeSpans() {
			for _, s := range ss.GetSpans() {
				span := convertSpan(s, projectID, now)
				out = append(out, span)
			}
		}
	}
	return out
}

func convertSpan(s *tracev1.Span, projectID string, now time.Time) domain.Span {
	attrs := attrsToMap(s.GetAttributes())

	span := domain.Span{
		ProjectID:     projectID,
		TraceID:       hex.EncodeToString(s.GetTraceId()),
		SpanID:        hex.EncodeToString(s.GetSpanId()),
		ParentSpanID:  hex.EncodeToString(s.GetParentSpanId()),
		Name:          s.GetName(),
		StartTime:     unixNanoToTime(s.GetStartTimeUnixNano()),
		EndTime:       unixNanoToTime(s.GetEndTimeUnixNano()),
		StatusCode:    domain.StatusCode(s.GetStatus().GetCode()),
		StatusMessage: s.GetStatus().GetMessage(),
		IngestTime:    now,
	}

	// gen_ai semconv attributes
	if v, ok := attrs["gen_ai.request.model"]; ok {
		span.Model = v
	}
	if v, ok := attrs["gen_ai.usage.input_tokens"]; ok {
		span.InputTokens = parseUint32(v)
	}
	if v, ok := attrs["gen_ai.usage.output_tokens"]; ok {
		span.OutputTokens = parseUint32(v)
	}

	span.CostUSD = pricing.Calculate(span.Model, span.InputTokens, span.OutputTokens)

	if b, err := json.Marshal(attrs); err == nil {
		span.Attributes = string(b)
	}

	return span
}

func attrsToMap(attrs []*commonv1.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[kv.GetKey()] = anyValueToString(kv.GetValue())
	}
	return m
}

func anyValueToString(v *commonv1.AnyValue) string {
	if v == nil {
		return ""
	}
	switch x := v.GetValue().(type) {
	case *commonv1.AnyValue_StringValue:
		return x.StringValue
	case *commonv1.AnyValue_IntValue:
		return intToString(x.IntValue)
	case *commonv1.AnyValue_DoubleValue:
		return floatToString(x.DoubleValue)
	case *commonv1.AnyValue_BoolValue:
		if x.BoolValue {
			return "true"
		}
		return "false"
	default:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
		return ""
	}
}

func unixNanoToTime(ns uint64) time.Time {
	if ns == 0 {
		return time.Time{}
	}
	return time.Unix(0, int64(ns)).UTC()
}

func parseUint32(s string) uint32 {
	var v uint32
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		v = v*10 + uint32(c-'0')
	}
	return v
}

func intToString(i int64) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

func floatToString(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}
