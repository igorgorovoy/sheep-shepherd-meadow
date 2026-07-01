import { PageHeader } from '../components/PageHeader'
import { EmptyState, ErrorState, LoadingRows } from '../components/states'
import { formatRelativeTime } from '../lib/format'
import { usePageData } from './context'
import type { Event } from '../api/types'

function sortEvents(events: Event[]): Event[] {
  return [...events].sort(
    (a, b) =>
      new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
  )
}

export function Events() {
  const { data, loading, error } = usePageData()
  const events = sortEvents(data?.events ?? [])

  if (loading && !data)
    return (
      <>
        <PageHeader title="Events" />
        <LoadingRows />
      </>
    )
  if (error && !data)
    return (
      <>
        <PageHeader title="Events" />
        <ErrorState message={error} />
      </>
    )

  return (
    <>
      <PageHeader
        title="Events"
        count={events.length}
        description="Most recent cluster activity, newest first. Warnings are marked with a dashed glyph."
      />
      {events.length === 0 ? (
        <EmptyState title="No events" sub="The cluster has not reported any activity." />
      ) : (
        <div className="card">
          <div className="events">
            {events.map((e, i) => {
              const warning = e.type === 'Warning'
              return (
                <div
                  className={`event ${warning ? 'event--warning' : 'event--normal'}`}
                  key={`${e.timestamp}-${i}`}
                >
                  <div
                    className="event__icon"
                    title={e.type}
                    aria-label={e.type}
                  >
                    {warning ? '!' : 'i'}
                  </div>
                  <div className="event__body">
                    <div className="event__head">
                      <span className="event__reason">{e.reason}</span>
                      <span className="event__object">{e.object}</span>
                    </div>
                    <div className="event__msg">{e.message}</div>
                  </div>
                  <div className="event__time" title={e.timestamp}>
                    {formatRelativeTime(e.timestamp)}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </>
  )
}
