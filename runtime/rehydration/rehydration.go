package rehydration

import (
	"fmt"
	"sort"
	"strings"
)

// Rehydrate deterministically derives run state, decision ledger, and integrity report.
// It is read-only and fails closed on fatal invariant violations.
func Rehydrate(runID string, events []Event, artifacts Artifacts) (RunState, DecisionLedger, IntegrityReport, error) {
	state := RunState{
		RunID:            runID,
		StartedSequence:  -1,
		TerminalSequence: -1,
	}
	ledger := DecisionLedger{RunID: runID}
	report := IntegrityReport{RunID: runID}

	if isBlank(runID) {
		addFatal(&report, Issue{Code: "run_id_missing", Message: "run_id is required", Sequence: -1})
	}
	validateArtifacts(artifacts, &report)

	if len(events) == 0 {
		addFatal(&report, Issue{Code: "event_stream_empty", Message: "event stream is empty", Sequence: -1})
		finalizeReport(&report)
		return state, ledger, report, reportError(report)
	}

	ordered := append([]Event(nil), events...)
	sortEvents(ordered)

	state.LastSequence = ordered[len(ordered)-1].Sequence

	eventTypes := make(map[string]bool, len(ordered))
	var (
		seenStarted        bool
		seenCompleted      bool
		seenAborted        bool
		firstTerminalSeq   = -1
		firstTerminalType  string
		firstGuardianSeq   = -1
		firstEngineerSeq   = -1
		firstAbortSeq      = -1
		guardianRejectSeen bool
	)

	for i, evt := range ordered {
		validateEvent(evt, runID, &report)
		eventTypes[evt.Type] = true

		if i == 0 && evt.Type != EventRunStarted {
			addFatal(&report, Issue{
				Code:      "run_started_not_first",
				Message:   "first event must be system.run.started",
				Sequence:  evt.Sequence,
				EventType: evt.Type,
			})
		}

		if evt.Sequence < 0 {
			addFatal(&report, Issue{
				Code:      "sequence_negative",
				Message:   "sequence must be zero or greater",
				Sequence:  evt.Sequence,
				EventType: evt.Type,
			})
		}

		if i > 0 {
			prev := ordered[i-1].Sequence
			if evt.Sequence == prev {
				addFatal(&report, Issue{
					Code:      "sequence_duplicate",
					Message:   fmt.Sprintf("duplicate sequence %d", evt.Sequence),
					Sequence:  evt.Sequence,
					EventType: evt.Type,
				})
			} else if evt.Sequence != prev+1 {
				addFatal(&report, Issue{
					Code:      "sequence_gap",
					Message:   fmt.Sprintf("sequence gap between %d and %d", prev, evt.Sequence),
					Sequence:  evt.Sequence,
					EventType: evt.Type,
				})
			}
		}

		switch evt.Type {
		case EventRunStarted:
			seenStarted = true
			if state.StartedSequence == -1 {
				state.StartedSequence = evt.Sequence
				state.StartedAt = evt.Timestamp
			}
			state.Lifecycle = "started"
		case EventRunCompleted:
			seenCompleted = true
			if firstTerminalSeq == -1 {
				firstTerminalSeq = evt.Sequence
				firstTerminalType = EventRunCompleted
				state.TerminalSequence = evt.Sequence
				state.TerminalAt = evt.Timestamp
				state.TerminalOutcome = "completed"
				state.Lifecycle = "completed"
			}
		case EventRunAborted:
			seenAborted = true
			if firstAbortSeq == -1 {
				firstAbortSeq = evt.Sequence
			}
			if firstTerminalSeq == -1 {
				firstTerminalSeq = evt.Sequence
				firstTerminalType = EventRunAborted
				state.TerminalSequence = evt.Sequence
				state.TerminalAt = evt.Timestamp
				state.TerminalOutcome = "aborted"
				state.Lifecycle = "aborted"
			}
		}

		if isEngineerEvent(evt) && firstEngineerSeq == -1 {
			firstEngineerSeq = evt.Sequence
		}

		if isGuardianDecisionEvent(evt) {
			outcome := normalizeOutcome(evt.Decision.Outcome)
			if outcome == "" {
				addFatal(&report, Issue{
					Code:      "guardian_outcome_missing",
					Message:   "guardian decision outcome is required",
					Sequence:  evt.Sequence,
					EventType: evt.Type,
				})
			} else {
				if firstGuardianSeq == -1 {
					firstGuardianSeq = evt.Sequence
				}
				if outcome == GuardianOutcomeReject {
					guardianRejectSeen = true
				}
				ledger.Decisions = append(ledger.Decisions, DecisionEntry{
					Sequence:  evt.Sequence,
					Outcome:   outcome,
					Timestamp: evt.Timestamp,
					Evidence:  decisionEvidence(evt),
					EventType: evt.Type,
					Emitter:   evt.Emitter,
				})
			}
		}
	}

	if !seenStarted {
		addFatal(&report, Issue{Code: "run_started_missing", Message: "missing system.run.started", Sequence: -1})
	}
	if !seenCompleted && !seenAborted {
		addFatal(&report, Issue{Code: "terminal_missing", Message: "missing terminal event", Sequence: -1})
	}
	if (seenCompleted && seenAborted) || (countTerminalEvents(ordered) > 1) {
		addFatal(&report, Issue{Code: "terminal_multiple", Message: "multiple terminal events present", Sequence: -1})
	}

	if firstTerminalSeq != -1 {
		if state.LastSequence > firstTerminalSeq {
			addWarning(&report, Issue{
				Code:      "post_terminal_events",
				Message:   fmt.Sprintf("events occur after terminal at sequence %d", firstTerminalSeq),
				Sequence:  firstTerminalSeq,
				EventType: firstTerminalType,
			})
		}
	}

	if guardianRejectSeen && !seenAborted {
		addFatal(&report, Issue{
			Code:      "guardian_reject_requires_abort",
			Message:   "guardian REJECT requires system.run.aborted",
			Sequence:  -1,
			EventType: EventRunAborted,
		})
	}

	if firstEngineerSeq != -1 {
		if firstGuardianSeq == -1 || firstEngineerSeq < firstGuardianSeq {
			if firstAbortSeq == -1 || firstAbortSeq > firstEngineerSeq {
				addFatal(&report, Issue{
					Code:      "engineer_before_guardian",
					Message:   "engineer event observed before guardian PASS/REJECT without abort",
					Sequence:  firstEngineerSeq,
					EventType: EngineerRole,
				})
			}
		}
	}

	addOptionalEventWarnings(eventTypes, artifacts.OptionalEventTypes, &report)

	finalizeReport(&report)
	return state, ledger, report, reportError(report)
}

