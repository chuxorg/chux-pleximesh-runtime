# Planning Decomposition — Mesh Runtime v0

**Status:** Planning Complete  
**Date:** 2025-12-16  
**PM Agent:** Decomposition Report

---

## ARCH-001 — Mesh Runtime Skeleton

### Subtask 1.1: Create Repository Directory Structure
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** None (foundational structure)

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Establish the minimal directory structure required for Mesh Runtime v0.

**Deliverables:**
- Create `runtime/` directory at repository root
- Create `agents/` directory at repository root
- Create `guardian/` directory at repository root
- Create `docs/` directory at repository root
- Ensure all directories are empty (no placeholder files)

**Acceptance Criteria:**
- All four directories exist at repository root
- Directory structure is visible in repository
- No files committed within these directories

**Constraints:**
- Do not create any implementation files
- Do not add configuration files
- Do not initialize package managers or dependency files
- Do not create README files in subdirectories

---

### Subtask 1.2: Implement Runtime Entrypoint
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create the main runtime entrypoint that initializes and shuts down the runtime cleanly.

**Deliverables:**
- Create `runtime/main.ts` (or equivalent entrypoint for chosen language)
- Implement `initialize()` function that returns a runtime instance
- Implement `shutdown()` function that cleans up resources
- Export both functions for external use
- Entrypoint must be executable (can be invoked directly)

**Acceptance Criteria:**
- Entrypoint file exists and is syntactically valid
- `initialize()` function can be called and returns without error
- `shutdown()` function can be called and completes without error
- No agent logic is implemented (only structure)

**Constraints:**
- Do not implement agent discovery or loading
- Do not implement lifecycle state machine
- Do not implement InitKit retrieval
- Do not add external dependencies beyond language runtime

---

### Subtask 1.3: Implement Configuration Loading
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a configuration loading mechanism that reads runtime settings from a config file.

**Deliverables:**
- Create `runtime/config.ts` (or equivalent)
- Implement `loadConfig(path: string)` function that reads and parses config file
- Support JSON format for configuration
- Return a typed configuration object
- Handle file not found errors gracefully (return error, do not throw)

**Acceptance Criteria:**
- Config loader reads valid JSON config file
- Returns typed configuration object on success
- Returns error object (not exception) when file is missing
- Returns error object when JSON is invalid
- No default values or fallback behavior (fail explicitly)

**Constraints:**
- Do not implement environment variable overrides
- Do not implement config validation beyond JSON parsing
- Do not implement config schema validation
- Do not create default config files

---

### Subtask 1.4: Implement Logging Baseline
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Establish a minimal logging system for runtime operations.

**Deliverables:**
- Create `runtime/logger.ts` (or equivalent)
- Implement `log(level: string, message: string, metadata?: object)` function
- Support levels: `info`, `warn`, `error`
- Output logs to stdout (structured format preferred but not required)
- Include timestamp in log output

**Acceptance Criteria:**
- Logger can emit logs at info, warn, and error levels
- Logs include timestamp
- Logs are written to stdout
- Logger does not require external dependencies (use language standard library)

**Constraints:**
- Do not implement log file rotation
- Do not implement log aggregation
- Do not add external logging libraries
- Do not implement log filtering or levels beyond the three specified

---

### Subtask 1.5: Verify Runtime Startup and Shutdown
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Verify that the runtime can start and shut down cleanly without errors.

**Deliverables:**
- Create test file `runtime/main.test.ts` (or equivalent)
- Test that `initialize()` completes without throwing
- Test that `shutdown()` completes without throwing
- Test that shutdown can be called after initialize
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- All tests pass
- No errors or warnings during test execution
- Tests verify both happy path (initialize → shutdown)

**Constraints:**
- Do not test agent functionality (not yet implemented)
- Do not test configuration loading (separate task)
- Do not test logging output format
- Focus only on entrypoint lifecycle

---

## ARCH-002 — InitKit Retrieval & Handling

