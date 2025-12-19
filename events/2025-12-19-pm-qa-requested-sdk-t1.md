# PM_QA_REQUESTED — SDK-T1

- **Timestamp:** 2025-12-19T23:15:00Z
- **Repository:** chuxorg/chux-agent-mesh
- **Branch:** phase2a-slice3-sdk-t1
- **Commit:** cb1522f8e96df9b0ab1e056445c40b73a30334bc
- **Task:** SDK-T1 — Agent Base & Harness (#59)
- **Authorized Scope:** SDK lifecycle, harness, manifest loader, interfaces

## QA Instructions (Forwarded Verbatim)

You are acting as QA for PlexiMesh Phase 2A — SDK-T1.

Validation Scope (Only)

Agent lifecycle ordering

Deterministic harness behavior

Manifest parsing + immutability

SDK error surfaces

Interface isolation (no bus / EMS / transport access)

Prohibited

Do not test message bus, EMS, gRPC, or event store

Do not introduce network listeners

Do not validate downstream agents

Required

Checkout:

git checkout phase2a-slice3-sdk-t1

Run:

go test ./internal/sdk/...

Inspect code paths for:

Lifecycle enforcement

Harness determinism

Capability immutability

Absence of forbidden imports

Reporting

Publish QA_TEST_RESULTS including:

Commit hash

PASS / FAIL

File + line references for:

Lifecycle hooks

Harness injection/capture

Manifest validation

Typed error paths

Begin validation only for SDK-T1.

🚦 Guardrails

SDK-T1 must pass QA before:

DummyAgent

EngineerAgent

Any runtime wiring

If defects are found:

Engineer remediates on a scoped branch

PM re-authorizes QA for remediation only

✅ Success Criteria

SDK-T1 is accepted when QA confirms:

Deterministic lifecycle

No forbidden dependencies

Harness reliably captures agent behavior

SDK constrains agents instead of empowering them

Why this matters (meta, for you)

This SDK is now the constitutional layer of PlexiMesh.
Once it’s accepted, every agent you ever build is boxed in safely.

You are doing this exactly right.

ChatGPT can make mistakes. Check important info.
