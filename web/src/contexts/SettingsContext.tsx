import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { fetchAuthStatus } from '../api/client'
import { fetchMeadowAuthStatus } from '../api/meadow'
import {
  DEFAULT_MEADOW_API,
  DEFAULT_SHEPHERD_API,
  getMeadowApiBase,
  getMeadowToken,
  getShepherdApiBase,
  getShepherdToken,
  setMeadowApiBase,
  setMeadowToken,
  setShepherdApiBase,
  setShepherdToken,
} from '../api/config'

interface SettingsContextValue {
  shepherdUrl: string
  shepherdToken: string
  meadowUrl: string
  meadowToken: string
  shepherdAuthRequired: boolean | null
  meadowAuthRequired: boolean | null
  setShepherdUrl: (v: string) => void
  setShepherdTokenValue: (v: string) => void
  setMeadowUrl: (v: string) => void
  setMeadowTokenValue: (v: string) => void
  refreshAuthStatus: () => void
  settingsOpen: boolean
  openSettings: () => void
  closeSettings: () => void
}

const SettingsContext = createContext<SettingsContextValue | null>(null)

export function SettingsProvider({ children }: { children: ReactNode }) {
  const [shepherdUrl, setShepherdUrlState] = useState(getShepherdApiBase)
  const [shepherdToken, setShepherdTokenState] = useState(
    () => getShepherdToken() ?? '',
  )
  const [meadowUrl, setMeadowUrlState] = useState(getMeadowApiBase)
  const [meadowToken, setMeadowTokenState] = useState(
    () => getMeadowToken() ?? '',
  )
  const [shepherdAuthRequired, setShepherdAuthRequired] = useState<boolean | null>(null)
  const [meadowAuthRequired, setMeadowAuthRequired] = useState<boolean | null>(null)
  const [settingsOpen, setSettingsOpen] = useState(false)

  const refreshAuthStatus = useCallback(() => {
    void fetchAuthStatus()
      .then((s) => setShepherdAuthRequired(s.auth_required))
      .catch(() => setShepherdAuthRequired(null))
    void fetchMeadowAuthStatus()
      .then((s) => setMeadowAuthRequired(s.auth_required))
      .catch(() => setMeadowAuthRequired(null))
  }, [])

  useEffect(() => {
    refreshAuthStatus()
  }, [refreshAuthStatus, shepherdUrl, meadowUrl])

  const setShepherdUrl = useCallback((v: string) => {
    const url = v.trim() || DEFAULT_SHEPHERD_API
    setShepherdApiBase(url)
    setShepherdUrlState(url)
  }, [])

  const setShepherdTokenValue = useCallback((v: string) => {
    setShepherdToken(v.trim())
    setShepherdTokenState(v.trim())
  }, [])

  const setMeadowUrl = useCallback((v: string) => {
    const url = v.trim() || DEFAULT_MEADOW_API
    setMeadowApiBase(url)
    setMeadowUrlState(url)
  }, [])

  const setMeadowTokenValue = useCallback((v: string) => {
    setMeadowToken(v.trim())
    setMeadowTokenState(v.trim())
  }, [])

  const value = useMemo(
    () => ({
      shepherdUrl,
      shepherdToken,
      meadowUrl,
      meadowToken,
      shepherdAuthRequired,
      meadowAuthRequired,
      setShepherdUrl,
      setShepherdTokenValue,
      setMeadowUrl,
      setMeadowTokenValue,
      refreshAuthStatus,
      settingsOpen,
      openSettings: () => setSettingsOpen(true),
      closeSettings: () => setSettingsOpen(false),
    }),
    [
      shepherdUrl,
      shepherdToken,
      meadowUrl,
      meadowToken,
      shepherdAuthRequired,
      meadowAuthRequired,
      setShepherdUrl,
      setShepherdTokenValue,
      setMeadowUrl,
      setMeadowTokenValue,
      refreshAuthStatus,
      settingsOpen,
    ],
  )

  return (
    <SettingsContext.Provider value={value}>{children}</SettingsContext.Provider>
  )
}

export function useSettings(): SettingsContextValue {
  const ctx = useContext(SettingsContext)
  if (!ctx) throw new Error('useSettings must be used within SettingsProvider')
  return ctx
}
