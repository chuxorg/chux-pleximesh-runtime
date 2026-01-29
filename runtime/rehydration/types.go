package rehydration

import "time"

const (
	EventRunStarted   = "system.run.started"
	EventRunCompleted = "system.run.completed"
	EventRunAborted   = "system.run.aborted"

	EventGuardianDecision = "guardian.decision"

	GuardianOutcomePass   = "PASS"
	GuardianOutcomeReject = "REJECT"

	SystemRole   = "system"
	GuardianRole = "guardian"
	EngineerRole = "engineer"
)

// Event is the minimal, canonical record needed for deterministic rehydration.
type Event struct {
	Sequence      int
	RunID         string
	Type          string
	SchemaVersion string
	Timestamp     time.Time
	Emitter       Emitter
	Decision      *GuardianDecision
	Metadata      map[string]string
}

// Emitter describes the agent or system that produced the event.
type Emitter struct {
	AgentID string
	Role    string
}

// GuardianDecision captures a PASS/REJECT outcome plus optional evidence text.
type GuardianDecision struct {
	Outcome  string
	Evidence string
}

// Artifacts lists references used for interpretation/reporting only.
type Artifacts struct {
	Laws               []string
	Charters           []string
	Prompts            []string
	OptionalReferences []string
	OptionalEventTypes []string
}

// RunState is the derived lifecycle state for the run.
type RunState struct {
	RunID            string
	Lifecycle        string
	TerminalOutcome  string
	StartedAt        time.Time
	TerminalAt       time.Time
	LastSequence     int
	StartedSequence  int
	TerminalSequence int
}

// DecisionLedger captures guardian decisions and supporting evidence references.
type DecisionLedger struct {
	RunID     string
	Decisions []DecisionEntry
}

// DecisionEntry records a single guardian decision.
type DecisionEntry struct {
	Sequence  int
	Outcome   string
	Timestamp time.Time
	Evidence  string
	EventType string
	Emitter   Emitter
}

// IntegrityReport summarizes fatal and warning findings.
type IntegrityReport struct {
	RunID        string
	Fatal        []Issue
	Warnings     []Issue
	FatalCount   int
	WarningCount int
	Summary      string
}

// Issue captures a single validation finding.
type Issue struct {
	Code      string
	Message   string
	Sequence  int
	EventType string
}
