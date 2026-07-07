import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { fetchService } from '../../api/client'
import { Badge } from '../../components/Badge'
import { Breadcrumb } from '../../components/Breadcrumb'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { DetailTabs } from '../../components/DetailTabs'
import { JsonBlock } from '../../components/JsonBlock'
import { PageHeader } from '../../components/PageHeader'
import { ErrorState, LoadingRows } from '../../components/states'
import { useToast } from '../../contexts/ToastContext'
import { useMutations } from '../../hooks/useMutations'
import { useResource } from '../../hooks/useResource'
import { usePageData } from '../context'

export function ServiceDetail() {
  const { ns = 'default', name = '' } = useParams()
  const navigate = useNavigate()
  const { refresh: refreshCluster } = usePageData()
  const { removeService } = useMutations()
  const { push } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const { data: svc, loading, error, refresh } = useResource(
    (signal) => fetchService(ns, name, signal),
    [ns, name],
  )

  async function handleDelete() {
    setBusy(true)
    try {
      await removeService(ns, name)
      refreshCluster()
      navigate('/services')
    } catch (err) {
      push(err instanceof Error ? err.message : 'Delete failed', 'err')
    } finally {
      setBusy(false)
      setConfirmOpen(false)
    }
  }

  if (loading && !svc) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Services', to: '/services' },
            { label: `${ns}/${name}` },
          ]}
        />
        <LoadingRows />
      </>
    )
  }

  if (error && !svc) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Services', to: '/services' },
            { label: `${ns}/${name}` },
          ]}
        />
        <ErrorState message={error} />
      </>
    )
  }

  if (!svc) return null

  const ports = svc.spec?.ports ?? []
  const endpoints = svc.status?.endpoints ?? []

  return (
    <>
      <Breadcrumb
        items={[
          { label: 'Services', to: '/services' },
          { label: svc.metadata.namespace },
          { label: svc.metadata.name },
        ]}
      />
      <PageHeader
        title={svc.metadata.name}
        description={`Namespace ${svc.metadata.namespace} · ${svc.spec?.type}`}
        actions={
          <>
            <button type="button" className="btn" onClick={refresh}>
              Refresh
            </button>
            <button
              type="button"
              className="btn btn--danger"
              onClick={() => setConfirmOpen(true)}
            >
              Delete
            </button>
          </>
        }
      />

      <div className="detail-cards">
        <div className="detail-card">
          <span className="detail-card__label">Type</span>
          <Badge variant={svc.spec?.type === 'NodePort' ? 'solid' : 'outline'}>
            {svc.spec?.type}
          </Badge>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Cluster IP</span>
          <span className="detail-card__value mono">{svc.status?.cluster_ip || '—'}</span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Endpoints</span>
          <span className="detail-card__value mono">{endpoints.length}</span>
        </div>
      </div>

      <div className="card">
        <div className="card__body">
          <DetailTabs
            tabs={[
              {
                id: 'ports',
                label: 'Ports',
                content:
                  ports.length === 0 ? (
                    <p className="muted">No ports defined.</p>
                  ) : (
                    <div className="table-wrap">
                      <table className="table">
                        <thead>
                          <tr>
                            <th>Name</th>
                            <th>Port</th>
                            <th>Target</th>
                            <th>Node port</th>
                            <th>Protocol</th>
                          </tr>
                        </thead>
                        <tbody>
                          {ports.map((p, i) => (
                            <tr key={`${p.port}-${i}`}>
                              <td className="td-name">{p.name || '—'}</td>
                              <td className="td-mono">{p.port}</td>
                              <td className="td-mono">{p.target_port}</td>
                              <td className="td-mono">{p.node_port ?? '—'}</td>
                              <td className="td-mono td-muted">{p.protocol || 'TCP'}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ),
              },
              {
                id: 'endpoints',
                label: 'Endpoints',
                content:
                  endpoints.length === 0 ? (
                    <p className="muted">No endpoints registered.</p>
                  ) : (
                    <ul className="endpoint-list">
                      {endpoints.map((ep) => (
                        <li key={ep} className="mono">
                          {ep}
                        </li>
                      ))}
                    </ul>
                  ),
              },
              {
                id: 'spec',
                label: 'Spec',
                content: <JsonBlock value={svc.spec} />,
              },
              {
                id: 'status',
                label: 'Status',
                content: <JsonBlock value={svc.status} />,
              },
            ]}
          />
        </div>
      </div>

      <ConfirmDialog
        open={confirmOpen}
        title="Delete service"
        description={`This removes service ${name} from namespace ${ns}.`}
        expectedName={name}
        busy={busy}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={handleDelete}
      />
    </>
  )
}
