import { useCallback, useEffect, useRef, useState } from 'react'
import { ApiError } from '../api/client'

export interface ResourceState<T> {
  data: T | null
  loading: boolean
  error: string | null
  refresh: () => void
}

export function useResource<T>(
  fetcher: (signal: AbortSignal) => Promise<T>,
  deps: unknown[],
): ResourceState<T> {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const mounted = useRef(true)

  const load = useCallback(async () => {
    const controller = new AbortController()
    setLoading(true)
    try {
      const result = await fetcher(controller.signal)
      if (!mounted.current || controller.signal.aborted) return
      setData(result)
      setError(null)
    } catch (err) {
      if (controller.signal.aborted || !mounted.current) return
      setError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      if (mounted.current && !controller.signal.aborted) {
        setLoading(false)
      }
    }
    return () => controller.abort()
  }, deps) // eslint-disable-line react-hooks/exhaustive-deps

  const refresh = useCallback(() => {
    void load()
  }, [load])

  useEffect(() => {
    mounted.current = true
    void load()
    return () => {
      mounted.current = false
    }
  }, [load])

  return { data, loading, error, refresh }
}

export function resourceNotFoundMessage(err: unknown): boolean {
  return err instanceof ApiError && err.status === 404
}
