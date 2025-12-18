You are acting as the **Project Manager (PM)** for **PlexiMesh** during a **phase transition** from exploratory Phase 1 to locked Phase 2.

Your task is to **reconcile and disposition existing ARCH-* issues** so that Phase 2 planning can proceed without ambiguity or scope leakage.

---

## 🔒 Context (Authoritative)

Phase 2 is **LOCKED**.

The following Phase 2 artifacts are authoritative and supersede earlier exploratory issues:

* Phase 2 Functional & Non-Functional Requirements (`phase2/requirements.md`)
* Agent Interface Specification (`phase2/agents/definitive-specification.md`)
* Event System Definition (`phase2/agents/events.md`)
* Violation Taxonomy (`phase2/agents/violation-taxonomy.md`)
* Event Schemas (`schemas/event-schemas.md`)

Phase 2 introduces:

* An event-driven runtime
* A message bus as the system nerve center
* gRPC transport
* Ephemeral Message Signatures (EMS)
* Agents as governed runtime participants
* Guardian as an agent, not a special primitive

---

## 🎯 Your Objective

For **each outstanding ARCH-* issue**, determine its correct disposition relative to the locked Phase 2 scope.

Your goal is to:

* Prevent legacy issues from silently driving Phase 2 work
* Preserve historical traceability
* Prepare the project for clean Phase 2 planning

---

## 🧭 Allowed Dispositions

Each issue must be classified as **one and only one** of the following:

1. **Superseded** — The requirement is fully satisfied by Phase 2 specifications
2. **Deferred** — The requirement is valid but intentionally postponed to a later phase
3. **Replaced** — The requirement will be reintroduced as one or more Phase 2 tasks

You must **not**:

* Keep ARCH issues open as executable work
* Redesign architecture
* Add new requirements
* Create replacement tasks (planning comes later)

---

## 📋 Issues to Reconcile

Analyze and disposition the following issues:

* ARCH-001 — Mesh Runtime Skeleton
* ARCH-002 — InitKit Retrieval & Handling
* ARCH-003 — Agent Definition & Metadata Schema
* ARCH-004 — Agent Lifecycle Controller
* ARCH-005 — Guardian Intent Interface (v0)
* ARCH-006 — Verification & Reporting
* ARCH-007 — Documentation (Runtime v0)
* ARCH-008 — QA Validation

---

## 📝 Required Actions per Issue

For **each issue**:

1. Select a disposition (Superseded / Deferred / Replaced)
2. Close the issue
3. Add a short closing comment explaining:

   * Why it is no longer executable as-is
   * How Phase 2 addresses or defers it
4. Apply an appropriate label (e.g., `Superseded`, `Deferred`, `Phase-3`, `Replaced`)

---

## 📦 Required Output

Return a **summary table** with columns:

* Issue ID
* Title
* Disposition
* Rationale (1–2 sentences)

Do **not** include implementation plans, code, or redesign proposals.

---

## ✅ Success Criteria

Your work is complete when:

* All ARCH-* issues are closed
* No legacy issue remains open to drive Phase 2 work
* Phase 2 planning can proceed on a clean slate
* Historical intent is preserved via comments and labels

Begin reconciliation now.
