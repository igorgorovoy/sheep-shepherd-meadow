// Sprite manifest loader.
//
// Fetches /sprites/manifest.json (see web/public/sprites/manifest.json) and
// resolves logical sprite keys -> absolute URLs the Phaser loader can consume.
// A missing/unknown key resolves to `null`, which the scene turns into a
// placeholder token so the world is always legible even before real art
// exists.

export interface DwarvesManifest {
  roles: string[]
  states: string[]
  path: string // e.g. "dwarves/{role}/{state}.svg"
  anchor: [number, number]
}

export interface SheepManifest {
  states: string[]
  path: string // e.g. "sheep/{state}.svg"
  anchor: [number, number]
}

export interface StructureEntry {
  anchor: [number, number]
  [state: string]: string | [number, number]
}

export interface StructuresManifest {
  station: StructureEntry
  core: StructureEntry
  vault: StructureEntry
  pen: StructureEntry
}

export interface TilesManifest {
  floor: string
  tileW: number
  tileH: number
}

export interface SpriteManifest {
  version: number
  tokenSize: number
  dwarves: DwarvesManifest
  sheep: SheepManifest
  structures: StructuresManifest
  tiles: TilesManifest
  fx: Record<string, string>
}

// A resolved sprite ready for Phaser: a stable texture key + its URL.
export interface ResolvedSprite {
  key: string
  url: string
  anchor: [number, number]
}

const DEFAULT_ANCHOR: [number, number] = [0.5, 0.9]

export class Manifest {
  readonly raw: SpriteManifest
  readonly baseUrl: string

  constructor(raw: SpriteManifest, baseUrl: string) {
    this.raw = raw
    this.baseUrl = baseUrl.replace(/\/+$/, '')
  }

  get tokenSize(): number {
    return this.raw.tokenSize ?? 96
  }

  get tileW(): number {
    return this.raw.tiles?.tileW ?? 256
  }

  get tileH(): number {
    return this.raw.tiles?.tileH ?? 128
  }

  private url(relative: string): string {
    return `${this.baseUrl}/${relative.replace(/^\/+/, '')}`
  }

  // Stable, filesystem-independent texture key for a relative path.
  private keyFor(relative: string): string {
    return relative.replace(/\.[a-z0-9]+$/i, '').replace(/[^a-z0-9]+/gi, '_')
  }

  private resolveTemplate(
    template: string,
    vars: Record<string, string>,
  ): string {
    return template.replace(/\{(\w+)\}/g, (_, k: string) => vars[k] ?? '')
  }

  // Resolve a dwarf sprite. Unknown roles fall back to "node-agent"
  // (the only fully-authored role for the §8 milestone).
  dwarf(role: string, state: string): ResolvedSprite | null {
    const d = this.raw.dwarves
    if (!d) return null
    const effRole = d.roles.includes(role) ? role : 'node-agent'
    const effState = d.states.includes(state) ? state : d.states[0]
    if (!effState) return null
    const rel = this.resolveTemplate(d.path, {
      role: effRole,
      state: effState,
    })
    return {
      key: this.keyFor(rel),
      url: this.url(rel),
      anchor: d.anchor ?? DEFAULT_ANCHOR,
    }
  }

  sheep(state: string): ResolvedSprite | null {
    const s = this.raw.sheep
    if (!s) return null
    const effState = s.states.includes(state) ? state : s.states[0]
    if (!effState) return null
    const rel = this.resolveTemplate(s.path, { state: effState })
    return {
      key: this.keyFor(rel),
      url: this.url(rel),
      anchor: s.anchor ?? DEFAULT_ANCHOR,
    }
  }

  structure(
    kind: keyof StructuresManifest,
    state: string,
  ): ResolvedSprite | null {
    const entry = this.raw.structures?.[kind]
    if (!entry) return null
    const rel = entry[state] ?? entry.idle
    if (typeof rel !== 'string') return null
    return {
      key: this.keyFor(rel),
      url: this.url(rel),
      anchor: (entry.anchor as [number, number]) ?? DEFAULT_ANCHOR,
    }
  }

  floor(): ResolvedSprite | null {
    const rel = this.raw.tiles?.floor
    if (!rel) return null
    return { key: this.keyFor(rel), url: this.url(rel), anchor: [0.5, 0.5] }
  }

  // Enumerate every sprite the scene might need, for a single preload pass.
  // Duplicate keys are collapsed. Callers still tolerate load failures.
  allSprites(): ResolvedSprite[] {
    const out = new Map<string, ResolvedSprite>()
    const add = (r: ResolvedSprite | null) => {
      if (r) out.set(r.key, r)
    }
    const d = this.raw.dwarves
    if (d) {
      for (const role of d.roles) {
        for (const state of d.states) add(this.dwarf(role, state))
      }
    }
    const s = this.raw.sheep
    if (s) for (const state of s.states) add(this.sheep(state))
    const st = this.raw.structures
    if (st) {
      ;(Object.keys(st) as (keyof StructuresManifest)[]).forEach((kind) => {
        const entry = st[kind]
        for (const k of Object.keys(entry)) {
          if (k === 'anchor') continue
          add(this.structure(kind, k))
        }
      })
    }
    add(this.floor())
    return [...out.values()]
  }
}

// Fetch + parse the manifest. Throws on network / parse failure so the caller
// can fall back gracefully.
export async function loadManifest(
  baseUrl = '/sprites',
  signal?: AbortSignal,
): Promise<Manifest> {
  const res = await fetch(`${baseUrl.replace(/\/+$/, '')}/manifest.json`, {
    signal,
    headers: { Accept: 'application/json' },
  })
  if (!res.ok) {
    throw new Error(`manifest fetch failed (${res.status})`)
  }
  const raw = (await res.json()) as SpriteManifest
  return new Manifest(raw, baseUrl)
}
