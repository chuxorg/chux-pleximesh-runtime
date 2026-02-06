package guardianservice

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxBodyBytes = 4 << 20

type Service struct {
	engine ProposalEngine
}

type namedInput struct {
	name     string
	artifact InputArtifact
}

func NewService(engine ProposalEngine) *Service {
	if engine == nil {
		engine = UnavailableEngine{Reason: "proposal engine not configured"}
	}
	return &Service{engine: engine}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/guardian/bundle", s.handleBundle)
	mux.HandleFunc("/guardian/propose-phase-close", s.handleProposePhaseClose)
	mux.HandleFunc("/guardian/diff-snapshots", s.handleDiffSnapshots)
	return mux
}

func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method_not_allowed"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true,
		"ts": time.Now().UTC(),
	})
}

func (s *Service) handleBundle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method_not_allowed"})
		return
	}
	var req BundleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_json", Message: err.Error()})
		return
	}
	inputs := []namedInput{
		{name: "snapshot", artifact: req.Snapshot},
		{name: "phase_artifact", artifact: req.PhaseArtifact},
		{name: "guardian_laws", artifact: req.GuardianLaws},
	}
	s.handleProposal(w, r, "bundle", inputs, req.Params)
}

func (s *Service) handleProposePhaseClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method_not_allowed"})
		return
	}
	var req ProposePhaseCloseRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_json", Message: err.Error()})
		return
	}
	inputs := []namedInput{
		{name: "snapshot", artifact: req.Snapshot},
		{name: "phase_artifact", artifact: req.PhaseArtifact},
		{name: "guardian_laws", artifact: req.GuardianLaws},
	}
	s.handleProposal(w, r, "propose-phase-close", inputs, req.Params)
}

func (s *Service) handleDiffSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method_not_allowed"})
		return
	}
	var req DiffSnapshotsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_json", Message: err.Error()})
		return
	}
	inputs := []namedInput{
		{name: "snapshot_base", artifact: req.SnapshotBase},
		{name: "snapshot_target", artifact: req.SnapshotTarget},
	}
	s.handleProposal(w, r, "diff-snapshots", inputs, req.Params)
}

func (s *Service) handleProposal(w http.ResponseWriter, r *http.Request, operation string, inputs []namedInput, params json.RawMessage) {
	missing, mismatched := validateInputs(inputs)
	if len(missing) > 0 {
		writeError(w, http.StatusBadRequest, ErrorResponse{Error: "missing_inputs", MissingInputs: missing})
		return
	}
	if len(mismatched) > 0 {
		writeError(w, http.StatusBadRequest, ErrorResponse{Error: "checksum_mismatch", MismatchedInputs: mismatched})
		return
	}

	ordered := make([]InputArtifact, 0, len(inputs))
	for _, item := range inputs {
		ordered = append(ordered, item.artifact)
	}

	prompt, err := BuildPrompt(operation, ordered, params)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorResponse{Error: "invalid_params", Message: err.Error()})
		return
	}

	resp, err := s.engine.Propose(r.Context(), ProposalInput{
		Operation: operation,
		Prompt:    prompt,
		Inputs:    ordered,
		Params:    params,
	})
	if err != nil {
		writeEngineError(w, err)
		return
	}
	resp = normalizeResponse(resp, ordered)
	writeJSON(w, http.StatusOK, resp)
}

func validateInputs(inputs []namedInput) ([]string, []string) {
	var missing []string
	var mismatched []string
	for _, item := range inputs {
		content := strings.TrimSpace(item.artifact.Content)
		sha := strings.TrimSpace(item.artifact.Sha256)
		id := strings.TrimSpace(item.artifact.ArtifactID)
		if content == "" || sha == "" || id == "" {
			missing = append(missing, item.name)
			continue
		}
		if !strings.EqualFold(sha256Hex(content), sha) {
			mismatched = append(mismatched, item.name)
		}
	}
	return missing, mismatched
}

func normalizeResponse(resp ProposalResponse, inputs []InputArtifact) ProposalResponse {
	if resp.ProposedArtifacts == nil {
		resp.ProposedArtifacts = []ProposedArtifact{}
	}
	if resp.ProposedPatches == nil {
		resp.ProposedPatches = []ProposedPatch{}
	}
	if resp.Plan == nil {
		resp.Plan = []string{}
	}
	if len(resp.Citations) == 0 {
		resp.Citations = buildCitations(inputs)
	}
	for i := range resp.ProposedArtifacts {
		if resp.ProposedArtifacts[i].Sha256 == "" && resp.ProposedArtifacts[i].Content != "" {
			resp.ProposedArtifacts[i].Sha256 = sha256Hex(resp.ProposedArtifacts[i].Content)
		}
	}
	return resp
}

func buildCitations(inputs []InputArtifact) []Citation {
	out := make([]Citation, 0, len(inputs))
	for _, input := range inputs {
		if input.ArtifactID == "" || input.Sha256 == "" {
			continue
		}
		out = append(out, Citation{
			InputArtifact: input.ArtifactID,
			Sha256:        input.Sha256,
		})
	}
	return out
}

func writeEngineError(w http.ResponseWriter, err error) {
	var proposalErr ProposalError
	if errors.As(err, &proposalErr) && proposalErr.Code == ErrProposalEngineUnavailable {
		writeError(w, http.StatusServiceUnavailable, ErrorResponse{
			Error:   proposalErr.Code,
			Message: proposalErr.Message,
		})
		return
	}
	writeError(w, http.StatusBadGateway, ErrorResponse{
		Error:   "proposal_engine_error",
		Message: err.Error(),
	})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("multiple json values")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

func writeError(w http.ResponseWriter, status int, resp ErrorResponse) {
	writeJSON(w, status, resp)
}

func sha256Hex(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