func validateEvent(evt Event, runID string, report *IntegrityReport) {
	if isBlank(evt.RunID) {
		addFatal(report, Issue{Code: "event_run_id_missing", Message: "event run_id is required", Sequence: evt.Sequence, EventType: evt.Type})
	} else if evt.RunID != runID {
		addFatal(report, Issue{
			Code:      "event_run_id_mismatch",
			Message:   fmt.Sprintf("event run_id %q does not match %q", evt.RunID, runID),
			Sequence:  evt.Sequence,
			EventType: evt.Type,
		})
	}
	if isBlank(evt.Type) {
		addFatal(report, Issue{Code: "event_type_missing", Message: "event type is required", Sequence: evt.Sequence, EventType: evt.Type})
	}
	if isBlank(evt.SchemaVersion) {
		addFatal(report, Issue{Code: "schema_version_missing", Message: "schema_version is required", Sequence: evt.Sequence, EventType: evt.Type})
	}
	if evt.Timestamp.IsZero() {
		addFatal(report, Issue{Code: "timestamp_missing", Message: "timestamp is required", Sequence: evt.Sequence, EventType: evt.Type})
	}
	if isBlank(evt.Emitter.AgentID) || isBlank(evt.Emitter.Role) {
		addFatal(report, Issue{Code: "emitter_missing", Message: "emitter agent_id and role are required", Sequence: evt.Sequence, EventType: evt.Type})
	}
	if isLifecycleEvent(evt.Type) && evt.Emitter.Role != SystemRole {
		addFatal(report, Issue{
			Code:      "lifecycle_emitter_invalid",
			Message:   "lifecycle events must be emitted by system",
			Sequence:  evt.Sequence,
			EventType: evt.Type,
		})
	}
	if evt.Type == EventGuardianDecision && evt.Decision == nil {
		addFatal(report, Issue{
			Code:      "guardian_decision_missing",
			Message:   "guardian decision payload is required",
			Sequence:  evt.Sequence,
			EventType: evt.Type,
		})
	}
}

