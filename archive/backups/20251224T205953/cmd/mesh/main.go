//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	runtimepkg "github.com/chuxorg/chux-agent-mesh/internal/runtime"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	instrumentor := newLibrarianInstrumentor(logger)
	commandLine := strings.Join(os.Args, " ")
	start := time.Now().UTC()

	if instrumentor.Enabled() {
		promptBody := buildPromptBody(instrumentor, commandLine, start)
		instrumentor.SubmitPrompt(ctx, promptBody)
	}

	rt := runtimepkg.New(runtimepkg.Config{
		Logger:  logger,
		DataDir: "data",
	})

	runErr := rt.Run(ctx)

	resultSummary := "runtime execution completed successfully"
	resultStatus := "success"
	if runErr != nil {
		resultSummary = fmt.Sprintf("runtime execution failed: %v", runErr)
		resultStatus = "failure"
	}

	if instrumentor.Enabled() {
		end := time.Now().UTC()
		resultBody := buildResultBody(resultSummary, start, end, commandLine, instrumentor)
		details := map[string]string{
			"command":              commandLine,
			"root_execution_id":    instrumentor.RootExecutionID(),
			"entrypoint_execution": instrumentor.ExecutionID(),
			"result_status":        resultStatus,
		}
		logs := []string{
			fmt.Sprintf("execution started: %s", start.Format(time.RFC3339Nano)),
			fmt.Sprintf("execution finished: %s", end.Format(time.RFC3339Nano)),
		}
		instrumentor.SubmitResult(ctx, resultBody, details, logs)
	}

	if runErr != nil {
		logger.Error("runtime execution failed", "error", runErr)
		os.Exit(1)
	}
	logger.Info("runtime execution completed")
}

func buildPromptBody(inst *librarianInstrumentor, commandLine string, ts time.Time) string {
	var b strings.Builder
	b.WriteString("# Mesh Runtime Execution Prompt\n\n")
	b.WriteString(fmt.Sprintf("- Role: %s\n", inst.AgentRole()))
	b.WriteString(fmt.Sprintf("- Root Execution ID: %s\n", inst.RootExecutionID()))
	b.WriteString(fmt.Sprintf("- Entrypoint Execution ID: %s\n", inst.ExecutionID()))
	if wd, err := os.Getwd(); err == nil {
		b.WriteString(fmt.Sprintf("- Working Directory: `%s`\n", wd))
	}
	b.WriteString(fmt.Sprintf("- Command: `%s`\n", commandLine))
	b.WriteString(fmt.Sprintf("- Timestamp: %s\n\n", ts.Format(time.RFC3339Nano)))
	b.WriteString("Runtime entrypoint invoked. Execute the linked agents and capture the resulting logs as evidence for Librarian.")
	return b.String()
}

func buildResultBody(summary string, started, finished time.Time, commandLine string, inst *librarianInstrumentor) string {
	var b strings.Builder
	b.WriteString("# Mesh Runtime Execution Result\n\n")
	b.WriteString(summary)
	b.WriteString("\n\n## Details\n")
	b.WriteString(fmt.Sprintf("- Role: %s\n", inst.AgentRole()))
	b.WriteString(fmt.Sprintf("- Root Execution ID: %s\n", inst.RootExecutionID()))
	b.WriteString(fmt.Sprintf("- Entrypoint Execution ID: %s\n", inst.ExecutionID()))
	b.WriteString(fmt.Sprintf("- Command: `%s`\n", commandLine))
	b.WriteString(fmt.Sprintf("- Started At: %s\n", started.Format(time.RFC3339Nano)))
	b.WriteString(fmt.Sprintf("- Finished At: %s\n", finished.Format(time.RFC3339Nano)))
	return b.String()
}

type librarianInstrumentor struct {
	logger          *slog.Logger
	client          *http.Client
	endpoint        string
	agentRole       string
	agentID         string
	rootExecutionID string
	executionID     string
	promptID        string
	enabled         bool
}

