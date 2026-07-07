# ADR-0003: API Bearer Token Auth for Dashboard Write Operations

- **Status:** Accepted
- **Date:** 2026-07-07
- **Related:** RFC-0003

## Context

The cluster dashboard gained write operations (apply, scale, delete) and Meadow
registry management. The Shepherd and Meadow APIs were previously unauthenticated.

## Decision

Use optional **Bearer token** middleware:

| Service | Env var | Behavior |
|---------|---------|----------|
| Shepherd | `SHEPHERD_API_TOKEN` | If set, all `/api/*` requests require `Authorization: Bearer <token>` |
| Meadow | `MEADOW_API_TOKEN` | If set, all requests except `GET /meadow/auth/status` require Bearer token |

When the env var is **empty**, auth is disabled (backward compatible).

Public endpoints (no token required when auth enabled):

- `GET /healthz`
- `GET /api/v1/auth/status` (Shepherd)
- `GET /meadow/auth/status` (Meadow)
- Embedded SPA static assets (non-`/api/` paths)

## Frontend

Tokens are stored in **runtime `localStorage`**, not `VITE_*` build vars:

- `shepherd:token`, `shepherd:apiUrl`
- `meadow:token`, `meadow:apiUrl`

Settings panel (gear icon) configures URLs and tokens.

## Consequences

- Trusted-network assumption remains; token in localStorage is not high security.
- CORS allows `Authorization` header for dev cross-origin (`:5173` → `:9876`).
- Operators enable auth by exporting `SHEPHERD_API_TOKEN` before starting shepherd.
