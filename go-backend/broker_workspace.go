package main

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saksham/self-healing-pipeline/agent"
)

var (
	errBrokerWorkspaceNotFound       = errors.New("sandbox workspace not found")
	errBrokerWorkspaceCapacity       = errors.New("sandbox workspace capacity reached")
	errInvalidBrokerWorkspaceArchive = errors.New("invalid sandbox workspace archive")
)

type brokerEvaluationResponse struct {
	TestsPassed      bool                    `json:"tests_passed"`
	Outcome          agent.EvaluationOutcome `json:"outcome"`
	Output           string                  `json:"output"`
	Duration         time.Duration           `json:"duration"`
	UpdatedFileBytes []byte                  `json:"updated_file_bytes"`
	Error            string                  `json:"error,omitempty"`
}

type brokerWorkspace struct {
	path          string
	expiresAt     time.Time
	active        int
	deletePending bool
}

type brokerWorkspaceManager struct {
	root       string
	ttl        time.Duration
	now        func() time.Time
	mu         sync.Mutex
	workspaces map[string]*brokerWorkspace
	creating   int
}

func newBrokerWorkspaceManager(root string, ttl time.Duration) (*brokerWorkspaceManager, error) {
	if root == "" {
		return nil, errors.New("sandbox workspace root is required")
	}
	if ttl <= 0 {
		return nil, errors.New("sandbox workspace TTL must be positive")
	}
	root = filepath.Join(root, "sessions")
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create sandbox workspace root: %w", err)
	}
	if err := os.Chmod(root, 0o700); err != nil {
		return nil, fmt.Errorf("secure sandbox workspace root: %w", err)
	}
	root, err := filepath.EvalSymlinks(root)
	if err != nil {
		return nil, fmt.Errorf("resolve sandbox workspace root: %w", err)
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("make sandbox workspace root absolute: %w", err)
	}
	return &brokerWorkspaceManager{
		root:       root,
		ttl:        ttl,
		now:        time.Now,
		workspaces: make(map[string]*brokerWorkspace),
	}, nil
}

func (m *brokerWorkspaceManager) CreateWorkspace(archive []byte) (string, error) {
	limits := agent.DefaultWorkspaceArchiveLimits()
	if int64(len(archive)) > limits.MaxArchiveBytes {
		return "", fmt.Errorf("workspace archive exceeds %d-byte limit", limits.MaxArchiveBytes)
	}

	m.mu.Lock()
	if len(m.workspaces)+m.creating >= 1 {
		m.mu.Unlock()
		return "", errBrokerWorkspaceCapacity
	}
	m.creating++
	m.mu.Unlock()

	finishCreation := func() {
		m.mu.Lock()
		m.creating--
		m.mu.Unlock()
	}
	workspacePath, err := os.MkdirTemp(m.root, "workspace-")
	if err != nil {
		finishCreation()
		return "", fmt.Errorf("create temporary sandbox workspace: %w", err)
	}
	if err := agent.ExtractWorkspaceArchive(archive, workspacePath, limits); err != nil {
		cleanupErr := os.RemoveAll(workspacePath)
		finishCreation()
		return "", errors.Join(fmt.Errorf("%w: %v", errInvalidBrokerWorkspaceArchive, err), cleanupErr)
	}

	workspaceID, err := newBrokerWorkspaceID()
	if err != nil {
		cleanupErr := os.RemoveAll(workspacePath)
		finishCreation()
		return "", errors.Join(fmt.Errorf("create sandbox workspace ID: %w", err), cleanupErr)
	}
	m.mu.Lock()
	m.creating--
	if _, exists := m.workspaces[workspaceID]; exists {
		m.mu.Unlock()
		cleanupErr := os.RemoveAll(workspacePath)
		return "", errors.Join(errors.New("generated duplicate sandbox workspace ID"), cleanupErr)
	}
	m.workspaces[workspaceID] = &brokerWorkspace{
		path:      workspacePath,
		expiresAt: m.now().Add(m.ttl),
	}
	m.mu.Unlock()
	return workspaceID, nil
}

