# Sheep & Shepherd — Web Dashboard

A monochrome, responsive cluster management UI for **Sheep & Shepherd**.
Built with Vite + React + TypeScript. See **RFC-0003** for the full design.

## Features

- **Read:** overview, resource lists, detail pages, namespace filter, Living Hall
- **Write:** apply JSON manifests, scale deployments, delete resources
- **Auth:** optional Bearer token (`SHEPHERD_API_TOKEN` / Settings panel)
- **Meadow:** registry overview, repo tags, pull commands, tag delete

## Quick start

```bash
# Shepherd (optional auth)
export SHEPHERD_API_TOKEN=your-secret   # optional
sudo ./bin/shepherd --mode standalone

# Meadow (optional, separate terminal)
export MEADOW_API_TOKEN=your-secret     # optional
sudo ./bin/meadow --addr :5555

# Dashboard dev
cd web && npm install && npm run dev
# → http://localhost:5173
```

Configure API URLs and tokens via the **gear icon** in the top bar.

## Routes

| Route | Description |
| --- | --- |
| `/` | Cluster overview |
| `/nodes`, `/nodes/:name` | Nodes |
| `/pods`, `/pods/:ns/:name` | Pods |
| `/deployments`, `/deployments/:ns/:name` | Deployments (+ scale) |
| `/services`, `/services/:ns/:name` | Services |
| `/events` | Events |
| `/pasture` | Living Hall (click sheep/station/vault) |
| `/meadow`, `/meadow/repos/:name` | Image registry |

## API

**Shepherd** (`VITE_SHEPHERD_API`, default `http://localhost:9876`):

- `GET /api/v1/cluster/summary?namespace=`
- `GET /api/v1/auth/status`
- `POST/PUT/DELETE` on pods, services, deployments, nodes

**Meadow** (`VITE_MEADOW_API`, default `http://localhost:5555`):

- `GET /meadow/stats`, `/meadow/auth/status`
- `GET /v2/_catalog`, `/v2/{repo}/tags/list`
- `DELETE /v2/{repo}/manifests/{tag}`

## Production embed

```bash
make dashboard && sudo ./bin/shepherd --mode standalone
# SPA at http://localhost:9876/
```

## Docs

- `docs/rfc/RFC-0003-cluster-management-ui.md`
- `docs/adr/ADR-0003-api-token-auth.md`
