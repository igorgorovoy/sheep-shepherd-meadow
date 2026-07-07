import { useMemo, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { ApplyDrawer } from './ApplyDrawer'
import { SettingsPanel } from './SettingsPanel'
import { Sidebar } from './Sidebar'
import { Topbar } from './Topbar'
import { useSettings } from '../contexts/SettingsContext'
import { useClusterData } from '../hooks/useClusterData'
import { useNamespace } from '../hooks/useNamespace'
import { useTheme } from '../hooks/useTheme'
import type { PageContext } from '../pages/context'

function titleFromPath(pathname: string): string {
  if (pathname === '/') return 'Overview'
  if (pathname.startsWith('/nodes/')) return 'Node'
  if (pathname === '/nodes') return 'Nodes'
  if (pathname.startsWith('/pods/')) return 'Pod'
  if (pathname === '/pods') return 'Pods'
  if (pathname.startsWith('/deployments/')) return 'Deployment'
  if (pathname === '/deployments') return 'Deployments'
  if (pathname.startsWith('/services/')) return 'Service'
  if (pathname === '/services') return 'Services'
  if (pathname === '/events') return 'Events'
  if (pathname === '/pasture') return 'Pasture'
  if (pathname.startsWith('/meadow/')) return 'Repository'
  if (pathname === '/meadow') return 'Meadow'
  return 'Dashboard'
}

export function Layout() {
  const { namespace, setNamespace } = useNamespace()
  const { theme, toggle } = useTheme()
  const { shepherdAuthRequired, shepherdToken, openSettings, shepherdUrl } =
    useSettings()
  const [menuOpen, setMenuOpen] = useState(false)
  const [applyOpen, setApplyOpen] = useState(false)
  const location = useLocation()

  const configKey = `${shepherdUrl}:${shepherdToken}`
  const cluster = useClusterData(namespace, configKey)

  const needsToken =
    shepherdAuthRequired === true && !shepherdToken.trim()

  const title = titleFromPath(location.pathname)
  const context: PageContext = useMemo(
    () => ({ ...cluster, namespace, setNamespace }),
    [cluster, namespace, setNamespace],
  )

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
          namespace={namespace}
          namespaces={cluster.data?.namespaces ?? ['default']}
          onNamespaceChange={setNamespace}
          onApply={() => setApplyOpen(true)}
          onSettings={openSettings}
        />
        {needsToken && (
          <div className="auth-banner" role="status">
            API token required.{' '}
            <button type="button" className="link-btn" onClick={openSettings}>
              Open Settings
            </button>{' '}
            and enter your Shepherd Bearer token.
          </div>
        )}
        <main className="content">
          <div className="content__inner">
            <Outlet context={context} />
          </div>
        </main>
      </div>

      <SettingsPanel />
      <ApplyDrawer
        open={applyOpen}
        onClose={() => setApplyOpen(false)}
        onSuccess={cluster.refresh}
      />
    </div>
  )
}
