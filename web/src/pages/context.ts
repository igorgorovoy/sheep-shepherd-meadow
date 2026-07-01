import { useOutletContext } from 'react-router-dom'
import type { ClusterDataState } from '../hooks/useClusterData'

// Context passed from the Layout <Outlet> down to each page.
export type PageContext = ClusterDataState

export function usePageData(): PageContext {
  return useOutletContext<PageContext>()
}
