import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  deleteMeadowManifest,
  fetchMeadowRepoStats,
  fetchMeadowTags,
} from '../../api/meadow'
import { Breadcrumb } from '../../components/Breadcrumb'
import { ConfirmDialog } from '../../components/ConfirmDialog'
import { PageHeader } from '../../components/PageHeader'
import { PullCommand } from '../../components/PullCommand'
import { ErrorState, LoadingRows } from '../../components/states'
import { useToast } from '../../contexts/ToastContext'
import { useResource } from '../../hooks/useResource'
import { formatBytes } from '../../lib/format'

export function RepoDetail() {
  const { name = '' } = useParams()
  const repo = decodeURIComponent(name)
  const navigate = useNavigate()
  const { push } = useToast()
  const [deleteTag, setDeleteTag] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  const { data: stats, loading, error, refresh } = useResource(
    (signal) => fetchMeadowRepoStats(repo, signal),
    [repo],
  )

  const { data: tagList, refresh: refreshTags } = useResource(
    (signal) => fetchMeadowTags(repo, signal),
    [repo],
  )

  const tags = tagList?.tags ?? stats?.tags ?? []

  async function confirmDelete() {
    if (!deleteTag) return
    setBusy(true)
    try {
      await deleteMeadowManifest(repo, deleteTag)
      push(`Tag ${deleteTag} deleted`)
      setDeleteTag(null)
      refresh()
      refreshTags()
    } catch (err) {
      push(err instanceof Error ? err.message : 'Delete failed', 'err')
    } finally {
      setBusy(false)
    }
  }

  if (loading && !stats) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Meadow', to: '/meadow' },
            { label: repo },
          ]}
        />
        <LoadingRows />
      </>
    )
  }

  if (error && !stats) {
    return (
      <>
        <Breadcrumb
          items={[
            { label: 'Meadow', to: '/meadow' },
            { label: repo },
          ]}
        />
        <ErrorState message={error} />
      </>
    )
  }

  return (
    <>
      <Breadcrumb
        items={[
          { label: 'Meadow', to: '/meadow' },
          { label: repo },
        ]}
      />
      <PageHeader
        title={repo}
        description={`${tags.length} tags · ${formatBytes(stats?.total_size ?? 0)}`}
        actions={
          <button type="button" className="btn" onClick={() => navigate('/meadow')}>
            Back
          </button>
        }
      />

      <div className="card">
        <div className="card__body">
          {tags.length === 0 ? (
            <p className="muted">No tags in this repository.</p>
          ) : (
            <div className="table-wrap">
              <table className="table">
                <thead>
                  <tr>
                    <th>Tag</th>
                    <th>Pull</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  {tags.map((tag) => (
                    <tr key={tag}>
                      <td className="td-name mono">{tag}</td>
                      <td>
                        <PullCommand repo={repo} tag={tag} />
                      </td>
                      <td className="td-actions">
                        <button
                          type="button"
                          className="btn btn--danger btn--sm"
                          onClick={() => setDeleteTag(tag)}
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      <ConfirmDialog
        open={deleteTag != null}
        title="Delete manifest tag"
        description={`Permanently remove tag ${deleteTag} from ${repo}?`}
        expectedName={deleteTag ?? ''}
        confirmLabel="Delete tag"
        busy={busy}
        onCancel={() => setDeleteTag(null)}
        onConfirm={confirmDelete}
      />
    </>
  )
}
