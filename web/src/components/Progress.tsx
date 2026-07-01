// Monochrome progress bar for replica readiness. When not fully ready the
// fill uses a hatched pattern so the "partial" state reads without color.
export function Progress({
  value,
  max,
}: {
  value: number
  max: number
}) {
  const safeMax = Math.max(max, 0)
  const pct = safeMax === 0 ? 0 : Math.min(100, (value / safeMax) * 100)
  const complete = safeMax > 0 && value >= safeMax
  return (
    <div className="progress" title={`${value} / ${safeMax} ready`}>
      <div className="progress__track">
        <div
          className={`progress__fill${complete ? '' : ' progress__fill--partial'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="progress__label">
        {value}/{safeMax}
      </span>
    </div>
  )
}
