package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// ErrBrokerAuthenticationFailed is returned when the sandbox broker rejects credentials.
	ErrBrokerAuthenticationFailed = errors.New("sandbox broker authentication failed: invalid or missing token")
)

// RemoteBrokerEvaluator evaluates candidates via a dedicated, authenticated Sandbox Broker service.
type RemoteBrokerEvaluator struct {
	brokerURL  string
	token      string
	httpClient *http.Client
}

// NewRemoteBrokerEvaluator initializes an evaluator pointing to a remote sandbox broker.
func NewRemoteBrokerEvaluator(brokerURL, token string) *RemoteBrokerEvaluator {
	return &RemoteBrokerEvaluator{
		brokerURL: strings.TrimRight(brokerURL, "/"),
		token:     strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// CheckAvailable verifies whether the remote sandbox broker is reachable and healthy.
func (r *RemoteBrokerEvaluator) CheckAvailable(ctx context.Context) error {
	if r.brokerURL == "" {
		return fmt.Errorf("%w: SANDBOX_SERVICE_URL is empty", ErrSandboxUnavailable)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.brokerURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("%w: failed to create broker health request: %v", ErrSandboxUnavailable, err)
	}
	if r.token != "" {
		req.Header.Set("X-Sandbox-Token", r.token)
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: sandbox broker unreachable at %s: %v", ErrSandboxUnavailable, r.brokerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w: sandbox broker rejected credentials (status %d)", ErrBrokerAuthenticationFailed, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%w: sandbox broker health failed (status %d): %s", ErrSandboxUnavailable, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// EvaluateCandidate serializes the request and dispatches it to the remote sandbox broker.
func (r *RemoteBrokerEvaluator) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
	start := time.Now()

	// Pre-validate locally to fast-fail before making a network call
	if strings.TrimSpace(req.RepoPath) == "" || strings.TrimSpace(req.FilePath) == "" {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: repository path and file path are required", ErrInvalidRequest),
		}, ErrInvalidRequest
	}

	// Reject path traversals locally
	cleanInput := strings.ReplaceAll(req.FilePath, "\\", "/")
	if strings.HasPrefix(cleanInput, "/") || strings.HasPrefix(cleanInput, "\\") ||
		strings.Contains(cleanInput, "../") || strings.Contains(cleanInput, "/..") {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: path traversal detected in file path '%s'", ErrSecurityViolation, req.FilePath),
		}, ErrSecurityViolation
	}

	// 1. Verify remote broker is reachable and healthy
	if err := r.CheckAvailable(ctx); err != nil {
		outcome := OutcomeInfrastructureFailure
		if errors.Is(err, ErrBrokerAuthenticationFailed) {
			outcome = OutcomeSecurityRejected
		}
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     outcome,
			Duration:    time.Since(start),
			Error:       err,
		}, err
	}

	// 2. Serialize request
	payload, err := json.Marshal(req)
	if err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to encode evaluation request: %w", err),
		}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, r.brokerURL+"/v1/evaluate", bytes.NewReader(payload))
	if err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to create evaluate HTTP request: %w", err),
		}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if r.token != "" {
		httpReq.Header.Set("X-Sandbox-Token", r.token)
	}

	resp, err := r.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return &EvaluationResult{
				TestsPassed: false,
				Outcome:     OutcomeTimeout,
				Duration:    time.Since(start),
				Error:       ErrExecutionTimeout,
			}, nil
		}
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("%w: failed to communicate with broker: %v", ErrSandboxUnavailable, err),
		}, ErrSandboxUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeSecurityRejected,
			Duration:    time.Since(start),
			Error:       ErrBrokerAuthenticationFailed,
		}, ErrBrokerAuthenticationFailed
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Output:      string(body),
			Error:       fmt.Errorf("%w: broker returned HTTP status %d", ErrSandboxUnavailable, resp.StatusCode),
		}, ErrSandboxUnavailable
	}

	var evalRes EvaluationResult
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&evalRes); err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to decode evaluation response from broker: %w", err),
		}, err
	}

	return &evalRes, evalRes.Error
}
