package agent

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	// ErrBrokerAuthenticationFailed is returned when the sandbox broker rejects credentials.
	ErrBrokerAuthenticationFailed = errors.New("sandbox broker authentication failed: invalid or missing token")
)

type remoteBrokerEvaluationResponse struct {
	TestsPassed      bool              `json:"tests_passed"`
	Outcome          EvaluationOutcome `json:"outcome"`
	Output           string            `json:"output"`
	Duration         time.Duration     `json:"duration"`
	UpdatedFileBytes []byte            `json:"updated_file_bytes"`
	Error            string            `json:"error,omitempty"`
}

// RemoteBrokerEvaluator evaluates candidates via a dedicated, authenticated Sandbox Broker service.
type RemoteBrokerEvaluator struct {
	brokerURL  string
	token      string
	httpClient *http.Client
	configErr  error
}

// NewRemoteBrokerEvaluator initializes an evaluator pointing to a remote sandbox broker.
func NewRemoteBrokerEvaluator(brokerURL, token string) *RemoteBrokerEvaluator {
	evaluator := &RemoteBrokerEvaluator{
		brokerURL: strings.TrimRight(brokerURL, "/"),
		token:     strings.TrimSpace(token),
	}
	parsedURL, err := url.Parse(evaluator.brokerURL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" || parsedURL.User != nil ||
		parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		evaluator.configErr = fmt.Errorf("%w: SANDBOX_SERVICE_URL must be an absolute HTTPS URL without user info or query", ErrSandboxUnavailable)
		return evaluator
	}
	if len(evaluator.token) < 16 {
		evaluator.configErr = fmt.Errorf("%w: SANDBOX_SERVICE_TOKEN must be at least 16 characters", ErrBrokerAuthenticationFailed)
		return evaluator
	}
	evaluator.httpClient, err = newRemoteBrokerHTTPClient()
	if err != nil {
		evaluator.configErr = fmt.Errorf("%w: configure broker TLS: %v", ErrSandboxUnavailable, err)
	}
	return evaluator
}

