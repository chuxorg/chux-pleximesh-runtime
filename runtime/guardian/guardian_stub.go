package guardian

import (
	"strings"
	"unicode/utf8"
)

const (
	decisionSchema      = "decision.result.v0"
	decisionMessageType = "decision.result"
	guardianActor       = "guardian"
	maestroActor        = "maestro"
)

// Guardian exposes the contract enforced by PlexiMesh Guardians.
type Guardian interface {
	EvaluateCompliance(msg ComplianceCheck) (DecisionResult, error)
}

// ComplianceCheck mirrors compliance.check.v0 and represents a parsed request.
type ComplianceCheck struct {
	Schema      string                 `json:"schema"`
	MessageType string                 `json:"message_type"`
	From        string                 `json:"from"`
	To          string                 `json:"to"`
	Artifact    ComplianceArtifact     `json:"artifact"`
	Expectation ComplianceExpectation  `json:"expectation"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceArtifact contains the submitted artifact under review.
type ComplianceArtifact struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// ComplianceExpectation captures caller intent for the check.
type ComplianceExpectation struct {
	Posture string `json:"posture"`
}

// DecisionResult mirrors decision.result.v0 and is emitted by the Guardian.
type DecisionResult struct {
	Schema      string          `json:"schema"`
	MessageType string          `json:"message_type"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Decision    DecisionPayload `json:"decision"`
	Guidance    GuidancePayload `json:"guidance"`
}

// DecisionPayload contains the approval state.
type DecisionPayload struct {
	Outcome Outcome `json:"outcome"`
}

// GuidancePayload contains actionable direction for the caller.
type GuidancePayload struct {
	Category    GuidanceCategory `json:"category"`
	Description string           `json:"description"`
	AutoFixable bool             `json:"auto_fixable"`
}

// Outcome expresses the final guardian decision.
type Outcome string

const (
	// OutcomeApprove indicates the artifact is accepted.
	OutcomeApprove Outcome = "APPROVE"
	// OutcomeReject indicates the artifact must be changed.
	OutcomeReject Outcome = "REJECT"
)

// GuidanceCategory specifies the reasoning lane for the guidance block.
type GuidanceCategory string

const (
	GuidanceCategoryStructural  GuidanceCategory = "structural"
	GuidanceCategoryCorrectness GuidanceCategory = "correctness"
	GuidanceCategoryPolicy      GuidanceCategory = "policy"
)

// Stub implements the Guardian interface with deterministic, in-memory rules.
type Stub struct{}

// NewStub creates a Guardian stub suitable for local runtime testing.
func NewStub() *Stub {
	return &Stub{}
}

// EvaluateCompliance applies deterministic guardrails to the provided artifact.
func (s *Stub) EvaluateCompliance(msg ComplianceCheck) (DecisionResult, error) {
	content := strings.TrimSpace(msg.Artifact.Content)
	contentLength := utf8.RuneCountInString(content)

	switch {
	case containsPolicyViolation(content):
		return newDecisionResult(
			OutcomeReject,
			GuidanceCategoryPolicy,
			"Artifact contains explicit [POLICY_VIOLATION] marker.",
			false,
		), nil
	case contentLength < 200:
		return newDecisionResult(
			OutcomeReject,
			GuidanceCategoryStructural,
			"Artifact content is shorter than the 200 character minimum.",
			true,
		), nil
	case !hasWhatVsEnablesDistinction(content):
		return newDecisionResult(
			OutcomeReject,
			GuidanceCategoryCorrectness,
			"Artifact must clearly explain what PlexiMesh is and what it enables.",
			true,
		), nil
	default:
		return newDecisionResult(
			OutcomeApprove,
			GuidanceCategoryCorrectness,
			"Artifact satisfies stub Guardian requirements.",
			false,
		), nil
	}
}

func newDecisionResult(outcome Outcome, category GuidanceCategory, description string, autoFixable bool) DecisionResult {
	return DecisionResult{
		Schema:      decisionSchema,
		MessageType: decisionMessageType,
		From:        guardianActor,
		To:          maestroActor,
		Decision: DecisionPayload{
			Outcome: outcome,
		},
		Guidance: GuidancePayload{
			Category:    category,
			Description: description,
			AutoFixable: autoFixable,
		},
	}
}

func containsPolicyViolation(content string) bool {
	return strings.Contains(content, "[POLICY_VIOLATION]")
}

func hasWhatVsEnablesDistinction(content string) bool {
	lower := strings.ToLower(content)
	hasWhat := strings.Contains(lower, "what pleximesh is")
	hasEnables := strings.Contains(lower, "what it enables") || strings.Contains(lower, "what pleximesh enables")
	return hasWhat && hasEnables
}
