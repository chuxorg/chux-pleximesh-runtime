package guardianservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const noMutationConstraints = "The Guardian Service MUST be proposal-only: produces artifacts + diffs + plans; does NOT mutate Librarian; does NOT mutate repos; no silent side effects."
const approvalConstraints = "Approvals are required for any tool-call with side effects."

func BuildPrompt(operation string, inputs []InputArtifact, params json.RawMessage) (string, error) {
	normalizedParams, err := normalizeJSON(params)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString("GUARDIAN_SERVICE_PROPOSAL_ONLY\n")
	buf.WriteString(fmt.Sprintf("Operation: %s\n", operation))
	buf.WriteString("No mutation constraints:\n")
	buf.WriteString(noMutationConstraints + "\n")
	buf.WriteString(approvalConstraints + "\n\n")
	buf.WriteString("Inputs (ordered):\n")
	for i, input := range inputs {
		buf.WriteString(fmt.Sprintf("%d) %s\n", i+1, input.ArtifactID))
		buf.WriteString(fmt.Sprintf("   sha256: %s\n", input.Sha256))
		if strings.TrimSpace(input.Version) != "" {
			buf.WriteString(fmt.Sprintf("   version: %s\n", input.Version))
		}
		if strings.TrimSpace(input.Kind) != "" {
			buf.WriteString(fmt.Sprintf("   kind: %s\n", input.Kind))
		}
		if strings.TrimSpace(input.Domain) != "" {
			buf.WriteString(fmt.Sprintf("   domain: %s\n", input.Domain))
		}
		buf.WriteString("   content:\n")
		buf.WriteString(input.Content)
		buf.WriteString("\n")
	}
	buf.WriteString("Params:\n")
	buf.WriteString(normalizedParams)
	buf.WriteString("\n")
	return buf.String(), nil
}

func normalizeJSON(raw json.RawMessage) (string, error) {
	if len(raw) == 0 {
		return "null", nil
	}
	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	normalized, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}
