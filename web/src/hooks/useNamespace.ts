import { useCallback, useEffect, useState } from 'react'
import type { NamespaceFilter } from '../api/types'

const STORAGE_KEY = 'shepherd:namespace'

function readStored(): NamespaceFilter {
  if (typeof window === 'undefined') return 'all'
  const v = localStorage.getItem(STORAGE_KEY)
  return v && v.length > 0 ? v : 'all'
}

export function useNamespace() {
  const [namespace, setNamespaceState] = useState<NamespaceFilter>(readStored)

  const setNamespace = useCallback((ns: NamespaceFilter) => {
    setNamespaceState(ns)
    localStorage.setItem(STORAGE_KEY, ns)
  }, [])

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, namespace)
  }, [namespace])

  return { namespace, setNamespace }
}