func validateArtifacts(artifacts Artifacts, report *IntegrityReport) {
	if !hasNonBlank(artifacts.Laws) {
		addFatal(report, Issue{Code: "laws_missing", Message: "laws artifact reference required", Sequence: -1})
	}
	if !hasNonBlank(artifacts.Charters) {
		addFatal(report, Issue{Code: "charters_missing", Message: "charters artifact reference required", Sequence: -1})
	}
	if hasBlank(artifacts.Prompts) {
		addWarning(report, Issue{Code: "prompts_reference_missing", Message: "prompt artifact reference missing", Sequence: -1})
	}
	if hasBlank(artifacts.OptionalReferences) {
		addWarning(report, Issue{Code: "optional_reference_missing", Message: "optional artifact reference missing", Sequence: -1})
	}
}

func addOptionalEventWarnings(present map[string]bool, optional []string, report *IntegrityReport) {
	if len(optional) == 0 {
		return
	}
	unique := uniqueSorted(optional)
	for _, eventType := range unique {
		if !present[eventType] {
			addWarning(report, Issue{
				Code:      "optional_event_missing",
				Message:   fmt.Sprintf("optional event %q missing", eventType),
				Sequence:  -1,
				EventType: eventType,
			})
		}
	}
}

func sortEvents(events []Event) {
	sort.SliceStable(events, func(i, j int) bool {
		a := events[i]
		b := events[j]
		if a.Sequence != b.Sequence {
			return a.Sequence < b.Sequence
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.RunID != b.RunID {
			return a.RunID < b.RunID
		}
		if a.Emitter.Role != b.Emitter.Role {
			return a.Emitter.Role < b.Emitter.Role
		}
		if a.Emitter.AgentID != b.Emitter.AgentID {
			return a.Emitter.AgentID < b.Emitter.AgentID
		}
		if a.Timestamp.Before(b.Timestamp) {
			return true
		}
		if b.Timestamp.Before(a.Timestamp) {
			return false
		}
		return a.SchemaVersion < b.SchemaVersion
	})
}

func isLifecycleEvent(eventType string) bool {
	return eventType == EventRunStarted || eventType == EventRunCompleted || eventType == EventRunAborted
}

func isEngineerEvent(evt Event) bool {
	return evt.Emitter.Role == EngineerRole
}

func isGuardianDecisionEvent(evt Event) bool {
	if evt.Decision == nil {
		return false
	}
	if evt.Type == EventGuardianDecision {
		return true
	}
	return evt.Emitter.Role == GuardianRole
}

func normalizeOutcome(outcome string) string {
	upper := strings.ToUpper(strings.TrimSpace(outcome))
	if upper == GuardianOutcomePass || upper == GuardianOutcomeReject {
		return upper
	}
	return ""
}

func decisionEvidence(evt Event) string {
	if evt.Decision == nil {
		return ""
	}
	if !isBlank(evt.Decision.Evidence) {
		return evt.Decision.Evidence
	}
	return fmt.Sprintf("event:%s:%d", evt.Type, evt.Sequence)
}

func countTerminalEvents(events []Event) int {
	count := 0
	for _, evt := range events {
		if evt.Type == EventRunCompleted || evt.Type == EventRunAborted {
			count++
		}
	}
	return count
}

func finalizeReport(report *IntegrityReport) {
	sortIssues(report.Fatal)
	sortIssues(report.Warnings)
	report.FatalCount = len(report.Fatal)
	report.WarningCount = len(report.Warnings)
	report.Summary = fmt.Sprintf("fatal=%d warnings=%d", report.FatalCount, report.WarningCount)
}

func sortIssues(issues []Issue) {
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Sequence != issues[j].Sequence {
			return issues[i].Sequence < issues[j].Sequence
		}
		if issues[i].Code != issues[j].Code {
			return issues[i].Code < issues[j].Code
		}
		return issues[i].EventType < issues[j].EventType
	})
}

func addFatal(report *IntegrityReport, issue Issue) {
	report.Fatal = append(report.Fatal, issue)
}

func addWarning(report *IntegrityReport, issue Issue) {
	report.Warnings = append(report.Warnings, issue)
}

func reportError(report IntegrityReport) error {
	if report.FatalCount == 0 {
		return nil
	}
	return fmt.Errorf("rehydration failed with %d fatal issue(s)", report.FatalCount)
}

func hasNonBlank(values []string) bool {
	for _, value := range values {
		if !isBlank(value) {
			return true
		}
	}
	return false
}

func hasBlank(values []string) bool {
	for _, value := range values {
		if isBlank(value) {
			return true
		}
	}
	return false
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		seen[value] = true
	}
	unique := make([]string, 0, len(seen))
	for value := range seen {
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}
