import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { fetchPod } from '../../api/client'
import type { Event } from '../../api/types'
import { Breadcrumb } from '../../components/Breadcrumb'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { DetailTabs } from '../../components/DetailTabs'
import { JsonBlock } from '../../components/JsonBlock'
import { PageHeader } from '../../components/PageHeader'
import { PodPhaseBadge } from '../../components/status'
import { ErrorState, LoadingRows } from '../../components/states'
import { useToast } from '../../contexts/ToastContext'
import { useMutations } from '../../hooks/useMutations'
import { useResource } from '../../hooks/useResource'
import { formatRelativeTime } from '../../lib/format'
import { usePageData } from '../context'

function filterPodEvents(events: Event[], podName: string): Event[] {
  return events.filter((e) => {
    const obj = e.object ?? ''
    return obj === podName || obj.endsWith(`/${podName}`) || obj.includes(podName)
  })
}

export function PodDetail() {
  const { ns = 'default', name = '' } = useParams()
  const navigate = useNavigate()
  const { data: cluster, refresh: refreshCluster } = usePageData()
  const { removePod } = useMutations()
  const { push } = useToast()
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [busy, setBusy] = useState(false)
  const { data: pod, loading, error, refresh } = useResource(
    (signal) => fetchPod(ns, name, signal),
    [ns, name],
  )

  const events = useMemo(
    () => filterPodEvents(cluster?.events ?? [], name),
    [cluster?.events, name],
  )

  async function handleDelete() {
    setBusy(true)
    try {
      await removePod(ns, name)
      refreshCluster()
      navigate('/pods')
    } catch (err) {
      push(err instanceof Error ? err.message : 'Delete failed', 'err')
    } finally {
      setBusy(false)
      setConfirmOpen(false)
    }
  }

  if (loading && !pod) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Pods', to: '/pods' },
            { label: `${ns}/${name}` },
          ]}
        />
        <LoadingRows />
      </>
    )
  }

  if (error && !pod) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Pods', to: '/pods' },
            { label: `${ns}/${name}` },
          ]}
        />
        <ErrorState message={error} />
      </>
    )
  }

  if (!pod) return null

  const containers = pod.status?.containers ?? []

  return (
    <>
      <Breadcrumb
        items={[
          { label: 'Pods', to: '/pods' },
          { label: pod.metadata.namespace },
          { label: pod.metadata.name },
        ]}
      />
      <PageHeader
        title={pod.metadata.name}
        description={`Namespace ${pod.metadata.namespace} · node ${pod.spec?.node_name || 'unscheduled'}`}
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
          <span className="detail-card__label">Phase</span>
          <PodPhaseBadge phase={pod.status?.phase} />
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Pod IP</span>
          <span className="detail-card__value mono">{pod.status?.pod_ip || '—'}</span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Host IP</span>
          <span className="detail-card__value mono">{pod.status?.host_ip || '—'}</span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Restart policy</span>
          <span className="detail-card__value mono">{pod.spec?.restart_policy || '—'}</span>
        </div>
        <div className="detail-card">
          <span className="detail-card__label">Started</span>
          <span className="detail-card__value">
            {formatRelativeTime(pod.status?.start_time)}
          </span>
        </div>
      </div>

      <div className="card">
        <div className="card__body">
          <DetailTabs
            tabs={[
              {
                id: 'containers',
                label: 'Containers',
                content:
                  containers.length === 0 ? (
                    <p className="muted">No container status reported yet.</p>
                  ) : (
                    <div className="table-wrap">
                      <table className="table">
                        <thead>
                          <tr>
                            <th>Name</th>
                            <th>State</th>
                            <th>Ready</th>
                            <th>Container ID</th>
                            <th>Exit</th>
                          </tr>
                        </thead>
                        <tbody>
                          {containers.map((c) => (
                            <tr key={c.name}>
                              <td className="td-name">{c.name}</td>
                              <td className="td-mono">{c.state}</td>
                              <td className="td-mono">{c.ready ? 'yes' : 'no'}</td>
                              <td className="td-mono td-muted">{c.container_id || '—'}</td>
                              <td className="td-mono">{c.exit_code ?? '—'}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ),
              },
              {
                id: 'spec',
                label: 'Spec',
                content: <JsonBlock value={pod.spec} />,
              },
              {
                id: 'status',
                label: 'Status',
                content: <JsonBlock value={pod.status} />,
              },
              {
                id: 'events',
                label: 'Events',
                content:
                  events.length === 0 ? (
                    <p className="muted">No events for this pod.</p>
                  ) : (
                    <div className="table-wrap">
                      <table className="table">
                        <thead>
                          <tr>
                            <th>Type</th>
                            <th>Reason</th>
                            <th>Message</th>
                            <th>Age</th>
                          </tr>
                        </thead>
                        <tbody>
                          {events.map((e, i) => (
                            <tr key={`${e.timestamp}-${i}`}>
                              <td className="td-mono">{e.type}</td>
                              <td className="td-name">{e.reason}</td>
                              <td>{e.message}</td>
                              <td className="td-muted">
                                {formatRelativeTime(e.timestamp)}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  ),
              },
            ]}
          />
        </div>
      </div>

      <ConfirmDialog
        open={confirmOpen}
        title="Delete pod"
        description={`This removes pod ${name} from namespace ${ns}.`}
        expectedName={name}
        busy={busy}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={handleDelete}
      />
    </>
  )
}
