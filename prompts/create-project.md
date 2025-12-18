Here’s a concise prompt you can feed into GitHub Copilot (or another Codex‑style assistant) to generate the project, board, and issues from your requirements document:

---

### 📋 Codex Prompt to Create GitHub Project & Issues

> **Goal:** Create a new GitHub Project for the `chux-agent-mesh` repository within the `chuxorg` organization, and populate it with the execution tasks defined in the Mesh Runtime v0 architecture document.
>
> **Steps to perform:**
>
> 1. **Create a Project Board**
>
>    * **Project name:** `Mesh Runtime v0`
>    * **Description:** “Execution plan for implementing the Mesh Runtime v0 architecture: agent lifecycle, InitKit handling, verification and reporting, and supporting documentation/testing.”
>    * **Visibility:** Private (within organization).
>    * **Default view:** Table.
>
> 2. **Create eight issues in `chuxorg/chux-agent-mesh` and add them to the project.**  Use the following titles and bodies:
>
>    1. **ARCH‑001 — Mesh Runtime Skeleton**
>
>       * *Body:* “Establish minimal runtime structure. Deliverables: repo layout (`runtime/`, `agents/`, `guardian/`, `docs/`), runtime entrypoint, config loading, and logging baseline. Acceptance: runtime starts and shuts down cleanly; no agent logic yet.”
>    2. **ARCH‑002 — InitKit Retrieval & Handling**
>
>       * *Body:* “Implement InitKit fetch mechanism (git/archive/URL), `.initkit/` directory handling, read‑only enforcement, and failure modes (missing/invalid InitKit). Acceptance: agent can retrieve InitKit, `.initkit/` is git‑ignored, clear error on failure.”
>    3. **ARCH‑003 — Agent Definition & Metadata Schema**
>
>       * *Body:* “Define declarative agent metadata schema (name, role, capabilities) and validation logic. Create sample agents (PM, Engineer, QA). Acceptance: agents load from metadata; no hard‑coded role logic.”
>    4. **ARCH‑004 — Agent Lifecycle Controller**
>
>       * *Body:* “Implement lifecycle state machine (Discover → Initialize → Execute → Verify → Report → Terminate) with transition rules and failure handling. Acceptance: lifecycle enforced; invalid transitions blocked.”
>    5. **ARCH‑005 — Guardian Intent Interface (v0)**
>
>       * *Body:* “Provide intent injection layer: define intent declaration format, implement Guardian approval checkpoints, and supply a simple CLI or config‑based interface. Acceptance: runtime requires Guardian intent to proceed; intent is visible in logs.”
>    6. **ARCH‑006 — Verification & Reporting**
>
>       * *Body:* “Add verification hooks and structured report output; implement success/failure signals. Acceptance: every execution ends with verification; Guardian receives a report.”
>    7. **ARCH‑007 — Documentation (Runtime v0)**
>
>       * *Body:* “Write documentation covering the runtime overview, agent lifecycle, InitKit usage, and an example flow. Acceptance: a new contributor can run v0 end‑to‑end.”
>    8. **ARCH‑008 — QA Validation**
>
>       * *Body:* “Create startup tests, InitKit failure tests, and lifecycle transition tests to prevent regressions. Acceptance: tests fail on lifecycle or InitKit regression.”
>
> 3. **Assign roles or labels** (optional) to reflect engineer, docs, and QA responsibilities as noted in each task.
>
> 4. **Link the issues** into the project board under the default view (Table) so that you can track progress per issue.

---

You can paste that prompt directly into Copilot or a similar tool; it describes exactly what to create and how each task should be structured. Once executed, you’ll have the project board and issues set up in the correct repository and organization.
