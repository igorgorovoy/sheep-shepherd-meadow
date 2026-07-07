import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { fetchNode } from '../../api/client'
import { Breadcrumb } from '../../components/Breadcrumb'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { DetailTabs } from '../../components/DetailTabs'
import { JsonBlock } from '../../components/JsonBlock'
import { PageHeader } from '../../components/PageHeader'
import { NodeConditionBadge } from '../../components/status'
import { ErrorState, LoadingRows } from '../../components/states'
import { useToast } from '../../contexts/ToastContext'
import { useMutations } from '../../hooks/useMutations'
import { useResource } from '../../hooks/useResource'
import { formatBytes, formatCpu, formatRelativeTime } from '../../lib/format'
import { usePageData } from '../context'

export function NodeDetail() {
  const { name = '' } = useParams()
  const navigate = useNavigate()
  const { refresh: refreshCluster } = usePageData()
  const { removeNode } = useMutations()
  const { push } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const { data: node, loading, error, refresh } = useResource(
    (signal) => fetchNode(name, signal),
    [name],
  )

  async function handleDelete() {
    setBusy(true)
    try {
      await removeNode(name)
      refreshCluster()
      navigate('/nodes')
    } catch (err) {
      push(err instanceof Error ? err.message : 'Delete failed', 'err')
    } finally {
      setBusy(false)
      setConfirmOpen(false)
    }
  }

  if (loading && !node) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Nodes', to: '/nodes' },
            { label: name },
          ]}
        />
        <LoadingRows />
      </>
    )
  }

  if (error && !node) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Nodes', to: '/nodes' },
            { label: name },
          ]}
        />
        <ErrorState message={error} />
      </>
    )
  }

  if (!node) return null

  return (
    <>
      <Breadcrumb
        items={[
          { label: 'Nodes', to: '/nodes' },
          { label: node.metadata.name },
        ]}
      />
      <PageHeader
        title={node.metadata.name}
        description={`Worker node · ${node.spec?.address || 'no address'}`}
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
          <span className="detail-card__label">Condition</span>
          <NodeConditionBadge condition={node.status?.condition} />
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Pods</span>
          <span className="detail-card__value mono">{node.status?.pod_count ?? 0}</span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">CPU capacity</span>
          <span className="detail-card__value mono">
            {formatCpu(node.status?.capacity?.cpu)}
          </span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Memory capacity</span>
          <span className="detail-card__value mono">
            {formatBytes(node.status?.capacity?.memory)}
          </span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Last heartbeat</span>
          <span className="detail-card__value">
            {formatRelativeTime(node.status?.last_heartbeat)}
          </span>
        </div>
      </div>

      <div className="card">
        <div className="card__body">
          <DetailTabs
            tabs={[
              {
                id: 'spec',
                label: 'Spec',
                content: <JsonBlock value={node.spec} />,
              },
              {
                id: 'status',
                label: 'Status',
                content: <JsonBlock value={node.status} />,
              },
              {
                id: 'full',
                label: 'Full object',
                content: <JsonBlock value={node} />,
              },
            ]}
          />
        </div>
      </div>

      <ConfirmDialog
        open={confirmOpen}
        title="Delete node"
        description={`Remove node ${name} from the cluster?`}
        expectedName={name}
        busy={busy}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={handleDelete}
      />
    </>
  )
}
