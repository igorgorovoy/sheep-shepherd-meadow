// The Living Hall Phaser scene.
//
// Renders an isometric (2:1) steampunk hall: a grid of stations across the
// floor, each with a node-agent dwarf and a flock of pod-sheep; a central
// steam-core whose tube fills with cluster load; a runic vault (deployments);
// and a stray pen for unscheduled pods.
//
// Data flows in via `applyWorld(world)` — the scene DIFFS the incoming
// WorldModel against the last one, spawning / moving / despawning actors
// rather than rebuilding the world each tick. Sprites are loaded from the
// manifest; a missing sprite key becomes a drawn placeholder token so the
// scene is always legible.

import Phaser from 'phaser'
import type { Manifest, ResolvedSprite } from './manifest'
import type {
  CoreModel,
  DwarfModel,
  SheepModel,
  StationModel,
  WorldModel,
} from './mapper'

// State colors (match the palette in the task brief / design system).
const COLOR = {
  running: 0x7de06a,
  pending: 0xebb24c,
  failed: 0xec4b33,
  succeeded: 0xb9b29c,
  brass: 0xb8863a,
  brassHi: 0xe8c67c,
  teal: 0x37c7e9,
  tealGlow: 0x9cf1ff,
  rune: 0x9a6be0,
  stone: 0x211a12,
  parchment: 0xe7d7a7,
} as const

const SHEEP_STATE_COLOR: Record<SheepModel['state'], number> = {
  running: COLOR.running,
  pending: COLOR.pending,
  failed: COLOR.failed,
  succeeded: COLOR.succeeded,
}

export interface SceneInitData {
  manifest: Manifest
  reducedMotion: boolean
}

// ---- iso helpers ----------------------------------------------------------

interface IsoPoint {
  x: number
  y: number
}

// An actor wraps a sprite-or-placeholder container plus bookkeeping for diffing.
interface Actor {
  container: Phaser.GameObjects.Container
  // tween handle for idle motion so we can stop it on reduced-motion / despawn
  idleTween?: Phaser.Tweens.Tween
  // last-known logical state, to avoid redundant texture swaps
  stateKey: string
  homeX: number
  homeY: number
}

export class LivingHallScene extends Phaser.Scene {
  private manifest!: Manifest
  private reducedMotion = false

  private tileW = 256
  private tileH = 128

  // Diffed actor pools keyed by logical id.
  private stationActors = new Map<string, Actor>()
  private dwarfActors = new Map<string, Actor>()
  private sheepActors = new Map<string, Actor>()

  // Layout: node index -> iso grid cell.
  private nodeOrder: string[] = []
  private cols = 3

  // World-space anchors.
  private origin: IsoPoint = { x: 0, y: 0 }
  private coreContainer?: Phaser.GameObjects.Container
  private coreTube?: Phaser.GameObjects.Rectangle
  private vaultContainer?: Phaser.GameObjects.Container
  private penContainer?: Phaser.GameObjects.Container

  // fx we've already fired (id set) so warnings don't re-trigger endlessly.
  private firedFx = new Set<string>()

  // The most recent world we were handed before preload finished.
  private pendingWorld: WorldModel | null = null
  private ready = false

  constructor() {
    super('living-hall')
  }

  init(data: SceneInitData) {
    this.manifest = data.manifest
    this.reducedMotion = data.reducedMotion
    this.tileW = this.manifest.tileW
    this.tileH = this.manifest.tileH
  }

  preload() {
    // Load every sprite the manifest knows about. Phaser can rasterize SVGs
    // via load.svg. Failures are tolerated (we check textures.exists later).
    for (const sprite of this.manifest.allSprites()) {
      if (this.textures.exists(sprite.key)) continue
      this.load.svg(sprite.key, sprite.url)
    }
    // If a single asset 404s, don't abort the whole scene.
    this.load.on('loaderror', (file: Phaser.Loader.File) => {
      // eslint-disable-next-line no-console
      console.warn('[living-hall] sprite failed to load:', file.key)
    })
  }

