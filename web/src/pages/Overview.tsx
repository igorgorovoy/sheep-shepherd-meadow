import { PageHeader } from '../components/PageHeader'
import { ErrorState, LoadingRows } from '../components/states'
import { usePageData } from './context'
import type { Pod, PodPhase } from '../api/types'

function StatTile({
  label,
  value,
  sub,
  icon,
}: {
  label: string
  value: string | number
  sub?: string
  icon: string
}) {
  return (
    <div className="stat">
      <span className="stat__label">
        <span aria-hidden>{icon}</span>
        {label}
      </span>
      <span className="stat__value">{value}</span>
      {sub && <span className="stat__sub">{sub}</span>}
    </div>
  )
}

// A single horizontal segmented meter. Each segment carries a grayscale
// fill class (solid / hatch / dot / mid) so it stays legible without color.
function SegmentMeter({
  segments,
}: {
  segments: { key: string; label: string; value: number; fill: string }[]
}) {
  const total = segments.reduce((s, x) => s + x.value, 0)
  return (
    <div className="meter">
      {segments.map((seg) => {
        const pct = total === 0 ? 0 : (seg.value / total) * 100
        return (
          <div className="meter__row" key={seg.key}>
            <span className="meter__key">
              <span className={`meter__seg ${seg.fill}`} style={{ width: 12, height: 12, borderRadius: 3, display: 'inline-block', border: '1px solid var(--border)' }} aria-hidden />
              {seg.label}
            </span>
            <span className="meter__bar">
              <span
                className={`meter__seg ${seg.fill}`}
                style={{ width: `${pct}%` }}
              />
            </span>
            <span className="meter__num">{seg.value}</span>
          </div>
        )
      })}
    </div>
  )
}

const PHASE_ORDER: PodPhase[] = ['Running', 'Pending', 'Succeeded', 'Failed']
const PHASE_FILL: Record<PodPhase, string> = {
  Running: 'fill-solid',
  Pending: 'fill-dot',
  Succeeded: 'fill-mid',
  Failed: 'fill-hatch',
}

function countPhases(pods: Pod[]): Record<PodPhase, number> {
  const out: Record<PodPhase, number> = {
    Running: 0,
    Pending: 0,
    Succeeded: 0,
    Failed: 0,
  }
  for (const p of pods) {
    const phase = p.status?.phase
    if (phase && phase in out) out[phase] += 1
  }
  return out
}

export function Overview() {
  const { data, loading, error, healthy } = usePageData()

  if (loading && !data) {
    return (
      <>
        <PageHeader title="Overview" />
        <LoadingRows rows={4} />
      </>
    )
  }
  if (error && !data) {
    return (
      <>
        <PageHeader title="Overview" />
        <ErrorState message={error} />
      </>
    )
  }

  const info = data?.info
  const nodes = data?.nodes ?? []
  const pods = data?.pods ?? []
  const deployments = data?.deployments ?? []
  const services = data?.services ?? []

  const nodesReady = nodes.filter((n) => n.status?.condition === 'Ready').length
  const nodesNotReady = nodes.length - nodesReady
  const phases = countPhases(pods)

  return (
    <>
      <PageHeader
        title="Overview"
        description={
          <>
            {info?.name ?? 'cluster'}
            {info?.version ? ` · v${info.version}` : ''} ·{' '}
            {healthy ? 'control plane reachable' : 'control plane unreachable'}
          </>
        }
      />

      <div className="stack">
        <div className="grid grid--stats">
          <StatTile
            icon="▦"
            label="Nodes"
            value={nodes.length}
            sub={`${nodesReady} ready · ${nodesNotReady} not ready`}
          />
          <StatTile
            icon="◧"
            label="Pods"
            value={pods.length}
            sub={`${phases.Running} running · ${phases.Pending} pending`}
          />
          <StatTile
            icon="⧉"
            label="Deployments"
            value={deployments.length}
          />
          <StatTile icon="⇄" label="Services" value={services.length} />
        </div>

        <div className="grid grid--halves">
          <div className="card">
            <div className="card__head">
              <span className="card__title">Node health</span>
            </div>
            <div className="card__body">
              {nodes.length === 0 ? (
                <p className="muted">No nodes registered.</p>
              ) : (
                <SegmentMeter
                  segments={[
                    {
                      key: 'ready',
                      label: 'Ready',
                      value: nodesReady,
                      fill: 'fill-solid',
                    },
                    {
                      key: 'notready',
                      label: 'NotReady',
                      value: nodesNotReady,
                      fill: 'fill-hatch',
                    },
                  ]}
                />
              )}
            </div>
          </div>

          <div className="card">
            <div className="card__head">
              <span className="card__title">Pods by phase</span>
            </div>
            <div className="card__body">
              {pods.length === 0 ? (
                <p className="muted">No pods scheduled.</p>
              ) : (
                <SegmentMeter
                  segments={PHASE_ORDER.map((phase) => ({
                    key: phase,
                    label: phase,
                    value: phases[phase],
                    fill: PHASE_FILL[phase],
                  }))}
                />
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
