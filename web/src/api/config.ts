const KEYS = {
  shepherdUrl: 'shepherd:apiUrl',
  shepherdToken: 'shepherd:token',
  meadowUrl: 'meadow:apiUrl',
  meadowToken: 'meadow:token',
} as const

export const DEFAULT_SHEPHERD_API = (
  import.meta.env.VITE_SHEPHERD_API ?? 'http://localhost:9876'
).replace(/\/+$/, '')

export const DEFAULT_MEADOW_API = (
  import.meta.env.VITE_MEADOW_API ?? 'http://localhost:5555'
).replace(/\/+$/, '')

export function getShepherdApiBase(): string {
  if (typeof window === 'undefined') return DEFAULT_SHEPHERD_API
  const v = localStorage.getItem(KEYS.shepherdUrl)?.trim()
  return v || DEFAULT_SHEPHERD_API
}

export function getShepherdToken(): string | null {
  if (typeof window === 'undefined') return null
  const v = localStorage.getItem(KEYS.shepherdToken)?.trim()
  return v || null
}

export function setShepherdApiBase(url: string) {
  localStorage.setItem(KEYS.shepherdUrl, url.replace(/\/+$/, ''))
}

export function setShepherdToken(token: string) {
  localStorage.setItem(KEYS.shepherdToken, token)
}

export function getMeadowApiBase(): string {
  if (typeof window === 'undefined') return DEFAULT_MEADOW_API
  const v = localStorage.getItem(KEYS.meadowUrl)?.trim()
  return v || DEFAULT_MEADOW_API
}

export function getMeadowToken(): string | null {
  if (typeof window === 'undefined') return null
  const v = localStorage.getItem(KEYS.meadowToken)?.trim()
  return v || null
}

export function setMeadowApiBase(url: string) {
  localStorage.setItem(KEYS.meadowUrl, url.replace(/\/+$/, ''))
}

export function setMeadowToken(token: string) {
  localStorage.setItem(KEYS.meadowToken, token)
}

export function shepherdAuthHeaders(): Record<string, string> {
  const h: Record<string, string> = { Accept: 'application/json' }
  const token = getShepherdToken()
  if (token) h.Authorization = `Bearer ${token}`
  return h
}

export function meadowAuthHeaders(): Record<string, string> {
  const h: Record<string, string> = { Accept: 'application/json' }
  const token = getMeadowToken()
  if (token) h.Authorization = `Bearer ${token}`
  return h
}
