package rehydration

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestRehydrateFixtures(t *testing.T) {
	artifacts := fixtureArtifacts()

	t.Run("completed", func(t *testing.T) {
		events := loadFixture(t, "completed.jsonl")
		state, ledger, report, err := Rehydrate("run-completed-001", events, artifacts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if state.TerminalOutcome != "completed" {
			t.Fatalf("expected terminal outcome completed, got %q", state.TerminalOutcome)
		}
		if len(ledger.Decisions) == 0 || ledger.Decisions[0].Outcome != GuardianOutcomePass {
			t.Fatalf("expected guardian PASS decision, got %+v", ledger.Decisions)
		}
		if report.FatalCount != 0 || len(report.Fatal) != 0 {
			t.Fatalf("expected no fatal issues, got %+v", report.Fatal)
		}
		if report.WarningCount != 0 {
			t.Fatalf("expected no warnings, got %+v", report.Warnings)
		}
	})

	t.Run("aborted", func(t *testing.T) {
		events := loadFixture(t, "aborted.jsonl")
		state, ledger, report, err := Rehydrate("run-aborted-001", events, artifacts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if state.TerminalOutcome != "aborted" {
			t.Fatalf("expected terminal outcome aborted, got %q", state.TerminalOutcome)
		}
		if len(ledger.Decisions) == 0 || ledger.Decisions[0].Outcome != GuardianOutcomeReject {
			t.Fatalf("expected guardian REJECT decision, got %+v", ledger.Decisions)
		}
		if report.FatalCount != 0 || len(report.Fatal) != 0 {
			t.Fatalf("expected no fatal issues, got %+v", report.Fatal)
		}
		if report.WarningCount != 0 {
			t.Fatalf("expected no warnings, got %+v", report.Warnings)
		}
	})
}

func TestRehydrateDeterministic(t *testing.T) {
	events := loadFixture(t, "completed.jsonl")
	artifacts := fixtureArtifacts()

	stateA, ledgerA, reportA, errA := Rehydrate("run-completed-001", events, artifacts)
	stateB, ledgerB, reportB, errB := Rehydrate("run-completed-001", events, artifacts)

	if errA != nil || errB != nil {
		t.Fatalf("expected no errors, got %v and %v", errA, errB)
	}
	if !reflect.DeepEqual(stateA, stateB) {
		t.Fatalf("state mismatch: %+v vs %+v", stateA, stateB)
	}
	if !reflect.DeepEqual(ledgerA, ledgerB) {
		t.Fatalf("ledger mismatch: %+v vs %+v", ledgerA, ledgerB)
	}
	if !reflect.DeepEqual(reportA, reportB) {
		t.Fatalf("report mismatch: %+v vs %+v", reportA, reportB)
	}
}

func TestRehydrateDuplicateSequenceFatal(t *testing.T) {
	runID := "run-dup-001"
	base := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)
	events := []Event{
		makeEvent(1, runID, EventRunStarted, SystemRole, base),
		makeGuardianDecision(2, runID, GuardianOutcomePass, base.Add(1*time.Second)),
		makeEvent(2, runID, EventRunCompleted, SystemRole, base.Add(2*time.Second)),
	}

	_, _, report, err := Rehydrate(runID, events, fixtureArtifacts())
	if err == nil {
		t.Fatal("expected fatal error for duplicate sequence")
	}
	if !hasFatalCode(report, "sequence_duplicate") {
		t.Fatalf("expected sequence_duplicate fatal, got %+v", report.Fatal)
	}
}

func TestRehydrateEngineerBeforeGuardianFatal(t *testing.T) {
	runID := "run-early-engineer-001"
	base := time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC)
	events := []Event{
		makeEvent(1, runID, EventRunStarted, SystemRole, base),
		makeEvent(2, runID, "engineer.action", EngineerRole, base.Add(1*time.Second)),
		makeGuardianDecision(3, runID, GuardianOutcomePass, base.Add(2*time.Second)),
		makeEvent(4, runID, EventRunCompleted, SystemRole, base.Add(3*time.Second)),
	}

	_, _, report, err := Rehydrate(runID, events, fixtureArtifacts())
	if err == nil {
		t.Fatal("expected fatal error for engineer before guardian")
	}
	if !hasFatalCode(report, "engineer_before_guardian") {
		t.Fatalf("expected engineer_before_guardian fatal, got %+v", report.Fatal)
	}
}

func loadFixture(t *testing.T, name string) []Event {
	t.Helper()

	path := filepath.Join("fixtures", name)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		var evt Event
		if err := json.Unmarshal([]byte(text), &evt); err != nil {
			t.Fatalf("decode fixture line %d: %v", line, err)
		}
		events = append(events, evt)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return events
}

func fixtureArtifacts() Artifacts {
	return Artifacts{
		Laws:               []string{"initkit/laws.md"},
		Charters:           []string{"initkit/charter.md"},
		Prompts:            []string{"prompts/prompt.md"},
		OptionalReferences: []string{"references/optional.md"},
	}
}

func makeEvent(sequence int, runID, eventType, role string, timestamp time.Time) Event {
	return Event{
		Sequence:      sequence,
		RunID:         runID,
		Type:          eventType,
		SchemaVersion: "v1",
		Timestamp:     timestamp,
		Emitter: Emitter{
			AgentID: role + "-agent",
			Role:    role,
		},
	}
}

func makeGuardianDecision(sequence int, runID, outcome string, timestamp time.Time) Event {
	event := makeEvent(sequence, runID, EventGuardianDecision, GuardianRole, timestamp)
	event.Decision = &GuardianDecision{
		Outcome:  outcome,
		Evidence: "fixture decision",
	}
	return event
}

func hasFatalCode(report IntegrityReport, code string) bool {
	for _, issue := range report.Fatal {
		if issue.Code == code {
			return true
		}
	}
	return false
}
