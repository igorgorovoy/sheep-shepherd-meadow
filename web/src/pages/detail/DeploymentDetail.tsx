import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { fetchDeployment } from '../../api/client'
import { Breadcrumb } from '../../components/Breadcrumb'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { DetailTabs } from '../../components/DetailTabs'
import { JsonBlock } from '../../components/JsonBlock'
import { PageHeader } from '../../components/PageHeader'
import { Progress } from '../../components/Progress'
import { ScaleControl } from '../../components/ScaleControl'
import { ErrorState, LoadingRows } from '../../components/states'
import { useToast } from '../../contexts/ToastContext'
import { useMutations } from '../../hooks/useMutations'
import { useResource } from '../../hooks/useResource'
import { formatSelector } from '../../lib/format'
import { usePageData } from '../context'

export function DeploymentDetail() {
  const { ns = 'default', name = '' } = useParams()
  const navigate = useNavigate()
  const { refresh: refreshCluster } = usePageData()
  const { removeDeployment } = useMutations()
  const { push } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const { data: dep, loading, error, refresh } = useResource(
    (signal) => fetchDeployment(ns, name, signal),
    [ns, name],
  )

  async function handleDelete() {
    setBusy(true)
    try {
      await removeDeployment(ns, name)
      refreshCluster()
      navigate('/deployments')
    } catch (err) {
      push(err instanceof Error ? err.message : 'Delete failed', 'err')
    } finally {
      setBusy(false)
      setConfirmOpen(false)
    }
  }

  if (loading && !dep) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Deployments', to: '/deployments' },
            { label: `${ns}/${name}` },
          ]}
        />
        <LoadingRows />
      </>
    )
  }

  if (error && !dep) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Deployments', to: '/deployments' },
            { label: `${ns}/${name}` },
          ]}
        />
        <ErrorState message={error} />
      </>
    )
  }

  if (!dep) return null

  const desired = dep.spec?.replicas ?? 0
  const ready = dep.status?.ready_replicas ?? 0

  return (
    <>
      <Breadcrumb
        items={[
          { label: 'Deployments', to: '/deployments' },
          { label: dep.metadata.namespace },
          { label: dep.metadata.name },
        ]}
      />
      <PageHeader
        title={dep.metadata.name}
        description={`Namespace ${dep.metadata.namespace} · selector ${formatSelector(dep.spec?.selector)}`}
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
        <div className="detail-card detail-card--wide">
          <span className="detail-card__label">Ready replicas</span>
          <Progress value={ready} max={desired} />
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Desired</span>
          <span className="detail-card__value mono">{desired}</span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Available</span>
          <span className="detail-card__value mono">
            {dep.status?.available_replicas ?? 0}
          </span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Updated</span>
          <span className="detail-card__value mono">
            {dep.status?.updated_replicas ?? 0}
          </span>
        </div>
      </div>

      <div className="card" style={{ marginBottom: 'var(--sp-4)' }}>
        <div className="card__body">
          <ScaleControl
            deployment={dep}
            onScaled={() => {
              refresh()
              refreshCluster()
            }}
          />
        </div>
      </div>

      <div className="card">
        <div className="card__body">
          <DetailTabs
            tabs={[
              {
                id: 'spec',
                label: 'Spec',
                content: <JsonBlock value={dep.spec} />,
              },
              {
                id: 'status',
                label: 'Status',
                content: <JsonBlock value={dep.status} />,
              },
              {
                id: 'template',
                label: 'Pod template',
                content: <JsonBlock value={dep.spec?.template} />,
              },
            ]}
          />
        </div>
      </div>

      <ConfirmDialog
        open={confirmOpen}
        title="Delete deployment"
        description={`This removes deployment ${name} and its managed pods.`}
        expectedName={name}
        busy={busy}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={handleDelete}
      />
    </>
  )
}
