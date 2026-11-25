package execution

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/saksham/self-healing-pipeline/agent"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
	"github.com/saksham/self-healing-pipeline/agent/repo_intel/planning"
)

// MVPTarget defines an evaluation or live repository target.
type MVPTarget struct {
	RuleID            string  `json:"rule_id,omitempty"`
	Symbol            string  `json:"symbol,omitempty"`
	LineEnd           int     `json:"line_end,omitempty"`
	Confidence        float64 `json:"confidence"`
	UncertaintyClass  string  `json:"uncertainty_class,omitempty"`
	SourceEvidence    string  `json:"source_evidence,omitempty"`
	DetectionEvidence string  `json:"detection_evidence,omitempty"`
	EligibilityReason string  `json:"eligibility_reason,omitempty"`
	TargetID          string  `json:"target_id"`
	Repository        string  `json:"repository"`
	RelativeSource    string  `json:"relative_source"`
	TargetFile        string  `json:"target_file"`
	DefectCategory    string  `json:"defect_category"`
	Stratum           string  `json:"stratum"`
	ExpectedAction    string  `json:"expected_action"`
	Description       string  `json:"description"`
	UpstreamCommit    string  `json:"upstream_commit,omitempty"`
	UpstreamURL       string  `json:"upstream_url,omitempty"`
	VerifiedBy        string  `json:"verified_by"`
	WorkspaceID       string  `json:"workspace_id,omitempty"`
	LineNumber        int     `json:"line_number,omitempty"`
	IsDemoTarget      bool    `json:"is_demo_target"`
	SnapshotID        string  `json:"snapshot_id,omitempty"`
	SourceFileHash    string  `json:"source_file_hash,omitempty"`
}

// MVPAnalysisResult stores diagnostic and semantic findings for a target.
type MVPAnalysisResult struct {
	AnalysisID           string   `json:"analysis_id"`
	TargetID             string   `json:"target_id"`
	Repository           string   `json:"repository"`
	TargetFile           string   `json:"target_file"`
	DefectCategory       string   `json:"defect_category"`
	Severity             string   `json:"severity"`
	LineNumber           int      `json:"line_number"`
	CodeSnippet          string   `json:"code_snippet"`
	SemanticEvidence     []string `json:"semantic_evidence"`
	UncertaintyScore     float64  `json:"uncertainty_score"`
	AutomaticEligibility bool     `json:"automatic_eligibility"`
	EligibilityRationale string   `json:"eligibility_rationale"`
	RequiresHumanReview  bool     `json:"requires_human_review"`
	FindingID            string   `json:"finding_id"`
	Timestamp            string   `json:"timestamp"`
	SnapshotID           string   `json:"snapshot_id,omitempty"`
	SourceFileHash       string   `json:"source_file_hash,omitempty"`
}

