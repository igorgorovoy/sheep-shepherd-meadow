import { PageHeader } from '../components/PageHeader'
import { Progress } from '../components/Progress'
import { EmptyState, ErrorState, LoadingRows } from '../components/states'
import { formatSelector } from '../lib/format'
import { usePageData } from './context'

export function Deployments() {
  const { data, loading, error } = usePageData()
  const deployments = data?.deployments ?? []

  if (loading && !data)
    return (
      <>
        <PageHeader title="Deployments" />
        <LoadingRows />
      </>
    )
  if (error && !data)
    return (
      <>
        <PageHeader title="Deployments" />
        <ErrorState message={error} />
      </>
    )

  return (
    <>
      <PageHeader
        title="Deployments"
        count={deployments.length}
        description="Declarative rollouts that keep a desired number of pod replicas running."
      />
      {deployments.length === 0 ? (
        <EmptyState title="No deployments" sub="Apply a deployment manifest to see it here." />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Namespace</th>
                <th style={{ minWidth: 200 }}>Ready</th>
                <th>Available</th>
                <th>Selector</th>
              </tr>
            </thead>
            <tbody>
              {deployments.map((d) => {
                const desired = d.spec?.replicas ?? 0
                const ready = d.status?.ready_replicas ?? 0
                return (
                  <tr key={d.metadata.uid || d.metadata.name}>
                    <td className="td-name">{d.metadata.name}</td>
                    <td className="td-muted">{d.metadata.namespace}</td>
                    <td>
                      <Progress value={ready} max={desired} />
                    </td>
                    <td className="td-mono">
                      {d.status?.available_replicas ?? 0}
                    </td>
                    <td className="td-mono td-muted">
                      {formatSelector(d.spec?.selector)}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}
