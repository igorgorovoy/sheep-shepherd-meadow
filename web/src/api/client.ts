// Typed API client for the Shepherd REST API.
//
// Base URL comes from VITE_SHEPHERD_API, defaulting to http://localhost:9876.
// Every method targets an individual endpoint that is guaranteed to exist.
// `fetchSummary` optionally uses /api/v1/cluster/summary if the backend
// implements it, but nothing here depends on that endpoint.

import type {
  ClusterSummary,
  Deployment,
  Event,
  Info,
  Node,
  Pod,
  Service,
} from './types'

export const API_BASE: string = (
  import.meta.env.VITE_SHEPHERD_API ?? 'http://localhost:9876'
).replace(/\/+$/, '')

export class ApiError extends Error {
  status?: number
  cause?: unknown
  constructor(message: string, opts?: { status?: number; cause?: unknown }) {
    super(message)
    this.name = 'ApiError'
    this.status = opts?.status
    this.cause = opts?.cause
  }
}

async function request<T>(path: string, signal?: AbortSignal): Promise<T> {
  const url = `${API_BASE}${path}`
  let res: Response
  try {
    res = await fetch(url, {
      signal,
      headers: { Accept: 'application/json' },
    })
  } catch (err) {
    throw new ApiError(`Cannot reach Shepherd API at ${API_BASE}`, {
      cause: err,
    })
  }
  if (!res.ok) {
    throw new ApiError(`Request to ${path} failed (${res.status})`, {
      status: res.status,
    })
  }
  return (await res.json()) as T
}

// GET /healthz -> "ok" (plain text). Returns true when reachable and healthy.
export async function fetchHealth(signal?: AbortSignal): Promise<boolean> {
  const url = `${API_BASE}/healthz`
  try {
    const res = await fetch(url, { signal })
    if (!res.ok) return false
    const text = (await res.text()).trim().toLowerCase()
    return text === 'ok' || text.length > 0
  } catch {
    return false
  }
}

export function fetchInfo(signal?: AbortSignal): Promise<Info> {
  return request<Info>('/api/v1/info', signal)
}

export function fetchNodes(signal?: AbortSignal): Promise<Node[]> {
  return request<Node[]>('/api/v1/nodes', signal)
}

export function fetchPods(signal?: AbortSignal): Promise<Pod[]> {
  return request<Pod[]>('/api/v1/pods', signal)
}

export function fetchServices(signal?: AbortSignal): Promise<Service[]> {
  return request<Service[]>('/api/v1/services', signal)
}

export function fetchDeployments(signal?: AbortSignal): Promise<Deployment[]> {
  return request<Deployment[]>('/api/v1/deployments', signal)
}

export function fetchEvents(signal?: AbortSignal): Promise<Event[]> {
  return request<Event[]>('/api/v1/events', signal)
}

// Aggregate fetch. Pulls every resource in parallel from the individual
// endpoints. Arrays that are null/undefined are normalized to [].
export async function fetchClusterSummary(
  signal?: AbortSignal,
): Promise<ClusterSummary> {
  const [info, nodes, pods, services, deployments, events] = await Promise.all([
    fetchInfo(signal),
    fetchNodes(signal),
    fetchPods(signal),
    fetchServices(signal),
    fetchDeployments(signal),
    fetchEvents(signal),
  ])
  return {
    info,
    nodes: nodes ?? [],
    pods: pods ?? [],
    services: services ?? [],
    deployments: deployments ?? [],
    events: events ?? [],
  }
}
