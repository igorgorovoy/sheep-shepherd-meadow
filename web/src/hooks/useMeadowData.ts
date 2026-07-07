import { useCallback, useEffect, useRef, useState } from 'react'
import { fetchMeadowStats } from '../api/meadow'
import type { MeadowStats } from '../api/types'

const POLL_MS = 10000

export function useMeadowData() {
  const [data, setData] = useState<MeadowStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const mounted = useRef(true)

  const load = useCallback(async () => {
    const controller = new AbortController()
    try {
      const stats = await fetchMeadowStats(controller.signal)
      if (!mounted.current) return
      setData(stats)
      setError(null)
    } catch (err) {
      if (!mounted.current) return
      setError(err instanceof Error ? err.message : 'Meadow unreachable')
    } finally {
      if (mounted.current) setLoading(false)
    }
  }, [])

  useEffect(() => {
    mounted.current = true
    void load()
    const id = window.setInterval(() => void load(), POLL_MS)
    return () => {
      mounted.current = false
      window.clearInterval(id)
    }
  }, [load])

  return { data, loading, error, refresh: load }
}
