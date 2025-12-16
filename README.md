# Chux Agent Mesh Runtime

This repository contains the scaffolding for the PlexiMesh runtime implemented in Go.
It intentionally focuses on structure over features so each capability can evolve inside its own vertical slice.

## Project Layout

- `cmd/mesh`: single entry point that wires the runtime binary.
- `internal/guardian`: guardian slice that will enforce runtime guardrails.
- `internal/engineer`: engineer slice responsible for build and adaptation workflows.
- `internal/compliance`: compliance slice that owns audits and constitutional enforcement.
- `pkg/agent` and `pkg/event`: shared contracts that multiple slices can consume.
- `initkit`: frozen constitutional artifacts referenced by every slice.

The runtime currently exposes only a placeholder `main` to prove the build works.
Future changes must preserve vertical slice boundaries and keep shared code minimal.
