import { useState } from 'react'

export interface DetailTab {
  id: string
  label: string
  content: React.ReactNode
}

export function DetailTabs({ tabs }: { tabs: DetailTab[] }) {
  const [active, setActive] = useState(tabs[0]?.id ?? '')

  if (tabs.length === 0) return null

  const current = tabs.find((t) => t.id === active) ?? tabs[0]

  return (
    <div className="detail-tabs">
      <div className="detail-tabs__bar" role="tablist">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            type="button"
            role="tab"
            aria-selected={tab.id === current.id}
            className={`detail-tabs__tab${tab.id === current.id ? ' is-active' : ''}`}
            onClick={() => setActive(tab.id)}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="detail-tabs__panel" role="tabpanel">
        {current.content}
      </div>
    </div>
  )
}
