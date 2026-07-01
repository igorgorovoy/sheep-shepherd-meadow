# Cluster Dashboard (backend)

The `shepherd` control plane serves a single-page dashboard and exposes a few
conveniences for it.

## Endpoints

- `GET /api/v1/cluster/summary` — aggregate snapshot of the cluster in one
  response:

  ```json
  {
    "info":        { "version": "...", "name": "shepherd", "node_count": 0, "pod_count": 0 },
    "nodes":       [ /* Node */ ],
    "pods":        [ /* Pod */ ],
    "deployments": [ /* Deployment */ ],
    "services":    [ /* Service */ ],
    "events":      [ /* Event */ ]
  }
  ```

  This is a convenience endpoint; the SPA can also use the individual
  `/api/v1/*` resource endpoints.

## CORS

All API responses include permissive CORS headers (`Access-Control-Allow-Origin: *`,
methods `GET, OPTIONS`) and preflight `OPTIONS` requests are answered with
`204 No Content`. This lets the Vite dev server (http://localhost:5173) call the
API (http://localhost:9876) during development.

## Static SPA serving

The SPA is embedded into the `shepherd` binary via `//go:embed` in
`internal/dashboard`. Requests that do not match an API route
(`/api/`, `/healthz`) fall through to the dashboard handler, which serves the
embedded files with SPA fallback: any path that is not a real file returns
`index.html` so client-side routing works.

A placeholder `internal/dashboard/static/index.html` is committed so the package
always compiles. The real assets are produced by the frontend worker under
`web/` (built to `web/dist`).

## Building with the real dashboard

```sh
make dashboard
```

This runs the SPA build (`web-build`), copies `web/dist/*` over the placeholder
in `internal/dashboard/static/`, then compiles the `shepherd` binary with the
real assets embedded. `web-build` skips gracefully if `web/` or `npm` is absent,
so `make build` / `go build ./...` never fail on machines without the SPA
toolchain.
