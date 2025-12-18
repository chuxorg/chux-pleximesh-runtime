package event

// Lifecycle states that agents report.
type AgentState string

const (
	AgentStateCreated        AgentState = "CREATED"
	AgentStateInitialized    AgentState = "INITIALIZED"
	AgentStateIdle           AgentState = "IDLE"
	AgentStateWorking        AgentState = "WORKING"
	AgentStateBlocked        AgentState = "BLOCKED"
	AgentStateAwaitingReview AgentState = "AWAITING_REVIEW"
	AgentStateCompleted      AgentState = "COMPLETED"
	AgentStateFailed         AgentState = "FAILED"
	AgentStateTerminated     AgentState = "TERMINATED"
)

// IntentOrigin identifies who produced the intent.
type IntentOrigin string

const (
	IntentOriginHuman    IntentOrigin = "human"
	IntentOriginGuardian IntentOrigin = "guardian"
)

// IntentPriority expresses urgency.
type IntentPriority string

const (
	IntentPriorityLow      IntentPriority = "low"
	IntentPriorityNormal   IntentPriority = "normal"
	IntentPriorityHigh     IntentPriority = "high"
	IntentPriorityCritical IntentPriority = "critical"
)

// VerificationResult captures verification outcomes.
type VerificationResult string

const (
	VerificationResultPassed VerificationResult = "passed"
	VerificationResultFailed VerificationResult = "failed"
)

// TaskStatus expresses a task terminal status.
type TaskStatus string