func (m *brokerWorkspaceManager) AcquireWorkspace(workspaceID string) (string, func() error, error) {
	if !validBrokerWorkspaceID(workspaceID) {
		return "", nil, errBrokerWorkspaceNotFound
	}
	m.mu.Lock()
	workspace, exists := m.workspaces[workspaceID]
	if !exists {
		m.mu.Unlock()
		return "", nil, errBrokerWorkspaceNotFound
	}
	now := m.now()
	if workspace.deletePending || !now.Before(workspace.expiresAt) {
		workspace.deletePending = true
		var removeErr error
		if workspace.active == 0 {
			removeErr = m.removeWorkspaceLocked(workspaceID, workspace)
		}
		m.mu.Unlock()
		if removeErr != nil {
			return "", nil, fmt.Errorf("remove expired sandbox workspace: %w", removeErr)
		}
		return "", nil, errBrokerWorkspaceNotFound
	}
	workspace.active++
	workspace.expiresAt = now.Add(m.ttl)
	path := workspace.path
	m.mu.Unlock()

	var once sync.Once
	var releaseErr error
	release := func() error {
		once.Do(func() {
			m.mu.Lock()
			current, exists := m.workspaces[workspaceID]
			if !exists || current != workspace {
				m.mu.Unlock()
				releaseErr = errBrokerWorkspaceNotFound
				return
			}
			current.active--
			if current.active < 0 {
				current.active = 0
				m.mu.Unlock()
				releaseErr = errors.New("sandbox workspace lease count underflow")
				return
			}
			if current.active == 0 && (current.deletePending || !m.now().Before(current.expiresAt)) {
				releaseErr = m.removeWorkspaceLocked(workspaceID, current)
			}
			m.mu.Unlock()
		})
		return releaseErr
	}
	return path, release, nil
}

func (m *brokerWorkspaceManager) DeleteWorkspace(workspaceID string) error {
	if !validBrokerWorkspaceID(workspaceID) {
		return errBrokerWorkspaceNotFound
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	workspace, exists := m.workspaces[workspaceID]
	if !exists {
		return errBrokerWorkspaceNotFound
	}
	workspace.deletePending = true
	if workspace.active != 0 {
		return nil
	}
	return m.removeWorkspaceLocked(workspaceID, workspace)
}

func (m *brokerWorkspaceManager) CleanupExpired() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	removed := 0
	var cleanupErrors []error
	now := m.now()
	for workspaceID, workspace := range m.workspaces {
		if now.Before(workspace.expiresAt) {
			continue
		}
		workspace.deletePending = true
		if workspace.active != 0 {
			continue
		}
		if err := m.removeWorkspaceLocked(workspaceID, workspace); err != nil {
			cleanupErrors = append(cleanupErrors, err)
			continue
		}
		removed++
	}
	return removed, errors.Join(cleanupErrors...)
}

func (m *brokerWorkspaceManager) removeWorkspaceLocked(workspaceID string, workspace *brokerWorkspace) error {
	if err := os.RemoveAll(workspace.path); err != nil {
		return fmt.Errorf("remove sandbox workspace %s: %w", workspaceID, err)
	}
	delete(m.workspaces, workspaceID)
	return nil
}

func (m *brokerWorkspaceManager) RunCleanup(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		log.Printf("[ERROR] sandbox workspace cleanup interval must be positive")
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := m.CleanupExpired(); err != nil {
				log.Printf("[ERROR] clean expired sandbox workspaces: %v", err)
			}
		}
	}
}

func newBrokerWorkspaceID() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func validBrokerWorkspaceID(workspaceID string) bool {
	if len(workspaceID) != 64 {
		return false
	}
	_, err := hex.DecodeString(workspaceID)
	return err == nil
}

