# Chux Agent Mesh Runtime

This repository contains the scaffolding for the PlexiMesh runtime implemented in Go.
It intentionally focuses on structure over features so each capability can evolve inside its own vertical slice.

## Project Layout

- `cmd/mesh`: single entry point that wires the runtime binary.
- `cmd/runtime-daemon`: standalone HTTP + SSE daemon for AWACS integration.
- `internal/guardian`: guardian slice that will enforce runtime guardrails.
- `internal/engineer`: engineer slice responsible for build and adaptation workflows.
- `internal/compliance`: compliance slice that owns audits and constitutional enforcement.
- `pkg/agent` and `pkg/event`: shared contracts that multiple slices can consume.
- `initkit`: frozen constitutional artifacts referenced by every slice.

The runtime currently exposes only a placeholder `main` to prove the build works.
Future changes must preserve vertical slice boundaries and keep shared code minimal.

## Runtime Daemon (HTTP + SSE)

Purpose: standalone runtime process for AWACS integration.

Entrypoint: `cmd/runtime-daemon/main.go`

Default port: 8787 (config is env-only via `RUNTIME_PORT`, `EVENT_BUFFER_SIZE`,
`BUS_SUBSCRIBER_BUFFER`, and `SNAPSHOT_LIMIT`).

Endpoints:

- `GET /health`
- `GET /events` (SSE; one JSON `EventEnvelope` per event)
- `POST /emit` (smoke path)
- `GET /snapshot`

Docker usage:

```sh
docker compose up --build
```

Local run:

```sh
go run ./cmd/runtime-daemon
```

Curl examples:

```sh
curl -s http://localhost:8787/health
curl -N http://localhost:8787/events
curl -s -X POST -H 'Content-Type: application/json' \
  -d '{"source":"qa","kind":"test","name":"ping","payload":{"ok":true}}' \
  http://localhost:8787/emit
```

Not included:

- library persistence
- auth
- prompt execution
- tool execution