// MVPProposalResult contains the synthesized patch diff or explicit abstention reason.
type MVPProposalResult struct {
	ProposalID            string `json:"proposal_id"`
	AnalysisID            string `json:"analysis_id"`
	TargetID              string `json:"target_id"`
	Status                string `json:"status"` // "PROPOSED", "ABSTAINED"
	IntentType            string `json:"intent_type,omitempty"`
	UnifiedDiff           string `json:"unified_diff,omitempty"`
	DiffHash              string `json:"diff_hash,omitempty"`
	ModifiedSource        string `json:"modified_source,omitempty"`
	AbstainCategory       string `json:"abstain_category,omitempty"`
	AbstainReason         string `json:"abstain_reason,omitempty"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
	ExecutionClass        string `json:"execution_class"`
	FileHash              string `json:"file_hash"`
	WorkspaceID           string `json:"workspace_id"`
	TargetAbsPath         string `json:"target_abs_path"`
	SandboxDir            string `json:"sandbox_dir"`
	DurationMs            int64  `json:"duration_ms"`
	SnapshotID            string `json:"snapshot_id,omitempty"`
}

// MVPApprovalResult contains the issued cryptographic authorization token.
type MVPApprovalResult struct {
	TokenID    string `json:"token_id"`
	ProposalID string `json:"proposal_id"`
	Principal  string `json:"principal"`
	Role       string `json:"role"`
	DiffHash   string `json:"diff_hash"`
	RiskTier   string `json:"risk_tier"`
	ExpiresAt  string `json:"expires_at"`
	Signature  string `json:"signature"`
	Valid      bool   `json:"valid"`
	IssuedAt   string `json:"issued_at"`
}

// MVPOracleOutcome records single-oracle verification details.
type MVPOracleOutcome struct {
	OracleType    string                      `json:"oracle_type"`
	Status        string                      `json:"status"` // "PASSED", "FAILED", "NOT_RUN", "UNAVAILABLE"
	Description   string                      `json:"description"`
	Passed        bool                        `json:"passed"`
	Details       string                      `json:"details"`
	ContainerInfo *ContainerExecutionEvidence `json:"container_info,omitempty"`
}

// MVPExecutionResult compiles sandbox execution, multi-oracle checks, and source immutability.
type MVPExecutionResult struct {
	ExecutionID            string                      `json:"execution_id"`
	ProposalID             string                      `json:"proposal_id"`
	TargetID               string                      `json:"target_id"`
	TokenID                string                      `json:"token_id"`
	SandboxDir             string                      `json:"sandbox_dir"`
	AppliedInSandbox       bool                        `json:"applied_in_sandbox"`
	TokenConsumed          bool                        `json:"token_consumed"`
	Oracles                map[string]MVPOracleOutcome `json:"oracles"`
	AllOraclesPassed       bool                        `json:"all_oracles_passed"`
	VerificationVerdict    string                      `json:"verification_verdict"`
	OriginalRepoPreDigest  string                      `json:"original_repo_pre_digest"`
	OriginalRepoPostDigest string                      `json:"original_repo_post_digest"`
	OriginalRepoImmutable  bool                        `json:"original_repo_immutable"`
	CleanupOperation       string                      `json:"cleanup_operation"`
	CleanupStatus          string                      `json:"cleanup_status,omitempty"`
	RollbackAvailable      bool                        `json:"rollback_available"`
	DurationMs             int64                       `json:"duration_ms"`
	Timestamp              string                      `json:"timestamp"`
}

// MVPAuditRecord stores an immutable record of an end-to-end workflow trial.
type MVPAuditRecord struct {
	RecordID       string              `json:"record_id"`
	TargetID       string              `json:"target_id"`
	Repository     string              `json:"repository"`
	DefectCategory string              `json:"defect_category"`
	Analysis       *MVPAnalysisResult  `json:"analysis"`
	Proposal       *MVPProposalResult  `json:"proposal"`
	Approval       *MVPApprovalResult  `json:"approval,omitempty"`
	Execution      *MVPExecutionResult `json:"execution,omitempty"`
	FinalStatus    string              `json:"final_status"`
	Timestamp      string              `json:"timestamp"`
}

// MVPAuditReport compiles the complete audit trail.
type MVPAuditReport struct {
	BaselineCommit        string           `json:"baseline_commit"`
	ExportTimestamp       string           `json:"export_timestamp"`
	TotalSessionsRecorded int              `json:"total_sessions_recorded"`
	VerifiedRepairsCount  int              `json:"verified_repairs_count"`
	AbstainedRepairsCount int              `json:"abstained_repairs_count"`
	FailedRepairsCount    int              `json:"failed_repairs_count"`
	SourceImmutability100 bool             `json:"source_immutability_100_percent"`
	AuditRecords          []MVPAuditRecord `json:"audit_records"`
}

// WorkspaceSnapshot stores immutable snapshot metadata bound to a scanned workspace.
type WorkspaceSnapshot struct {
	SnapshotID   string            `json:"snapshot_id"`
	WorkspaceID  string            `json:"workspace_id"`
	OwnerID      string            `json:"owner_id"`
	SourceCommit string            `json:"source_commit"`
	RootDigest   string            `json:"root_digest"`
	FileHashes   map[string]string `json:"file_hashes"`
	TargetCount  int               `json:"target_count"`
	CreatedAt    string            `json:"created_at"`
	ExpiresAt    string            `json:"expires_at"`
}

// ValidateWorkspaceID validates that a workspace identifier is well-formed and safe.
func ValidateWorkspaceID(workspaceID string) error {
	trimmed := strings.TrimSpace(workspaceID)
	if trimmed == "" {
		return fmt.Errorf("MALFORMED_WORKSPACE_ID: workspace ID cannot be empty")
	}
	if strings.ContainsRune(trimmed, 0) {
		return fmt.Errorf("MALFORMED_WORKSPACE_ID: workspace ID contains null bytes")
	}
	if filepath.IsAbs(trimmed) || strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "\\") {
		return fmt.Errorf("MALFORMED_WORKSPACE_ID: absolute workspace ID forbidden: %s", workspaceID)
	}
	if strings.HasPrefix(trimmed, `\\`) || strings.HasPrefix(trimmed, `//`) ||
		strings.HasPrefix(trimmed, `\/`) || strings.HasPrefix(trimmed, `/\`) {
		return fmt.Errorf("MALFORMED_WORKSPACE_ID: UNC or device workspace ID forbidden: %s", workspaceID)
	}
	if len(trimmed) >= 2 && trimmed[1] == ':' {
		return fmt.Errorf("MALFORMED_WORKSPACE_ID: drive-letter workspace ID forbidden: %s", workspaceID)
	}
	norm := strings.ReplaceAll(trimmed, "\\", "/")
	for _, part := range strings.Split(norm, "/") {
		if part == ".." {
			return fmt.Errorf("MALFORMED_WORKSPACE_ID: path traversal in workspace ID: %s", workspaceID)
		}
	}
	clean := filepath.Clean(norm)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("MALFORMED_WORKSPACE_ID: path traversal in workspace ID: %s", workspaceID)
	}
	return nil
}

// ValidateSnapshotID validates that a snapshot identifier is well-formed (SNAP-<64 hex chars>).
func ValidateSnapshotID(snapshotID string) error {
	trimmed := strings.TrimSpace(snapshotID)
	if trimmed == "" {
		return fmt.Errorf("MALFORMED_SNAPSHOT_ID: snapshot ID cannot be empty")
	}
	if !strings.HasPrefix(trimmed, "SNAP-") {
		return fmt.Errorf("MALFORMED_SNAPSHOT_ID: snapshot ID must start with SNAP-: %s", snapshotID)
	}
	digest := strings.TrimPrefix(trimmed, "SNAP-")
	if len(digest) != 64 {
		return fmt.Errorf("MALFORMED_SNAPSHOT_ID: snapshot digest length must be 64 hex characters: got %d in %s", len(digest), snapshotID)
	}
	for _, c := range digest {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("MALFORMED_SNAPSHOT_ID: snapshot digest contains non-hex character: %c", c)
		}
	}
	return nil
}

// MVPAnalysisOperator is the shared authenticated admin principal used by the analysis gateway.
const MVPAnalysisOperator = "authenticated-analysis-client"

// MVPService provides the unified business logic for the Autonomous Repair MVP.
type MVPService struct {
	WorkspaceRoot     string
	Authority         *ApprovalAuthority
	Registry          *ApprovalConsumptionRegistry
	Synthesizer       *AutonomousPatchSynthesizer
	Oracles           *IndependentOracleSuite
	BaselineSHA       string
	WorkspaceResolver func(workspaceID string) (string, error)
	Runner            ContainerRunner

	mu              sync.Mutex
	snapshots       map[string]*WorkspaceSnapshot
	dynamicTargets  map[string]*MVPTarget
	tokenObjs       map[string]*ApprovalToken
	analyses        map[string]*MVPAnalysisResult
	proposals       map[string]*MVPProposalResult
	approvals       map[string]*MVPApprovalResult
	executions      map[string]*MVPExecutionResult
	activeProposals map[string]bool
	auditTrail      []MVPAuditRecord
}

// DefaultPreconfiguredTargets returns the audited pilot corpus targets.
func DefaultPreconfiguredTargets() []MVPTarget {
	return []MVPTarget{
		{
			TargetID:       "INJ-EXT2-RES-02",
			Repository:     "fatih_color",
			RelativeSource: "benchmarks/external_v2/fatih_color",
			TargetFile:     "file_printer.go",
			DefectCategory: "RESOURCE_LEAK",
			Stratum:        "AUTHENTIC_REPO_INJECTED",
			ExpectedAction: "REPAIR",
			Description:    "[DEMO] Authentic Repo Injected Defect: os.Open file descriptor leak in DumpColorProfile",
			UpstreamCommit: "12126ed593697635c525b302836b292b657ea573",
			UpstreamURL:    "https://github.com/fatih/color",
			VerifiedBy:     "External v2 Injection Ledger",
			IsDemoTarget:   true,
		},
		{
			TargetID:       "INJ-EXT2-SQL-01",
			Repository:     "stretchr_objx",
			RelativeSource: "benchmarks/external_v2/stretchr_objx",
			TargetFile:     "sql_helper.go",
			DefectCategory: "SQL_INJECTION",
			Stratum:        "AUTHENTIC_REPO_INJECTED",
			ExpectedAction: "REPAIR",
			Description:    "[DEMO] Authentic Repo Injected Defect: Dynamic SQL string formatting via fmt.Sprintf (Requires parameterization)",
			UpstreamCommit: "753177893c52a48b9487b7a1cf711d948cfbb20e",
			UpstreamURL:    "https://github.com/stretchr/objx",
			VerifiedBy:     "External v2 Injection Ledger",
			IsDemoTarget:   true,
		},
		{
			TargetID:       "INJ-EXT2-CMD-01",
			Repository:     "schollz_progressbar",
			RelativeSource: "benchmarks/external_v2/schollz_progressbar",
			TargetFile:     "cmd_runner.go",
			DefectCategory: "COMMAND_INJECTION",
			Stratum:        "AUTHENTIC_REPO_INJECTED",
			ExpectedAction: "ABSTAIN",
			Description:    "[DEMO] Authentic Repo Injected Defect: Unsanitized command executed via sh -c (Must safely abstain)",
			UpstreamCommit: "17ae25cdf22bc402ef34618a2ddddbfdb832ab81",
			UpstreamURL:    "https://github.com/schollz/progressbar",
			VerifiedBy:     "External v2 Injection Ledger",
			IsDemoTarget:   true,
		},
		{
			TargetID:       "DEF-HOLD-01",
			Repository:     "microservice_api",
			RelativeSource: "benchmarks/heldout_v1/microservice_api",
			TargetFile:     "config/config.go",
			DefectCategory: "UNHANDLED_ERROR",
			Stratum:        "CONSTRUCTED_MICROSERVICE",
			ExpectedAction: "ABSTAIN",
			Description:    "[DEMO] Constructed Microservice Defect: json.Unmarshal error discarded in non-error return signature",
			VerifiedBy:     "Held-Out Suite Manifest",
			IsDemoTarget:   true,
		},
		{
			TargetID:       "NAT-03",
			Repository:     "gorilla_websocket",
			RelativeSource: "benchmarks/external_v2/gorilla_websocket",
			TargetFile:     "client.go",
			DefectCategory: "RESOURCE_LEAK",
			Stratum:        "GROUND_TRUTH_DISQUALIFIED",
			ExpectedAction: "ABSTAIN",
			Description:    "[DEMO] Disqualified / False Alarm: Pristine DialContext manages socket via defer at line 330; response body does not require application close (Safely abstained)",
			UpstreamCommit: "9111bb834a68b893cebbbaed5060bdbc1d9ab7d2",
			UpstreamURL:    "https://github.com/gorilla/websocket",
			VerifiedBy:     "Disqualified in ground_truth_amendment_v2_1.json; Issue #789 unrelated",
			IsDemoTarget:   true,
		},
	}
}

// NewMVPService constructs an initialized MVP service instance.
func NewMVPService(workspaceRoot, persistLedgerPath string, runner ...ContainerRunner) (*MVPService, error) {
	if workspaceRoot == "" {
		workspaceRoot = ResolveWorkspaceRoot()
	}
	tmpDir, err := EnsureTmpDir(workspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure tmp directory: %w", err)
	}

	if persistLedgerPath == "" {
		persistLedgerPath = filepath.Join(tmpDir, "consumed_approvals_mvp.json")
	} else {
		if err := os.MkdirAll(filepath.Dir(persistLedgerPath), 0755); err != nil {
			persistLedgerPath = filepath.Join(tmpDir, filepath.Base(persistLedgerPath))
		}
	}

	registry, err := NewApprovalConsumptionRegistry(persistLedgerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to init approval registry: %w", err)
	}

	secretKey := []byte(os.Getenv("MVP_APPROVAL_SIGNING_KEY"))
	if len(secretKey) == 0 {
		// The approval objects are process-local, so a process-local key can protect them.
		secretKey = make([]byte, 32)
		if _, err := rand.Read(secretKey); err != nil {
			return nil, fmt.Errorf("generate approval signing key: %w", err)
		}
	} else if len(secretKey) < 32 || strings.TrimSpace(string(secretKey)) == "" {
		return nil, fmt.Errorf("MVP_APPROVAL_SIGNING_KEY must contain at least 32 bytes")
	}
	authority := NewApprovalAuthority(secretKey, registry)
	synthesizer := NewAutonomousPatchSynthesizer()

	var r ContainerRunner
	if len(runner) > 0 && runner[0] != nil {
		r = runner[0]
	} else {
		r = NewDockerContainerRunner()
	}
	oracles := NewIndependentOracleSuite(20, r)

	return &MVPService{
		WorkspaceRoot:  workspaceRoot,
		Authority:      authority,
		Registry:       registry,
		Synthesizer:    synthesizer,
		Oracles:        oracles,
		BaselineSHA:    "de06b6e568e0cce07d20b319d6618c15a9d3c899",
		Runner:         r,
		snapshots:      make(map[string]*WorkspaceSnapshot),
		dynamicTargets: make(map[string]*MVPTarget),
		tokenObjs:      make(map[string]*ApprovalToken),
		analyses:       make(map[string]*MVPAnalysisResult),
		proposals:      make(map[string]*MVPProposalResult),
		approvals:      make(map[string]*MVPApprovalResult),
		executions:     make(map[string]*MVPExecutionResult),
		auditTrail:     make([]MVPAuditRecord, 0),
	}, nil
}

// SetContainerRunner configures or overrides the container isolation runtime for verification.
func (s *MVPService) SetContainerRunner(runner ContainerRunner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Runner = runner
	if s.Oracles != nil {
		s.Oracles.Runner = runner
	}
}

// SetWorkspaceResolver configures the callback used to locate retained workspace directories.
func (s *MVPService) SetWorkspaceResolver(resolver func(workspaceID string) (string, error)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.WorkspaceResolver = resolver
}

func (s *MVPService) resolveWorkspaceDirLocked(workspaceID string) (string, error) {
	if workspaceID != "" {
		if err := ValidateWorkspaceID(workspaceID); err != nil {
			return "", err
		}
	}
	if s.WorkspaceResolver != nil && workspaceID != "" {
		dir, err := s.WorkspaceResolver(workspaceID)
		if err == nil && dir != "" {
			if fi, statErr := os.Lstat(dir); statErr == nil && fi.IsDir() {
				return dir, nil
			}
		}
	}
	return "", fmt.Errorf("workspace %q cannot be resolved or is not a directory", workspaceID)
}

// ListTargets is the legacy source-only accessor. HTTP callers use ListLiveTargets
// so discovery errors cannot be mistaken for an empty successful scan.
func (s *MVPService) ListTargets(workspaceID ...string) []MVPTarget {
	if len(workspaceID) == 0 || strings.TrimSpace(workspaceID[0]) == "" {
		return []MVPTarget{}
	}
	targets, _ := s.ListLiveTargets(workspaceID[0])
	return targets
}

func (s *MVPService) ListLiveTargets(workspaceID string) ([]MVPTarget, error) {
	return s.DiscoverWorkspaceTargets(workspaceID, MVPAnalysisOperator)
}

// AnalyzeTarget inspects a target and returns diagnostic findings with semantic evidence.
func (s *MVPService) AnalyzeTarget(targetID string) (*MVPAnalysisResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var target *MVPTarget
	if dyn, ok := s.dynamicTargets[targetID]; ok {
		target = dyn
	} else {
		for _, t := range DefaultPreconfiguredTargets() {
			if t.TargetID == targetID {
				target = &t
				break
			}
		}
	}
	if target == nil {
		return nil, fmt.Errorf("target %q not recognized", targetID)
	}

	if !target.IsDemoTarget {
		snap := s.snapshots[target.SnapshotID]
		if snap == nil || snap.WorkspaceID != target.WorkspaceID {
			return nil, fmt.Errorf("SNAPSHOT_MISSING: invalid target snapshot")
		}
		expiry, err := time.Parse(time.RFC3339, snap.ExpiresAt)
		if err != nil || !time.Now().Before(expiry) {
			return nil, fmt.Errorf("SNAPSHOT_EXPIRED: target snapshot expired")
		}
		dir, err := s.resolveWorkspaceDirLocked(target.WorkspaceID)
		if err != nil {
			return nil, err
		}
		digest, err := ComputeRepoDigest(dir)
		if err != nil || digest != snap.RootDigest {
			return nil, fmt.Errorf("STALE_WORKSPACE: source changed or unavailable")
		}
	}

	analysisID := fmt.Sprintf("ANALYSIS-%s-%d", target.TargetID, time.Now().UnixNano()%1000000)

	var evidence []string
	var snippet string
	var lineNo int
	var uncertainty float64
	var autoEligible bool
	var rationale string

	if target.IsDemoTarget {
		switch target.TargetID {
		case "INJ-EXT2-RES-02":
			lineNo = 21
			snippet = "f, err := os.Open(profilePath)\nif err != nil {\n    return err\n}\n// Missing defer f.Close() -> RESOURCE_LEAK\ndata, err := io.ReadAll(f)"
			evidence = []string{
				"Reaching Definitions: File handle 'f' allocated via os.Open at line 22.",
				"Lifecycle Sink: Function exits at lines 25, 30, and 33 without closing file handle 'f'.",
				"AST Mutator: Insert 'defer f.Close()' immediately following error check at line 25.",
			}
			uncertainty = 0.05
			autoEligible = true
			rationale = "Supported deterministic AST mutation: Single-file intra-procedural deferred close."
		case "INJ-EXT2-SQL-01":
			lineNo = 25
			snippet = "query := fmt.Sprintf(\"SELECT id, payload FROM data_store WHERE role = '%s'\", role)\nreturn e.db.QueryContext(ctx, query)"
			evidence = []string{
				"Taint Propagation: Function parameter 'role' formatted into raw SQL string via fmt.Sprintf.",
				"Sink: db.QueryContext executes untrusted dynamic string.",
				"AST Mutator: Replace fmt.Sprintf with ? placeholder; bind role argument to QueryContext.",
			}
			uncertainty = 0.08
			autoEligible = true
			rationale = "Supported deterministic AST mutation: SQL placeholder parameterization."
		case "INJ-EXT2-CMD-01":
			lineNo = 15
			snippet = "cmd := exec.Command(\"sh\", \"-c\", \"echo hook: \"+customHook)\nreturn cmd.Run()"
			evidence = []string{
				"Taint Sink: exec.Command argument passed directly to shell interpreter 'sh -c'.",
				"Arbitrary Shell Expansion: customHook allows semicolon/pipe command injection.",
				"AST Boundary Bound: Parameterizing shell string requires architectural redesign (split args).",
			}
			uncertainty = 0.95
			autoEligible = false
			rationale = "Unsupported AST repair: Shell command injection requires architectural restructuring. Safe fail-closed."
		case "NAT-03":
			lineNo = 371
			snippet = "resp, err := http.ReadResponse(conn.br, req)\nif err != nil {\n    return nil, nil, err\n}"
			evidence = []string{
				"Ground-Truth Disqualification: Upstream gorilla/websocket#789 is an issue titled 'Investigating Error code 1006', not a resource leak PR.",
				"Socket Lifecycle: Underline netConn is already managed and closed via defer func() at line 330 on error exits.",
				"Upstream Contract: Official DialContext docstring explicitly states response body does not need to be closed by application.",
				"Escaping Return Signature: DialContext returns (*Conn, *http.Response, error); callee cleanup cannot be safely applied.",
			}
			uncertainty = 0.95
			autoEligible = false
			rationale = "Ground-Truth Disqualified: Pristine upstream code manages socket lifecycle via defer at line 330. Callee cleanup on escaping response body safely abstained."
		default:
			lineNo = 40
			snippet = "// Candidate source location"
			evidence = []string{"Diagnostic summary confirmed."}
			uncertainty = 0.15
			autoEligible = (target.ExpectedAction == "REPAIR")
			rationale = "Standard candidate finding."
		}
	} else {
		lineNo = target.LineNumber
		if lineNo <= 0 {
			lineNo = 1
		}

		wsDir, wsErr := s.resolveWorkspaceDirLocked(target.WorkspaceID)
		if wsErr != nil {
			uncertainty = 0.99
			autoEligible = false
			rationale = fmt.Sprintf("Workspace resolution failed: %v. Safe fail-closed.", wsErr)
			evidence = []string{"Workspace unavailable or expired."}
			snippet = "// Source code unavailable"
		} else {
			if content, readErr := readWorkspaceFile(wsDir, target.TargetFile); readErr == nil {
				snippet = extractSourceSnippet(string(content), lineNo)
			} else {
				snippet = fmt.Sprintf("// Target file %s could not be read: %v", target.TargetFile, readErr)
			}

			switch target.DefectCategory {
			case "COMMAND_INJECTION":
				evidence = []string{
					fmt.Sprintf("Taint Sink: Arbitrary command execution pattern identified in %s:%d.", target.TargetFile, lineNo),
					"Architectural Constraint: Splitting command into discrete arguments requires semantic understanding of invocation intent.",
					"Safety Policy: Autonomous modification of shell execution boundaries is strictly prohibited.",
				}
				uncertainty = 0.95
				autoEligible = false
				rationale = "Unsupported AST repair: Arbitrary shell execution requires human architectural review. Conservative fail-closed abstention."

			case "HIGH_COMPLEXITY":
				evidence = []string{
					fmt.Sprintf("Cognitive/Cyclomatic Complexity Hotspot in %s:%d.", target.TargetFile, lineNo),
					"Metric Alert: Function exceeds maintainability threshold (Cyclomatic Complexity >= 15).",
					"Refactoring Requirement: Method extraction and state decomposition cannot be reliably automated without specification.",
				}
				uncertainty = 0.85
				autoEligible = false
				rationale = "High complexity hotspot requires decomposition and human architectural review. Safe fail-closed abstention."

			case "RESOURCE_LEAK":
				escapes := checkResourceEscapes(wsDir, target.TargetFile, lineNo)
				if escapes {
					evidence = []string{
						fmt.Sprintf("Resource handle allocated at %s:%d.", target.TargetFile, lineNo),
						"Escaping Lifecycle: Resource is returned to caller or transferred across functional boundary.",
						"Safety Policy: Callee cannot insert defer close when caller takes ownership of resource.",
					}
					uncertainty = 0.90
					autoEligible = false
					rationale = "Resource handle escapes enclosing function; caller owns lifecycle. Conservative fail-closed abstention."
				} else {
					evidence = []string{
						fmt.Sprintf("Reaching Definitions: Resource handle allocated at %s:%d.", target.TargetFile, lineNo),
						"Intra-procedural Sink: Resource handle does not escape enclosing function.",
						"AST Mutator: Candidate for deterministic deferred Close() insertion following acquisition check.",
					}
					uncertainty = 0.05
					autoEligible = true
					rationale = "Supported deterministic AST mutation: Single-file intra-procedural deferred cleanup."
				}

			case "SQL_INJECTION":
				evidence = []string{
					fmt.Sprintf("Taint Propagation: Dynamic SQL string constructed at %s:%d.", target.TargetFile, lineNo),
					"Sink: Database query execution function invoked with dynamically formatted query string.",
					"AST Mutator: Replace dynamic formatting with parameter placeholder binding.",
				}
				uncertainty = 0.08
				autoEligible = true
				rationale = "Supported deterministic AST mutation: SQL placeholder parameterization."

			case "UNHANDLED_ERROR":
				evidence = []string{
					fmt.Sprintf("Error Discarded: Return error ignored or assigned to blank identifier at %s:%d.", target.TargetFile, lineNo),
					"Control Flow: Propagating error requires modifying function return signature or error handling strategy.",
				}
				uncertainty = 0.60
				autoEligible = false
				rationale = "Unhandled error propagation requires signature or control flow changes. Human review required."

			default:
				evidence = []string{
					fmt.Sprintf("Diagnostic finding in %s:%d (%s).", target.TargetFile, lineNo, target.DefectCategory),
				}
				uncertainty = 0.50
				autoEligible = (target.ExpectedAction == "REPAIR")
				rationale = fmt.Sprintf("Candidate finding in category %s.", target.DefectCategory)
			}
		}
	}

	if !target.IsDemoTarget && target.ExpectedAction != "REPAIR" {
		autoEligible = false
		rationale = "Scanner requires manual review; automated repair is not eligible. " + rationale
	}

	res := &MVPAnalysisResult{
		AnalysisID:           analysisID,
		TargetID:             target.TargetID,
		Repository:           target.Repository,
		TargetFile:           target.TargetFile,
		DefectCategory:       target.DefectCategory,
		Severity:             "HIGH",
		LineNumber:           lineNo,
		CodeSnippet:          snippet,
		SemanticEvidence:     evidence,
		UncertaintyScore:     uncertainty,
		AutomaticEligibility: autoEligible,
		EligibilityRationale: rationale,
		RequiresHumanReview:  true,
		FindingID:            fmt.Sprintf("FIND-%s", target.TargetID),
		Timestamp:            time.Now().UTC().Format(time.RFC3339),
		SnapshotID:           target.SnapshotID,
		SourceFileHash:       target.SourceFileHash,
	}

	s.analyses[analysisID] = res
	return res, nil
}

// ProposeRepair synthesizes a patch via deterministic AST mutator or yields explicit abstention.
func (s *MVPService) ProposeRepair(analysisID string) (*MVPProposalResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	analysis, exists := s.analyses[analysisID]
	if !exists {
		return nil, fmt.Errorf("analysis %q not found", analysisID)
	}

	start := time.Now()
	proposalID := fmt.Sprintf("PROP-%s-%d", analysis.TargetID, time.Now().UnixNano()%1000000)

	var target *MVPTarget
	if dyn, ok := s.dynamicTargets[analysis.TargetID]; ok {
		target = dyn
	} else {
		for _, t := range DefaultPreconfiguredTargets() {
			if t.TargetID == analysis.TargetID {
				target = &t
				break
			}
		}
	}
	if target == nil {
		return nil, fmt.Errorf("target %q not recognized", analysis.TargetID)
	}

	// Create isolated ephemeral sandbox directory for this proposal
	tmpDir, _ := EnsureTmpDir(s.WorkspaceRoot)
	tmpSandbox, err := os.MkdirTemp(tmpDir, "mvp-sandbox-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create sandbox dir: %w", err)
	}

	var srcRepo string
	if target.WorkspaceID != "" {
		resolved, wsErr := s.resolveWorkspaceDirLocked(target.WorkspaceID)
		if wsErr != nil {
			os.RemoveAll(tmpSandbox)
			return nil, fmt.Errorf("cannot resolve workspace %s: %w", target.WorkspaceID, wsErr)
		}
		srcRepo = resolved
	} else {
		srcRepo = filepath.Join(s.WorkspaceRoot, filepath.FromSlash(target.RelativeSource))
		if fi, err := os.Stat(srcRepo); err != nil || !fi.IsDir() {
			altRoot := ResolveWorkspaceRoot()
			altRepo := filepath.Join(altRoot, filepath.FromSlash(target.RelativeSource))
			if altFi, altErr := os.Stat(altRepo); altErr == nil && altFi.IsDir() {
				srcRepo = altRepo
			} else {
				os.RemoveAll(tmpSandbox)
				return nil, fmt.Errorf("source repository not found at %s: %w", srcRepo, err)
			}
		}
	}
	if err := CopyDirectory(srcRepo, tmpSandbox); err != nil {
		os.RemoveAll(tmpSandbox)
		return nil, fmt.Errorf("failed to clone repo to sandbox: %w", err)
	}

	task := &planning.RemediationTask{
		TaskID:                fmt.Sprintf("TASK-MVP-%s", target.TargetID),
		Title:                 fmt.Sprintf("Remediate %s in %s", target.DefectCategory, target.TargetFile),
		Description:           target.Description,
		TargetFiles:           []string{target.TargetFile},
		ExecutionClass:        planning.ClassHumanApprovalRequired,
		HumanApprovalRequired: true,
	}
	finding := &repo_intel.Finding{
		FindingID:       analysis.FindingID,
		Category:        repo_intel.CategorySecurity,
		Title:           target.DefectCategory,
		AffectedSymbols: []string{target.Symbol},
		Evidence:        []repo_intel.EvidenceLocation{{File: target.TargetFile, LineStart: target.LineNumber}},
	}

	var decision *SynthesisDecision
	if !target.IsDemoTarget && !analysis.AutomaticEligibility {
		decision = &SynthesisDecision{Status: DecisionAbstained, Intent: &RepairIntent{IntentType: IntentAbstain, AbstainCategory: AbstainUnrecognizedPattern}, Reason: analysis.EligibilityRationale}
	} else {
		decision = s.Synthesizer.Synthesize(tmpSandbox, task, finding)
	}

	targetAbs := filepath.Join(tmpSandbox, filepath.FromSlash(target.TargetFile))
	fBytes, _ := os.ReadFile(targetAbs)
	fHash := sha256.Sum256(fBytes)
	fileHashStr := hex.EncodeToString(fHash[:])

	// If target has a pinned source file hash from scan, verify it hasn't mutated
	if target.SourceFileHash != "" && target.SourceFileHash != fileHashStr {
		os.RemoveAll(tmpSandbox)
		return nil, fmt.Errorf("STALE_WORKSPACE: source file %s was modified on disk after scan snapshot (expected: %s, found: %s)", target.TargetFile, target.SourceFileHash, fileHashStr)
	}

	// Verify snapshot record integrity and expiration
	if target.SnapshotID != "" {
		if err := ValidateSnapshotID(target.SnapshotID); err != nil {
			os.RemoveAll(tmpSandbox)
			return nil, err
		}
		snap, snapOk := s.snapshots[target.SnapshotID]
		if !snapOk || snap == nil {
			os.RemoveAll(tmpSandbox)
			return nil, fmt.Errorf("SNAPSHOT_MISSING: snapshot record %s does not exist", target.SnapshotID)
		}
		exp, parseErr := time.Parse(time.RFC3339, snap.ExpiresAt)
		if parseErr == nil && !exp.IsZero() && time.Now().UTC().After(exp) {
			os.RemoveAll(tmpSandbox)
			return nil, fmt.Errorf("SNAPSHOT_EXPIRED: snapshot record %s expired at %s", target.SnapshotID, snap.ExpiresAt)
		}
		if snap.WorkspaceID != target.WorkspaceID {
			os.RemoveAll(tmpSandbox)
			return nil, fmt.Errorf("SNAPSHOT_WORKSPACE_MISMATCH: snapshot %s is bound to workspace %s, not %s", target.SnapshotID, snap.WorkspaceID, target.WorkspaceID)
		}
		currRepoDigest, dErr := ComputeRepoDigest(srcRepo)
		if dErr != nil || currRepoDigest != snap.RootDigest {
			os.RemoveAll(tmpSandbox)
			return nil, fmt.Errorf("STALE_WORKSPACE: repository root digest mismatch against immutable scan snapshot %s (snapshot: %s, current: %s)", target.SnapshotID, snap.RootDigest, currRepoDigest)
		}
	}

	workspaceID := target.WorkspaceID
	if workspaceID == "" {
		workspaceID = fmt.Sprintf("ws-mvp-%s", target.TargetID)
	}

	var res *MVPProposalResult
	if decision.Status == DecisionAbstained {
		res = &MVPProposalResult{
			ProposalID:            proposalID,
			AnalysisID:            analysisID,
			TargetID:              target.TargetID,
			Status:                "ABSTAINED",
			AbstainCategory:       string(decision.Intent.AbstainCategory),
			AbstainReason:         decision.Reason,
			RequiresHumanApproval: false,
			ExecutionClass:        string(planning.ClassHumanApprovalRequired),
			SandboxDir:            tmpSandbox,
			DurationMs:            time.Since(start).Milliseconds(),
			SnapshotID:            target.SnapshotID,
		}
	} else if decision.Status == DecisionProposed && decision.Patch != nil {
		res = &MVPProposalResult{
			ProposalID:            proposalID,
			AnalysisID:            analysisID,
			TargetID:              target.TargetID,
			Status:                "PROPOSED",
			IntentType:            string(decision.Intent.IntentType),
			UnifiedDiff:           decision.Patch.UnifiedDiff,
			DiffHash:              decision.Patch.DiffHash,
			ModifiedSource:        decision.Patch.ModifiedSource,
			RequiresHumanApproval: true,
			ExecutionClass:        string(planning.ClassHumanApprovalRequired),
			FileHash:              fileHashStr,
			WorkspaceID:           workspaceID,
			TargetAbsPath:         targetAbs,
			SandboxDir:            tmpSandbox,
			DurationMs:            time.Since(start).Milliseconds(),
			SnapshotID:            target.SnapshotID,
		}
	} else {
		os.RemoveAll(tmpSandbox)
		return nil, fmt.Errorf("synthesis yielded unexpected status: %s (reason: %s)", decision.Status, decision.Reason)
	}

	s.proposals[proposalID] = res
	return res, nil
}

// ApproveRepair issues a cryptographic HMAC-SHA256 authorization token.
func (s *MVPService) ApproveRepair(proposalID, principal, role string) (*MVPApprovalResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	proposal, exists := s.proposals[proposalID]
	if !exists {
		return nil, fmt.Errorf("proposal %q not found", proposalID)
	}
	if proposal.Status != "PROPOSED" {
		return nil, fmt.Errorf("cannot approve proposal with status %q", proposal.Status)
	}

	// 1. Verify proposal diff integrity (anti-tamper)
	diffH := sha256.Sum256([]byte(proposal.UnifiedDiff))
	expectedDiffHash := hex.EncodeToString(diffH[:])
	if proposal.DiffHash != "" && proposal.DiffHash != expectedDiffHash {
		return nil, fmt.Errorf("PROPOSAL_TAMPERED: proposal diff hash mismatch (expected: %s, found: %s)", expectedDiffHash, proposal.DiffHash)
	}

	if principal == "" {
		principal = "operator@code-vitals.internal"
	}
	if role == "" {
		role = "Lead Reliability Engineer"
	}

	// 2. Cross-user workspace access check
	if proposal.SnapshotID != "" {
		snap := s.snapshots[proposal.SnapshotID]
		if snap == nil || snap.WorkspaceID != proposal.WorkspaceID {
			return nil, fmt.Errorf("SNAPSHOT_WORKSPACE_MISMATCH: proposal workspace does not match its snapshot")
		}
		if snap.OwnerID != "" && snap.OwnerID != principal {
			return nil, fmt.Errorf("CROSS_USER_WORKSPACE_FORBIDDEN: approval requires the workspace owner")
		}
	}

	fileHashes := map[string]string{
		filepath.Base(proposal.TargetAbsPath): proposal.FileHash,
	}

	token, err := s.Authority.IssueToken(
		principal,
		role,
		fmt.Sprintf("TASK-MVP-%s", proposal.TargetID),
		"TIER_2_CONTROLLED",
		planning.ExecutionClass(proposal.ExecutionClass),
		s.BaselineSHA,
		fileHashes,
		proposal.DiffHash,
		"APPLY_PATCH",
		proposal.WorkspaceID,
		30*time.Minute,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to issue cryptographic approval token: %w", err)
	}

	s.tokenObjs[token.TokenID] = token

	res := &MVPApprovalResult{
		TokenID:    token.TokenID,
		ProposalID: proposalID,
		Principal:  token.ApproverID,
		Role:       token.ApproverRole,
		DiffHash:   token.PatchDiffHash,
		RiskTier:   token.RiskTier,
		ExpiresAt:  token.ExpiresAt.Format(time.RFC3339),
		Signature:  token.HMACSignature,
		Valid:      true,
		IssuedAt:   token.IssuedAt.Format(time.RFC3339),
	}

	s.approvals[token.TokenID] = res
	return res, nil
}

// ExecuteRepair applies the patch in sandbox, consumes the token atomically, and runs 5 independent oracles.
func (s *MVPService) ExecuteRepair(tokenID string) (*MVPExecutionResult, error) {
	s.mu.Lock()

	start := time.Now()
	approval, exists := s.approvals[tokenID]
	if !exists {
		s.mu.Unlock()
		return nil, fmt.Errorf("approval token %q not found", tokenID)
	}

	proposal, exists := s.proposals[approval.ProposalID]
	if !exists {
		s.mu.Unlock()
		return nil, fmt.Errorf("associated proposal %q not found", approval.ProposalID)
	}
	if s.activeProposals == nil {
		s.activeProposals = make(map[string]bool)
	}
	if s.activeProposals[proposal.ProposalID] {
		s.mu.Unlock()
		return nil, fmt.Errorf("proposal %q is already executing", proposal.ProposalID)
	}
	tokenObj, ok := s.tokenObjs[tokenID]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("token object missing from authority memory")
	}

	var target *MVPTarget
	if dyn, ok := s.dynamicTargets[proposal.TargetID]; ok {
		target = dyn
	} else {
		for _, t := range DefaultPreconfiguredTargets() {
			if t.TargetID == proposal.TargetID {
				target = &t
				break
			}
		}
	}
	snap := s.snapshots[proposal.SnapshotID]
	s.activeProposals[proposal.ProposalID] = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.activeProposals, proposal.ProposalID)
		s.mu.Unlock()
	}()

	if target == nil {
		return nil, fmt.Errorf("target %q not recognized", proposal.TargetID)
	}

	// 1. Verify proposal diff integrity (anti-tamper)
	diffH := sha256.Sum256([]byte(proposal.UnifiedDiff))
	expectedDiffHash := hex.EncodeToString(diffH[:])
	if proposal.DiffHash != "" && proposal.DiffHash != expectedDiffHash {
		return nil, fmt.Errorf("PROPOSAL_TAMPERED: proposal diff hash mismatch (expected: %s, found: %s)", expectedDiffHash, proposal.DiffHash)
	}

	var origRepoPath string
	if target.WorkspaceID != "" {
		resolved, wsErr := s.resolveWorkspaceDirLocked(target.WorkspaceID)
		if wsErr != nil {
			return nil, fmt.Errorf("cannot resolve workspace %s: %w", target.WorkspaceID, wsErr)
		}
		origRepoPath = resolved
	} else {
		origRepoPath = filepath.Join(s.WorkspaceRoot, filepath.FromSlash(target.RelativeSource))
		if fi, err := os.Stat(origRepoPath); err != nil || !fi.IsDir() {
			altRoot := ResolveWorkspaceRoot()
			altRepo := filepath.Join(altRoot, filepath.FromSlash(target.RelativeSource))
			if altFi, altErr := os.Stat(altRepo); altErr == nil && altFi.IsDir() {
				origRepoPath = altRepo
			} else {
				return nil, fmt.Errorf("original repository directory %s does not exist: %w", origRepoPath, err)
			}
		}
	}
	srcBytes, err := readWorkspaceFile(origRepoPath, target.TargetFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read target source file %s: %w", target.TargetFile, err)
	}
	srcHash := sha256.Sum256(srcBytes)
	srcHashStr := hex.EncodeToString(srcHash[:])
	if proposal.FileHash != "" && srcHashStr != proposal.FileHash {
		return nil, fmt.Errorf("STALE_WORKSPACE: original target file %s has been modified on disk since proposal was generated (pre-hash: %s, current: %s)", target.TargetFile, proposal.FileHash, srcHashStr)
	}

	preDigest, err := ComputeRepoDigest(origRepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute original repo pre-digest: %w", err)
	}

	if target.SnapshotID != "" {
		if err := ValidateSnapshotID(target.SnapshotID); err != nil {
			return nil, err
		}
		if snap == nil {
			return nil, fmt.Errorf("SNAPSHOT_MISSING: snapshot record %s does not exist", target.SnapshotID)
		}
		exp, parseErr := time.Parse(time.RFC3339, snap.ExpiresAt)
		if parseErr == nil && !exp.IsZero() && time.Now().UTC().After(exp) {
			return nil, fmt.Errorf("SNAPSHOT_EXPIRED: snapshot record %s expired at %s", target.SnapshotID, snap.ExpiresAt)
		}
		if snap.WorkspaceID != target.WorkspaceID || snap.WorkspaceID != proposal.WorkspaceID {
			return nil, fmt.Errorf("SNAPSHOT_WORKSPACE_MISMATCH: snapshot workspace %s does not match proposal workspace %s", snap.WorkspaceID, proposal.WorkspaceID)
		}
		if snap.RootDigest != "" && preDigest != snap.RootDigest {
			return nil, fmt.Errorf("STALE_WORKSPACE: workspace tree digest mismatch against immutable scan snapshot %s (snapshot: %s, current: %s)", target.SnapshotID, snap.RootDigest, preDigest)
		}
	}

	task := &planning.RemediationTask{
		TaskID:                fmt.Sprintf("TASK-MVP-%s", proposal.TargetID),
		Title:                 fmt.Sprintf("Remediate %s in %s", target.DefectCategory, target.TargetFile),
		Description:           target.Description,
		TargetFiles:           []string{target.TargetFile},
		ExecutionClass:        planning.ExecutionClass(proposal.ExecutionClass),
		HumanApprovalRequired: true,
	}

	currentSource, err := readWorkspaceFile(proposal.SandboxDir, target.TargetFile)
	if err != nil {
		return nil, fmt.Errorf("read approved file: %w", err)
	}
	currentHash := sha256.Sum256(currentSource)
	approvedFileHash := hex.EncodeToString(currentHash[:])
	if approvedFileHash != proposal.FileHash {
		return nil, fmt.Errorf("STALE_WORKSPACE: approved sandbox file was modified on disk")
	}
	fileHashes := map[string]string{filepath.Base(proposal.TargetAbsPath): approvedFileHash}
	eligible, _, rejectReason := CheckTaskEligibilitySecure(
		task,
		tokenObj,
		s.Authority,
		s.BaselineSHA,
		fileHashes,
		proposal.DiffHash,
		proposal.WorkspaceID,
	)
	if !eligible {
		return nil, fmt.Errorf("execution boundary rejected token: %s", rejectReason)
	}

	// 2. Token is atomically verified and consumed via CheckTaskEligibilitySecure -> Authority.VerifyAndConsumeToken.

	// Apply patch inside sandbox ONLY
	if err := agent.AtomicWriteFileWithConflictDetection(proposal.TargetAbsPath, []byte(proposal.ModifiedSource), proposal.FileHash, 0644); err != nil {
		return nil, fmt.Errorf("failed to write patched file to sandbox: %w", err)
	}

	// Evaluate 5 Independent Oracles
	patchObj := &CandidatePatch{
		UnifiedDiff:    proposal.UnifiedDiff,
		DiffHash:       proposal.DiffHash,
		ModifiedSource: proposal.ModifiedSource,
		TargetFile:     target.TargetFile,
	}

	oracleReport := s.Oracles.VerifyAll(proposal.SandboxDir, task, patchObj, "", "")

	oracleOutcomes := make(map[string]MVPOracleOutcome)
	hasFailure := false
	hasUnavailable := false

	for oType, oRes := range oracleReport.OracleResults {
		oracleOutcomes[string(oType)] = MVPOracleOutcome{
			OracleType:    string(oType),
			Status:        string(oRes.Status),
			Description:   oRes.Details,
			Passed:        oRes.Passed,
			Details:       oRes.FailureReason,
			ContainerInfo: oRes.ContainerInfo,
		}
		if oRes.Status == OracleStatusFailed {
			hasFailure = true
		}
		if oRes.Status == OracleStatusUnavailable {
			hasUnavailable = true
		}
	}

	// Verify Original Repository Immutability
	postDigest, err := ComputeRepoDigest(origRepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to compute original repo post-digest: %w", err)
	}
	repoImmutable := (preDigest == postDigest)
	if !repoImmutable {
		return nil, fmt.Errorf("CRITICAL SAFETY FAILURE: original repository modified on disk!")
	}

	executionID := fmt.Sprintf("EXEC-%s-%d", proposal.TargetID, time.Now().UnixNano()%1000000)
	verdict := "FAILED"

	if hasFailure {
		verdict = "FAILED"
	} else {
		defectRep, hasRep := oracleReport.OracleResults[OracleDefectReproducer]
		dynamicReproduced := hasRep && defectRep.Status == OracleStatusPassed && strings.Contains(defectRep.Details, "isolated container")

		if dynamicReproduced && oracleReport.AllPassed && !hasUnavailable {
			verdict = "INDEPENDENTLY_VERIFIED"
		} else {
			verdict = "SYNTAX_ONLY_VERIFIED"
		}
	}

	res := &MVPExecutionResult{
		ExecutionID:            executionID,
		ProposalID:             proposal.ProposalID,
		TargetID:               proposal.TargetID,
		TokenID:                tokenID,
		SandboxDir:             proposal.SandboxDir,
		AppliedInSandbox:       true,
		TokenConsumed:          true,
		Oracles:                oracleOutcomes,
		AllOraclesPassed:       oracleReport.AllPassed,
		VerificationVerdict:    verdict,
		OriginalRepoPreDigest:  preDigest,
		OriginalRepoPostDigest: postDigest,
		OriginalRepoImmutable:  repoImmutable,
		RollbackAvailable:      true,
		CleanupOperation:       "DISCARD_SANDBOX",
		DurationMs:             time.Since(start).Milliseconds(),
		Timestamp:              time.Now().UTC().Format(time.RFC3339),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.executions[executionID] = res

	// Append to audit trail
	auditRec := MVPAuditRecord{
		RecordID:       fmt.Sprintf("AUDIT-%s-%d", proposal.TargetID, time.Now().UnixNano()%1000000),
		TargetID:       proposal.TargetID,
		Repository:     target.Repository,
		DefectCategory: target.DefectCategory,
		Analysis:       s.analyses[proposal.AnalysisID],
		Proposal:       proposal,
		Approval:       approval,
		Execution:      res,
		FinalStatus:    verdict,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
	s.auditTrail = append(s.auditTrail, auditRec)

	return res, nil
}

// RecordAbstention appends an explicit abstention decision to the audit trail.
func (s *MVPService) RecordAbstention(analysisID, proposalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	analysis, ok := s.analyses[analysisID]
	if !ok {
		return fmt.Errorf("analysis %q not found", analysisID)
	}
	prop, ok := s.proposals[proposalID]
	if !ok {
		return fmt.Errorf("proposal %q not found", proposalID)
	}

	rec := MVPAuditRecord{
		RecordID:       fmt.Sprintf("AUDIT-ABSTAIN-%s-%d", prop.TargetID, time.Now().UnixNano()%1000000),
		TargetID:       prop.TargetID,
		Repository:     analysis.Repository,
		DefectCategory: analysis.DefectCategory,
		Analysis:       analysis,
		Proposal:       prop,
		FinalStatus:    "EXPLICITLY_ABSTAINED",
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	}
	s.auditTrail = append(s.auditTrail, rec)
	return nil
}

// RollbackSandbox discards the sandbox workspace and records the rollback.
func (s *MVPService) RollbackSandbox(executionID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exec, exists := s.executions[executionID]
	if !exists {
		return false, fmt.Errorf("execution record %q not found", executionID)
	}
	if exec.SandboxDir == "" {
		return false, fmt.Errorf("DISCARD_SANDBOX: missing sandbox path")
	}
	root, err := EnsureTmpDir(s.WorkspaceRoot)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(root, exec.SandboxDir)
	if err != nil || filepath.IsAbs(rel) || strings.HasPrefix(rel, "..") || filepath.Dir(rel) != "." || !strings.HasPrefix(filepath.Base(rel), "mvp-sandbox-") {
		return false, fmt.Errorf("DISCARD_SANDBOX: path is outside managed sandbox root")
	}
	exec.CleanupOperation = "DISCARD_SANDBOX"
	if err := os.RemoveAll(exec.SandboxDir); err != nil {
		exec.CleanupStatus = "FAILED"
		return false, fmt.Errorf("DISCARD_SANDBOX: %w", err)
	}
	if _, err := os.Lstat(exec.SandboxDir); !os.IsNotExist(err) {
		exec.CleanupStatus = "FAILED"
		return false, fmt.Errorf("DISCARD_SANDBOX: deletion not verified: %v", err)
	}
	exec.CleanupStatus = "PASSED"
	exec.RollbackAvailable = false
	return true, nil
}

// ExportAuditReport returns the full machine-readable audit report.
func (s *MVPService) ExportAuditReport() *MVPAuditReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	verifiedCount := 0
	abstainedCount := 0
	failedCount := 0

	for _, rec := range s.auditTrail {
		if rec.FinalStatus == "INDEPENDENTLY_VERIFIED" {
			verifiedCount++
		} else if rec.FinalStatus == "EXPLICITLY_ABSTAINED" || (rec.Proposal != nil && rec.Proposal.Status == "ABSTAINED") {
			abstainedCount++
		} else {
			failedCount++
		}
	}

	return &MVPAuditReport{
		BaselineCommit:        s.BaselineSHA,
		ExportTimestamp:       time.Now().UTC().Format(time.RFC3339),
		TotalSessionsRecorded: len(s.auditTrail),
		VerifiedRepairsCount:  verifiedCount,
		AbstainedRepairsCount: abstainedCount,
		FailedRepairsCount:    failedCount,
		SourceImmutability100: true,
		AuditRecords:          s.auditTrail,
	}
}

// RunDeterministicDemo runs a repair candidate and an abstention case.
func (s *MVPService) RunDeterministicDemo() (map[string]interface{}, error) {
	// 1. Injected resource leak repair candidate.
	analysis1, err := s.AnalyzeTarget("INJ-EXT2-RES-02")
	if err != nil {
		return nil, fmt.Errorf("demo analysis failed: %w", err)
	}

	prop1, err := s.ProposeRepair(analysis1.AnalysisID)
	if err != nil {
		return nil, fmt.Errorf("demo proposal failed: %w", err)
	}

	appr1, err := s.ApproveRepair(prop1.ProposalID, "demo-engineer@code-vitals.internal", "Principal Reliability Engineer")
	if err != nil {
		return nil, fmt.Errorf("demo approval failed: %w", err)
	}

	exec1, err := s.ExecuteRepair(appr1.TokenID)
	if err != nil {
		return nil, fmt.Errorf("demo execution failed: %w", err)
	}

	// 2. Conservative Abstention Demo (INJ-EXT2-CMD-01: schollz/progressbar)
	analysis2, err := s.AnalyzeTarget("INJ-EXT2-CMD-01")
	if err != nil {
		return nil, fmt.Errorf("demo abstention analysis failed: %w", err)
	}

	prop2, err := s.ProposeRepair(analysis2.AnalysisID)
	if err != nil {
		return nil, fmt.Errorf("demo abstention proposal failed: %w", err)
	}

	// Record abstention in audit trail
	s.mu.Lock()
	s.auditTrail = append(s.auditTrail, MVPAuditRecord{
		RecordID:       fmt.Sprintf("AUDIT-ABSTAIN-%d", time.Now().UnixNano()%1000000),
		TargetID:       "INJ-EXT2-CMD-01",
		Repository:     "schollz_progressbar",
		DefectCategory: "COMMAND_INJECTION",
		Analysis:       analysis2,
		Proposal:       prop2,
		FinalStatus:    "EXPLICITLY_ABSTAINED",
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
	})
	s.mu.Unlock()

	auditReport := s.ExportAuditReport()

	return map[string]interface{}{
		"status": "DEMO_COMPLETED",
		"repair_candidate": map[string]interface{}{
			"target":               "INJ-EXT2-RES-02 (fatih/color)",
			"defect":               "Injected resource leak",
			"unified_diff":         prop1.UnifiedDiff,
			"diff_hash":            prop1.DiffHash,
			"approval_token":       appr1.TokenID,
			"all_oracles_passed":   exec1.AllOraclesPassed,
			"verification_verdict": exec1.VerificationVerdict,
			"oracles_evaluated":    exec1.Oracles,
			"source_immutable":     exec1.OriginalRepoImmutable,
		},
		"conservative_abstention": map[string]interface{}{
			"target":           "INJ-EXT2-CMD-01 (schollz/progressbar)",
			"defect":           "Command Injection (sh -c)",
			"abstain_status":   prop2.Status,
			"abstain_category": prop2.AbstainCategory,
			"abstain_reason":   prop2.AbstainReason,
		},
		"disqualified_natural_baseline": map[string]interface{}{
			"target":      "NAT-03 (gorilla/websocket)",
			"disposition": "GROUND_TRUTH_DISQUALIFIED",
			"rationale":   "DialContext manages socket in defer at line 330; response body does not require application close. 0 confirmed natural repairs.",
		},
		"audit_summary": map[string]interface{}{
			"total_sessions":  auditReport.TotalSessionsRecorded,
			"verified_count":  auditReport.VerifiedRepairsCount,
			"abstained_count": auditReport.AbstainedRepairsCount,
			"immutability":    auditReport.SourceImmutability100,
		},
	}, nil
}

func readWorkspaceFile(dir, relativePath string) ([]byte, error) {
	if err := ValidatePathSafety(relativePath); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	return readRegularFile(root, filepath.FromSlash(relativePath))
}

func readRegularFile(root *os.Root, path string) ([]byte, error) {
	info, err := root.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("refusing non-regular workspace file %q", path)
	}
	return root.ReadFile(path)
}

// ComputeRepoDigest walks dir and computes SHA-256 over relative path and contents.
func ComputeRepoDigest(dir string) (string, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	h := sha256.New()
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "node_modules" || base == "vendor" || base == ".tmp" || base == "uploads" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		data, err := readRegularFile(root, rel)
		if err != nil {
			return err
		}
		h.Write([]byte(filepath.ToSlash(rel)))
		h.Write(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CopyDirectory recursively copies srcDir to dstDir.
func CopyDirectory(srcDir, dstDir string) error {
	root, err := os.OpenRoot(srcDir)
	if err != nil {
		return err
	}
	defer root.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "node_modules" || base == "vendor" || base == ".tmp" || base == "uploads" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0755)
		}
		data, err := readRegularFile(root, rel)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}
