# DESIGN-0002 — Forge Sprite Kit (AI generation + integration)

- **Status:** Draft
- **Author(s):** i.gorovoy
- **Created At:** 2026-07-02
- **For:** RFC-0002 (Living Hall) · ADR-0002 (Phaser 3)

This is the production kit for the Living Hall art. You generate the sprites (Gemini / DALL·E / ChatGPT image), drop the PNGs into `web/public/sprites/` at the paths below, and the scene loads them via `manifest.json`. Until a real sprite exists, the engine renders a colored placeholder token, so the game runs at every stage.

---

## 1. How to use this kit

1. Copy the **Style Anchor** (§2) into *every* prompt, then append the specific **Sprite Prompt** (§4).
2. Generate at a **1:1 square** aspect, highest quality, **transparent background**.
3. To keep the look consistent across sprites, use one already-approved sprite (or a reference image from `docs/blog/posts/assets/`) as a **style/seed reference** on subsequent generations.
4. Trim/center, export **PNG with alpha**, name per §5, place in the directory in §5, and add/confirm the entry in `manifest.json` (§6).
5. Run the checklist in §7.

---

## 2. Style Anchor (prepend to every prompt)

> 2D game sprite, **Warcraft-III / Hearthstone hand-painted cartoon** style, **steampunk dwarven engineering** theme. Thick clean dark outlines, painterly shading, warm **brass/copper/gold metal** and dark stone, glowing **teal aether** energy and **arcane-purple runes** as accents. Dramatic top-left rim light, rich saturation, high detail. **Single centered subject, 3/4 top-down game-token view, fully transparent background (PNG alpha), no ground shadow baked in, no text, no logo, no border, no frame.** Consistent lighting and scale across the set.

Global generation rules:
- **Transparent background**, subject centered, generous even margin.
- **No text / UI / captions** baked into the art (labels are drawn by the app).
- Consistent **light from top-left**; consistent line weight; consistent color temperature.
- One subject per file (except explicit tile/fx sheets).

---

## 3. Palette lock (name these hexes in prompts when it helps)

| Role | Hex |
|---|---|
| Brass | `#B8863A` · highlight `#E8C67C` · shadow `#6E4E22` |
| Copper | `#8C5A2B` |
| Stone (dark) | `#211A12` / `#130F0A` |
| Parchment | `#E7D7A7` |
| Aether (teal) | `#37C7E9` · glow `#9CF1FF` |
| Rune (purple) | `#9A6BE0` · glow `#C9A6FF` |
| **Running** (green) | `#7DE06A` |
| **Pending** (amber) | `#EBB24C` |
| **Failed / alarm** (red) | `#EC4B33` |

Signal colors (green/amber/red) appear **only** on actor state (collars, lamps, alarms), never as chrome.

---

## 4. Sprite prompt catalog

Generate one PNG per row unless noted. Recommended canvas sizes in §5.

### 4.1 Dwarf engineers (heroes) — one per node/role, 3 state poses each

Base body prompt (reuse, swap the role kit + beard):
> …a stout **dwarf engineer**, thick braided beard, riveted **brass helmet with glowing teal goggles**, battered steel-and-brass armor with a leather tool-belt, sturdy boots. Heroic, characterful, friendly-gruff. 3/4 top-down token view.

Roles (change tool + beard color for identity):
| Role (node persona) | Tool / kit | Beard |
|---|---|---|
| Runtime Smith | forge hammer + glowing ingot | ginger `#c96a2c` |
| Scheduler Warden | multi-lever control rod, brass clipboard | grey `#9a9a9a` |
| Reconciler | oversized brass gear + oil can | gold `#a8863f` |
| Edge Tinker | wrench + backpack of copper pipes | dark-brown `#7a4a22` |
| Node Agent (generic) | lantern + satchel | auburn |

