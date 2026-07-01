import { NavLink } from 'react-router-dom'
import type { ClusterSummary } from '../api/types'

interface NavItem {
  to: string
  label: string
  icon: string
  count?: number
}

export function Sidebar({
  data,
  version,
  onNavigate,
}: {
  data: ClusterSummary | null
  version?: string
  onNavigate: () => void
}) {
  const items: NavItem[] = [
    { to: '/', label: 'Overview', icon: '◎' },
    { to: '/nodes', label: 'Nodes', icon: '▦', count: data?.nodes?.length },
    { to: '/pods', label: 'Pods', icon: '◧', count: data?.pods?.length },
    {
      to: '/deployments',
      label: 'Deployments',
      icon: '⧉',
      count: data?.deployments?.length,
    },
    {
      to: '/services',
      label: 'Services',
      icon: '⇄',
      count: data?.services?.length,
    },
    { to: '/events', label: 'Events', icon: '≡', count: data?.events?.length },
    { to: '/pasture', label: 'Pasture', icon: '⛰' },
  ]

  return (
    <aside className="sidebar">
      <div className="sidebar__brand">
        <div className="brand-mark" aria-hidden>
          🐑
        </div>
        <div className="brand-text">
          <strong>Sheep &amp; Shepherd</strong>
          <span>Cluster Dashboard</span>
        </div>
      </div>

      <nav className="nav">
        {items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            onClick={onNavigate}
            className={({ isActive }) =>
              `nav__link${isActive ? ' is-active' : ''}`
            }
          >
            <span className="nav__icon" aria-hidden>
              {item.icon}
            </span>
            <span className="nav__label">{item.label}</span>
            {item.count != null && (
              <span className="nav__count">{item.count}</span>
            )}
          </NavLink>
        ))}
      </nav>

      <div className="sidebar__foot">
        <div>Shepherd control plane</div>
        <div className="mono">{version ? `v${version}` : 'version —'}</div>
      </div>
    </aside>
  )
}
