// Typed API client for the Shepherd REST API.

import {
  getShepherdApiBase,
  shepherdAuthHeaders,
} from './config'
import type {
  AuthStatus,
  ClusterSummary,
  Deployment,
  Event,
  Info,
  NamespaceFilter,
  Node,
  Pod,
  Service,
} from './types'

export class ApiError extends Error {
  status?: number
  body?: string
  cause?: unknown
  constructor(message: string, opts?: { status?: number; body?: string; cause?: unknown }) {
    super(message)
    this.name = 'ApiError'
    this.status = opts?.status
    this.body = opts?.body
    this.cause = opts?.cause
  }
}

function enc(segment: string): string {
  return encodeURIComponent(segment)
}

function summaryPath(namespace: NamespaceFilter): string {
  const q =
    namespace && namespace !== 'all'
      ? `?namespace=${encodeURIComponent(namespace)}`
      : ''
  return `/api/v1/cluster/summary${q}`
}

async function parseError(res: Response, path: string): Promise<ApiError> {
  const body = (await res.text()).trim()
  return new ApiError(
    body ? `Request to ${path} failed (${res.status}): ${body}` : `Request to ${path} failed (${res.status})`,
    { status: res.status, body },
  )
}

async function request<T>(
  path: string,
  init?: RequestInit & { signal?: AbortSignal },
): Promise<T> {
  const base = getShepherdApiBase()
  const url = `${base}${path}`
  let res: Response
  try {
    res = await fetch(url, {
      ...init,
      headers: {
        ...shepherdAuthHeaders(),
        ...(init?.headers as Record<string, string> | undefined),
      },
    })
  } catch (err) {
    throw new ApiError(`Cannot reach Shepherd API at ${base}`, { cause: err })
  }
  if (!res.ok) throw await parseError(res, path)
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export async function fetchHealth(signal?: AbortSignal): Promise<boolean> {
  const url = `${getShepherdApiBase()}/healthz`
  try {
    const res = await fetch(url, { signal })
    if (!res.ok) return false
    const text = (await res.text()).trim().toLowerCase()
    return text === 'ok' || text.length > 0
  } catch {
    return false
  }
}

export function fetchAuthStatus(signal?: AbortSignal): Promise<AuthStatus> {
  return request<AuthStatus>('/api/v1/auth/status', { signal })
}

export function fetchInfo(signal?: AbortSignal): Promise<Info> {
  return request<Info>('/api/v1/info', { signal })
}

export function fetchNodes(signal?: AbortSignal): Promise<Node[]> {
  return request<Node[]>('/api/v1/nodes', { signal })
}

export function fetchPods(signal?: AbortSignal): Promise<Pod[]> {
  return request<Pod[]>('/api/v1/pods', { signal })
}

export function fetchServices(signal?: AbortSignal): Promise<Service[]> {
  return request<Service[]>('/api/v1/services', { signal })
}

export function fetchDeployments(signal?: AbortSignal): Promise<Deployment[]> {
  return request<Deployment[]>('/api/v1/deployments', { signal })
}

export function fetchEvents(signal?: AbortSignal): Promise<Event[]> {
  return request<Event[]>('/api/v1/events', { signal })
}

export function fetchClusterSummary(
  namespace: NamespaceFilter = 'all',
  signal?: AbortSignal,
): Promise<ClusterSummary> {
  return request<ClusterSummary>(summaryPath(namespace), { signal })
}

export function fetchNode(name: string, signal?: AbortSignal): Promise<Node> {
  return request<Node>(`/api/v1/nodes/${enc(name)}`, { signal })
}

export function fetchPod(ns: string, name: string, signal?: AbortSignal): Promise<Pod> {
  return request<Pod>(`/api/v1/namespaces/${enc(ns)}/pods/${enc(name)}`, { signal })
}

export function fetchDeployment(
  ns: string,
  name: string,
  signal?: AbortSignal,
): Promise<Deployment> {
  return request<Deployment>(
    `/api/v1/namespaces/${enc(ns)}/deployments/${enc(name)}`,
    { signal },
  )
}

export function fetchService(
  ns: string,
  name: string,
  signal?: AbortSignal,
): Promise<Service> {
  return request<Service>(
    `/api/v1/namespaces/${enc(ns)}/services/${enc(name)}`,
    { signal },
  )
}

export function createPod(body: unknown): Promise<Pod> {
  return request<Pod>('/api/v1/pods', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export function createService(body: unknown): Promise<Service> {
  return request<Service>('/api/v1/services', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export function createDeployment(body: unknown): Promise<Deployment> {
  return request<Deployment>('/api/v1/deployments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
}

export function updateDeployment(
  ns: string,
  name: string,
  body: Deployment,
): Promise<Deployment> {
  return request<Deployment>(
    `/api/v1/namespaces/${enc(ns)}/deployments/${enc(name)}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    },
  )
}

export function deletePod(ns: string, name: string): Promise<void> {
  return request<void>(`/api/v1/namespaces/${enc(ns)}/pods/${enc(name)}`, {
    method: 'DELETE',
  })
}

export function deleteService(ns: string, name: string): Promise<void> {
  return request<void>(`/api/v1/namespaces/${enc(ns)}/services/${enc(name)}`, {
    method: 'DELETE',
  })
}

export function deleteDeployment(ns: string, name: string): Promise<void> {
  return request<void>(
    `/api/v1/namespaces/${enc(ns)}/deployments/${enc(name)}`,
    { method: 'DELETE' },
  )
}

export function deleteNode(name: string): Promise<void> {
  return request<void>(`/api/v1/nodes/${enc(name)}`, { method: 'DELETE' })
}

export function normalizeSummary(raw: ClusterSummary): ClusterSummary {
  return {
    ...raw,
    nodes: raw.nodes ?? [],
    pods: raw.pods ?? [],
    services: raw.services ?? [],
    deployments: raw.deployments ?? [],
    events: raw.events ?? [],
    namespaces: raw.namespaces ?? ['default'],
  }
}
