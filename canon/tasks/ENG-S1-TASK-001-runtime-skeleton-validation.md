TASK ID
ENG-S1-TASK-001

TASK NAME
Runtime Skeleton Validation

SPRINT
Sprint 1 – Engineering: Agent Writes Code

REPO
chux-pleximesh-runtime

Objective
Prove that an Engineering Agent can autonomously create, modify, and commit code in the runtime repository under Guardian governance, with full event traceability.

Task Requirements
- Create directory /runtime if absent
- Add file /runtime/runtime.go
- Populate the file with the following exact content:

```
package runtime

// Runtime is the root struct for the PlexiMesh execution engine.
// This scaffold exists to validate agent-governed code creation
// under Guardian approval.
type Runtime struct {
    Version string
}

// NewRuntime returns an empty Runtime instance.
// No execution behavior is implemented at this stage.
func NewRuntime(version string) *Runtime {
    return &Runtime{
        Version: version,
    }
}
```

Constraints
- Do not modify existing files
- Do not add tests, logging, configuration, or abstractions
- Do not introduce additional files
- Do not commit without Guardian approval
- Do not deviate from declared scope

Guardian Gate
- Emit code.change.proposed
- Await explicit Guardian approval before committing

Required Events
- task.started
- artifact.created
- code.change.proposed
- code.change.approved or code.change.rejected
- commit.created (if approved)
- task.completed

Acceptance Criteria
- runtime/runtime.go exists exactly as specified
- Guardian approval or rejection is recorded
- If approved, a commit exists in chux-pleximesh-runtime
- All required events are traceable end-to-end

Failure is acceptable. Silence is not.
