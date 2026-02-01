//go:build ignore
// +build ignore

package librarian

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// ArtifactType identifies the librarian artifact variant.
type ArtifactType string

const (
	// ArtifactPrompt indicates an input prompt artifact.
	ArtifactPrompt ArtifactType = "prompt"
	// ArtifactResult indicates an execution result artifact.
	ArtifactResult ArtifactType = "result"
	defaultDir                  = "data/librarian"
)

// ExecutionHeaders capture the correlation context shared across artifacts.
type ExecutionHeaders struct {
	ExecutionID     string `json:"execution_id"`
	RootExecutionID string `json:"root_execution_id"`
	CorrelationID   string `json:"correlation_id"`
	AgentID         string `json:"agent_id"`
}

// PromptPayload is the structured body persisted for prompt artifacts.
type PromptPayload struct {
	Title        string            `json:"title"`
	Instructions string            `json:"instructions"`
	Inputs       map[string]string `json:"inputs,omitempty"`
}

// ResultPayload is the structured body persisted for result artifacts.
type ResultPayload struct {
	Outcome string            `json:"outcome"`
	Details map[string]string `json:"details,omitempty"`
	Logs    []string          `json:"logs,omitempty"`
}

// Writer persists prompt/result artifacts onto the mounted data directory.
type Writer struct {
	basePath string
	logger   *slog.Logger
	nowFunc  func() time.Time
}

// NewWriter constructs a writer rooted at basePath (defaults to ./data/librarian).
func NewWriter(basePath string, logger *slog.Logger) *Writer {
	if basePath == "" {
		basePath = defaultDir
	}
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return &Writer{
		basePath: basePath,
		logger:   logger,
		nowFunc:  time.Now,
	}
}

// SubmitPrompt writes a prompt artifact and returns the generated prompt ID.
func (w *Writer) SubmitPrompt(ctx context.Context, headers ExecutionHeaders, payload PromptPayload) (string, error) {
	promptID := fmt.Sprintf("prompt-%d", w.nowFunc().UnixNano())
	record := artifactRecord{
		Type:      ArtifactPrompt,
		PromptID:  promptID,
		Timestamp: w.nowFunc().UTC(),
		Headers:   headers,
		Payload:   payload,
	}
	err := w.persist(ctx, record)
	if err != nil {
		w.logger.Error("librarian.prompt.write_failed",
			"execution_id", headers.ExecutionID,
			"prompt_id", promptID,
			"error", err,
		)
		return "", err
	}
	w.logger.Info("librarian.prompt.recorded",
		"execution_id", headers.ExecutionID,
		"prompt_id", promptID,
		"path", w.artifactPath(record),
	)
	return promptID, nil
}

// SubmitResult writes a result artifact tied to the provided prompt ID.
func (w *Writer) SubmitResult(ctx context.Context, headers ExecutionHeaders, promptID string, payload ResultPayload) error {
	record := artifactRecord{
		Type:      ArtifactResult,
		PromptID:  promptID,
		Timestamp: w.nowFunc().UTC(),
		Headers:   headers,
		Payload:   payload,
	}
	if err := w.persist(ctx, record); err != nil {
		w.logger.Error("librarian.result.write_failed",
			"execution_id", headers.ExecutionID,
			"prompt_id", promptID,
			"error", err,
		)
		return err
	}
	w.logger.Info("librarian.result.recorded",
		"execution_id", headers.ExecutionID,
		"prompt_id", promptID,
		"path", w.artifactPath(record),
	)
	return nil
}

type artifactRecord struct {
	Type      ArtifactType     `json:"artifact_type"`
	PromptID  string           `json:"prompt_id"`
	Timestamp time.Time        `json:"timestamp"`
	Headers   ExecutionHeaders `json:"headers"`
	Payload   any              `json:"payload"`
}

func (w *Writer) persist(ctx context.Context, record artifactRecord) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	path := w.artifactPath(record)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(record)
}

func (w *Writer) artifactPath(record artifactRecord) string {
	execID := record.Headers.ExecutionID
	if execID == "" {
		execID = "unknown"
	}
	filename := fmt.Sprintf("%s-%s-%s.json", record.Timestamp.Format("20060102T150405Z0700"), record.Type, execID)
	return filepath.Join(w.basePath, filename)
}
