package guardianservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ErrProposalEngineUnavailable = "proposal_engine_unavailable"
const defaultAppServerEndpoint = "/v1/proposals"

type ProposalInput struct {
	Operation string          `json:"operation"`
	Prompt    string          `json:"prompt"`
	Inputs    []InputArtifact `json:"inputs"`
	Params    json.RawMessage `json:"params,omitempty"`
}

type ProposalEngine interface {
	Propose(ctx context.Context, input ProposalInput) (ProposalResponse, error)
}

type ProposalError struct {
	Code    string
	Message string
}

func (e ProposalError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

type UnavailableEngine struct {
	Reason string
}

func (e UnavailableEngine) Propose(ctx context.Context, input ProposalInput) (ProposalResponse, error) {
	reason := e.Reason
	if reason == "" {
		reason = "proposal engine unavailable"
	}
	return ProposalResponse{}, ProposalError{Code: ErrProposalEngineUnavailable, Message: reason}
}

type AppServerEngine struct {
	baseURL  string
	token    string
	endpoint string
	client   *http.Client
}

func NewAppServerEngine(baseURL, token, endpoint string) *AppServerEngine {
	if endpoint == "" {
		endpoint = defaultAppServerEndpoint
	}
	return &AppServerEngine{
		baseURL:  baseURL,
		token:    token,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

type appServerRequest struct {
	Operation string          `json:"operation"`
	Prompt    string          `json:"prompt"`
	Inputs    []InputArtifact `json:"inputs"`
	Params    json.RawMessage `json:"params,omitempty"`
}

func (e *AppServerEngine) Propose(ctx context.Context, input ProposalInput) (ProposalResponse, error) {
	if e == nil || strings.TrimSpace(e.baseURL) == "" {
		return ProposalResponse{}, ProposalError{Code: ErrProposalEngineUnavailable, Message: "CODEX_APP_SERVER_URL not set"}
	}

	payload := appServerRequest{
		Operation: input.Operation,
		Prompt:    input.Prompt,
		Inputs:    input.Inputs,
		Params:    input.Params,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return ProposalResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url(), bytes.NewReader(body))
	if err != nil {
		return ProposalResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(e.token) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(e.token))
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return ProposalResponse{}, ProposalError{Code: ErrProposalEngineUnavailable, Message: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProposalResponse{}, ProposalError{
			Code:    ErrProposalEngineUnavailable,
			Message: fmt.Sprintf("app-server status %d", resp.StatusCode),
		}
	}

	var out ProposalResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return ProposalResponse{}, err
	}
	return out, nil
}

func (e *AppServerEngine) url() string {
	base := strings.TrimRight(e.baseURL, "/")
	endpoint := e.endpoint
	if endpoint == "" {
		endpoint = defaultAppServerEndpoint
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return base + endpoint
}
