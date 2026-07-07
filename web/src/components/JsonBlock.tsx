export function JsonBlock({ value }: { value: unknown }) {
  const text =
    value === undefined
      ? '—'
      : JSON.stringify(value, null, 2)

  return (
    <pre className="json-block">
      <code>{text}</code>
    </pre>
  )
}
