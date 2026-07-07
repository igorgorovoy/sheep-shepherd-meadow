import { useCallback, useEffect, useRef, useState } from 'react'
import {
  fetchClusterSummary,
  fetchHealth,
  normalizeSummary,
} from '../api/client'
import type { ClusterSummary, NamespaceFilter } from '../api/types'

const POLL_INTERVAL_MS = 5000

export interface ClusterDataState {
  data: ClusterSummary | null
  healthy: boolean
  loading: boolean
  refreshing: boolean
  error: string | null
  lastUpdated: Date | null
  refresh: () => void
}

export function useClusterData(
  namespace: NamespaceFilter,
  configKey = '',
): ClusterDataState {
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

    void fetchHealth(controller.signal).then((h) => {
      if (mounted.current) setHealthy(h)
    })

    try {
      const summary = normalizeSummary(
        await fetchClusterSummary(namespace, controller.signal),
      )
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
  }, [namespace, configKey])

  const refresh = useCallback(() => {
    void load()
  }, [load])

  useEffect(() => {
    mounted.current = true
    setLoading((prev) => (data ? prev : true))
    void load()
    const id = window.setInterval(() => void load(), POLL_INTERVAL_MS)
    return () => {
      mounted.current = false
      window.clearInterval(id)
      inFlight.current?.abort()
    }
  }, [load]) // eslint-disable-line react-hooks/exhaustive-deps

  return { data, healthy, loading, refreshing, error, lastUpdated, refresh }
}
