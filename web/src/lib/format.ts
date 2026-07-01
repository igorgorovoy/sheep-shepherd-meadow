// Formatting helpers for the dashboard.

// Human-readable memory from a byte count (binary units).
export function formatBytes(bytes: number | undefined | null): string {
  if (bytes == null || Number.isNaN(bytes)) return '—'
  if (bytes === 0) return '0 B'
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB']
  const i = Math.min(
    units.length - 1,
    Math.floor(Math.log(Math.abs(bytes)) / Math.log(1024)),
  )
  const value = bytes / Math.pow(1024, i)
  const rounded = value >= 100 || i === 0 ? Math.round(value) : Number(value.toFixed(1))
  return `${rounded} ${units[i]}`
}

// CPU millicores -> human-readable. 1000m == 1 core.
export function formatCpu(millicores: number | undefined | null): string {
  if (millicores == null || Number.isNaN(millicores)) return '—'
  if (millicores >= 1000) {
    const cores = millicores / 1000
    const rounded = Number.isInteger(cores) ? cores : Number(cores.toFixed(2))
    return `${rounded} ${rounded === 1 ? 'core' : 'cores'}`
  }
  return `${millicores}m`
}

// Relative time, e.g. "3m ago", "just now", "in 2h".
export function formatRelativeTime(iso: string | undefined | null): string {
  if (!iso) return '—'
  const t = new Date(iso).getTime()
  if (Number.isNaN(t)) return '—'
  const diffMs = t - Date.now()
  const abs = Math.abs(diffMs)
  const sec = Math.round(abs / 1000)
  const min = Math.round(sec / 60)
  const hr = Math.round(min / 60)
  const day = Math.round(hr / 24)

  const suffix = (s: string) => (diffMs < 0 ? `${s} ago` : `in ${s}`)

  if (sec < 5) return 'just now'
  if (sec < 60) return suffix(`${sec}s`)
  if (min < 60) return suffix(`${min}m`)
  if (hr < 24) return suffix(`${hr}h`)
  if (day < 30) return suffix(`${day}d`)
  return new Date(iso).toLocaleDateString()
}

// Absolute timestamp for tooltips / detail.
export function formatTimestamp(iso: string | undefined | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString()
}

// key=value list from a label map.
export function formatLabels(
  labels: Record<string, string> | undefined,
): string {
  if (!labels) return '—'
  const entries = Object.entries(labels)
  if (entries.length === 0) return '—'
  return entries.map(([k, v]) => `${k}=${v}`).join(', ')
}

export function formatSelector(
  selector: Record<string, string> | undefined,
): string {
  return formatLabels(selector)
}
