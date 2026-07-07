import { useState } from 'react'
import type { Deployment } from '../api/types'
import { useToast } from '../contexts/ToastContext'
import { useMutations } from '../hooks/useMutations'

export function ScaleControl({
  deployment,
  onScaled,
}: {
  deployment: Deployment
  onScaled: () => void
}) {
  const [replicas, setReplicas] = useState(deployment.spec?.replicas ?? 1)
  const [busy, setBusy] = useState(false)
  const { scaleDeployment } = useMutations()
  const { push } = useToast()

  async function submit() {
    if (replicas < 0) {
      push('Replicas must be ≥ 0', 'err')
      return
    }
    setBusy(true)
    try {
      await scaleDeployment(deployment, replicas)
      onScaled()
    } catch (err) {
      push(err instanceof Error ? err.message : 'Scale failed', 'err')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="scale-control">
      <label className="scale-control__label">
        Scale replicas
        <input
          type="number"
          min={0}
          className="scale-control__input mono"
          value={replicas}
          onChange={(e) => setReplicas(Number(e.target.value))}
        />
      </label>
      <button
        type="button"
        className="btn btn--primary"
        onClick={() => void submit()}
        disabled={busy}
      >
        {busy ? 'Scaling…' : 'Scale'}
      </button>
    </div>
  )
}
