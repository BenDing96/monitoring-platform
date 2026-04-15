import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api, type Run } from '../../lib/api'

const STATUS_LABEL: Record<number, string> = { 0: 'Unset', 1: 'OK', 2: 'Error' }
const STATUS_COLOR: Record<number, string> = {
  0: 'text-gray-400',
  1: 'text-green-400',
  2: 'text-red-400',
}

export function RunsList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['runs'],
    queryFn: () => api.listRuns(),
    refetchInterval: 15_000,
  })

  if (isLoading) return <p className="text-gray-500">Loading…</p>
  if (error) return <p className="text-red-400">Failed to load runs.</p>

  const runs = data?.runs ?? []

  return (
    <div>
      <h1 className="text-xl font-semibold mb-6">Runs</h1>
      {runs.length === 0 ? (
        <p className="text-gray-500 text-sm">No runs yet. Send some traces to the collector.</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-gray-800">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-800 text-gray-400 text-left">
                <Th>Name</Th>
                <Th>Status</Th>
                <Th>Spans</Th>
                <Th>Tokens in / out</Th>
                <Th>Cost</Th>
                <Th>Started</Th>
                <Th>Duration</Th>
              </tr>
            </thead>
            <tbody>
              {runs.map((run) => (
                <RunRow key={run.TraceID} run={run} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function RunRow({ run }: { run: Run }) {
  const start = new Date(run.StartTime)
  const end = new Date(run.EndTime)
  const durMs = end.getTime() - start.getTime()

  return (
    <tr className="border-b border-gray-800/60 hover:bg-gray-800/30 transition-colors">
      <td className="px-4 py-3 font-medium">
        <Link to={`/runs/${run.TraceID}`} className="hover:text-blue-400 transition-colors">
          {run.Name || '(unnamed)'}
        </Link>
      </td>
      <td className={`px-4 py-3 ${STATUS_COLOR[run.StatusCode]}`}>
        {STATUS_LABEL[run.StatusCode] ?? 'Unknown'}
      </td>
      <td className="px-4 py-3 text-gray-300">{run.SpanCount}</td>
      <td className="px-4 py-3 text-gray-300">
        {run.TotalInputTokens.toLocaleString()} / {run.TotalOutputTokens.toLocaleString()}
      </td>
      <td className="px-4 py-3 text-gray-300">
        ${run.TotalCostUSD.toFixed(6)}
      </td>
      <td className="px-4 py-3 text-gray-400">{start.toLocaleString()}</td>
      <td className="px-4 py-3 text-gray-400">{formatDuration(durMs)}</td>
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
