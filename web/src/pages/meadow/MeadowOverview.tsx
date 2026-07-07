import { formatBytes } from '../../lib/format'
import { useRowNavigate } from '../../lib/navigation'
import { PageHeader } from '../../components/PageHeader'
import { EmptyState, ErrorState, LoadingRows } from '../../components/states'
import { useSettings } from '../../contexts/SettingsContext'
import { useMeadowData } from '../../hooks/useMeadowData'

export function MeadowOverview() {
  const { data, loading, error, refresh } = useMeadowData()
  const { openSettings } = useSettings()
  const repos = data?.details ?? []
  const onRow = useRowNavigate()

  if (loading && !data) {
    return (
      <>
        <PageHeader title="Meadow" />
        <LoadingRows />
      </>
    )
  }

  if (error && !data) {
    return (
      <>
        <PageHeader title="Meadow" />
        <ErrorState message={error} />
      </>
    )
  }

  const totalSize = repos.reduce((n, r) => n + (r.total_size ?? 0), 0)

  return (
    <>
      <PageHeader
        title="Meadow"
        count={data?.repositories ?? repos.length}
        description={`OCI image registry · ${formatBytes(totalSize)} total storage`}
        actions={
          <button type="button" className="btn" onClick={() => void refresh()}>
            Refresh
          </button>
        }
      />
      {repos.length === 0 ? (
        <EmptyState
          title="No repositories"
          sub="Push an image to Meadow to see it here."
        />
      ) : (
        <div className="table-wrap">
          <table className="table">
            <thead>
              <tr>
                <th>Repository</th>
                <th>Tags</th>
                <th>Size</th>
              </tr>
            </thead>
            <tbody>
              {repos.map((r) => (
                <tr
                  key={r.name}
                  className="table-row--clickable"
                  onClick={onRow(`/meadow/repos/${encodeURIComponent(r.name)}`)}
                >
                  <td className="td-name">{r.name}</td>
                  <td className="td-mono">{r.tags?.length ?? 0}</td>
                  <td className="td-mono">{formatBytes(r.total_size)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <p className="muted meadow-hint">
        Configure Meadow URL and token in{' '}
        <button type="button" className="link-btn" onClick={openSettings}>
          Settings
        </button>
        .
      </p>
    </>
  )
}
