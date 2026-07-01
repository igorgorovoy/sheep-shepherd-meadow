import { PageHeader } from '../components/PageHeader'
import { PodPhaseBadge } from '../components/status'
import { EmptyState, ErrorState, LoadingRows } from '../components/states'
import { usePageData } from './context'
import type { Pod } from '../api/types'

function images(pod: Pod): string[] {
  return (pod.spec?.containers ?? []).map((c) => c.image).filter(Boolean)
}

export function Pods() {
  const { data, loading, error } = usePageData()
  const pods = data?.pods ?? []

  if (loading && !data)
    return (
      <>
        <PageHeader title="Pods" />
        <LoadingRows />
      </>
    )
  if (error && !data)
    return (
      <>
        <PageHeader title="Pods" />
        <ErrorState message={error} />
      </>
    )

  return (
    <>
      <PageHeader
        title="Pods"
        count={pods.length}
        description="Smallest deployable units. Each pod runs one or more containers on a node."
      />
      {pods.length === 0 ? (
        <EmptyState title="No pods" sub="Nothing is scheduled on the cluster yet." />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Namespace</th>
                <th>Phase</th>
                <th>Node</th>
                <th>Pod IP</th>
                <th>Images</th>
                <th>Restart</th>
              </tr>
            </thead>
            <tbody>
              {pods.map((p) => (
                <tr key={p.metadata.uid || `${p.metadata.namespace}/${p.metadata.name}`}>
                  <td className="td-name">{p.metadata.name}</td>
                  <td className="td-muted">{p.metadata.namespace}</td>
                  <td>
                    <PodPhaseBadge phase={p.status?.phase} />
                  </td>
                  <td className="td-mono">{p.spec?.node_name || <span className="td-muted">unscheduled</span>}</td>
                  <td className="td-mono">{p.status?.pod_ip || '—'}</td>
                  <td>
                    <div className="chips">
                      {images(p).length === 0 ? (
                        <span className="td-muted">—</span>
                      ) : (
                        images(p).map((img, i) => (
                          <span className="chip" key={`${img}-${i}`} title={img}>
                            {img}
                          </span>
                        ))
                      )}
                    </div>
                  </td>
                  <td className="td-mono td-muted">{p.spec?.restart_policy || '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  )
}
