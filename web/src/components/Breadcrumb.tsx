import { Link } from 'react-router-dom'

export function Breadcrumb({
  items,
}: {
  items: { label: string; to?: string }[]
}) {
  return (
    <nav className="breadcrumb" aria-label="Breadcrumb">
      {items.map((item, i) => (
        <span key={`${item.label}-${i}`} className="breadcrumb__item">
          {i > 0 && <span className="breadcrumb__sep" aria-hidden>/</span>}
          {item.to ? (
            <Link to={item.to} className="breadcrumb__link">
              {item.label}
            </Link>
          ) : (
            <span className="breadcrumb__current">{item.label}</span>
          )}
        </span>
      ))}
    </nav>
  )
}
