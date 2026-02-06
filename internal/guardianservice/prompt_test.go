package guardianservice

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBuildPromptDeterministic(t *testing.T) {
	snapshot := InputArtifact{
		ArtifactID: "snapshot-v2",
		Version:    "v2",
		Kind:       "snapshot",
		Domain:     "runtime",
		Content:    "snapshot v2 content",
		Sha256:     sha256Hex("snapshot v2 content"),
	}
	phase := InputArtifact{
		ArtifactID: "phase-artifact-v2",
		Version:    "v2",
		Kind:       "phase",
		Domain:     "runtime",
		Content:    "phase artifact content",
		Sha256:     sha256Hex("phase artifact content"),
	}
	laws := InputArtifact{
		ArtifactID: "guardian-laws-v1",
		Version:    "v1",
		Kind:       "laws",
		Domain:     "policy",
		Content:    "phase close law v1",
		Sha256:     sha256Hex("phase close law v1"),
	}
	params := json.RawMessage(`{"phase":"close","attempt":1}`)

	got, err := BuildPrompt("bundle", []InputArtifact{snapshot, phase, laws}, params)
	if err != nil {
		t.Fatalf("BuildPrompt error: %v", err)
	}

	want := fmt.Sprintf(
		"GUARDIAN_SERVICE_PROPOSAL_ONLY\nOperation: bundle\nNo mutation constraints:\n%s\n%s\n\nInputs (ordered):\n"+
			"1) %s\n   sha256: %s\n   version: %s\n   kind: %s\n   domain: %s\n   content:\n%s\n"+
			"2) %s\n   sha256: %s\n   version: %s\n   kind: %s\n   domain: %s\n   content:\n%s\n"+
			"3) %s\n   sha256: %s\n   version: %s\n   kind: %s\n   domain: %s\n   content:\n%s\n"+
			"Params:\n{\"attempt\":1,\"phase\":\"close\"}\n",
		noMutationConstraints,
		approvalConstraints,
		snapshot.ArtifactID,
		snapshot.Sha256,
		snapshot.Version,
		snapshot.Kind,
		snapshot.Domain,
		snapshot.Content,
		phase.ArtifactID,
		phase.Sha256,
		phase.Version,
		phase.Kind,
		phase.Domain,
		phase.Content,
		laws.ArtifactID,
		laws.Sha256,
		laws.Version,
		laws.Kind,
		laws.Domain,
		laws.Content,
	)

	if got != want {
		t.Fatalf("prompt mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
