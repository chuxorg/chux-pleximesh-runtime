package guardianservice

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type staticEngine struct {
	response  ProposalResponse
	err       error
	lastInput ProposalInput
}

func (s *staticEngine) Propose(ctx context.Context, input ProposalInput) (ProposalResponse, error) {
	s.lastInput = input
	if s.err != nil {
		return ProposalResponse{}, s.err
	}
	return s.response, nil
}

func validInput(id, content string) InputArtifact {
	return InputArtifact{
		ArtifactID: id,
		Content:    content,
		Sha256:     sha256Hex(content),
	}
}

func TestBundleMissingInputs(t *testing.T) {
	req := BundleRequest{
		Snapshot:      InputArtifact{},
		PhaseArtifact: validInput("phase-v2", "phase content"),
		GuardianLaws:  validInput("laws-v1", "laws content"),
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	service := NewService(&staticEngine{})
	rec := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/guardian/bundle", bytes.NewReader(body))
	service.Handler().ServeHTTP(rec, request)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "missing_inputs" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if !contains(resp.MissingInputs, "snapshot") {
		t.Fatalf("missing snapshot entry: %+v", resp.MissingInputs)
	}
}

func TestEngineUnavailable(t *testing.T) {
	req := BundleRequest{
		Snapshot:      validInput("snapshot-v2", "snapshot v2"),
		PhaseArtifact: validInput("phase-v2", "phase v2"),
		GuardianLaws:  validInput("laws-v1", "laws v1"),
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	service := NewService(UnavailableEngine{Reason: "down"})
	rec := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/guardian/bundle", bytes.NewReader(body))
	service.Handler().ServeHTTP(rec, request)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != ErrProposalEngineUnavailable {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}

func TestProposePhaseCloseGolden(t *testing.T) {
	engine := &staticEngine{
		response: ProposalResponse{
			ProposedArtifacts: []ProposedArtifact{
				{
					ArtifactID: "snapshot-v3",
					Version:    "v3",
					Kind:       "snapshot",
					Domain:     "runtime",
					Content:    "snapshot v3 placeholder",
				},
			},
			ProposedPatches: []ProposedPatch{
				{
					TargetPath: "seeds/phase-close.md",
					UnifiedDiff: "diff --git a/seeds/phase-close.md b/seeds/phase-close.md\n" +
						"--- a/seeds/phase-close.md\n+++ b/seeds/phase-close.md\n@@\n+seed delta\n",
				},
			},
			Plan: []string{
				"Generate snapshot v3 candidate",
				"Adjust seed deltas",
				"Archive phase artifacts",
			},
		},
	}

	req := ProposePhaseCloseRequest{
		Snapshot:      validInput("snapshot-v2", "snapshot v2"),
		PhaseArtifact: validInput("phase-v2", "phase v2"),
		GuardianLaws:  validInput("phase-close-law-v1", "law v1"),
		Params:        json.RawMessage(`{"phase":"close"}`),
	}

	body, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	service := NewService(engine)
	rec := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/guardian/propose-phase-close", bytes.NewReader(body))
	service.Handler().ServeHTTP(rec, request)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var resp ProposalResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.ProposedArtifacts) != 1 || resp.ProposedArtifacts[0].ArtifactID != "snapshot-v3" {
		t.Fatalf("unexpected artifacts: %+v", resp.ProposedArtifacts)
	}
	if resp.ProposedArtifacts[0].Sha256 == "" {
		t.Fatalf("expected sha256 for proposed artifact")
	}
	if !contains(resp.Plan, "Archive phase artifacts") {
		t.Fatalf("missing archive step: %+v", resp.Plan)
	}
	if !containsCitation(resp.Citations, "snapshot-v2") {
		t.Fatalf("missing citation: %+v", resp.Citations)
	}
	if engine.lastInput.Operation != "propose-phase-close" {
		t.Fatalf("unexpected operation: %s", engine.lastInput.Operation)
	}
	if !strings.Contains(engine.lastInput.Prompt, noMutationConstraints) {
		t.Fatalf("prompt missing constraints")
	}
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func containsCitation(items []Citation, artifactID string) bool {
	for _, item := range items {
		if item.InputArtifact == artifactID {
			return true
		}
	}
	return false
}
