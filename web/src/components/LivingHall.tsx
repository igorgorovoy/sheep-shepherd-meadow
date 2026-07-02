import { useEffect, useRef, useState } from 'react'
import type { ClusterSummary } from '../api/types'
import type { HallHandle } from '../game/bootstrap'

// LivingHall mounts a lazily-loaded Phaser scene into a <div> and streams
// cluster snapshots into it WITHOUT re-mounting. Phaser and the scene are
// pulled in via dynamic import() from ../game/bootstrap so they code-split out
// of the initial bundle.
//
// Data flow:
//   - `summary` (the polled ClusterSummary) is pushed into the running scene
//     on every change via handle.apply(); the scene diffs it internally.
//   - The component NEVER re-creates the game when data changes.
//
// Graceful degradation:
//   - If Phaser/WebGL init or the sprite manifest fetch fails, we surface a
//     fallback notice and keep the rest of the app usable.
//   - prefers-reduced-motion disables idle motion (state is still shown).

type Status = 'loading' | 'ready' | 'error'

function prefersReducedMotion(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
  )
}

export function LivingHall({ summary }: { summary: ClusterSummary | null }) {
  const mountRef = useRef<HTMLDivElement | null>(null)
  const handleRef = useRef<HallHandle | null>(null)
  const [status, setStatus] = useState<Status>('loading')
  const [errorMsg, setErrorMsg] = useState<string | null>(null)

  // Keep the freshest summary in a ref so the async bootstrap can apply it
  // as soon as the scene is live, even if data arrived first.
  const latest = useRef<ClusterSummary | null>(summary)
  latest.current = summary

  // Mount the game exactly once.
  useEffect(() => {
    let cancelled = false
    const controller = new AbortController()
    const parent = mountRef.current
    if (!parent) return

    setStatus('loading')
    setErrorMsg(null)

    // Dynamic import -> Phaser lands in its own chunk.
    import('../game/bootstrap')
      .then(({ bootstrapHall }) =>
        bootstrapHall({
          parent,
          reducedMotion: prefersReducedMotion(),
          signal: controller.signal,
        }),
      )
      .then((handle) => {
        if (cancelled) {
          handle.destroy()
          return
        }
        handleRef.current = handle
        handle.apply(latest.current)
        setStatus('ready')
      })
      .catch((err: unknown) => {
        if (cancelled) return
        // eslint-disable-next-line no-console
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

  // Push new snapshots into the live scene (no re-mount).
  useEffect(() => {
    if (status === 'ready') {
      handleRef.current?.apply(summary)
    }
  }, [summary, status])

  return (
    <div className="living-hall">
      <div
        ref={mountRef}
        className="living-hall__stage"
        role="img"
        aria-label="Animated cluster visualization: nodes as forge stations, pods as sheep"
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
