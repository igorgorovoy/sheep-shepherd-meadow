import { useState } from 'react'
import { useToast } from '../contexts/ToastContext'
import { useMutations } from '../hooks/useMutations'

export function ApplyDrawer({
  open,
  onClose,
  onSuccess,
}: {
  open: boolean
  onClose: () => void
  onSuccess: () => void
}) {
  const [json, setJson] = useState('')
  const [busy, setBusy] = useState(false)
  const { apply } = useMutations()
  const { push } = useToast()

  if (!open) return null

  async function submit() {
    setBusy(true)
    try {
      await apply(json)
      setJson('')
      onSuccess()
      onClose()
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Apply failed'
      push(msg, 'err')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="drawer-scrim" onClick={onClose}>
      <aside
        className="drawer drawer--wide"
        role="dialog"
        aria-labelledby="apply-title"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="drawer__head">
          <h2 id="apply-title">Apply resource</h2>
          <button
            type="button"
            className="btn btn--icon"
            onClick={onClose}
            aria-label="Close"
          >
            ×
          </button>
        </header>
        <div className="drawer__body">
          <p className="muted">
            Paste a JSON manifest with <code>kind</code>: Pod, Service, or
            Deployment.
          </p>
          <textarea
            className="apply-textarea mono"
            value={json}
            onChange={(e) => setJson(e.target.value)}
            rows={16}
            placeholder='{"kind":"Pod", ...}'
            spellCheck={false}
          />
        </div>
        <footer className="drawer__foot">
          <button type="button" className="btn" onClick={onClose} disabled={busy}>
            Cancel
          </button>
          <button
            type="button"
            className="btn btn--primary"
            onClick={() => void submit()}
            disabled={busy || !json.trim()}
          >
            {busy ? 'Applying…' : 'Apply'}
          </button>
        </footer>
      </aside>
    </div>
  )
}
