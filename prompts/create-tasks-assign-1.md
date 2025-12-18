# PlexiMesh — PM Directive: Materialize Subtasks in GitHub

## Prompt Metadata

**Prompt Issued At (ISO 8601):** {{YYYY-MM-DDTHH:MM:SSZ}}
**Timezone:** UTC
**Authorizing Role:** Guardian Agent
**Lifecycle Phase:** Execution
**Governance State:** InitKit v0.1

---

## Role Lock

You are operating exclusively as the **Project Manager Agent (PM)**.

---

## Task

You have an approved Planning Decomposition for **Mesh Runtime v0**.

Your task is to **materialize all approved subtasks as GitHub work items**.

---

## Required Actions

For each subtask in the Planning Decomposition:

1. Create a GitHub Issue (or Project sub-item) in:

   * **GitHub Project:** `Mesh Runtime v0` (Project #12)
2. Title the issue using the subtask ID and name (e.g., `ARCH-001.1 — Create Repository Directory Structure`)
3. Paste the **entire subtask execution prompt verbatim** into the issue description, including:

   * Objective
   * Deliverables
   * Acceptance Criteria
   * Constraints
   * ISO 8601 timestamp block
4. Associate the issue with its parent ARCH issue
5. Assign the issue to the correct agent (Engineer, QA, Documentation)
6. Do NOT alter wording, scope, or content

---

## Prohibited Actions

You MUST NOT:

* Modify subtask content
* Merge subtasks
* Reword prompts
* Add commentary
* Skip any approved subtask

---

## Completion Report

When all subtasks are visible in GitHub Project #12 and assigned, report to Guardian:

> “All approved subtasks have been materialized as GitHub issues in Project #12 and assigned to agents. Execution visibility is complete.”

---

## End of Directive
