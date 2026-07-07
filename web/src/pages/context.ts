import { useOutletContext } from 'react-router-dom'
import type { ClusterDataState } from '../hooks/useClusterData'
import type { NamespaceFilter } from '../api/types'

export interface PageContext extends ClusterDataState {
  namespace: NamespaceFilter
  setNamespace: (ns: NamespaceFilter) => void
}

export function usePageData(): PageContext {
  return useOutletContext<PageContext>()
}