func newRemoteBrokerHTTPClient() (*http.Client, error) {
	roots, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("load system TLS roots: %w", err)
	}
	if roots == nil {
		return nil, errors.New("system TLS root pool is unavailable")
	}
	if caPath := strings.TrimSpace(os.Getenv("SANDBOX_TLS_CA_FILE")); caPath != "" {
		caPEM, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("read sandbox TLS trust certificate: %w", err)
		}
		if !roots.AppendCertsFromPEM(caPEM) {
			return nil, errors.New("sandbox TLS trust file contains no certificates")
		}
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
	}
	return &http.Client{
		Timeout:   120 * time.Second,
		Transport: transport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

// CheckAvailable verifies whether the remote sandbox broker is reachable and healthy.
func (r *RemoteBrokerEvaluator) CheckAvailable(ctx context.Context) error {
	if r.configErr != nil {
		return r.configErr
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.brokerURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("%w: failed to create broker health request: %v", ErrSandboxUnavailable, err)
	}
	req.Header.Set("X-Sandbox-Token", r.token)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: sandbox broker unreachable at %s: %v", ErrSandboxUnavailable, r.brokerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w: sandbox broker rejected credentials (status %d)", ErrBrokerAuthenticationFailed, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 512))
		if readErr != nil {
			return fmt.Errorf("%w: sandbox broker health returned HTTP %d and response read failed: %v", ErrSandboxUnavailable, resp.StatusCode, readErr)
		}
		return fmt.Errorf("%w: sandbox broker health failed (status %d): %s", ErrSandboxUnavailable, resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (r *RemoteBrokerEvaluator) PrepareWorkspace(ctx context.Context, repoPath string) (string, error) {
	if r.configErr != nil {
		return "", r.configErr
	}
	if strings.TrimSpace(repoPath) == "" {
		return "", fmt.Errorf("%w: repository path is required", ErrInvalidRequest)
	}
	archive, err := CreateWorkspaceArchive(repoPath, maxWorkspaceArchiveBytes)
	if err != nil {
		return "", fmt.Errorf("create sandbox workspace archive: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.brokerURL+"/v1/workspaces", bytes.NewReader(archive))
	if err != nil {
		return "", fmt.Errorf("create sandbox workspace request: %w", err)
	}
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("X-Sandbox-Token", r.token)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: create sandbox workspace: %v", ErrSandboxUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return "", ErrBrokerAuthenticationFailed
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			return "", fmt.Errorf("sandbox workspace creation returned HTTP %d and response read failed: %w", resp.StatusCode, readErr)
		}
		return "", fmt.Errorf("sandbox workspace creation failed with HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		WorkspaceID string `json:"workspace_id"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("decode sandbox workspace response: %w", err)
	}
	if !validRemoteWorkspaceID(result.WorkspaceID) {
		return "", errors.New("sandbox broker returned an invalid workspace ID")
	}
	return result.WorkspaceID, nil
}

func (r *RemoteBrokerEvaluator) CleanupWorkspace(ctx context.Context, workspaceID string) error {
	if r.configErr != nil {
		return r.configErr
	}
	if !validRemoteWorkspaceID(workspaceID) {
		return errors.New("invalid sandbox workspace ID")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, r.brokerURL+"/v1/workspaces/"+workspaceID, nil)
	if err != nil {
		return fmt.Errorf("create sandbox workspace cleanup request: %w", err)
	}
	req.Header.Set("X-Sandbox-Token", r.token)
	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: cleanup sandbox workspace: %v", ErrSandboxUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrBrokerAuthenticationFailed
	}
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("sandbox workspace cleanup returned HTTP %d and response read failed: %w", resp.StatusCode, err)
	}
	return fmt.Errorf("sandbox workspace cleanup failed with HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
}

func validRemoteWorkspaceID(workspaceID string) bool {
	if len(workspaceID) != 64 {
		return false
	}
	_, err := hex.DecodeString(workspaceID)
	return err == nil
}

// EvaluateCandidate serializes the request and dispatches it to the remote sandbox broker.
func (r *RemoteBrokerEvaluator) EvaluateCandidate(ctx context.Context, req EvaluationRequest) (*EvaluationResult, error) {
	start := time.Now()
	// Pre-validate locally to fast-fail before making a network call
	if !validRemoteWorkspaceID(req.WorkspaceID) || strings.TrimSpace(req.FilePath) == "" {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: valid workspace ID and file path are required", ErrInvalidRequest),
		}, ErrInvalidRequest
	}
	maxFileBytes := DefaultWorkspaceArchiveLimits().MaxFileBytes
	if int64(len(req.OriginalBytes)) > maxFileBytes || int64(len(req.MutatedCode)) > maxFileBytes ||
		req.StartByte < 0 || req.EndByte < req.StartByte || req.EndByte > len(req.OriginalBytes) ||
		int64(len(req.OriginalBytes)-(req.EndByte-req.StartByte)+len(req.MutatedCode)) > maxFileBytes {
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    fmt.Errorf("%w: candidate source exceeds workspace file limits", ErrInvalidRequest),
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

	if r.configErr != nil {
		outcome := OutcomeInfrastructureFailure
		if errors.Is(r.configErr, ErrBrokerAuthenticationFailed) {
			outcome = OutcomeSecurityRejected
		}
		return &EvaluationResult{Outcome: outcome, Duration: time.Since(start), Error: r.configErr}, r.configErr
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
	remoteRequest := req
	remoteRequest.RepoPath = ""
	payload, err := json.Marshal(remoteRequest)
	if err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to encode evaluation request: %w", err),
		}, err
	}
	if int64(len(payload)) > MaxBrokerEvaluationRequestBytes {
		err := fmt.Errorf("%w: evaluation request exceeds broker size limit", ErrInvalidRequest)
		return &EvaluationResult{
			Outcome:  OutcomeSecurityRejected,
			Duration: time.Since(start),
			Error:    err,
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
	httpReq.Header.Set("X-Sandbox-Token", r.token)

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
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1024))
		if readErr != nil {
			return &EvaluationResult{
				TestsPassed: false,
				Outcome:     OutcomeInfrastructureFailure,
				Duration:    time.Since(start),
				Error:       fmt.Errorf("%w: broker returned HTTP %d and response read failed: %v", ErrSandboxUnavailable, resp.StatusCode, readErr),
			}, ErrSandboxUnavailable
		}
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Output:      string(body),
			Error:       fmt.Errorf("%w: broker returned HTTP status %d", ErrSandboxUnavailable, resp.StatusCode),
		}, ErrSandboxUnavailable
	}

	responseBytes, err := io.ReadAll(io.LimitReader(resp.Body, MaxBrokerEvaluationResponseBytes+1))
	if err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("read evaluation response from broker: %w", err),
		}, err
	}
	if int64(len(responseBytes)) > MaxBrokerEvaluationResponseBytes {
		err := errors.New("broker evaluation response exceeds size limit")
		return &EvaluationResult{Outcome: OutcomeInfrastructureFailure, Duration: time.Since(start), Error: err}, err
	}
	var remoteResponse remoteBrokerEvaluationResponse
	if err := json.Unmarshal(responseBytes, &remoteResponse); err != nil {
		return &EvaluationResult{
			TestsPassed: false,
			Outcome:     OutcomeInfrastructureFailure,
			Duration:    time.Since(start),
			Error:       fmt.Errorf("failed to decode evaluation response from broker: %w", err),
		}, err
	}
	evalRes := &EvaluationResult{
		TestsPassed:      remoteResponse.TestsPassed,
		Outcome:          remoteResponse.Outcome,
		Output:           remoteResponse.Output,
		Duration:         remoteResponse.Duration,
		UpdatedFileBytes: remoteResponse.UpdatedFileBytes,
	}
	if remoteResponse.Error != "" {
		evalRes.Error = errors.New(remoteResponse.Error)
	}
	return evalRes, nil
}