func newLibrarianInstrumentor(logger *slog.Logger) *librarianInstrumentor {
	role := normalizedRole(firstNonEmpty(os.Getenv("AGENT_ROLE"), os.Getenv("ROLE"), "engineer"))
	if role == "librarian" || role == "sandbox" {
		logger.Info("librarian.instrumentation_skipped", "reason", "role_excluded", "role", role)
		return &librarianInstrumentor{enabled: false}
	}
	if isSandboxExecution() {
		logger.Info("librarian.instrumentation_skipped", "reason", "sandbox_execution")
		return &librarianInstrumentor{enabled: false}
	}
	rawURL := normalizeLibrarianURL(os.Getenv("LIBRARIAN_URL"))
	if rawURL == "" {
		logger.Info("librarian.instrumentation_skipped", "reason", "missing_librarian_url")
		return &librarianInstrumentor{enabled: false}
	}
	rootID := os.Getenv("ROOT_EXECUTION_ID")
	if rootID == "" {
		rootID = fmt.Sprintf("root-%d", time.Now().UTC().UnixNano())
	}
	execID := fmt.Sprintf("%s-entrypoint", rootID)

	return &librarianInstrumentor{
		logger:          logger,
		client:          &http.Client{Timeout: 5 * time.Second},
		endpoint:        rawURL,
		agentRole:       role,
		agentID:         firstNonEmpty(os.Getenv("AGENT_ID"), os.Getenv("USER"), "mesh-entrypoint"),
		rootExecutionID: rootID,
		executionID:     execID,
		enabled:         true,
	}
}

func (l *librarianInstrumentor) Enabled() bool {
	return l != nil && l.enabled
}

func (l *librarianInstrumentor) RootExecutionID() string {
	if l == nil {
		return ""
	}
	return l.rootExecutionID
}

func (l *librarianInstrumentor) ExecutionID() string {
	if l == nil {
		return ""
	}
	return l.executionID
}

func (l *librarianInstrumentor) AgentRole() string {
	if l == nil {
		return ""
	}
	return l.agentRole
}

func (l *librarianInstrumentor) SubmitPrompt(ctx context.Context, body string) {
	if !l.Enabled() {
		return
	}
	payload := artifactRequest{
		ArtifactType: "prompt",
		Agent:        artifactAgent{Role: l.agentRole, ID: l.agentID},
		Execution: artifactExecution{
			ExecutionID: l.executionID,
			Phase:       "runtime",
		},
		Correlation: artifactCorrelation{
			RootExecutionID: l.rootExecutionID,
		},
		Source: artifactSource{
			Repo:   "github.com/chuxorg/chux-agent-mesh",
			Branch: os.Getenv("GIT_BRANCH"),
			Commit: os.Getenv("GIT_COMMIT"),
		},
		Content: artifactContent{
			Format: "markdown",
			Body:   body,
		},
		Metadata: artifactMetadata{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Tags:      []string{"runtime-entrypoint", "prompt"},
			Notes:     "Recorded via cmd/mesh harness",
		},
	}
	resp, err := l.postArtifact(ctx, payload)
	if err != nil {
		l.logger.Warn("librarian.prompt_submission_failed", "error", err)
		l.ensurePromptFallback()
		return
	}
	if resp != nil && resp.PromptID != "" {
		l.promptID = resp.PromptID
	} else {
		l.ensurePromptFallback()
	}
}