### Subtask 2.1: Implement InitKit Fetch Mechanism
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`, `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a function that retrieves the InitKit from a git repository, archive, or URL.

**Deliverables:**
- Create `runtime/initkit/fetcher.ts` (or equivalent)
- Implement `fetchInitKit(source: string, targetPath: string)` function
- Support git repository URLs (clone operation)
- Support archive URLs (download and extract)
- Support local file paths (copy operation)
- Return success/error result (typed error, not exception)

**Acceptance Criteria:**
- Function can clone from git repository URL
- Function can download and extract from archive URL
- Function can copy from local file path
- All operations write to specified `targetPath`
- Errors are returned as typed error objects, not thrown
- Function handles network failures gracefully

**Constraints:**
- Do not implement InitKit validation (separate task)
- Do not implement caching or version checking
- Do not implement authentication (assume public or pre-authenticated)
- Do not implement retry logic

---

### Subtask 2.2: Implement .initkit/ Directory Handling
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`, `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create functions to manage the `.initkit/` directory, including creation and read-only enforcement.

**Deliverables:**
- Create `runtime/initkit/directory.ts` (or equivalent)
- Implement `ensureInitKitDirectory(path: string)` function that creates `.initkit/` if missing
- Implement `enforceReadOnly(path: string)` function that sets directory permissions to read-only
- Implement `validateInitKitStructure(path: string)` function that checks for required `constitution/` and `runrules.md`
- All functions return success/error results (typed errors)

**Acceptance Criteria:**
- `.initkit/` directory is created if it does not exist
- Directory permissions are set to read-only after InitKit is loaded
- Structure validation checks for `constitution/` directory
- Structure validation checks for `runrules.md` file
- All operations return typed error objects on failure

**Constraints:**
- Do not implement InitKit content validation (only structure)
- Do not implement InitKit version checking
- Do not modify InitKit contents
- Do not implement InitKit update mechanisms

---

### Subtask 2.3: Implement InitKit Failure Modes
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Handle failure cases when InitKit is missing or invalid.

**Deliverables:**
- Create `runtime/initkit/errors.ts` (or equivalent)
- Define typed error classes: `InitKitNotFoundError`, `InitKitInvalidError`, `InitKitStructureError`
- Each error must include: error type, message, and path where failure occurred
- Create `handleInitKitFailure(error: Error)` function that formats error for Guardian reporting
- Errors must be logged using runtime logger

**Acceptance Criteria:**
- Three typed error classes exist and are exported
- Errors include descriptive messages
- Errors include the path where failure occurred
- Failure handler formats errors for Guardian consumption
- All errors are logged using runtime logger

**Constraints:**
- Do not implement error recovery or fallback behavior
- Do not implement InitKit auto-retrieval on failure
- Do not create default InitKit content
- Errors must cause runtime to halt (no silent failures)

---

### Subtask 2.4: Verify .initkit/ is Git-Ignored
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Verify that `.initkit/` directory is properly excluded from git tracking.

**Deliverables:**
- Check that `.gitignore` file exists in repository root
- Verify `.initkit/` is listed in `.gitignore`
- Create test that verifies git does not track files in `.initkit/`
- Test must fail if `.initkit/` is not ignored

**Acceptance Criteria:**
- `.gitignore` file exists
- `.initkit/` entry exists in `.gitignore`
- Test confirms git ignores `.initkit/` directory
- Test passes

**Constraints:**
- Do not modify `.gitignore` (only verify)
- Do not test InitKit retrieval functionality
- Focus only on git ignore verification

---

## ARCH-003 — Agent Definition & Metadata Schema

### Subtask 3.1: Define Agent Metadata Schema
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a TypeScript type (or equivalent) that defines the structure of agent metadata.

**Deliverables:**
- Create `agents/metadata.ts` (or equivalent)
- Define `AgentMetadata` type with fields: `name: string`, `role: string`, `capabilities: string[]`
- Define `AgentDefinition` type that includes `AgentMetadata` and optional fields as needed
- Export types for use by other modules
- Types must be strict (no `any` or untyped fields)

