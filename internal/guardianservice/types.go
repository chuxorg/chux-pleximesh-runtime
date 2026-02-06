package guardianservice

import "encoding/json"

type InputArtifact struct {
	ArtifactID string `json:"artifact_id"`
	Version    string `json:"version,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Domain     string `json:"domain,omitempty"`
	Content    string `json:"content"`
	Sha256     string `json:"sha256"`
}

type BundleRequest struct {
	Snapshot      InputArtifact   `json:"snapshot"`
	PhaseArtifact InputArtifact   `json:"phase_artifact"`
	GuardianLaws  InputArtifact   `json:"guardian_laws"`
	Params        json.RawMessage `json:"params,omitempty"`
}

type ProposePhaseCloseRequest struct {
	Snapshot      InputArtifact   `json:"snapshot"`
	PhaseArtifact InputArtifact   `json:"phase_artifact"`
	GuardianLaws  InputArtifact   `json:"guardian_laws"`
	Params        json.RawMessage `json:"params,omitempty"`
}

type DiffSnapshotsRequest struct {
	SnapshotBase   InputArtifact   `json:"snapshot_base"`
	SnapshotTarget InputArtifact   `json:"snapshot_target"`
	Params         json.RawMessage `json:"params,omitempty"`
}

type ProposedArtifact struct {
	ArtifactID string `json:"artifact_id"`
	Version    string `json:"version"`
	Kind       string `json:"kind"`
	Domain     string `json:"domain"`
	Content    string `json:"content"`
	Sha256     string `json:"sha256"`
}

type ProposedPatch struct {
	TargetPath  string `json:"target_path"`
	UnifiedDiff string `json:"unified_diff"`
}

type Citation struct {
	InputArtifact string `json:"input_artifact"`
	Sha256        string `json:"sha256"`
}

type ProposalResponse struct {
	ProposedArtifacts []ProposedArtifact `json:"proposed_artifacts"`
	ProposedPatches   []ProposedPatch    `json:"proposed_patches"`
	Plan              []string           `json:"plan"`
	Citations         []Citation         `json:"citations"`
}

type ErrorResponse struct {
	Error            string   `json:"error"`
	Message          string   `json:"message,omitempty"`
	MissingInputs    []string `json:"missing_inputs,omitempty"`
	MismatchedInputs []string `json:"mismatched_inputs,omitempty"`
}
