const BASE = import.meta.env.VITE_API_URL ?? ''

export interface Run {
  ProjectID: string
  TraceID: string
  Name: string
  StartTime: string
  EndTime: string
  StatusCode: number
  TotalInputTokens: number
  TotalOutputTokens: number
  TotalCostUSD: number
  SpanCount: number
  Attributes: string
}

export interface Span {
  ProjectID: string
  TraceID: string
  SpanID: string
  ParentSpanID: string
  Name: string
  StartTime: string
  EndTime: string
  StatusCode: number
  StatusMessage: string
  Model: string
  InputTokens: number
  OutputTokens: number
  CostUSD: number
  Attributes: string
}

async function get<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`)
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json() as Promise<T>
}

export const api = {
  listRuns: () => get<{ runs: Run[] }>('/v1/runs'),
  getRunSpans: (traceID: string) =>
    get<{ spans: Span[] }>(`/v1/runs/${traceID}`),
}