  create() {
    const cam = this.cameras.main
    cam.setBackgroundColor('#130f0a')

    // Origin roughly centers the hall; recomputed on resize.
    this.recomputeOrigin()

    this.scale.on('resize', () => {
      this.recomputeOrigin()
      this.relayout()
    })

    this.ready = true
    if (this.pendingWorld) {
      const w = this.pendingWorld
      this.pendingWorld = null
      this.applyWorld(w)
    }
  }

  private recomputeOrigin() {
    const { width, height } = this.scale
    this.origin = { x: width / 2, y: height * 0.32 }
  }

  // Grid cell (col,row) -> screen coords for a 2:1 iso projection.
  private cellToScreen(col: number, row: number): IsoPoint {
    const halfW = this.tileW / 2
    const halfH = this.tileH / 2
    return {
      x: this.origin.x + (col - row) * halfW,
      y: this.origin.y + (col + row) * halfH,
    }
  }

  private nodeCell(index: number): { col: number; row: number } {
    return { col: index % this.cols, row: Math.floor(index / this.cols) }
  }

  // ---- public API: called by React on each snapshot ----------------------

  applyWorld(world: WorldModel) {
    if (!this.ready) {
      this.pendingWorld = world
      return
    }
    this.nodeOrder = world.stations.map((s) => s.nodeName)
    this.cols = Math.max(1, Math.ceil(Math.sqrt(world.stations.length || 1)))

    this.diffStations(world.stations)
    this.diffDwarves(world.dwarves)
    this.diffSheep(world.sheep)
    this.updateCore(world.core)
    this.updateVault(world.vaultActive)
    this.updatePen(world.sheep.filter((s) => s.nodeName == null).length)
    this.fireFx(world)
  }

  // Recompute all actor home positions after a resize.
  private relayout() {
    this.nodeOrder.forEach((nodeName, i) => {
      const cell = this.nodeCell(i)
      const p = this.cellToScreen(cell.col, cell.row)
      const st = this.stationActors.get(`station:${nodeName}`)
      if (st) this.moveActor(st, p.x, p.y, true)
      const dw = this.dwarfActors.get(`dwarf:${nodeName}`)
      if (dw) this.moveActor(dw, p.x - 46, p.y - 10, true)
    })
    this.placeStructures()
    // Re-home sheep to their stations.
    for (const [, actor] of this.sheepActors) {
      // sheep re-home on next diff; nothing precise needed here.
      void actor
    }
  }

  // ---- sprite / placeholder factory --------------------------------------

  // Build a display object for a resolved sprite, or a placeholder token if the
  // texture is unavailable. Returns a Container so callers get a uniform anchor.
  private makeActorContainer(
    resolved: ResolvedSprite | null,
    fallbackLabel: string,
    fallbackColor: number,
    fallbackW = 48,
    fallbackH = 60,
  ): Phaser.GameObjects.Container {
    const container = this.add.container(0, 0)
    if (resolved && this.textures.exists(resolved.key)) {
      const img = this.add.image(0, 0, resolved.key)
      img.setOrigin(resolved.anchor[0], resolved.anchor[1])
      container.add(img)
    } else {
      container.add(
        this.makePlaceholder(fallbackLabel, fallbackColor, fallbackW, fallbackH),
      )
    }
    return container
  }

  // A rounded-rect token in the state color with a short label. Used whenever a
  // sprite key is missing, so the world stays readable pre-art.
  private makePlaceholder(
    label: string,
    color: number,
    w: number,
    h: number,
  ): Phaser.GameObjects.Container {
    const c = this.add.container(0, 0)
    const g = this.add.graphics()
    g.fillStyle(color, 0.85)
    g.lineStyle(2, 0x130f0a, 1)
    g.fillRoundedRect(-w / 2, -h, w, h, 8)
    g.strokeRoundedRect(-w / 2, -h, w, h, 8)
    c.add(g)
    const short = label.length > 8 ? label.slice(0, 7) + '…' : label
    const txt = this.add.text(0, -h / 2, short, {
      fontFamily: 'monospace',
      fontSize: '11px',
      color: '#130f0a',
      align: 'center',
    })
    txt.setOrigin(0.5, 0.5)
    c.add(txt)
    return c
  }

