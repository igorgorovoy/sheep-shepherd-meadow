import { PageHeader } from '../components/PageHeader'
import { NodeConditionBadge } from '../components/status'
import { EmptyState, ErrorState, LoadingRows } from '../components/states'
import { formatBytes, formatCpu, formatRelativeTime } from '../lib/format'
import { useRowNavigate } from '../lib/navigation'
import { usePageData } from './context'

export function Nodes() {
  const { data, loading, error } = usePageData()
  const nodes = data?.nodes ?? []
  const onRow = useRowNavigate()

  if (loading && !data) return withHead(<LoadingRows />)
  if (error && !data) return withHead(<ErrorState message={error} />)

  return (
    <>
      <PageHeader
        title="Nodes"
        count={nodes.length}
        description="Worker machines (shepherd + sheep agents) registered with the control plane."
      />
      {nodes.length === 0 ? (
        <EmptyState title="No nodes registered" sub="Start a sheep agent to join the cluster." />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Address</th>
                <th>Condition</th>
                <th>Pods</th>
                <th>CPU</th>
                <th>Memory</th>
                <th>Heartbeat</th>
              </tr>
            </thead>
            <tbody>
              {nodes.map((n) => (
                <tr
                  key={n.metadata.uid || n.metadata.name}
                  className="table-row--clickable"
                  onClick={onRow(`/nodes/${n.metadata.name}`)}
                >
                  <td className="td-name">{n.metadata.name}</td>
                  <td className="td-mono">{n.spec?.address || '—'}</td>
                  <td>
                    <NodeConditionBadge condition={n.status?.condition} />
                  </td>
                  <td className="td-mono">{n.status?.pod_count ?? 0}</td>
                  <td className="td-mono" title="capacity">
                    {formatCpu(n.status?.capacity?.cpu)}
                  </td>
                  <td className="td-mono" title="capacity">
                    {formatBytes(n.status?.capacity?.memory)}
                  </td>
                  <td className="td-muted" title={n.status?.last_heartbeat}>
                    {formatRelativeTime(n.status?.last_heartbeat)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}

function withHead(body: React.ReactNode) {
  return (
    <>
      <PageHeader title="Nodes" />
      {body}
    </>
  )
}
