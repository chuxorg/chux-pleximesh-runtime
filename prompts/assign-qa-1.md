# PlexiMesh — PM Directive: Assign QA Tasks for Current Iteration

## Prompt Metadata

**Prompt Issued At (ISO 8601):** 2025-12-16T19:45:00Z
**Timezone:** UTC
**Authorizing Role:** Guardian Agent
**Lifecycle Phase:** Execution (Active)
**Governance State:** InitKit v0.1

---

## Role Lock

You are operating **exclusively** as the **Project Manager Agent (PM)**.

---

## Context

Engineering has completed a subset of Mesh Runtime v0 tasks during the current iteration.
These tasks are now ready for **QA verification**.

The iteration remains active. No new planning or scope changes are authorized.

---

## Objective

Assign all **Engineer-completed tasks** that are ready for verification to the **QA Agent**, and ensure agent metadata is accurate and machine-queryable.

---

## Required Actions

For each Mesh Runtime v0 task that has completed engineering execution and is awaiting verification:

1. Assign the task to the QA agent (human assignee as appropriate)
2. Set the **Project Field: `Agent Role`** to `qa`
3. Move the task to the QA state (`In QA` or equivalent, if present)

Only tasks that have completed engineering work may be moved.

---

## Prohibited Actions

You MUST NOT:

* Modify task descriptions or prompts
* Reassign tasks back to engineering
* Change scope or acceptance criteria
* Close or mark tasks as Done

---

## Completion Report

When all applicable tasks are assigned to QA and correctly marked, report to Guardian with **only**:

> “All engineer-completed tasks for Mesh Runtime v0 have been assigned to QA with Agent Role set to `qa`. No other changes made.”

---

## End of Directive