State poses (generate all three per role — same dwarf, same seed):
- `idle` — standing, tool at side, calm, subtle idle.
- `work` — mid-swing / operating the tool, focused (the "healthy/active" pose).
- `offline` — slumped, goggles dark, one knee down, faint grey — for **node NotReady**.

### 4.2 Sheep (pods) — phase states, 1 per state + 1 walk

Base:
> …a plump cartoon **sheep** with thick cream wool `#EDE6D2`, small dark face, wearing a slim **brass collar** with a glowing gem. Steampunk farm animal, cute but sturdy. 3/4 token view.

| File | Prompt add | Reads as |
|---|---|---|
| `running` | lively, head up, **green `#7DE06A` glowing collar gem**, faint aether wisp | pod Running |
| `pending` | sitting, half-asleep, **amber `#EBB24C` collar**, slightly dim wool | pod Pending |
| `failed` | collapsed/tipped, **red `#EC4B33` collar**, small alarm spark, wisp of smoke | pod Failed |
| `succeeded` | resting contentedly, **faded/greyer wool**, dim collar | pod Succeeded |
| `walk` | mid-stride walking pose, green collar | transitions (spawn/move) |

### 4.3 Structures

| File | Prompt add (after anchor) | Represents |
|---|---|---|
| `stations/forge-idle` | a **dwarven forge workbench / brass control console** with pipes, gauges, an anvil, cold furnace | a Node (empty/idle) |
| `stations/forge-active` | same console, **furnace glowing, gauges lit, teal steam** rising | a Node (active) |
| `core/core-idle` | a **great central steam-engine core**, tall brass boiler with a vertical **teal aether tube**, gears, pressure valves | Shepherd control plane |
| `core/core-load-high` | same core, **tube brighter/fuller, more steam, valves venting** | high cluster load |
| `vault/vault-idle` | a **stone archway sealed with purple glowing runes**, faint nebula within | Meadow registry |
| `vault/vault-active` | same arch, **portal open, arcane-purple light pouring out, floating crystals** | registry active / pull-push |
| `pen/stray-pen` | a small **fenced holding pen with an open gate**, empty, dim lantern | unscheduled pods area |

### 4.4 Environment tiles (isometric, seamless)

> …**isometric floor tile**, dark dwarven **stone flagstones** with faint brass inlay, seamless/tileable, top-down 2:1 isometric diamond, transparent edges.

| File | Note |
|---|---|
| `tiles/floor` | base diamond floor tile (2:1 iso) |
| `tiles/floor-edge` | edge/rim tile with a low stone kerb |
| `tiles/wall` | back wall segment with pipes/gears |

### 4.5 FX (small, transparent)

| File | Prompt add |
|---|---|
| `fx/smoke` | a soft **puff of grey steam/smoke**, a few frames of a small cloud |
| `fx/spark` | a bright **electric/aether spark burst** |
| `fx/alarm` | a red **rotating alarm glyph / warning klaxon light** |
| `fx/aether-pulse` | a **ring of teal energy** expanding |

### 4.6 HUD chrome (optional — can stay CSS/SVG)

`ui/plaque` (brass sign, 9-slice), `ui/scroll` (parchment, 9-slice), `ui/gauge-frame`, `ui/panel` — only if you prefer painted chrome over the CSS chrome in the mockup. Not required for Phase 3.

---

## 5. Files, sizes, directory layout

Place under `web/public/sprites/`. PNG + alpha. Sizes are the source canvas; the engine scales down.

```
web/public/sprites/
  manifest.json
  dwarves/
    runtime-smith/{idle,work,offline}.png        # 256×256
    scheduler-warden/{idle,work,offline}.png
    reconciler/{idle,work,offline}.png
    edge-tinker/{idle,work,offline}.png
    node-agent/{idle,work,offline}.png
  sheep/{running,pending,failed,succeeded,walk}.png   # 192×192
  stations/{forge-idle,forge-active}.png          # 512×384
  core/{core-idle,core-load-high}.png             # 512×640
  vault/{vault-idle,vault-active}.png             # 384×448
  pen/stray-pen.png                               # 384×256
  tiles/{floor,floor-edge,wall}.png               # floor 256×128 (2:1 iso)
  fx/{smoke,spark,alarm,aether-pulse}.png         # 128×128
  ui/{plaque,scroll,gauge-frame,panel}.png        # optional
```