**Acceptance Criteria:**
- `AgentMetadata` type exists and is exported
- Type includes required fields: name, role, capabilities
- Type is strict (fully typed, no `any`)
- Type can be used to validate agent definitions

**Constraints:**
- Do not implement validation logic (separate task)
- Do not create sample agent definitions (separate task)
- Do not implement agent loading (separate task)
- Types only, no runtime behavior

---

### Subtask 3.2: Implement Agent Metadata Validation
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`, `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create validation logic that ensures agent metadata conforms to the schema.

**Deliverables:**
- Create `agents/validator.ts` (or equivalent)
- Implement `validateAgentMetadata(data: unknown): Result<AgentMetadata, ValidationError>` function
- Validate that `name` is non-empty string
- Validate that `role` is non-empty string
- Validate that `capabilities` is array of non-empty strings
- Return typed error if validation fails

**Acceptance Criteria:**
- Validator accepts unknown input and returns typed result
- Validator rejects invalid name (empty, missing, wrong type)
- Validator rejects invalid role (empty, missing, wrong type)
- Validator rejects invalid capabilities (not array, empty array, non-string elements)
- Validator returns typed error on failure

**Constraints:**
- Do not implement role-specific validation (e.g., valid role names)
- Do not implement capability validation (e.g., valid capability names)
- Do not implement agent loading from files (separate task)
- Focus only on schema structure validation

---

### Subtask 3.3: Create Sample Agent Definitions
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** None (data only)

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create sample agent metadata files for PM, Engineer, and QA agents.

**Deliverables:**
- Create `agents/samples/pm-agent.json` with valid PM agent metadata
- Create `agents/samples/engineer-agent.json` with valid Engineer agent metadata
- Create `agents/samples/qa-agent.json` with valid QA agent metadata
- All files must conform to `AgentMetadata` schema
- All files must pass validation

**Acceptance Criteria:**
- Three sample agent JSON files exist
- PM agent has role "pm" or "project-manager"
- Engineer agent has role "engineer" or "codex-engineer"
- QA agent has role "qa" or "qa-agent"
- All files are valid JSON
- All files pass metadata validation

**Constraints:**
- Do not implement agent loading logic (separate task)
- Do not implement agent execution logic
- Do not create agent implementation code
- Metadata files only, no behavior

---

### Subtask 3.4: Verify Agent Loading from Metadata
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Verify that agents can be loaded from metadata files without hard-coded role logic.

**Deliverables:**
- Create test file `agents/metadata.test.ts` (or equivalent)
- Test that sample agent files can be loaded and parsed
- Test that loaded agents conform to `AgentMetadata` type
- Test that validation passes for all sample agents
- Test that invalid metadata files are rejected
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests load all three sample agents successfully
- Tests verify metadata structure matches schema
- Tests reject invalid metadata files
- All tests pass
- No hard-coded role checks in test code (use schema validation only)

**Constraints:**
- Do not test agent execution (not yet implemented)
- Do not test agent lifecycle (separate task)
- Focus only on metadata loading and validation

---

## ARCH-004 — Agent Lifecycle Controller

### Subtask 4.1: Define Lifecycle State Machine
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a type system that defines the lifecycle state machine with states and valid transitions.

**Deliverables:**
- Create `runtime/lifecycle/types.ts` (or equivalent)
- Define `LifecycleState` enum/type with values: `Discover`, `Initialize`, `Execute`, `Verify`, `Report`, `Terminate`
- Define `LifecycleTransition` type that maps from state to allowed next states
- Define transition rules: Discover → Initialize → Execute → Verify → Report → Terminate
- Export types for use by controller

**Acceptance Criteria:**
- All six lifecycle states are defined
- Transition rules are explicitly defined
- Types are strict (no `any`)
- Types can be used to validate state transitions

**Constraints:**
- Do not implement transition logic (separate task)
- Do not implement state persistence
- Types only, no runtime behavior

---