  private moveActor(actor: Actor, x: number, y: number, instant = false) {
    actor.homeX = x
    actor.homeY = y
    if (instant || this.reducedMotion) {
      actor.container.setPosition(x, y)
    } else {
      this.tweens.add({
        targets: actor.container,
        x,
        y,
        duration: 420,
        ease: 'Sine.easeInOut',
      })
    }
  }

  private destroyActor(actor: Actor) {
    actor.idleTween?.stop()
    actor.container.destroy()
  }

  // ---- stations -----------------------------------------------------------

  private diffStations(stations: StationModel[]) {
    const seen = new Set<string>()
    stations.forEach((st, i) => {
      seen.add(st.id)
      const cell = this.nodeCell(i)
      const p = this.cellToScreen(cell.col, cell.row)
      let actor = this.stationActors.get(st.id)
      const stateKey = `${st.active}|${st.dim}|${st.lampRed}`

      if (!actor) {
        // floor tile beneath the station (drawn first, sits lower in z).
        const floor = this.manifest.floor()
        const container = this.add.container(0, 0)
        if (floor && this.textures.exists(floor.key)) {
          const tile = this.add.image(0, this.tileH * 0.25, floor.key)
          tile.setOrigin(0.5, 0.5)
          container.add(tile)
        }
        const resolved = this.manifest.structure(
          'station',
          st.active ? 'active' : 'idle',
        )
        const stationVisual = this.makeActorContainer(
          resolved,
          st.label,
          COLOR.brass,
          120,
          80,
        )
        container.add(stationVisual)

        // lamp
        const lamp = this.add.circle(46, -70, 5, COLOR.running)
        lamp.setData('lamp', true)
        container.add(lamp)

        // name plate
        const plate = this.add.text(0, 8, st.label, {
          fontFamily: 'monospace',
          fontSize: '11px',
          color: '#e7d7a7',
        })
        plate.setOrigin(0.5, 0)
        container.add(plate)

        actor = { container, stateKey: '', homeX: p.x, homeY: p.y }
        container.setPosition(p.x, p.y)
        container.setDepth(p.y)
        this.stationActors.set(st.id, actor)
      } else {
        this.moveActor(actor, p.x, p.y)
        actor.container.setDepth(p.y)
      }

      if (actor.stateKey !== stateKey) {
        this.applyStationState(actor, st)
        actor.stateKey = stateKey
      }
    })

    for (const [id, actor] of this.stationActors) {
      if (!seen.has(id)) {
        this.destroyActor(actor)
        this.stationActors.delete(id)
      }
    }
  }

  private applyStationState(actor: Actor, st: StationModel) {
    actor.container.setAlpha(st.dim ? 0.45 : 1)
    const lamp = actor.container.list.find(
      (o) => (o as Phaser.GameObjects.GameObject).getData?.('lamp'),
    ) as Phaser.GameObjects.Arc | undefined
    if (lamp) {
      if (st.lampRed) {
        lamp.setFillStyle(COLOR.failed)
        if (!this.reducedMotion) {
          this.tweens.add({
            targets: lamp,
            alpha: { from: 1, to: 0.2 },
            duration: 600,
            yoyo: true,
            repeat: -1,
          })
        }
      } else {
        this.tweens.killTweensOf(lamp)
        lamp.setAlpha(1)
        lamp.setFillStyle(st.active ? COLOR.running : COLOR.brass)
      }
    }
  }

  // ---- dwarves ------------------------------------------------------------

