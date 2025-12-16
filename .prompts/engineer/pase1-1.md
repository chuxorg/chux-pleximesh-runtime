# ROLE: Engineer Agent (Codex)

You are acting as the **Engineer Agent**.

Your task is to **initialize a new Go project** for an AI agent mesh runtime.
This task is **scaffolding only**.

No runtime logic.
No agent implementations.
No business behavior.

---

## NON-NEGOTIABLE ARCHITECTURAL RULE

This project MUST follow **Vertical Slice Architecture (VSA)**.

This means:

* Code is organized by **capability / slice**, not by technical layer
* No global `services`, `handlers`, or `utils` directories
* Shared code must be minimal and explicitly justified
* Each slice owns its:

  * Interfaces
  * Logic
  * Tests (when added later)

If something does not belong to a specific slice, it probably does not belong in the project yet.

---

## GOALS

1. Create a **standard Go module**
2. Define a **clear, minimal directory structure**
3. Establish **where slices will live**
4. Add README files explaining intent
5. Ensure the project builds (empty main is acceptable)

---

## REQUIRED DIRECTORY STRUCTURE

Create the following structure exactly (names matter):

```
/cmd/
  mesh/
    main.go

/internal/
  guardian/
    README.md
  engineer/
    README.md
  compliance/
    README.md

/pkg/
  agent/
    README.md
  event/
    README.md

/initkit/
  README.md

/go.mod
/README.md
```

### Notes

* `/cmd/mesh` is the entry point (single binary)
* `/internal/*` directories represent **vertical slices**
* `/pkg/*` is for **shared contracts only** (agent interface, event types)
* `/initkit` will later contain read-only constitutional artifacts
* No additional directories are allowed

---

## FILE REQUIREMENTS

### main.go

* Minimal `package main`
* Prints a startup message
* Does not start agents
* Does not contain logic

### README.md files

Each README should briefly explain:

* The purpose of the directory
* What *belongs* there
* What explicitly does *not* belong there

Keep READMEs short and factual.

---

## CONSTRAINTS

You MUST:

* Use Go module conventions
* Use idiomatic Go formatting
* Keep code intentionally boring
* Avoid abstractions
* Avoid future-proofing
* Avoid placeholders like TODOs

You MUST NOT:

* Implement agent logic
* Add configuration systems
* Add logging frameworks
* Add dependencies beyond the Go standard library
* Add tests yet
* Add concurrency

---

## SUCCESS CRITERIA

The project:

* Builds with `go build ./...`
* Has the exact directory structure above
* Clearly communicates Vertical Slice intent
* Contains no unnecessary code

When this is complete, **stop**.

Begin now.
