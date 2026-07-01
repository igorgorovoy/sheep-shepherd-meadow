import type { ReactNode } from 'react'

export function PageHeader({
  title,
  count,
  description,
}: {
  title: string
  count?: number
  description?: ReactNode
}) {
  return (
    <>
      <div className="page-head">
        <h2>{title}</h2>
        {count != null && <span className="page-head__count">{count}</span>}
      </div>
      {description && <p className="page-desc">{description}</p>}
    </>
  )
}
