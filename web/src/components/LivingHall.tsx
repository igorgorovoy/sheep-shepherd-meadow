import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ClusterSummary } from '../api/types'
import type { HallHandle } from '../game/bootstrap'

type Status = 'loading' | 'ready' | 'error'

function prefersReducedMotion(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

export function LivingHall({
  summary,
  meadowRepoCount = 0,
}: {
  summary: ClusterSummary | null
  meadowRepoCount?: number
}) {
  const navigate = useNavigate()
  const mountRef = useRef<HTMLDivElement | null>(null)
  const handleRef = useRef<HallHandle | null>(null)
  const [status, setStatus] = useState<Status>('loading')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  const latest = useRef({ summary, meadowRepoCount })
  latest.current = { summary, meadowRepoCount }

  const onNavigateRef = useRef(navigate)
  onNavigateRef.current = navigate

  useEffect(() => {
    let cancelled = false
    const controller = new AbortController()
    const parent = mountRef.current
    if (!parent) return

    setStatus('loading')
    setErrorMsg(null)

    import('../game/bootstrap')
      .then(({ bootstrapHall }) =>
        bootstrapHall({
          parent,
          reducedMotion: prefersReducedMotion(),
          signal: controller.signal,
          onNavigate: (path) => onNavigateRef.current(path),
        }),
      )
      .then((handle) => {
        if (cancelled) {
          handle.destroy()
          return
        }
        handleRef.current = handle
        const { summary: s, meadowRepoCount: m } = latest.current
        handle.apply(s, m)
        setStatus('ready')
      })
      .catch((err: unknown) => {
        if (cancelled) return
        console.error('[living-hall] failed to initialise', err)
        setErrorMsg(err instanceof Error ? err.message : 'Unknown error')
        setStatus('error')
      })

    return () => {
      cancelled = true
      controller.abort()
      handleRef.current?.destroy()
      handleRef.current = null
    }
  }, [])

  useEffect(() => {
    if (status === 'ready') {
      handleRef.current?.apply(summary, meadowRepoCount)
    }
  }, [summary, meadowRepoCount, status])

  return (
    <div className="living-hall">
      <div
        ref={mountRef}
        className="living-hall__stage"
        role="img"
        aria-label="Animated cluster visualization. Click stations, sheep, or the vault for details."
      />
      {status === 'loading' && (
        <div className="living-hall__overlay" aria-live="polite">
          <span className="living-hall__spinner" aria-hidden />
          <span>Firing up the hall…</span>
        </div>
      )}
      {status === 'error' && (
        <div className="living-hall__overlay living-hall__overlay--error">
          <strong>The Living Hall could not start.</strong>
          <p className="muted">
            The animated view needs WebGL/Canvas and the sprite manifest. The
            rest of the dashboard is unaffected — use the resource tables to
            inspect the cluster.
          </p>
          {errorMsg && (
            <code className="living-hall__err">{errorMsg}</code>
          )}
        </div>
      )}
    </div>
  )
}
