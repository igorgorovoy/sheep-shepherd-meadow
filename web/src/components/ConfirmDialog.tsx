import { useEffect, useState } from 'react'

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel,
  expectedName,
  onConfirm,
  onCancel,
  busy,
}: {
  open: boolean
  title: string
  description: string
  confirmLabel?: string
  expectedName: string
  onConfirm: () => void | Promise<void>
  onCancel: () => void
  busy?: boolean
}) {
  const [typed, setTyped] = useState('')

  useEffect(() => {
    if (open) setTyped('')
  }, [open, expectedName])

  if (!open) return null

  const ok = typed === expectedName

  return (
    <div className="modal-scrim" onClick={onCancel}>
      <div
        className="modal"
        role="alertdialog"
        aria-labelledby="confirm-title"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="confirm-title">{title}</h2>
        <p className="modal__desc">{description}</p>
        <label className="field">
          <span className="field__label">
            Type <strong className="mono">{expectedName}</strong> to confirm
          </span>
          <input
            className="field__input mono"
            value={typed}
            onChange={(e) => setTyped(e.target.value)}
            autoFocus
          />
        </label>
        <div className="modal__actions">
          <button type="button" className="btn" onClick={onCancel} disabled={busy}>
            Cancel
          </button>
          <button
            type="button"
            className="btn btn--danger"
            disabled={!ok || busy}
            onClick={() => void onConfirm()}
          >
            {busy ? 'Deleting…' : (confirmLabel ?? 'Delete')}
          </button>
        </div>
      </div>
    </div>
  )
}
