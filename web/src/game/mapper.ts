// Pure domain -> scene mapping.
//
// Takes a ClusterSummary snapshot and produces a plain, serializable
// description of the world (stations, dwarves, sheep, core, vault, pen, fx).
// No Phaser imports here — this stays unit-testable and framework-free. The
// scene consumes a WorldModel and diffs it against the previous one.

import type {
  ClusterSummary,
  Event as ClusterEvent,
  Node,
  Pod,
} from '../api/types'

export type DwarfState = 'idle' | 'work' | 'offline'
export type SheepState = 'running' | 'pending' | 'failed' | 'succeeded'
export type CoreIntensity = 'idle' | 'low' | 'mid' | 'high'

// Round-robin dwarf roles. node-agent is the only fully-authored role for the
// §8 milestone; the rest resolve to node-agent art via the manifest fallback,
// but we still assign distinct roles so labels/behaviour can differentiate.
export const DWARF_ROLES = [
  'node-agent',
  'runtime-smith',
  'scheduler-warden',
  'reconciler',
  'edge-tinker',
] as const

export interface DwarfModel {
  id: string // stable: `dwarf:<nodeName>`
  nodeName: string
  role: string
  state: DwarfState
  label: string
}

export interface StationModel {
  id: string // stable: `station:<nodeName>`
  nodeName: string
  active: boolean // furnace glow when node Ready and has running pods
  dim: boolean // NotReady -> dimmed
  lampRed: boolean // NotReady -> red blinking lamp
  label: string
}

export interface SheepModel {
  id: string // stable: pod uid (falls back to name)
  podName: string
  nodeName: string | null // null => stray pen
  state: SheepState
  label: string
}

export interface CoreModel {
  intensity: CoreIntensity
  fill: number // 0..1 tube fill (running / total pods)
  running: number
  total: number
}

export interface FxModel {
  id: string // stable per (object,reason) so we don't re-fire endlessly
  kind: 'alarm' | 'smoke'
  stationId: string | null // target station, or null for hall-level
  message: string
}

export interface WorldModel {
  stations: StationModel[]
  dwarves: DwarfModel[]
  sheep: SheepModel[]
  core: CoreModel
  vaultActive: boolean // any deployments present -> rune vault glows
  fx: FxModel[]
}

function podPhaseToSheep(phase: string | undefined): SheepState {
  switch (phase) {
    case 'Running':
      return 'running'
    case 'Failed':
      return 'failed'
    case 'Succeeded':
      return 'succeeded'
    case 'Pending':
    default:
      return 'pending'
  }
}

function coreIntensity(fill: number, total: number): CoreIntensity {
  if (total === 0) return 'idle'
  if (fill >= 0.85) return 'high'
  if (fill >= 0.5) return 'mid'
  if (fill > 0) return 'low'
  return 'idle'
}

// Short, stable id for a sheep. Prefer uid; fall back to namespaced name.
function sheepId(pod: Pod): string {
  const uid = pod.metadata?.uid
  if (uid) return uid
  const ns = pod.metadata?.namespace ?? 'default'
  return `${ns}/${pod.metadata?.name ?? 'pod'}`
}

function stationId(node: Node): string {
  return `station:${node.metadata.name}`
}

// Extract the latest N Warning events and attach them to the station whose
// object name they reference, when resolvable. Basic best-effort mapping.
function mapWarningFx(
  events: ClusterEvent[],
  nodeNames: Set<string>,
  podToNode: Map<string, string>,
): FxModel[] {
  const warnings = events
    .filter((e) => e.type === 'Warning')
    .slice(-6) // most recent tail
  const out: FxModel[] = []
  for (const e of warnings) {
    // e.object is typically "kind/name" or "name"; try to resolve to a node.
    const rawName = (e.object ?? '').split('/').pop() ?? ''
    let stationTarget: string | null = null
    if (nodeNames.has(rawName)) {
      stationTarget = `station:${rawName}`
    } else if (podToNode.has(rawName)) {
      stationTarget = `station:${podToNode.get(rawName)!}`
    }
    out.push({
      id: `fx:${e.object}:${e.reason}:${e.timestamp}`,
      kind: 'alarm',
      stationId: stationTarget,
      message: `${e.reason}: ${e.message}`,
    })
  }
  return out
}

export function mapClusterToWorld(summary: ClusterSummary | null): WorldModel {
  const nodes = summary?.nodes ?? []
  const pods = summary?.pods ?? []
  const events = summary?.events ?? []
  const deployments = summary?.deployments ?? []

  // Pods grouped by node, plus running counts per node.
  const runningByNode = new Map<string, number>()
  const totalByNode = new Map<string, number>()
  const podToNode = new Map<string, string>()
  for (const pod of pods) {
    const nn = pod.spec?.node_name
    if (!nn) continue
    totalByNode.set(nn, (totalByNode.get(nn) ?? 0) + 1)
    if (pod.status?.phase === 'Running') {
      runningByNode.set(nn, (runningByNode.get(nn) ?? 0) + 1)
    }
    const name = pod.metadata?.name
    if (name) podToNode.set(name, nn)
  }

  const nodeNames = new Set(nodes.map((n) => n.metadata.name))

  const stations: StationModel[] = nodes.map((node) => {
    const notReady = node.status?.condition === 'NotReady'
    const running = runningByNode.get(node.metadata.name) ?? 0
    return {
      id: stationId(node),
      nodeName: node.metadata.name,
      active: !notReady && running > 0,
      dim: notReady,
      lampRed: notReady,
      label: node.metadata.name,
    }
  })

  const dwarves: DwarfModel[] = nodes.map((node, i) => {
    const notReady = node.status?.condition === 'NotReady'
    const running = runningByNode.get(node.metadata.name) ?? 0
    // Role: node label `sheep.sh/role` if present, else round-robin.
    const labeledRole = node.metadata.labels?.['sheep.sh/role']
    const role =
      labeledRole && DWARF_ROLES.includes(labeledRole as (typeof DWARF_ROLES)[number])
        ? labeledRole
        : DWARF_ROLES[i % DWARF_ROLES.length]
    const state: DwarfState = notReady
      ? 'offline'
      : running > 0
        ? 'work'
        : 'idle'
    return {
      id: `dwarf:${node.metadata.name}`,
      nodeName: node.metadata.name,
      role,
      state,
      label: role,
    }
  })

  const sheep: SheepModel[] = pods.map((pod) => {
    const nn = pod.spec?.node_name ?? null
    return {
      id: sheepId(pod),
      podName: pod.metadata?.name ?? 'pod',
      nodeName: nn,
      state: podPhaseToSheep(pod.status?.phase),
      label: pod.metadata?.name ?? 'pod',
    }
  })

  const total = pods.length
  const running = pods.filter((p) => p.status?.phase === 'Running').length
  const fill = total > 0 ? running / total : 0
  const core: CoreModel = {
    intensity: coreIntensity(fill, total),
    fill,
    running,
    total,
  }

  const fx = mapWarningFx(events, nodeNames, podToNode)

  return {
    stations,
    dwarves,
    sheep,
    core,
    vaultActive: deployments.length > 0,
    fx,
  }
}