### Subtask 4.2: Implement Lifecycle State Controller
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`, `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a controller that enforces lifecycle state transitions according to the state machine rules.

**Deliverables:**
- Create `runtime/lifecycle/controller.ts` (or equivalent)
- Implement `LifecycleController` class with methods: `getCurrentState()`, `transitionTo(nextState: LifecycleState)`
- Enforce transition rules (block invalid transitions)
- Return typed error when transition is invalid
- Log all state transitions using runtime logger

**Acceptance Criteria:**
- Controller maintains current state
- Valid transitions succeed (Discover → Initialize → Execute → Verify → Report → Terminate)
- Invalid transitions are blocked and return typed error
- All transitions are logged
- Controller starts in `Discover` state

**Constraints:**
- Do not implement state persistence (in-memory only)
- Do not implement rollback or undo
- Do not implement parallel state machines
- Single agent lifecycle only

---

### Subtask 4.3: Implement Lifecycle Failure Handling
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Handle failure cases during lifecycle transitions and execution.

**Deliverables:**
- Create `runtime/lifecycle/errors.ts` (or equivalent)
- Define typed error: `InvalidTransitionError` with current state and attempted next state
- Define typed error: `LifecycleFailureError` with state where failure occurred
- Implement `handleLifecycleFailure(error: Error, currentState: LifecycleState)` function
- Failures must transition to `Terminate` state
- All failures must be logged

**Acceptance Criteria:**
- Two typed error classes exist
- Errors include state information
- Failure handler transitions to `Terminate` on any failure
- All failures are logged with state context
- Errors are formatted for Guardian reporting

**Constraints:**
- Do not implement failure recovery or retry
- Do not implement partial rollback
- Failures cause immediate termination
- No silent error handling

---

### Subtask 4.4: Verify Lifecycle Transition Enforcement
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Verify that invalid lifecycle transitions are blocked and valid transitions succeed.

**Deliverables:**
- Create test file `runtime/lifecycle/controller.test.ts` (or equivalent)
- Test valid transition sequence: Discover → Initialize → Execute → Verify → Report → Terminate
- Test that invalid transitions are blocked (e.g., Execute → Discover)
- Test that controller starts in Discover state
- Test that failures transition to Terminate
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests verify all valid transitions succeed
- Tests verify invalid transitions are blocked
- Tests verify failure handling transitions to Terminate
- All tests pass

**Constraints:**
- Do not test agent execution (not yet implemented)
- Do not test InitKit integration (separate task)
- Focus only on state machine behavior

---

## ARCH-005 — Guardian Intent Interface (v0)

### Subtask 5.1: Define Intent Declaration Format
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a type system that defines the structure of Guardian intent declarations.

**Deliverables:**
- Create `guardian/intent/types.ts` (or equivalent)
- Define `GuardianIntent` type with fields: `projectId: string`, `scope: string`, `approved: boolean`, `timestamp: string`
- Define intent declaration format (JSON schema or TypeScript type)
- Export types for use by interface

**Acceptance Criteria:**
- `GuardianIntent` type exists and is exported
- Type includes required fields: projectId, scope, approved, timestamp
- Type is strict (fully typed, no `any`)
- Type can be used to validate intent declarations

**Constraints:**
- Do not implement intent validation (separate task)
- Do not implement intent storage
- Types only, no runtime behavior

---

### Subtask 5.2: Implement Guardian Approval Checkpoints
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`, `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create checkpoints that require Guardian intent before proceeding with lifecycle transitions.

**Deliverables:**
- Create `guardian/intent/checkpoint.ts` (or equivalent)
- Implement `requireGuardianIntent(state: LifecycleState): Result<GuardianIntent, Error>` function
- Checkpoint must block transition from `Discover` to `Initialize` without approved intent
- Checkpoint must block transition from `Execute` to `Verify` without approved intent
- Return typed error if intent is missing or not approved

**Acceptance Criteria:**
- Checkpoint function exists and is callable
- Function blocks Discover → Initialize without approved intent
- Function blocks Execute → Verify without approved intent
- Function returns typed error when intent is missing
- Function returns typed error when intent is not approved

