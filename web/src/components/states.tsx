import type { ReactNode } from 'react'
import { getShepherdApiBase } from '../api/config'

// Reusable loading / empty / error presentational states.

export function EmptyState({
  icon = '∅',
  title,
  sub,
}: {
  icon?: string
  title: string
  sub?: ReactNode
}) {
  return (
    <div className="state">
      <div className="state__icon" aria-hidden>
        {icon}
      </div>
      <div className="state__title">{title}</div>
      {sub && <div className="state__sub">{sub}</div>}
    </div>
  )
}

export function ErrorState({ message }: { message: string }) {
  return (
    <div className="state">
      <div className="state__icon" aria-hidden>
        ⚠
      </div>
      <div className="state__title">Cannot reach the cluster</div>
      <div className="state__sub">
        {message}. Confirm Shepherd is running at{' '}
        <span className="mono">{getShepherdApiBase()}</span> (configure in Settings
        if needed).
      </div>
    </div>
  )
}

export function LoadingRows({ rows = 5 }: { rows?: number }) {
  return (
    <div className="stack" style={{ gap: 'var(--sp-2)' }}>
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="skeleton"
          style={{ height: 40, opacity: 1 - i * 0.12 }}
        />
      ))}
    </div>
  )
}
