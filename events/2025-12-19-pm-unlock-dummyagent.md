# PM_UNLOCK — DummyAgent (S-206)

- **Timestamp:** 2025-12-19T23:55:00Z
- **Repository:** chuxorg/chux-agent-mesh
- **Story:** S-206 — DummyAgent (#47)
- **Authorization State:** In Progress (Engineer execution approved)

## Rationale

- Message bus, EMS, gRPC (mTLS), event store, and SDK (agent constitution) are merged and carry PASS QA evidence.
- Slice 3 is formally closed; SDK now defines the authoritative agent contract, enabling downstream integration agents.
- DummyAgent is required to provide the first end-to-end runtime proof without unlocking EngineerAgent or tooling agents.

## Authorized Scope

- Implement DummyAgent strictly as a reference integration that exercises:
  - SDK lifecycle hooks and harness instrumentation
  - Bus publish/subscribe paths
  - EMS ingress validation
  - gRPC transport bindings
  - Event store append and replay observation
- Emit deterministic, schema-valid events captured by existing telemetry.

## Explicit Prohibitions

- No business logic, retries, scheduling, or policy engines.
- No external tool calls, shell execution, network listeners, or LLM usage.
- Do not introduce heuristics, optimizations, or side-channel capabilities.
- EngineerAgent (S-207) and any tooling agents remain locked pending future PM authorization.

## Execution Notes

- Project #12 card for S-206 is now `In Progress`; engineer may branch from `master` following slice naming conventions.
- Engineering must provide QA-ready evidence demonstrating SDK conformance and end-to-end event flow once implementation completes.
- Guardian oversight remains unchanged; any scope drift requires renewed PM/Guardian approval.