  private diffDwarves(dwarves: DwarfModel[]) {
    const seen = new Set<string>()
    dwarves.forEach((dw, i) => {
      seen.add(dw.id)
      const cell = this.nodeCell(i)
      const p = this.cellToScreen(cell.col, cell.row)
      const dx = p.x - 46
      const dy = p.y - 10
      let actor = this.dwarfActors.get(dw.id)

      if (!actor || actor.stateKey.split('|')[0] !== dw.state) {
        // (re)build container to swap texture cleanly
        if (actor) this.destroyActor(actor)
        const resolved = this.manifest.dwarf(dw.role, dw.state)
        const color =
          dw.state === 'offline' ? COLOR.stone : COLOR.brass
        const container = this.makeActorContainer(
          resolved,
          dw.label,
          color,
          40,
          52,
        )
        container.setPosition(dx, dy)
        container.setDepth(p.y + 1)
        actor = { container, stateKey: `${dw.state}`, homeX: dx, homeY: dy }
        this.dwarfActors.set(dw.id, actor)
        this.startDwarfIdle(actor, dw.state)
      } else {
        this.moveActor(actor, dx, dy)
        actor.container.setDepth(p.y + 1)
      }
    })

    for (const [id, actor] of this.dwarfActors) {
      if (!seen.has(id)) {
        this.destroyActor(actor)
        this.dwarfActors.delete(id)
      }
    }
  }

  private startDwarfIdle(actor: Actor, state: DwarfModel['state']) {
    actor.idleTween?.stop()
    if (this.reducedMotion || state === 'offline') return
    // work = quicker bob; idle = slow breathe
    const dur = state === 'work' ? 380 : 1400
    const amp = state === 'work' ? 3 : 1.5
    actor.idleTween = this.tweens.add({
      targets: actor.container,
      y: actor.homeY - amp,
      duration: dur,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.easeInOut',
    })
  }

  // ---- sheep --------------------------------------------------------------

  private diffSheep(sheep: SheepModel[]) {
    const seen = new Set<string>()

    // Assign each sheep a slot around its station (or the pen).
    const slotCounter = new Map<string, number>()
    const slotFor = (key: string): number => {
      const n = slotCounter.get(key) ?? 0
      slotCounter.set(key, n + 1)
      return n
    }

    for (const sh of sheep) {
      seen.add(sh.id)
      const key = sh.nodeName ?? '__stray__'
      const slot = slotFor(key)
      const target = this.sheepHome(sh.nodeName, slot)

      let actor = this.sheepActors.get(sh.id)
      if (!actor || actor.stateKey !== sh.state) {
        if (actor) this.destroyActor(actor)
        const resolved = this.manifest.sheep(sh.state)
        const container = this.makeActorContainer(
          resolved,
          sh.label,
          SHEEP_STATE_COLOR[sh.state],
          40,
          34,
        )
        container.setPosition(target.x, target.y)
        container.setDepth(target.y + 2)
        actor = {
          container,
          stateKey: sh.state,
          homeX: target.x,
          homeY: target.y,
        }
        this.sheepActors.set(sh.id, actor)
        this.startSheepBob(actor, sh.state)
      } else {
        this.moveActor(actor, target.x, target.y)
        actor.container.setDepth(target.y + 2)
      }
    }

    for (const [id, actor] of this.sheepActors) {
      if (!seen.has(id)) {
        this.destroyActor(actor)
        this.sheepActors.delete(id)
      }
    }
  }

  private sheepHome(nodeName: string | null, slot: number): IsoPoint {
    // arrange sheep in a small cluster; 4 per row.
    const perRow = 4
    const gx = (slot % perRow) * 22 - 33
    const gy = Math.floor(slot / perRow) * 16

    if (nodeName == null) {
      // stray pen — fixed area lower-left of the hall.
      const base = this.penHome()
      return { x: base.x + gx, y: base.y + gy - 24 }
    }
    const idx = this.nodeOrder.indexOf(nodeName)
    const cell = this.nodeCell(idx < 0 ? 0 : idx)
    const p = this.cellToScreen(cell.col, cell.row)
    return { x: p.x + gx, y: p.y + 30 + gy }
  }

  private startSheepBob(actor: Actor, state: SheepModel['state']) {
    actor.idleTween?.stop()
    if (this.reducedMotion) return
    if (state === 'succeeded') return // resting, no bob
    const speed = state === 'running' ? 520 : state === 'pending' ? 900 : 700
    const amp = state === 'failed' ? 0.8 : 2
    actor.idleTween = this.tweens.add({
      targets: actor.container,
      y: actor.homeY - amp,
      duration: speed,
      yoyo: true,
      repeat: -1,
      ease: 'Sine.easeInOut',
    })
  }

  // ---- core / vault / pen -------------------------------------------------

