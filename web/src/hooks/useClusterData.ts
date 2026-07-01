import { useCallback, useEffect, useRef, useState } from 'react'
import {
  fetchClusterSummary,
  fetchHealth,
} from '../api/client'
import type { ClusterSummary } from '../api/types'

const POLL_INTERVAL_MS = 5000

export interface ClusterDataState {
  data: ClusterSummary | null
  healthy: boolean
  loading: boolean // true only on the very first load
  refreshing: boolean // true whenever a fetch is in-flight
  error: string | null
  lastUpdated: Date | null
  refresh: () => void
}

// Polls the whole cluster state every 5s, exposes a manual refresh, and
// keeps the last successful snapshot on screen while refetching.
export function useClusterData(): ClusterDataState {
  const [data, setData] = useState<ClusterSummary | null>(null)
  const [healthy, setHealthy] = useState(false)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  const inFlight = useRef<AbortController | null>(null)
  const mounted = useRef(true)

  const load = useCallback(async () => {
    inFlight.current?.abort()
    const controller = new AbortController()
    inFlight.current = controller
    setRefreshing(true)

    // Health check is independent and never throws.
    void fetchHealth(controller.signal).then((h) => {
      if (mounted.current) setHealthy(h)
    })

    try {
      const summary = await fetchClusterSummary(controller.signal)
      if (!mounted.current || controller.signal.aborted) return
      setData(summary)
      setError(null)
      setLastUpdated(new Date())
    } catch (err) {
      if (controller.signal.aborted) return
      if (!mounted.current) return
      setHealthy(false)
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      if (mounted.current && !controller.signal.aborted) {
        setLoading(false)
        setRefreshing(false)
      }
    }
  }, [])

  const refresh = useCallback(() => {
    void load()
  }, [load])

  useEffect(() => {
    mounted.current = true
    void load()
    const id = window.setInterval(() => void load(), POLL_INTERVAL_MS)
    return () => {
      mounted.current = false
      window.clearInterval(id)
      inFlight.current?.abort()
    }
  }, [load])

  return { data, healthy, loading, refreshing, error, lastUpdated, refresh }
}
