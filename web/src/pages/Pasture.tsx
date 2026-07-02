import { PageHeader } from '../components/PageHeader'
import { LivingHall } from '../components/LivingHall'
import { usePageData } from './context'

/*
 * ==========================================================================
 * The "Living Hall" — gamified cluster visualization (RFC-0002 / ADR-0002)
 * ==========================================================================
 *
 * An isometric steampunk hall rendered with Phaser 3:
 *   - Each NODE is a forge STATION with a node-agent DWARF.
 *   - Each POD is a SHEEP at its node's station (grouped by spec.node_name);
 *     unscheduled pods gather in the STRAY PEN. Pod phase drives sheep state.
 *   - Cluster load (running/total pods) drives the central STEAM-CORE tube.
 *   - Deployments light the runic VAULT; Warning events flash alarms.
 *
 * Data comes straight from the existing polled cluster hook via usePageData()
 * — no new fetch path. The <LivingHall> component lazy-loads Phaser and pushes
 * each new snapshot into the running scene (it diffs internally, never
 * re-mounts). If WebGL or the sprite manifest fail, it degrades to a notice
 * and the rest of the dashboard keeps working.
 * ==========================================================================
 */

export function Pasture() {
  const { data, error } = usePageData()

  return (
    <>
      <PageHeader
        title="Pasture"
        description="The Living Hall — nodes are forge stations, pods are sheep, and the steam-core burns with cluster load."
      />
      {error && (
        <p className="muted" role="status">
          Showing the last known cluster state. Live updates are paused:{' '}
          {error}
        </p>
      )}
      <div className="card">
        <div className="card__body card__body--flush">
          <LivingHall summary={data} />
        </div>
      </div>
    </>
  )
}