  private hallEdges(): { left: number; bottom: number; centerX: number } {
    const { width, height } = this.scale
    return { left: width * 0.14, bottom: height * 0.82, centerX: width / 2 }
  }

  private penHome(): IsoPoint {
    const { left, bottom } = this.hallEdges()
    return { x: left, y: bottom }
  }

  private placeStructures() {
    const { centerX } = this.hallEdges()
    const { height } = this.scale
    if (this.coreContainer) {
      this.coreContainer.setPosition(centerX, height * 0.2)
      this.coreContainer.setDepth(height * 0.2)
    }
    if (this.vaultContainer) {
      this.vaultContainer.setPosition(this.scale.width * 0.84, height * 0.24)
      this.vaultContainer.setDepth(height * 0.24)
    }
    if (this.penContainer) {
      const pen = this.penHome()
      this.penContainer.setPosition(pen.x, pen.y + 24)
      this.penContainer.setDepth(pen.y)
    }
  }

  private updateCore(core: CoreModel) {
    if (!this.coreContainer) {
      const resolved = this.manifest.structure('core', 'idle')
      this.coreContainer = this.makeActorContainer(
        resolved,
        'core',
        COLOR.teal,
        60,
        120,
      )
      // overlay tube fill rectangle (grows from bottom).
      this.coreTube = this.add.rectangle(0, -14, 14, 0, COLOR.tealGlow, 0.75)
      this.coreTube.setOrigin(0.5, 1)
      this.coreContainer.add(this.coreTube)
      this.placeStructures()
    }
    if (this.coreTube) {
      const maxH = 96
      const h = Math.round(maxH * Math.max(0, Math.min(1, core.fill)))
      const targetColor =
        core.intensity === 'high'
          ? COLOR.tealGlow
          : core.intensity === 'idle'
            ? COLOR.brass
            : COLOR.teal
      this.coreTube.setFillStyle(targetColor, 0.8)
      if (this.reducedMotion) {
        this.coreTube.height = h
      } else {
        this.tweens.add({
          targets: this.coreTube,
          height: h,
          duration: 600,
          ease: 'Sine.easeOut',
        })
      }
    }
  }

  private updateVault(active: boolean) {
    if (!this.vaultContainer) {
      const resolved = this.manifest.structure('vault', 'idle')
      this.vaultContainer = this.makeActorContainer(
        resolved,
        'vault',
        COLOR.rune,
        60,
        90,
      )
      this.placeStructures()
    }
    this.vaultContainer.setAlpha(active ? 1 : 0.55)
  }

  private updatePen(strayCount: number) {
    if (!this.penContainer) {
      const resolved = this.manifest.structure('pen', 'idle')
      this.penContainer = this.makeActorContainer(
        resolved,
        'stray',
        COLOR.brass,
        140,
        70,
      )
      this.placeStructures()
    }
    // dim the pen when empty.
    this.penContainer.setAlpha(strayCount > 0 ? 1 : 0.6)
  }

  // ---- fx -----------------------------------------------------------------

  private fireFx(world: WorldModel) {
    for (const fx of world.fx) {
      if (this.firedFx.has(fx.id)) continue
      this.firedFx.add(fx.id)
      // keep the fired set bounded
      if (this.firedFx.size > 256) {
        this.firedFx = new Set([...this.firedFx].slice(-128))
      }
      const target = fx.stationId
        ? this.stationActors.get(fx.stationId)
        : undefined
      const x = target?.homeX ?? this.scale.width / 2
      const y = (target?.homeY ?? this.scale.height / 2) - 90
      this.spawnAlarm(x, y)
    }
  }

  private spawnAlarm(x: number, y: number) {
    const ring = this.add.circle(x, y, 6, COLOR.failed, 0.9)
    ring.setDepth(1e6)
    if (this.reducedMotion) {
      this.time.delayedCall(900, () => ring.destroy())
      return
    }
    this.tweens.add({
      targets: ring,
      scale: { from: 0.4, to: 3 },
      alpha: { from: 0.9, to: 0 },
      duration: 900,
      ease: 'Cubic.easeOut',
      onComplete: () => ring.destroy(),
    })
  }
}