const (
	TaskStatusBlocked   TaskStatus = "blocked"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// StatusState expresses subjective status signals.
type StatusState string

const (
	StatusIdle     StatusState = "idle"
	StatusWorking  StatusState = "working"
	StatusWaiting  StatusState = "waiting"
	StatusDegraded StatusState = "degraded"
)

// HealthState enumerates health states.
type HealthState string

const (
	HealthAlive    HealthState = "alive"
	HealthReady    HealthState = "ready"
	HealthNotReady HealthState = "not_ready"
	HealthDegraded HealthState = "degraded"
)

// ViolationSeverity describes the severity of a violation.
type ViolationSeverity string

const (
	ViolationSeverityLow      ViolationSeverity = "low"
	ViolationSeverityMedium   ViolationSeverity = "medium"
	ViolationSeverityHigh     ViolationSeverity = "high"
	ViolationSeverityCritical ViolationSeverity = "critical"
)

// LifecycleStateReportedPayload reflects Lifecycle.StateReported.
type LifecycleStateReportedPayload struct {
	AgentState    AgentState `json:"agent_state"`
	PreviousState *string    `json:"previous_state,omitempty"`
	Reason        *string    `json:"reason,omitempty"`
}

// LifecycleTransitionRequestedPayload reflects Lifecycle.TransitionRequested.
type LifecycleTransitionRequestedPayload struct {
	FromState     string  `json:"from_state"`
	ToState       string  `json:"to_state"`
	Justification string  `json:"justification"`
	RelatedTaskID *string `json:"related_task_id,omitempty"`
}

// LifecycleTransitionApprovedPayload reflects Lifecycle.TransitionApproved.
type LifecycleTransitionApprovedPayload struct {
	FromState  string  `json:"from_state"`
	ToState    string  `json:"to_state"`
	ApprovedBy string  `json:"approved_by"`
	Reason     *string `json:"reason,omitempty"`
}

// LifecycleTransitionRejectedPayload reflects Lifecycle.TransitionRejected.
type LifecycleTransitionRejectedPayload = LifecycleTransitionApprovedPayload

// IntentCreatedPayload reflects Intent.Created.
type IntentCreatedPayload struct {
	IntentID string         `json:"intent_id"`
	Origin   IntentOrigin   `json:"origin"`
	Summary  string         `json:"summary"`
	Details  *string        `json:"details,omitempty"`
	Priority IntentPriority `json:"priority"`
}

// IntentClarifiedPayload reflects Intent.Clarified.
type IntentClarifiedPayload struct {
	IntentID      string `json:"intent_id"`
	Clarification string `json:"clarification"`
	ClarifiedBy   string `json:"clarified_by"`
}

// TaskCreatedPayload reflects Task.Created.
type TaskCreatedPayload struct {
	TaskID       string   `json:"task_id"`
	IntentID     string   `json:"intent_id"`
	Summary      string   `json:"summary"`
	AssignedRole string   `json:"assigned_role"`
	Constraints  []string `json:"constraints"`
}

// TaskAssignedPayload reflects Task.Assigned.
type TaskAssignedPayload struct {
	TaskID     string  `json:"task_id"`
	AssignedTo string  `json:"assigned_to"`
	AssignedBy string  `json:"assigned_by"`
	Reason     *string `json:"reason,omitempty"`
}

// TaskStatusPayload reflects Task.Blocked/Completed/Failed.
type TaskStatusPayload struct {
	TaskID    string     `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Reason    string     `json:"reason"`
	Artifacts []string   `json:"artifacts"`
}

// VerificationRequestedPayload reflects Verification.Requested.
type VerificationRequestedPayload struct {
	TaskID            string `json:"task_id"`
	RequestedBy       string `json:"requested_by"`
	VerificationScope string `json:"verification_scope"`
}

// VerificationResultPayload reflects Verification.Passed/Failed.
type VerificationResultPayload struct {
	TaskID     string             `json:"task_id"`
	VerifiedBy string             `json:"verified_by"`
	Result     VerificationResult `json:"result"`
	Notes      *string            `json:"notes,omitempty"`
}

// PromptCreatedPayload reflects Prompt.Created.
type PromptCreatedPayload struct {
	PromptID       string  `json:"prompt_id"`
	Purpose        string  `json:"purpose"`
	ContextSummary string  `json:"context_summary"`
	PromptText     string  `json:"prompt_text"`
	ModelHint      *string `json:"model_hint,omitempty"`
}

// PromptSentPayload reflects Prompt.Sent.
type PromptSentPayload struct {
	PromptID    string `json:"prompt_id"`
	LLMProvider string `json:"llm_provider"`
	Model       string `json:"model"`
}

// LLMResponseReceivedPayload reflects LLM.ResponseReceived.
type LLMResponseReceivedPayload struct {
	PromptID    string        `json:"prompt_id"`
	ResponseID  string        `json:"response_id"`
	RawResponse string        `json:"raw_response"`
	TokenUsage  LLMTokenUsage `json:"token_usage"`
}

// LLMTokenUsage describes token accounting.
type LLMTokenUsage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// PromptEvaluatedPayload reflects Prompt.Evaluated.
type PromptEvaluatedPayload struct {
	PromptID          string `json:"prompt_id"`
	ResponseID        string `json:"response_id"`
	EvaluationSummary string `json:"evaluation_summary"`
	Confidence        string `json:"confidence"`
}

// TelemetryMetricEmittedPayload reflects Telemetry.MetricEmitted.
type TelemetryMetricEmittedPayload struct {
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// TelemetryDurationMeasuredPayload reflects Telemetry.DurationMeasured.
type TelemetryDurationMeasuredPayload struct {
	Operation  string `json:"operation"`
	DurationMS int64  `json:"duration_ms"`
}

// StatusReportedPayload reflects Status.Reported.
type StatusReportedPayload struct {
	Status  StatusState `json:"status"`
	Message *string     `json:"message,omitempty"`
}

// HealthReportedPayload reflects Health.Reported.
type HealthReportedPayload struct {
	State   HealthState `json:"state"`
	Details *string     `json:"details,omitempty"`
}

// ViolationLawBreachPayload reflects Violation.LawBreach.
type ViolationLawBreachPayload struct {
	LawID       string            `json:"law_id"`
	Description string            `json:"description"`
	Severity    ViolationSeverity `json:"severity"`
	DetectedBy  string            `json:"detected_by"`
}

// ViolationCapabilityExceededPayload reflects Violation.CapabilityExceeded.
type ViolationCapabilityExceededPayload struct {
	AttemptedAction    string `json:"attempted_action"`
	RequiredCapability string `json:"required_capability"`
}

// DebugTracePayload reflects Debug.Trace.
type DebugTracePayload struct {
	TraceID string            `json:"trace_id"`
	Message string            `json:"message"`
	Context map[string]string `json:"context,omitempty"`
}
