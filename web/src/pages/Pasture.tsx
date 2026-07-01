import { PageHeader } from '../components/PageHeader'
import { usePageData } from './context'

/*
 * ==========================================================================
 * EXTENSIBILITY SEAM — farm-themed cluster visualization ("Pasture")
 * ==========================================================================
 *
 * This route is intentionally a placeholder. A LATER task will implement an
 * animated, farm-themed view where:
 *
 *   - Each NODE is drawn as a "pen" / "yard" (a bordered enclosure).
 *   - Each POD scheduled on that node is drawn as a "sheep" inside the pen.
 *   - Pod phase drives the sheep's appearance (grazing / resting / stray),
 *     conveyed in MONOCHROME only (fill, stroke, dashed outlines, motion) —
 *     never color, matching the rest of the design system.
 *   - Unscheduled pods sit in a "stray" holding area outside the pens.
 *
 * WHERE THE REAL COMPONENT PLUGS IN:
 *   Replace the <PasturePreview> below with e.g. <PastureCanvas nodes={nodes}
 *   pods={pods} />. The data it needs is ALREADY available here via
 *   usePageData(): `data.nodes` and `data.pods`. Group pods by
 *   `pod.spec.node_name` to assign each sheep to its pen; pods with no
 *   node_name are strays. Poll/refresh is handled upstream, so the component
 *   only needs to be a pure function of (nodes, pods).
 *
 * Do NOT implement the animation now — this is the seam, not the feature.
 * ==========================================================================
 */

function PasturePreview() {
  const { data } = usePageData()
  const nodes = data?.nodes ?? []
  const pods = data?.pods ?? []

  // Group pods by node so the eventual visualization has a data shape to
  // build on. This mirrors the node->pod grouping the real component will use.
  const byNode = new Map<string, number>()
  for (const pod of pods) {
    const node = pod.spec?.node_name
    if (!node) continue
    byNode.set(node, (byNode.get(node) ?? 0) + 1)
  }

  if (nodes.length === 0) {
    return <p className="muted">No nodes to graze yet.</p>
  }

  return (
    <div className="pasture__pens">
      {nodes.map((n) => {
        const count = byNode.get(n.metadata.name) ?? 0
        return (
          <div className="pen" key={n.metadata.uid || n.metadata.name}>
            <div className="pen__name">
              <span>{n.metadata.name}</span>
              <span>{count} 🐑</span>
            </div>
            <div className="pen__flock">
              {count === 0 ? (
                <span className="muted" style={{ fontSize: 'var(--fs-xs)' }}>
                  empty pen
                </span>
              ) : (
                Array.from({ length: Math.min(count, 24) }).map((_, i) => (
                  <span className="sheep" key={i} aria-hidden>
                    🐑
                  </span>
                ))
              )}
            </div>
          </div>
        )
      })}
    </div>
  )
}

export function Pasture() {
  return (
    <>
      <PageHeader title="Pasture" />
      <div className="pasture">
        <div className="pasture__hero">
          <div className="pasture__mark" aria-hidden>
            🐑
          </div>
          <h3 className="pasture__title">Pasture view — coming soon</h3>
          <p className="pasture__desc">
            A farm-themed cluster visualization is planned here: nodes become
            pens and pods become sheep grazing inside them. The static preview
            below shows the node → pod grouping the animated view will build on.
          </p>
        </div>

        <div className="card">
          <div className="card__head">
            <span className="card__title">Pens preview (static)</span>
          </div>
          <div className="card__body">
            <PasturePreview />
          </div>
        </div>
      </div>
    </>
  )
}
