import { Badge, type BadgeVariant } from './Badge'
import type { NodeCondition, PodPhase } from '../api/types'

// Maps domain statuses to monochrome badge variants.
// Running/Ready -> solid ink. Pending -> outline. Failed/NotReady -> dashed.
// Succeeded -> muted.

const POD_PHASE: Record<PodPhase, BadgeVariant> = {
  Running: 'solid',
  Pending: 'outline',
  Succeeded: 'muted',
  Failed: 'dashed',
}

export function PodPhaseBadge({ phase }: { phase: PodPhase }) {
  return <Badge variant={POD_PHASE[phase] ?? 'outline'}>{phase}</Badge>
}

export function NodeConditionBadge({ condition }: { condition: NodeCondition }) {
  return (
    <Badge variant={condition === 'Ready' ? 'solid' : 'dashed'}>
      {condition}
    </Badge>
  )
}
