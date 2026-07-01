import { useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'
import { useClusterData } from '../hooks/useClusterData'
import { useTheme } from '../hooks/useTheme'
import type { PageContext } from '../pages/context'

const TITLES: Record<string, string> = {
  '/': 'Overview',
  '/nodes': 'Nodes',
  '/pods': 'Pods',
  '/deployments': 'Deployments',
  '/services': 'Services',
  '/events': 'Events',
  '/pasture': 'Pasture',
}

// Owns polling + theme and passes the data state down to routed pages via the
// router <Outlet> context. Handles the responsive sidebar drawer.
export function Layout() {
  const cluster = useClusterData()
  const { theme, toggle } = useTheme()
  const [menuOpen, setMenuOpen] = useState(false)
  const location = useLocation()

  const title = TITLES[location.pathname] ?? 'Dashboard'
  const context: PageContext = cluster

  return (
    <div className={`app${menuOpen ? ' is-open' : ''}`}>
      <Sidebar
        data={cluster.data}
        version={cluster.data?.info?.version}
        onNavigate={() => setMenuOpen(false)}
      />
      <div className="scrim" onClick={() => setMenuOpen(false)} />

      <div className="main">
        <Topbar
          title={title}
          healthy={cluster.healthy}
          refreshing={cluster.refreshing}
          lastUpdated={cluster.lastUpdated}
          onRefresh={cluster.refresh}
          onToggleMenu={() => setMenuOpen((v) => !v)}
          theme={theme}
          onToggleTheme={toggle}
        />
        <main className="content">
          <div className="content__inner">
            <Outlet context={context} />
          </div>
        </main>
      </div>
    </div>
  )
}
