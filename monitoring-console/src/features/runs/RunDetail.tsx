import { useQuery } from '@tanstack/react-query'
import { useParams, Link } from 'react-router-dom'
import { api, type Span } from '../../lib/api'

export function RunDetail() {
  const { traceID } = useParams<{ traceID: string }>()
  const { data, isLoading, error } = useQuery({
    queryKey: ['spans', traceID],
    queryFn: () => api.getRunSpans(traceID!),
    enabled: !!traceID,
  })

  if (isLoading) return <p className="text-gray-500">Loading…</p>
  if (error) return <p className="text-red-400">Failed to load spans.</p>

  const spans = data?.spans ?? []
  const root = spans.find((s) => !s.ParentSpanID || isZeroHex(s.ParentSpanID))

  return (
    <div>
      <Link to="/runs" className="text-sm text-gray-500 hover:text-gray-300 mb-4 inline-block">
        ← Back to runs
      </Link>
      <h1 className="text-xl font-semibold mb-1">{root?.Name ?? traceID}</h1>
      <p className="text-xs text-gray-500 font-mono mb-6">{traceID}</p>

      <div className="overflow-x-auto rounded-lg border border-gray-800">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-gray-800 text-gray-400 text-left">
              <Th>Span</Th>
              <Th>Model</Th>
              <Th>Tokens in / out</Th>
              <Th>Cost</Th>
              <Th>Duration</Th>
              <Th>Status</Th>
            </tr>
          </thead>
          <tbody>
            {spans.map((span) => (
              <SpanRow key={span.SpanID} span={span} />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function SpanRow({ span }: { span: Span }) {
  const start = new Date(span.StartTime)
  const end = new Date(span.EndTime)
  const durMs = end.getTime() - start.getTime()
  const isRoot = !span.ParentSpanID || isZeroHex(span.ParentSpanID)

  return (
    <tr className="border-b border-gray-800/60 hover:bg-gray-800/30 transition-colors">
      <td className="px-4 py-3">
        <span className={isRoot ? 'font-semibold' : 'pl-4 text-gray-400'}>
          {span.Name}
        </span>
        <div className="text-xs text-gray-600 font-mono">{span.SpanID.slice(0, 12)}…</div>
      </td>
      <td className="px-4 py-3 text-gray-300 text-xs font-mono">{span.Model || '—'}</td>
      <td className="px-4 py-3 text-gray-300">
        {span.InputTokens} / {span.OutputTokens}
      </td>
      <td className="px-4 py-3 text-gray-300">${span.CostUSD.toFixed(6)}</td>
      <td className="px-4 py-3 text-gray-400">{formatDuration(durMs)}</td>
      <td className={`px-4 py-3 ${span.StatusCode === 2 ? 'text-red-400' : 'text-green-400'}`}>
        {span.StatusCode === 2 ? 'Error' : 'OK'}
      </td>
    </tr>
  )
}

function Th({ children }: { children: React.ReactNode }) {
  return <th className="px-4 py-3 font-medium text-xs uppercase tracking-wider">{children}</th>
}

function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  return `${(ms / 60_000).toFixed(1)}m`
}

function isZeroHex(s: string): boolean {
  return /^0*$/.test(s)
}
