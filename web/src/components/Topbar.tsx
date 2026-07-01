import { formatRelativeTime } from '../lib/format'
import type { Theme } from '../hooks/useTheme'

export function Topbar({
  title,
  healthy,
  refreshing,
  lastUpdated,
  onRefresh,
  onToggleMenu,
  theme,
  onToggleTheme,
}: {
  title: string
  healthy: boolean
  refreshing: boolean
  lastUpdated: Date | null
  onRefresh: () => void
  onToggleMenu: () => void
  theme: Theme
  onToggleTheme: () => void
}) {
  return (
    <header className="topbar">
      <button
        className="menu-toggle"
        onClick={onToggleMenu}
        aria-label="Toggle navigation"
      >
        ≡
      </button>

      <h1 className="topbar__title">{title}</h1>

      <div className="topbar__spacer" />

      <div className="topbar__meta">
        <span
          className={`conn ${healthy ? 'conn--up' : 'conn--down'}`}
          title={healthy ? 'Shepherd API reachable' : 'Shepherd API unreachable'}
        >
          <span className="conn__dot" aria-hidden />
          <span className="conn__text">
            {healthy ? 'Connected' : 'Offline'}
          </span>
        </span>

        <span className="topbar__updated nowrap">
          {lastUpdated ? `Updated ${formatRelativeTime(lastUpdated.toISOString())}` : 'Never updated'}
        </span>

        <button
          className="btn btn--icon"
          onClick={onRefresh}
          disabled={refreshing}
          aria-label="Refresh"
          title="Refresh now"
        >
          <span className={refreshing ? 'btn__spin' : ''} aria-hidden>
            ↻
          </span>
        </button>

        <button
          className="btn btn--icon"
          onClick={onToggleTheme}
          aria-label="Toggle light / dark theme"
          title={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
        >
          <span aria-hidden>{theme === 'light' ? '☾' : '☀'}</span>
        </button>
      </div>
    </header>
  )
}
