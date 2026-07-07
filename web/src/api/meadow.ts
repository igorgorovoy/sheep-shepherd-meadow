import {
  getMeadowApiBase,
  meadowAuthHeaders,
} from './config'
import type { MeadowCatalog, MeadowStats } from './types'

export class MeadowError extends Error {
  status?: number
  constructor(message: string, opts?: { status?: number }) {
    super(message)
    this.name = 'MeadowError'
    this.status = opts?.status
  }
}

async function meadowRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const base = getMeadowApiBase()
  const url = `${base}${path}`
  let res: Response
  try {
    res = await fetch(url, {
      ...init,
      headers: {
        ...meadowAuthHeaders(),
        ...(init?.headers as Record<string, string> | undefined),
      },
    })
  } catch (err) {
    throw new MeadowError(`Cannot reach Meadow API at ${base}`, {
      status: undefined,
    })
  }
  if (!res.ok) {
    const body = (await res.text()).trim()
    throw new MeadowError(
      body ? `Meadow ${path} failed (${res.status}): ${body}` : `Meadow ${path} failed (${res.status})`,
      { status: res.status },
    )
  }
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export function fetchMeadowAuthStatus(signal?: AbortSignal): Promise<{ auth_required: boolean }> {
  return meadowRequest('/meadow/auth/status', { signal })
}

export function fetchMeadowStats(signal?: AbortSignal): Promise<MeadowStats> {
  return meadowRequest('/meadow/stats', { signal })
}

export function fetchMeadowRepoStats(
  repo: string,
  signal?: AbortSignal,
): Promise<MeadowStats['details'][number]> {
  return meadowRequest(`/meadow/stats/${encodeURIComponent(repo)}`, { signal })
}

export function fetchMeadowCatalog(signal?: AbortSignal): Promise<MeadowCatalog> {
  return meadowRequest('/v2/_catalog', { signal })
}

export function fetchMeadowTags(
  repo: string,
  signal?: AbortSignal,
): Promise<{ name: string; tags: string[] }> {
  return meadowRequest(`/v2/${encodeURIComponent(repo)}/tags/list`, { signal })
}

export function deleteMeadowManifest(repo: string, tag: string): Promise<void> {
  return meadowRequest(
    `/v2/${encodeURIComponent(repo)}/manifests/${encodeURIComponent(tag)}`,
    { method: 'DELETE' },
  )
}