func newSandboxBrokerHandler(token string, evaluator agent.CandidateEvaluator, workspaces *brokerWorkspaceManager, limiter *EvolutionAdmissionController) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	authMiddleware := func(c *gin.Context) {
		provided := c.GetHeader("X-Sandbox-Token")
		if len(token) < minSandboxServiceTokenBytes || len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing sandbox token"})
			return
		}
		c.Next()
	}

	r.GET("/health", func(c *gin.Context) {
		provided := c.GetHeader("X-Sandbox-Token")
		isLoopback := c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1"
		if len(token) < minSandboxServiceTokenBytes || (!isLoopback && (len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing sandbox token"})
			return
		}
		if evaluator == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": "sandbox evaluator is not configured"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := evaluator.CheckAvailable(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/v1/workspaces", authMiddleware, func(c *gin.Context) {
		mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
		if err != nil || mediaType != "application/zip" {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "application/zip is required"})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, agent.DefaultWorkspaceArchiveLimits().MaxArchiveBytes)
		archive, err := io.ReadAll(c.Request.Body)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "workspace archive exceeds upload limit"})
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "could not read workspace archive"})
			return
		}
		workspaceID, err := workspaces.CreateWorkspace(archive)
		switch {
		case errors.Is(err, errBrokerWorkspaceCapacity):
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "sandbox workspace capacity reached"})
		case errors.Is(err, errInvalidBrokerWorkspaceArchive):
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid workspace archive"})
		case err != nil:
			log.Printf("[ERROR] create sandbox workspace: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "workspace creation failed"})
		default:
			c.JSON(http.StatusCreated, gin.H{"workspace_id": workspaceID})
		}
	})

	r.DELETE("/v1/workspaces/:id", authMiddleware, func(c *gin.Context) {
		if err := workspaces.DeleteWorkspace(c.Param("id")); err != nil {
			if errors.Is(err, errBrokerWorkspaceNotFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
				return
			}
			log.Printf("[ERROR] delete sandbox workspace: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "workspace cleanup failed"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.POST("/v1/evaluate", authMiddleware, func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, agent.MaxBrokerEvaluationRequestBytes)
		var req agent.EvaluationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "evaluation request exceeds body limit"})
				return
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid evaluation request"})
			return
		}
		if req.RepoPath != "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "repo_path is not accepted"})
			return
		}
		if err := validateBrokerEvaluationRequest(req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		workspacePath, releaseWorkspace, err := workspaces.AcquireWorkspace(req.WorkspaceID)
		if err != nil {
			if errors.Is(err, errBrokerWorkspaceNotFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
				return
			}
			log.Printf("[ERROR] acquire sandbox workspace: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "workspace acquisition failed"})
			return
		}
		workspaceReleased := false
		defer func() {
			if !workspaceReleased {
				if releaseErr := releaseWorkspace(); releaseErr != nil {
					log.Printf("[ERROR] release sandbox workspace lease: %v", releaseErr)
				}
			}
		}()
		if limiter == nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "sandbox admission controller unavailable"})
			return
		}
		releaseAdmission, err := limiter.Acquire(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "sandbox execution engine at maximum capacity", "code": "server_busy"})
			return
		}
		admissionReleased := false
		defer func() {
			if !admissionReleased {
				releaseAdmission()
			}
		}()

		req.RepoPath = workspacePath
		req.Limits = clampLimits(req.Limits)
		result, evalErr := evaluator.EvaluateCandidate(c.Request.Context(), req)
		releaseAdmission()
		admissionReleased = true
		releaseErr := releaseWorkspace()
		workspaceReleased = true
		if evalErr != nil {
			log.Printf("[ERROR] sandbox candidate evaluation failed: %v", evalErr)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "sandbox candidate evaluation failed"})
			return
		}
		if releaseErr != nil {
			log.Printf("[ERROR] sandbox workspace cleanup failed: %v", releaseErr)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "sandbox workspace cleanup failed"})
			return
		}
		if result == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "sandbox returned an empty evaluation result"})
			return
		}
		if int64(len(result.UpdatedFileBytes)) > agent.DefaultWorkspaceArchiveLimits().MaxFileBytes {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "sandbox result exceeds file size limit"})
			return
		}
		response := brokerEvaluationResponse{
			TestsPassed:      result.TestsPassed,
			Outcome:          result.Outcome,
			Output:           result.Output,
			Duration:         result.Duration,
			UpdatedFileBytes: result.UpdatedFileBytes,
		}
		if result.Error != nil {
			response.Error = result.Error.Error()
		}
		responseBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("[ERROR] encode sandbox evaluation response: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "sandbox response encoding failed"})
			return
		}
		if int64(len(responseBytes)) > agent.MaxBrokerEvaluationResponseBytes {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "sandbox response exceeds size limit"})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", responseBytes)
	})

	return r
}

func validateBrokerEvaluationRequest(req agent.EvaluationRequest) error {
	maxFileBytes := agent.DefaultWorkspaceArchiveLimits().MaxFileBytes
	if strings.TrimSpace(req.FilePath) == "" {
		return errors.New("file_path is required")
	}
	if int64(len(req.OriginalBytes)) > maxFileBytes || int64(len(req.MutatedCode)) > maxFileBytes {
		return errors.New("candidate source exceeds workspace file size limit")
	}
	if req.StartByte < 0 || req.EndByte < req.StartByte || req.EndByte > len(req.OriginalBytes) {
		return errors.New("candidate byte range is invalid")
	}
	finalSize := len(req.OriginalBytes) - (req.EndByte - req.StartByte) + len(req.MutatedCode)
	if int64(finalSize) > maxFileBytes {
		return errors.New("updated candidate source exceeds workspace file size limit")
	}
	return nil
}