func (l *librarianInstrumentor) SubmitResult(ctx context.Context, body string, details map[string]string, logs []string) {
	if !l.Enabled() {
		return
	}
	promptID := l.promptID
	if promptID == "" {
		promptID = fmt.Sprintf("prompt-%s", l.executionID)
	}
	payload := artifactRequest{
		ArtifactType: "result",
		Agent:        artifactAgent{Role: l.agentRole, ID: l.agentID},
		Execution: artifactExecution{
			ExecutionID: l.executionID,
			Phase:       "runtime",
		},
		Correlation: artifactCorrelation{
			RootExecutionID: l.rootExecutionID,
			PromptID:        promptID,
		},
		Source: artifactSource{
			Repo:   "github.com/chuxorg/chux-agent-mesh",
			Branch: os.Getenv("GIT_BRANCH"),
			Commit: os.Getenv("GIT_COMMIT"),
		},
		Content: artifactContent{
			Format: "markdown",
			Body:   body + "\n\n## Execution Details\n" + formatDetails(details) + formatLogs(logs),
		},
		Metadata: artifactMetadata{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Tags:      []string{"runtime-entrypoint", "result"},
			Notes:     "Runtime execution outcome",
			AgentID:   l.agentID,
		},
	}
	if _, err := l.postArtifact(ctx, payload); err != nil {
		l.logger.Warn("librarian.result_submission_failed", "error", err)
	}
}

func (l *librarianInstrumentor) ensurePromptFallback() {
	if l.promptID == "" {
		l.promptID = fmt.Sprintf("prompt-%s", l.executionID)
	}
}

func (l *librarianInstrumentor) postArtifact(ctx context.Context, payload artifactRequest) (*artifactResponse, error) {
	if !l.Enabled() {
		return nil, nil
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, l.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("librarian returned status %d: %s", resp.StatusCode, strings.TrimSpace(buf.String()))
	}
	var out artifactResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

type artifactRequest struct {
	ArtifactType string              `json:"artifact_type"`
	Agent        artifactAgent       `json:"agent"`
	Execution    artifactExecution   `json:"execution"`
	Correlation  artifactCorrelation `json:"correlation"`
	Source       artifactSource      `json:"source"`
	Content      artifactContent     `json:"content"`
	Metadata     artifactMetadata    `json:"metadata"`
}

type artifactAgent struct {
	Role string `json:"role"`
	ID   string `json:"id,omitempty"`
}

type artifactExecution struct {
	ExecutionID       string `json:"execution_id"`
	ParentExecutionID string `json:"parent_execution_id,omitempty"`
	Phase             string `json:"phase,omitempty"`
}

type artifactCorrelation struct {
	RootExecutionID string `json:"root_execution_id"`
	PromptID        string `json:"prompt_id,omitempty"`
}

type artifactSource struct {
	Repo   string `json:"repo,omitempty"`
	Branch string `json:"branch,omitempty"`
	Commit string `json:"commit,omitempty"`
}

type artifactContent struct {
	Format string `json:"format"`
	Body   string `json:"body"`
}

type artifactMetadata struct {
	Timestamp string   `json:"timestamp,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Notes     string   `json:"notes,omitempty"`
	AgentID   string   `json:"agent_id,omitempty"`
}

type artifactResponse struct {
	PromptID string `json:"prompt_id"`
}

func normalizeLibrarianURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimRight(trimmed, "/")
	if strings.HasSuffix(trimmed, "/artifacts") {
		return trimmed
	}
	return trimmed + "/artifacts"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func normalizedRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func isSandboxExecution() bool {
	for _, key := range []string{"SANDBOX_RUN", "SANDBOX", "SANDBOX_MODE"} {
		val := normalizedRole(os.Getenv(key))
		if val == "true" || val == "1" || val == "yes" {
			return true
		}
	}
	return false
}

func formatDetails(details map[string]string) string {
	if len(details) == 0 {
		return ""
	}
	var b strings.Builder
	keys := make([]string, 0, len(details))
	for key := range details {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		b.WriteString(fmt.Sprintf("- %s: %s\n", key, details[key]))
	}
	return b.String()
}

func formatLogs(logs []string) string {
	if len(logs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n## Timeline\n")
	for _, entry := range logs {
		b.WriteString(fmt.Sprintf("- %s\n", entry))
	}
	return b.String()
}
