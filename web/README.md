# Sheep & Shepherd — Web Dashboard

A monochrome, responsive single-page cluster dashboard for the **Sheep &
Shepherd** container platform. Built with Vite + React + TypeScript.

It reads live cluster state from the **Shepherd REST API** and presents nodes,
pods, deployments, services, and events, plus a cluster overview.

## Design

- **Strictly grayscale.** Black, white, and grays only — no accent hues. Status
  is conveyed through weight, borders, fills, and stroke style (solid /
  outline / dashed / hatched), never color.
- **Light & dark**, both monochrome (light = black-on-white, dark =
  white-on-black). Toggle in the top bar; choice persists in `localStorage`.
  Default is light.
- **Design system** built on CSS custom properties (gray ramp, spacing scale,
  type scale, radii) in `src/styles/theme.css`; components in
  `src/styles/app.css`.
- **Responsive** from ~360px up: the sidebar collapses into a drawer on narrow
  screens; content flows into responsive card grids and scrollable tables.
- System sans-serif for UI, monospace for IDs / IPs / values.

## Requirements

- Node.js 18+ and npm.

## Getting started

```bash
cd web
npm install

# Point the dashboard at your Shepherd API (optional; default shown).
export VITE_SHEPHERD_API=http://localhost:9876

npm run dev        # start the dev server (http://localhost:5173)
npm run build      # type-check + production build into web/dist/
npm run preview    # preview the production build
npm run typecheck  # tsc --noEmit only
```

`VITE_SHEPHERD_API` may also be placed in `web/.env.local`
(see `.env.example`).

## API

The dashboard consumes these Shepherd endpoints (all GET, JSON, under the base
URL), calling each individual endpoint directly:

| Endpoint | Purpose |
| --- | --- |
| `GET /healthz` | connection indicator (plain-text `ok`) |
| `GET /api/v1/info` | cluster name, version, counts |
| `GET /api/v1/nodes` | node list |
| `GET /api/v1/pods` | pod list |
| `GET /api/v1/services` | service list |
| `GET /api/v1/deployments` | deployment list |
| `GET /api/v1/events` | recent events |

Data is polled every ~5s. A manual **Refresh** control and a "last updated"
indicator live in the top bar. Loading, empty, and unreachable-API states are
all handled.

## Project layout

```
web/
  index.html
  vite.config.ts
  tsconfig*.json
  src/
    main.tsx            # entry + router
    App.tsx             # routes
    api/
      types.ts          # TS interfaces mirroring the API JSON
      client.ts         # typed fetch client (VITE_SHEPHERD_API)
    hooks/
      useClusterData.ts # 5s polling + manual refresh
      useTheme.ts       # light/dark, localStorage-backed
    lib/format.ts       # bytes / cpu / relative-time formatting
    components/          # Layout, Sidebar, Topbar, Badge, Progress, states…
    pages/               # Overview, Nodes, Pods, Deployments, Services, Events, Pasture
    styles/              # theme.css (design tokens) + app.css (components)
```

## Pasture (planned)

`src/pages/Pasture.tsx` is a placeholder route for a future farm-themed
visualization where nodes are rendered as pens and pods as sheep grazing
inside them. The file contains a documented **extensibility seam** describing
exactly where the animated component plugs in and what data it receives
(`nodes` + `pods`, grouped by `pod.spec.node_name`). The animation is
intentionally not implemented yet.
