// Lazily-imported Phaser bootstrap.
//
// This module (and everything it imports, including Phaser and the scene) is
// pulled in via a dynamic import() from LivingHall.tsx, so Phaser code-splits
// out of the initial SPA bundle. It exposes a tiny handle the React component
// drives: feed snapshots via `apply`, tear down via `destroy`.

import Phaser from 'phaser'
import { loadManifest } from './manifest'
import { LivingHallScene } from './scene'
import { mapClusterToWorld } from './mapper'
import type { ClusterSummary } from '../api/types'

export interface HallHandle {
  apply: (summary: ClusterSummary | null) => void
  destroy: () => void
}

export interface BootstrapOptions {
  parent: HTMLElement
  reducedMotion: boolean
  spritesBase?: string
  signal?: AbortSignal
}

// Create the Phaser game inside `parent`. Rejects if WebGL/Canvas or the
// manifest fail — the caller renders a fallback in that case.
export async function bootstrapHall(
  opts: BootstrapOptions,
): Promise<HallHandle> {
  const manifest = await loadManifest(opts.spritesBase ?? '/sprites', opts.signal)

  const scene = new LivingHallScene()

  const game = new Phaser.Game({
    type: Phaser.AUTO, // WebGL with Canvas fallback
    parent: opts.parent,
    backgroundColor: '#130f0a',
    scale: {
      mode: Phaser.Scale.RESIZE,
      autoCenter: Phaser.Scale.CENTER_BOTH,
    },
    render: {
      antialias: true,
      // SVG textures look best without pixelArt snapping.
      pixelArt: false,
    },
    scene,
  })

  // Wait until the scene has actually started so applyWorld isn't dropped.
  await new Promise<void>((resolve, reject) => {
    let settled = false
    const t = window.setTimeout(() => {
      if (!settled) {
        settled = true
        reject(new Error('Phaser scene failed to start'))
      }
    }, 8000)
    game.events.once(Phaser.Core.Events.READY, () => {
      // start (or restart) our scene with the manifest + motion prefs.
      game.scene.start('living-hall', {
        manifest,
        reducedMotion: opts.reducedMotion,
      })
      if (!settled) {
        settled = true
        window.clearTimeout(t)
        resolve()
      }
    })
  })

  const getScene = (): LivingHallScene | undefined =>
    game.scene.getScene('living-hall') as LivingHallScene | undefined

  return {
    apply(summary: ClusterSummary | null) {
      const s = getScene()
      if (!s) return
      s.applyWorld(mapClusterToWorld(summary))
    },
    destroy() {
      game.destroy(true)
    },
  }
}
