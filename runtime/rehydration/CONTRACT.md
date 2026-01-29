# Rehydration Contract (Phase 3, Single-Run)

This contract defines the minimal inputs and invariants for rehydrating a single run from
historical events. Phase 0 contracts (lifecycle, envelope, guardian, abort) are frozen,
and Phase 1 enforcement is authoritative. Rehydration fails closed when required inputs
or fatal invariants are missing.

## Required Inputs

- `run_id`: minimal run identifier for the run being rehydrated.
- Event stream: all events for `run_id`, ordered by sequence. The sequence must be a
  monotonic integer associated with each event (explicit field or stream position).
- Library artifacts required for interpretation: laws and charters (prompts are optional).

## Optional Inputs

- Prompts (read-only context for traceability).
- Non-critical artifact references (links/attachments that do not affect invariants).

## Invariants

### Fatal (rehydration fails)

- Missing `system.run.started`.
- Missing terminal event: `system.run.completed` or `system.run.aborted`.
- Multiple terminal events (more than one terminal event, or both completed and aborted).
- Non-monotonic or missing sequence (sequence must be strictly increasing with no gaps).
- Engineer events before guardian PASS/REJECT without corresponding abort:
  if any event emitted by role `engineer` occurs before the first guardian decision event
  with outcome PASS or REJECT, rehydration fails unless a `system.run.aborted` event occurs
  at or before the first engineer event.

### Warnings (rehydration succeeds with report)

- Missing optional events.
- Extra events after terminal (if already rejected by gateway, classify as warning only
  when present in historical data).
- Missing non-critical artifact references.

## Read-Only Rule

Rehydration MUST NOT:
- Execute agents
- Emit events
- Mutate runtime state

Output is derived state and reports only.

## Outputs

- Rehydrated run state: lifecycle state + terminal outcome for `run_id`.
- Decision ledger: guardian decisions and supporting evidence (event references).
- Integrity report: fatal violations (if any) and warnings.

## Fixture Location

Synthetic canonical event streams will be authored under `runtime/rehydration/fixtures/`:

- `completed.jsonl`
- `aborted.jsonl`
