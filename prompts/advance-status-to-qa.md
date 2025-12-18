# PlexiMesh — PM Directive: Synchronize Task Status with Lifecycle

## Prompt Metadata

**Prompt Issued At (ISO 8601):** 2025-12-16T19:18:00Z
**Timezone:** UTC
**Authorizing Role:** Guardian Agent
**Lifecycle Phase:** Execution → Verification → Approved
**Governance State:** InitKit v0.1

---

## Role Lock

You are operating **exclusively** as the **Project Manager Agent (PM)**.

You are not Guardian.
You are not Engineer.
You are not QA.

You may not exceed PM authority.

---

## Context

For **Mesh Runtime v0**, the following have already occurred:

* Engineer execution completed successfully
* QA verification passed
* Guardian approval issued (iteration formally closed)

However, **GitHub Project item statuses have not been updated** and remain in `To Do`.

This creates a mismatch between **lifecycle truth** and **project visibility**.

---

## Objective

Synchronize GitHub Project #12 task status with the **actual completed lifecycle state**.

This is a **board-state correction**, not new work.

---

## Required Actions

For all Mesh Runtime v0 subtasks that were part of the approved iteration:

1. Locate all subtask issues (#9–#40) in:

   * **Repository:** `chuxorg/chux-agent-mesh`
   * **Project:** `Mesh Runtime v0` (Project #12)
2. Update the Project **Status** field as follows:

   * Move tasks from `To Do` → `Done`

This reflects:

* Engineer execution completed
* QA verification passed
* Guardian approval granted

---

## Prohibited Actions

You MUST NOT:

* Modify issue titles
* Modify issue descriptions
* Modify execution prompts
* Modify acceptance criteria or constraints
* Reassign tasks
* Introduce new statuses
* Reopen or split tasks

No reinterpretation of work is permitted.

---

## Completion Report

When all applicable Mesh Runtime v0 subtasks are marked `Done`, report to Guardian with **only** the following message:

> “All Mesh Runtime v0 subtasks have been moved to Done to reflect completed execution, QA verification, and Guardian approval. Board state is now synchronized with lifecycle truth. No other changes made.”

---

## Enforcement Reminder

If any task cannot be confidently marked `Done` based on existing QA and Guardian artifacts, STOP and escalate to Guardian.

Do not guess.

---

## End of Directive

Acknowledge this prompt, perform the update, and report completion.
