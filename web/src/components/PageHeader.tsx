import type { ReactNode } from 'react'

export function PageHeader({
  title,
  count,
  description,
  actions,
}: {
  title: string
  count?: number
  description?: ReactNode
  actions?: ReactNode
}) {
  return (
    <>
      <div className="page-head">
        <div className="page-head__main">
          <h2>{title}</h2>
          {count != null && <span className="page-head__count">{count}</span>}
        </div>
        {actions && <div className="page-head__actions">{actions}</div>}
      </div>
      {description && <p className="page-desc">{description}</p>}
    </>
  )
}