**Constraints:**
- Do not implement intent storage or persistence
- Do not implement intent UI or CLI (separate task)
- Focus only on checkpoint enforcement

---

### Subtask 5.3: Implement CLI Intent Interface
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/command-handler.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a simple CLI interface that allows Guardian to provide intent declarations.

**Deliverables:**
- Create `guardian/intent/cli.ts` (or equivalent)
- Implement CLI command: `guardian intent approve --project-id <id> --scope <scope>`
- CLI must accept project ID and scope as arguments
- CLI must output intent declaration in JSON format
- CLI must validate input (non-empty project ID and scope)

**Acceptance Criteria:**
- CLI command exists and is executable
- Command accepts required arguments (project-id, scope)
- Command outputs valid GuardianIntent JSON
- Command validates input and returns error if invalid
- Command can be invoked from command line

**Constraints:**
- Do not implement intent storage (output only)
- Do not implement interactive prompts
- Do not implement intent history
- Simple CLI only, no GUI

---

### Subtask 5.4: Verify Intent Visibility in Logs
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Verify that Guardian intent is visible in runtime logs at checkpoints.

**Deliverables:**
- Create test file `guardian/intent/checkpoint.test.ts` (or equivalent)
- Test that intent approval is logged when checkpoint passes
- Test that intent rejection is logged when checkpoint fails
- Test that log output includes project ID and scope
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests verify intent is logged on approval
- Tests verify intent rejection is logged on failure
- Tests verify log includes project ID and scope
- All tests pass

**Constraints:**
- Do not test CLI interface (separate task)
- Do not test intent storage
- Focus only on log visibility

---

## ARCH-006 — Verification & Reporting

### Subtask 6.1: Implement Verification Hooks
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create hooks that allow agents to register verification functions that run before lifecycle transitions.

**Deliverables:**
- Create `runtime/verification/hooks.ts` (or equivalent)
- Implement `registerVerificationHook(name: string, hook: () => Promise<VerificationResult>)` function
- Implement `runVerificationHooks(): Promise<VerificationResult[]>` function
- Verification result must include: `success: boolean`, `message: string`, `details?: object`
- Hooks must run before `Verify` → `Report` transition

**Acceptance Criteria:**
- Hook registration function exists
- Hook execution function exists
- Multiple hooks can be registered
- Hooks return verification results
- Hooks are executed before Verify → Report transition

**Constraints:**
- Do not implement verification logic (agents provide hooks)
- Do not implement hook persistence
- Do not implement hook ordering or dependencies
- Simple hook system only

---

### Subtask 6.2: Implement Structured Report Output
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create a reporting system that generates structured reports for Guardian.

**Deliverables:**
- Create `runtime/reporting/reporter.ts` (or equivalent)
- Implement `generateReport(verificationResults: VerificationResult[], lifecycleState: LifecycleState): Report` function
- Report must include: `projectId: string`, `status: 'success' | 'failure'`, `verificationResults: VerificationResult[]`, `timestamp: string`
- Implement `sendReportToGuardian(report: Report): Promise<void>` function (output to stdout for v0)
- Report must be in JSON format

**Acceptance Criteria:**
- Report generation function exists
- Report includes all required fields
- Report is valid JSON
- Report is output to stdout (Guardian consumption)
- Report includes verification results

**Constraints:**
- Do not implement report storage or persistence
- Do not implement report delivery (stdout only for v0)
- Do not implement report formatting beyond JSON
- Simple reporting only

---

### Subtask 6.3: Implement Success/Failure Signals
**Assigned Agent:** Engineer Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/backend/domain-service.md`, `.initkit/initkit/constitution/patterns/errors/typed-error.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create signals that indicate execution success or failure for Guardian consumption.

**Deliverables:**
- Create `runtime/reporting/signals.ts` (or equivalent)
- Implement `emitSuccessSignal(projectId: string, details?: object): void` function
- Implement `emitFailureSignal(projectId: string, error: Error, details?: object): void` function
- Signals must be logged using runtime logger
- Signals must be included in Guardian report

