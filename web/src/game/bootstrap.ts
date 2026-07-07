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
  apply: (summary: ClusterSummary | null, meadowRepoCount?: number) => void
  destroy: () => void
}

export interface BootstrapOptions {
  parent: HTMLElement
  reducedMotion: boolean
  spritesBase?: string
  signal?: AbortSignal
  onNavigate?: (path: string) => void
}

// Create the Phaser game inside `parent`. Rejects if WebGL/Canvas or the
// manifest fail — the caller renders a fallback in that case.
export async function bootstrapHall(
  opts: BootstrapOptions,
): Promise<HallHandle> {
  const manifest = await loadManifest(opts.spritesBase ?? '/sprites', opts.signal)

  const scene = new LivingHallScene()

  // NOTE: do NOT pass the scene in `config.scene` — that auto-starts it during
  // boot with no init data, so the scene's init() dereferences an undefined
  // manifest and throws, and the game never reaches READY. Instead we boot with
  // no scenes, then add + start it once with the manifest after READY.
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
  })

  // Wait until the game has booted, then add + start our scene with data so
  // applyWorld isn't dropped. Guard the race where boot already completed.
  await new Promise<void>((resolve, reject) => {
    let settled = false
    const t = window.setTimeout(() => {
      if (!settled) {
        settled = true
        reject(new Error('Phaser scene failed to start'))
      }
    }, 8000)
    const startScene = () => {
      if (settled) return
      settled = true
      window.clearTimeout(t)
      // add(key, scene, autoStart=true, data) → runs init/preload/create with data.
      game.scene.add('living-hall', scene, true, {
        manifest,
        reducedMotion: opts.reducedMotion,
        onNavigate: opts.onNavigate,
      })
      resolve()
    }
    if (game.isBooted) startScene()
    else game.events.once(Phaser.Core.Events.READY, startScene)
  })

  const getScene = (): LivingHallScene | undefined =>
    game.scene.getScene('living-hall') as LivingHallScene | undefined

  return {
    apply(summary: ClusterSummary | null, meadowRepoCount = 0) {
      const s = getScene()
      if (!s) return
      s.applyWorld(mapClusterToWorld(summary, meadowRepoCount))
    },
    destroy() {
      game.destroy(true)
    },
  }
}
