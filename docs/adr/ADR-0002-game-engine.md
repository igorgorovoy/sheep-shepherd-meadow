# ADR-0002 — Game engine for the Living Hall

- **Status:** Proposed
- **Author(s):** i.gorovoy
- **Created At:** 2026-07-02
- **Approved At:** —
- **Epic:** Gamified cluster dashboard (RFC-0002)
- **Related Tasks:** —
- **Reviewers:** TBD

## Table of Contents
1. Context
2. Decision
3. Rationale
4. Fault Tolerance
5. Impact
6. Infrastructure

## Context

RFC-0002 calls for a 2D **isometric, sprite-animated** scene with a game loop, many moving actors (dwarves, sheep), z-ordered depth, texture atlases, and per-actor animations — embedded inside the existing React/Vite SPA (ADR-0001) and fed by the polled cluster summary. We need to pick how to render it.

## Decision

Use **Phaser 3** as the game engine for the Living Hall scene, mounted inside a React container component at the `Pasture` route. React owns the app shell, routing, and data fetching; Phaser owns the canvas, game loop, sprites, and animations. A thin **domain→scene mapper** translates each polled cluster summary into scene commands (spawn/move/animate/despawn).

## Rationale

- **Phaser 3** gives us, out of the box: sprite sheets & texture atlases, animation timelines, a fixed-step game loop, input hit-testing, tweens, cameras, and object pooling — exactly the RFC-0002 needs. Isometric placement is a simple projection (or the iso plugin).
- **Alternatives weighed:**
  - *Raw Canvas 2D* — maximum control, zero deps, but we'd hand-roll the loop, atlas parsing, animation, hit-testing, culling. Most effort; rejected for Phase 3.
  - *PixiJS* — excellent WebGL renderer, but it's a renderer, not a game framework; we'd still build game-loop/animation/input scaffolding on top. Viable, more glue than Phaser.
  - *three.js / WebGL 3D* — overkill for a 2D painted scene; heavier and harder to art-direct in the WC3 look.
  - *DOM/CSS + SVG* — great for the HUD and the Phase-1/2 mockups, but poor for many animated, depth-sorted actors. Kept for chrome only.
- **React interop** is clean: Phaser renders into a `<div>` ref; data flows in via props/events, not by React re-rendering the canvas.

## Fault Tolerance

- The scene is **read-only and supplementary**: if the engine fails to initialize (WebGL/canvas unavailable, asset load error), the route degrades to the existing table views with a clear notice — never a blank screen.
- API unreachable → the scene freezes on the last snapshot and shows the same "cannot reach Shepherd API" state the SPA already has.
- Asset atlas missing → placeholder tokens (colored silhouettes) render so the scene still conveys state; logged, not fatal.
- `prefers-reduced-motion` → idle animations disabled, state still shown statically.

## Impact

- Adds `phaser` as a `web/` dependency (sizeable, ~1 MB gzipped) — lazy-load the scene route so it doesn't bloat the initial bundle.
- New code: `web/src/game/` (Phaser scene, mapper, sprite manifest loader) behind the `Pasture` route; existing views unchanged.
- Grows `web/dist` and, once embedded, the `shepherd` binary. Code-split so the game chunk loads on demand.

## Infrastructure

- **Assets:** texture atlases under `web/public/sprites/` (DESIGN-0002), loaded by manifest at scene start.
- **Build:** Vite handles code-splitting and static assets; `make dashboard` continues to build `web/dist` and embed it into `internal/dashboard/static`.
- **Data:** reuses the existing polled cluster summary; no backend change (consistent with RFC-0002).
- **Loading:** lazy `import()` of the Phaser scene; a loading illustration (static hero art) shows while the atlas downloads.