**Acceptance Criteria:**
- Success signal function exists and is callable
- Failure signal function exists and is callable
- Signals are logged with appropriate log level
- Signals are included in Guardian reports
- Signals include project ID

**Constraints:**
- Do not implement signal delivery beyond logging and reporting
- Do not implement signal persistence
- Do not implement signal retry or acknowledgment
- Simple signal emission only

---

### Subtask 6.4: Verify Report Generation and Delivery
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Verify that reports are generated correctly and delivered to Guardian (stdout).

**Deliverables:**
- Create test file `runtime/reporting/reporter.test.ts` (or equivalent)
- Test that report is generated with all required fields
- Test that report includes verification results
- Test that report is valid JSON
- Test that report is written to stdout
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests verify report structure matches schema
- Tests verify report includes verification results
- Tests verify report is valid JSON
- Tests verify report is written to stdout
- All tests pass

**Constraints:**
- Do not test report delivery beyond stdout
- Do not test report storage
- Focus only on report generation and stdout output

---

## ARCH-007 — Documentation (Runtime v0)

### Subtask 7.1: Write Runtime Overview Documentation
**Assigned Agent:** Documentation Agent  
**Pattern Reference:** None (documentation task)

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create documentation that explains the Mesh Runtime v0 architecture and purpose.

**Deliverables:**
- Create `docs/runtime-overview.md`
- Document the purpose of Mesh Runtime v0
- Document the high-level architecture (runtime, agents, guardian)
- Document the key concepts (InitKit, lifecycle, intent)
- Documentation must be clear and accessible to new contributors

**Acceptance Criteria:**
- Documentation file exists in `docs/` directory
- Documentation explains runtime purpose
- Documentation describes high-level architecture
- Documentation introduces key concepts
- Documentation is readable and well-structured

**Constraints:**
- Do not include implementation details (separate task)
- Do not include code examples (separate task)
- Focus on conceptual overview only

---

### Subtask 7.2: Write Agent Lifecycle Documentation
**Assigned Agent:** Documentation Agent  
**Pattern Reference:** None (documentation task)

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Document the agent lifecycle state machine and transition rules.

**Deliverables:**
- Create `docs/agent-lifecycle.md`
- Document all lifecycle states (Discover, Initialize, Execute, Verify, Report, Terminate)
- Document valid state transitions
- Document invalid transition handling
- Include state diagram or table showing transitions

**Acceptance Criteria:**
- Documentation file exists
- All six lifecycle states are documented
- Valid transitions are clearly described
- Invalid transition handling is explained
- Documentation includes visual representation (diagram or table)

**Constraints:**
- Do not include code examples
- Do not include implementation details
- Focus on lifecycle behavior only

---

### Subtask 7.3: Write InitKit Usage Documentation
**Assigned Agent:** Documentation Agent  
**Pattern Reference:** None (documentation task)

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Document how agents retrieve and consume the InitKit.

**Deliverables:**
- Create `docs/initkit-usage.md`
- Document InitKit retrieval process (git clone, archive download, local copy)
- Document `.initkit/` directory structure and purpose
- Document which InitKit files agents must load (constitution/, runrules.md)
- Document how to verify InitKit is canonical (Status: Canonical)
- Include step-by-step instructions for manual InitKit retrieval

**Acceptance Criteria:**
- Documentation file exists
- Retrieval process is documented with clear steps
- Directory structure is explained
- Required files are listed
- Canonical verification process is described
- Instructions are executable by humans

**Constraints:**
- Do not include code examples
- Do not include implementation details
- Focus on usage instructions only

---

### Subtask 7.4: Write Example Flow Documentation
**Assigned Agent:** Documentation Agent  
**Pattern Reference:** None (documentation task)

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create an end-to-end example that demonstrates how to run Mesh Runtime v0.

