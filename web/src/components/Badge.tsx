import type { ReactNode } from 'react'

export type BadgeVariant = 'solid' | 'outline' | 'dashed' | 'muted'

// Monochrome status badge. The visual difference between states comes from
// fill and border style (solid / outline / dashed / muted), never color.
export function Badge({
  variant = 'outline',
  children,
}: {
  variant?: BadgeVariant
  children: ReactNode
}) {
  return <span className={`badge badge--${variant}`}>{children}</span>
}
