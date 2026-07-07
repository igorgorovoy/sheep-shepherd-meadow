import { useCallback } from 'react'
import {
  createDeployment,
  createPod,
  createService,
  deleteDeployment,
  deleteNode,
  deletePod,
  deleteService,
  updateDeployment,
} from '../api/client'
import type { Deployment } from '../api/types'
import { useToast } from '../contexts/ToastContext'

export function useMutations() {
  const { push } = useToast()

  const apply = useCallback(
    async (raw: string) => {
      let parsed: Record<string, unknown>
      try {
        parsed = JSON.parse(raw) as Record<string, unknown>
      } catch {
        throw new Error('Invalid JSON')
      }
      const kind = String(parsed.kind ?? '').toLowerCase()
      switch (kind) {
        case 'pod': {
          const pod = await createPod(parsed)
          push(`Pod ${pod.metadata.name} created`)
          return
        }
        case 'service': {
          const svc = await createService(parsed)
          push(`Service ${svc.metadata.name} created`)
          return
        }
        case 'deployment': {
          const dep = await createDeployment(parsed)
          push(`Deployment ${dep.metadata.name} created`)
          return
        }
        default:
          throw new Error(`Unsupported kind: ${kind || '(missing)'}`)
      }
    },
    [push],
  )

  const scaleDeployment = useCallback(
    async (dep: Deployment, replicas: number) => {
      const updated = structuredClone(dep)
      updated.spec = { ...updated.spec, replicas }
      await updateDeployment(dep.metadata.namespace, dep.metadata.name, updated)
      push(`Deployment ${dep.metadata.name} scaled to ${replicas}`)
    },
    [push],
  )

  const removePod = useCallback(
    async (ns: string, name: string) => {
      await deletePod(ns, name)
      push(`Pod ${name} deleted`)
    },
    [push],
  )

  const removeService = useCallback(
    async (ns: string, name: string) => {
      await deleteService(ns, name)
      push(`Service ${name} deleted`)
    },
    [push],
  )

  const removeDeployment = useCallback(
    async (ns: string, name: string) => {
      await deleteDeployment(ns, name)
      push(`Deployment ${name} deleted`)
    },
    [push],
  )

  const removeNode = useCallback(
    async (name: string) => {
      await deleteNode(name)
      push(`Node ${name} deleted`)
    },
    [push],
  )

  return {
    apply,
    scaleDeployment,
    removePod,
    removeService,
    removeDeployment,
    removeNode,
  }
}