Naming: lowercase-kebab, state as the filename. One pose per file (AI tools produce clean single poses far more reliably than multi-frame sheets). The engine composes motion from poses + tweens.

**Weight budget:** keep the whole `sprites/` tree under ~2–3 MB. Export at the sizes above, run through an optimizer (`pngquant`/`oxipng`, or convert to WebP if kept lossless-ish). Oversized generations → downscale before commit.

---

## 6. `manifest.json` (what the engine loads)

The scene loads this once, then the domain→scene mapper (RFC-0002) references sprites by key. Missing `file` → placeholder token.

```json
{
  "version": 1,
  "tokenSize": 96,
  "dwarves": {
    "roles": ["runtime-smith", "scheduler-warden", "reconciler", "edge-tinker", "node-agent"],
    "states": ["idle", "work", "offline"],
    "path": "dwarves/{role}/{state}.png",
    "anchor": [0.5, 0.9]
  },
  "sheep": {
    "states": ["running", "pending", "failed", "succeeded", "walk"],
    "path": "sheep/{state}.png",
    "anchor": [0.5, 0.85]
  },
  "structures": {
    "station": { "idle": "stations/forge-idle.png", "active": "stations/forge-active.png", "anchor": [0.5, 0.85] },
    "core":    { "idle": "core/core-idle.png", "high": "core/core-load-high.png", "anchor": [0.5, 0.9] },
    "vault":   { "idle": "vault/vault-idle.png", "active": "vault/vault-active.png", "anchor": [0.5, 0.9] },
    "pen":     { "idle": "pen/stray-pen.png", "anchor": [0.5, 0.9] }
  },
  "tiles": { "floor": "tiles/floor.png", "edge": "tiles/floor-edge.png", "wall": "tiles/wall.png", "tileW": 256, "tileH": 128 },
  "fx": { "smoke": "fx/smoke.png", "spark": "fx/spark.png", "alarm": "fx/alarm.png", "pulse": "fx/aether-pulse.png" }
}
```

Mapping (engine side, no backend change):
- node → one `station` + one dwarf (role picked by node label or round-robin); `status.condition === "NotReady"` → dwarf `offline`, station `idle` dimmed, lamp red.
- pod → one sheep at its node's station, grouped by `spec.node_name`; `status.phase` → sheep state; no `node_name` → sheep in the stray `pen`.
- running/total → `core` `idle`↔`high` + tube fill.
- Warning event → `fx/alarm` + `fx/smoke` at the relevant station.

---

## 7. Delivery checklist (per sprite)

- [ ] Style Anchor used; transparent background; single centered subject; no baked text/shadow/border.
- [ ] Correct file name + directory (§5); PNG with alpha.
- [ ] Consistent light (top-left), line weight, and scale with the rest of the set.
- [ ] State reads correctly at token size (~96px) — signal color visible.
- [ ] Optimized; within the weight budget.
- [ ] Entry present/valid in `manifest.json`.

---

## 8. Suggested generation order (unblocks the engine fastest)

1. `sheep/{running,pending,failed,succeeded}` — the most-repeated, most information-dense actor.
2. `dwarves/node-agent/{idle,work,offline}` — one generic dwarf to man every station.
3. `stations/{forge-idle,forge-active}`, `tiles/floor`.
4. `core/core-idle`, `vault/vault-idle`, `pen/stray-pen`.
5. Remaining dwarf roles, `*-active/high` variants, `fx/*`, `sheep/walk`.

With #1–#3 in place the Living Hall is already legible; the rest is polish.