**Deliverables:**
- Create `docs/example-flow.md`
- Document complete workflow from InitKit retrieval to runtime execution
- Include example commands for each step
- Include example output or expected behavior
- Document how to verify successful execution

**Acceptance Criteria:**
- Documentation file exists
- Complete workflow is documented step-by-step
- Example commands are provided and executable
- Expected output is described
- Verification steps are included
- A new contributor can follow the example successfully

**Constraints:**
- Do not include implementation code
- Focus on executable example only
- Example must be complete and runnable

---

## ARCH-008 — QA Validation

### Subtask 8.1: Create Startup Tests
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create tests that verify the runtime starts correctly with all required components.

**Deliverables:**
- Create test file `tests/startup.test.ts` (or equivalent)
- Test that runtime initializes without errors
- Test that config loading works with valid config
- Test that logging system is operational
- Test that all required directories exist
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests verify runtime initialization
- Tests verify config loading
- Tests verify logging functionality
- Tests verify directory structure
- All tests pass

**Constraints:**
- Do not test agent functionality (separate task)
- Do not test InitKit retrieval (separate task)
- Focus only on startup validation

---

### Subtask 8.2: Create InitKit Failure Tests
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create tests that verify the runtime handles InitKit failures correctly.

**Deliverables:**
- Create test file `tests/initkit-failure.test.ts` (or equivalent)
- Test that missing InitKit causes runtime to halt with error
- Test that invalid InitKit structure causes runtime to halt with error
- Test that missing `constitution/` directory is detected
- Test that missing `runrules.md` file is detected
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests verify missing InitKit is handled
- Tests verify invalid structure is detected
- Tests verify required files are checked
- Tests verify runtime halts on failure (does not continue)
- All tests pass

**Constraints:**
- Do not test InitKit retrieval (separate task)
- Do not test InitKit validation beyond structure
- Focus only on failure handling

---

### Subtask 8.3: Create Lifecycle Transition Tests
**Assigned Agent:** QA Agent  
**Pattern Reference:** `.initkit/initkit/constitution/patterns/testing/unit-test.md`

Prompt Issued At (ISO 8601): 2025-12-16T18:23:58Z
Authorizing Role: Project Manager Agent
Lifecycle Phase: Execution (Pending Guardian Approval)


**Objective:**  
Create tests that verify lifecycle transitions work correctly and prevent regressions.

**Deliverables:**
- Create test file `tests/lifecycle-transition.test.ts` (or equivalent)
- Test that valid transition sequence succeeds (Discover → Initialize → Execute → Verify → Report → Terminate)
- Test that invalid transitions are blocked
- Test that lifecycle controller enforces state machine rules
- Test that failures transition to Terminate
- All tests must pass

**Acceptance Criteria:**
- Test suite exists and is executable
- Tests verify all valid transitions succeed
- Tests verify invalid transitions are blocked
- Tests verify state machine enforcement
- Tests verify failure handling
- All tests pass
- Tests would fail if lifecycle rules are regressed

**Constraints:**
- Do not test agent execution (not yet implemented)
- Do not test InitKit integration
- Focus only on lifecycle state machine behavior

---

## Planning Summary

**Total Subtasks:** 32  
**Assigned to Engineer Agent:** 20  
**Assigned to QA Agent:** 7  
**Assigned to Documentation Agent:** 4  
**Unassigned:** 1 (this planning task)

**Planning Status:** Complete  
**Ready for Guardian Review:** Yes

---

## Next Steps

1. Guardian reviews this decomposition
2. Guardian approves transition to Execution phase
3. Subtasks are assigned to agents
4. Execution begins

---

**PM Agent Report to Guardian:**

Planning complete. All 8 high-level tasks have been decomposed into 32 atomic, executable subtasks. Each subtask includes:
- Clear objective
- Explicit deliverables
- Acceptance criteria
- Constraints (what is NOT to be done)
- Assigned agent
- Pattern references where applicable

Subtasks are written as execution prompts that agents can execute without inference. No gaps or ambiguities remain.

**Ready for Execution gate review.**

