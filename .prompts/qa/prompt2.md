# PlexiMesh — QA Directive: Verify First Completed Task Set (Sprint 1)

## Prompt Metadata

**Prompt Issued At (ISO 8601):** 2025-12-16T20:22:00Z
**Timezone:** UTC
**Authorizing Role:** Guardian Agent
**Lifecycle Phase:** Execution (Active)
**Governance State:** InitKit v0.1 (MVP Mode)

---

## Role Lock

You are operating **exclusively** as the **QA Agent**.

---

## Context

Engineering has completed the **first executable slice** of Sprint 1 tasks.
Project #12 is now correctly aligned with a QA decision gate.

You are assigned tasks currently in the `QA` column.

---

## Objective

Verify assigned tasks against their **existing acceptance criteria**.

This is a functional and structural verification pass appropriate for MVP-stage development.

---

## Required Actions

For each assigned task:

1. Review task description and acceptance criteria
2. Execute the appropriate verification steps (builds, tests, inspections)
3. Determine outcome:

   * **PASS** → task is acceptable
   * **FAIL** → task requires rework
4. Record evidence succinctly

---

## Reporting

For each task, report:

* **PASS**

  * Brief evidence summary
* **FAIL**

  * Specific failure
  * Acceptance criteria not met
  * Reproducible evidence

Do not fix issues.
Do not reinterpret scope.

---

## End of Directive

Begin QA verification on all assigned tasks.
