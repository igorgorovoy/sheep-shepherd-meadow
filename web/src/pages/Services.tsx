import { Badge } from '../components/Badge'
import { PageHeader } from '../components/PageHeader'
import { EmptyState, ErrorState, LoadingRows } from '../components/states'
import { usePageData } from './context'
import type { Service } from '../api/types'

function portLabel(service: Service): string[] {
  return (service.spec?.ports ?? []).map((p) => {
    const base = `${p.port}→${p.target_port}`
    return p.node_port ? `${base} (node ${p.node_port})` : base
  })
}

export function Services() {
  const { data, loading, error } = usePageData()
  const services = data?.services ?? []

  if (loading && !data)
    return (
      <>
        <PageHeader title="Services" />
        <LoadingRows />
      </>
    )
  if (error && !data)
    return (
      <>
        <PageHeader title="Services" />
        <ErrorState message={error} />
      </>
    )

  return (
    <>
      <PageHeader
        title="Services"
        count={services.length}
        description="Stable virtual IPs that load-balance traffic across a set of pods."
      />
      {services.length === 0 ? (
        <EmptyState title="No services" sub="No services are exposed on the cluster." />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Namespace</th>
                <th>Type</th>
                <th>Cluster IP</th>
                <th>Ports</th>
                <th>Endpoints</th>
              </tr>
            </thead>
            <tbody>
              {services.map((s) => {
                const ports = portLabel(s)
                const endpoints = s.status?.endpoints?.length ?? 0
                return (
                  <tr key={s.metadata.uid || s.metadata.name}>
                    <td className="td-name">{s.metadata.name}</td>
                    <td className="td-muted">{s.metadata.namespace}</td>
                    <td>
                      <Badge
                        variant={s.spec?.type === 'NodePort' ? 'solid' : 'outline'}
                      >
                        {s.spec?.type}
                      </Badge>
                    </td>
                    <td className="td-mono">{s.status?.cluster_ip || '—'}</td>
                    <td>
                      <div className="chips">
                        {ports.length === 0 ? (
                          <span className="td-muted">—</span>
                        ) : (
                          ports.map((p, i) => (
                            <span className="chip" key={`${p}-${i}`}>
                              {p}
                            </span>
                          ))
                        )}
                      </div>
                    </td>
                    <td className="td-mono" title={s.status?.endpoints?.join(', ')}>
                      {endpoints}
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
