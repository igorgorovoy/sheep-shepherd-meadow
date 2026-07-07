import type { NamespaceFilter } from '../api/types'

export function NamespaceSelect({
  value,
  namespaces,
  onChange,
}: {
  value: NamespaceFilter
  namespaces: string[]
  onChange: (ns: NamespaceFilter) => void
}) {
  const options = ['all', ...namespaces.filter((n) => n !== 'all')]

  return (
    <label className="ns-select">
      <span className="ns-select__label">Namespace</span>
      <select
        className="ns-select__input"
        value={value}
        onChange={(e) => onChange(e.target.value as NamespaceFilter)}
        aria-label="Filter by namespace"
      >
        {options.map((ns) => (
          <option key={ns} value={ns}>
            {ns === 'all' ? 'All namespaces' : ns}
          </option>
        ))}
      </select>
    </label>
  )
}
