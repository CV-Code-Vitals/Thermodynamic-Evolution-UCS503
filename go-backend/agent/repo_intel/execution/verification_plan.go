package execution

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// VerificationCheck is deliberately narrower than the general subprocess policy.
// No caller-controlled executable, flags, environment or shell is supported.
type VerificationCheck struct {
	Package string `json:"package"`
	Test    string `json:"test"`
}

func (c VerificationCheck) Command() (string, error) {
	if !regexp.MustCompile(`^(\.|\./[A-Za-z0-9_/-]+)$`).MatchString(c.Package) || strings.Contains(c.Package, "..") && c.Package != "." || !regexp.MustCompile(`^Test[A-Za-z0-9_]+$`).MatchString(c.Test) {
		return "", fmt.Errorf("SECURITY_REJECTION: invalid structured check")
	}
	return "go test -count=1 -v -run ^" + c.Test + "$ " + c.Package, nil
}

type VerificationPlan struct {
	ModuleDir          string              `json:"module_dir"`
	Repository         string              `json:"repository"`
	Revision           string              `json:"revision"`
	Workspace          string              `json:"workspace"`
	Snapshot           string              `json:"snapshot"`
	RootDigest         string              `json:"root_digest"`
	SourceHash         string              `json:"source_hash"`
	PatchDigest        string              `json:"patch_digest"`
	CandidateHash      string              `json:"candidate_hash"`
	Reproducer         VerificationCheck   `json:"reproducer"`
	Regression         []VerificationCheck `json:"regression"`
	Behavioral         VerificationCheck   `json:"behavioral"`
	TestFiles          map[string]string   `json:"test_files"`
	FailureAssertion   string              `json:"failure_assertion"`
	BeforeExit         int                 `json:"before_exit"`
	AfterExit          int                 `json:"after_exit"`
	NegativeProperties string              `json:"negative_properties"`
	Issuer             string              `json:"issuer"`
	Provenance         string              `json:"provenance"`
	Limits             ContainerLimits     `json:"limits"`
}

func (p VerificationPlan) Digest() string {
	b, _ := json.Marshal(p)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func sourceSHA(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func (p VerificationPlan) validate(dir string) error {
	if p.Issuer == "" || p.Provenance == "" || p.Repository == "" || p.Revision == "" || p.Snapshot == "" || p.RootDigest == "" || p.FailureAssertion == "" || p.BeforeExit != 1 || p.AfterExit != 0 || len(p.Regression) == 0 || len(p.TestFiles) == 0 || p.NegativeProperties != "AST_SAFETY_V1" || p.Limits != DefaultContainerLimits() {
		return fmt.Errorf("SECURITY_REJECTION: incomplete verification plan")
	}
	checks := append([]VerificationCheck{p.Reproducer, p.Behavioral}, p.Regression...)
	for _, c := range checks {
		if _, e := c.Command(); e != nil {
			return e
		}
	}
	for path, hash := range p.TestFiles {
		if !strings.HasSuffix(path, "_test.go") {
			return fmt.Errorf("SECURITY_REJECTION: reproducer must be a test file")
		}
		b, e := readWorkspaceFile(dir, path)
		if e != nil || sourceSHA(b) != hash {
			return fmt.Errorf("SECURITY_REJECTION: independent test changed: %s", path)
		}
	}
	_, err := p.modulePath(dir)
	return err
}

// AttachTrustedVerificationPlan is an operator-only in-process provisioning API.
// It is intentionally NOT exposed as a REST route or populated from repository metadata.
// Issuers must independently review test semantics before calling this method.
func (s *MVPService) AttachTrustedVerificationPlan(proposalID string, plan VerificationPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.proposals[proposalID]
	if p == nil {
		return fmt.Errorf("proposal missing")
	}
	for _, a := range s.approvals {
		if a.ProposalID == proposalID {
			return fmt.Errorf("SECURITY_REJECTION: already approved")
		}
	}
	snap := s.snapshots[p.SnapshotID]
	if snap == nil {
		return fmt.Errorf("snapshot missing")
	}
	if plan.Workspace != p.WorkspaceID || plan.Snapshot != p.SnapshotID || plan.RootDigest != snap.RootDigest || plan.SourceHash != p.FileHash || plan.PatchDigest != p.DiffHash || plan.CandidateHash != sourceSHA([]byte(p.ModifiedSource)) || plan.Revision != snap.SourceCommit || plan.Repository != snap.Repository {
		return fmt.Errorf("SECURITY_REJECTION: plan scope mismatch")
	}
	if e := plan.validate(p.SandboxDir); e != nil {
		return e
	}
	b, _ := json.Marshal(plan)
	var frozen VerificationPlan
	_ = json.Unmarshal(b, &frozen)
	p.VerificationPlan = &frozen
	p.VerificationPlanDigest = frozen.Digest()
	return nil
}
func validateProposalPlan(p *MVPProposalResult) error {
	if p.VerificationPlan == nil {
		return fmt.Errorf("REPRODUCER_NOT_AVAILABLE")
	}
	v := p.VerificationPlan
	if v.Digest() != p.VerificationPlanDigest || v.PatchDigest != p.DiffHash || v.SourceHash != p.FileHash || v.Workspace != p.WorkspaceID || v.Snapshot != p.SnapshotID || v.CandidateHash != sourceSHA([]byte(p.ModifiedSource)) {
		return fmt.Errorf("SECURITY_REJECTION: verification plan or candidate changed")
	}
	return v.validate(p.SandboxDir)
}

type ReproductionEvidence struct {
	SourceDigest       string              `json:"source_digest"`
	Command            string              `json:"command"`
	Result             *ContainerRunResult `json:"result"`
	Status             string              `json:"status"`
	BehavioralBaseline *ContainerRunResult `json:"behavioral_baseline,omitempty"`
	CleanupStatus      string              `json:"cleanup_status"`
}

func (s *MVPService) reproduceBefore(p *MVPProposalResult) (evidence *ReproductionEvidence, proofErr error) {
	v := p.VerificationPlan
	cmd, _ := v.Reproducer.Command()
	ev := &ReproductionEvidence{Command: cmd, Status: "DYNAMIC_VERIFICATION_UNAVAILABLE"}
	ctx, cancel := context.WithTimeout(context.Background(), v.Limits.Timeout)
	defer cancel()
	if s.Runner == nil {
		return ev, fmt.Errorf("DYNAMIC_VERIFICATION_UNAVAILABLE: runner missing")
	}
	if e := s.Runner.CheckAvailable(ctx); e != nil {
		return ev, fmt.Errorf("DYNAMIC_VERIFICATION_UNAVAILABLE: %w", e)
	}
	root, e := EnsureTmpDir(s.WorkspaceRoot)
	if e != nil {
		return ev, e
	}
	dir, e := os.MkdirTemp(root, "mvp-baseline-")
	if e != nil {
		return ev, e
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			ev.CleanupStatus = err.Error()
			proofErr = fmt.Errorf("EXECUTION_FAILED: baseline disposal: %w", err)
			return
		}
		if _, err := os.Lstat(dir); !os.IsNotExist(err) {
			ev.CleanupStatus = "FAILED"
			proofErr = fmt.Errorf("EXECUTION_FAILED: baseline disposal unverified")
			return
		}
		ev.CleanupStatus = "VERIFIED_REMOVAL"
	}()
	if e = CopyDirectory(p.SandboxDir, dir); e != nil {
		return ev, e
	}
	ev.SourceDigest, e = ComputeRepoDigest(dir)
	if e != nil {
		return ev, e
	}
	if ev.SourceDigest != v.RootDigest {
		return ev, fmt.Errorf("SECURITY_REJECTION: baseline digest mismatch")
	}
	moduleDir, e := v.modulePath(dir)
	if e != nil {
		return ev, e
	}
	ev.Result, e = s.Runner.RunCommand(ctx, moduleDir, cmd)
	r := ev.Result
	// A Go testing assertion and completed container are mandatory; compile/dependency failures do not reproduce defects.
	if e != nil || r == nil || !validContainerEvidence(r.Evidence, 1) || r.ExitCode != 1 || r.Evidence.TimeoutCause != "" || !strings.Contains(r.Output, "--- FAIL: "+v.Reproducer.Test+" (") || !strings.Contains(r.Output, v.FailureAssertion) || strings.Contains(r.Output, "[build failed]") || isOfflineDependencyMissing(r.Output) {
		ev.Status = "EXECUTION_FAILED"
		return ev, fmt.Errorf("EXECUTION_FAILED: expected defect assertion not established: %v", e)
	}
	if e = v.validate(dir); e != nil {
		return ev, e
	}
	behaviorCmd, _ := v.Behavioral.Command()
	baselineRunner := &proofRunner{ContainerRunner: s.Runner}
	ev.BehavioralBaseline, e = baselineRunner.RunCommand(ctx, moduleDir, behaviorCmd)
	if e != nil || ev.BehavioralBaseline == nil || ev.BehavioralBaseline.ExitCode != 0 {
		return ev, fmt.Errorf("BEHAVIORAL_INVARIANCE_FAILED: baseline: %v", e)
	}
	afterDigest, e := ComputeRepoDigest(dir)
	if e != nil || afterDigest != v.RootDigest {
		return ev, fmt.Errorf("SECURITY_REJECTION: baseline source mutated")
	}
	ev.Status = "PASSED"
	return ev, nil
}
func genuineDocker(r ContainerRunner) bool {
	d, ok := r.(*DockerContainerRunner)
	return ok && d.ExecCommand == nil && d.Checker == nil && d.Limits == DefaultContainerLimits()
}

// ConfigureTrustedVerificationPlans is bootstrap-only. Its input is operator configuration,
// never a request body or a file from the scanned repository. Templates are pinned to
// repository content; only proposal-specific identities and generated patch hashes are filled.
func (s *MVPService) ConfigureTrustedVerificationPlans(plans []VerificationPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.proposals) > 0 {
		return fmt.Errorf("SECURITY_REJECTION: configure plans before serving proposals")
	}
	b, err := json.Marshal(plans)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &s.trustedPlans)
}
func (s *MVPService) provisionPlanLocked(p *MVPProposalResult) error {
	snap := s.snapshots[p.SnapshotID]
	if snap == nil {
		return nil
	}
	for _, template := range s.trustedPlans {
		if template.RootDigest != snap.RootDigest || template.SourceHash != p.FileHash || template.Revision != snap.SourceCommit || template.Repository != snap.Repository {
			continue
		}
		v := template
		v.Workspace = p.WorkspaceID
		v.Snapshot = p.SnapshotID
		v.PatchDigest = p.DiffHash
		v.CandidateHash = sourceSHA([]byte(p.ModifiedSource))
		if err := v.validate(p.SandboxDir); err != nil {
			return err
		}
		b, _ := json.Marshal(v)
		var frozen VerificationPlan
		_ = json.Unmarshal(b, &frozen)
		p.VerificationPlan = &frozen
		p.VerificationPlanDigest = frozen.Digest()
		return nil
	}
	return nil
}

// Synthetic scope keys are signed as part of the authority's existing canonical
// file-hash map. They cannot collide with filepath.Base(target), and include the
// entire command set, issuer, limits, root digest and candidate bytes transitively.
func approvalScope(p *MVPProposalResult, hash string) map[string]string {
	return map[string]string{filepath.Base(p.TargetAbsPath): hash, "@verification/manifest": p.VerificationPlanDigest}
}
func completeDynamicProof(r ContainerRunner, p *MVPProposalResult, b *ReproductionEvidence, report *MultiOracleReport) bool {
	if !genuineDocker(r) || b == nil || b.Status != "PASSED" || !validContainerEvidence(b.Result.Evidence, 1) || !report.AllPassed {
		return false
	}
	for _, kind := range []OracleType{OracleSyntaxCompilation, OracleRegressionTest, OracleDefectReproducer, OracleBehavioralInvariance, OracleNegativeProperties} {
		o := report.OracleResults[kind]
		if o == nil || !o.Passed || o.Status != OracleStatusPassed {
			return false
		}
		if kind != OracleNegativeProperties && !validContainerEvidence(o.ContainerInfo, 0) {
			return false
		}
	}
	return true
}
func validContainerEvidence(e *ContainerExecutionEvidence, exit int) bool {
	return e != nil && e.ContainerID != "" && e.ImageDigest != "" && e.ExitStatus == exit && e.TimeoutCause == "" && strings.HasPrefix(e.CleanupResult, "VERIFIED_REMOVAL:") && strings.HasPrefix(e.IsolationPolicy, "DOCKER_HARDENED_CONTAINER") && e.NetworkPolicy == "DISABLED (--network none)" && e.RootFSPolicy == "READ_ONLY (--read-only)" && e.MemoryLimitMB == DefaultContainerLimits().MaxMemoryMB && e.CPULimit == DefaultContainerLimits().MaxCPUs
}
func failureVerdict(err error) string {
	msg := err.Error()
	for _, v := range []string{"REPRODUCER_NOT_AVAILABLE", "DYNAMIC_VERIFICATION_UNAVAILABLE", "REGRESSION_FAILED", "BEHAVIORAL_INVARIANCE_FAILED"} {
		if strings.Contains(msg, v) {
			return v
		}
	}
	if strings.Contains(msg, "SECURITY") || strings.Contains(msg, "STALE") || strings.Contains(msg, "TAMPER") || strings.Contains(msg, "token") || strings.Contains(msg, "SNAPSHOT") {
		return "SECURITY_REJECTION"
	}
	return "EXECUTION_FAILED"
}
func (s *MVPService) recordExecutionFailure(tokenID string, err error, b *ReproductionEvidence, consumed, applied bool) *MVPExecutionResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := failureVerdict(err)
	res := &MVPExecutionResult{ExecutionID: fmt.Sprintf("EXEC-REJECT-%d", time.Now().UnixNano()), TokenID: tokenID, VerificationVerdict: v, FailureReason: err.Error(), TokenConsumed: consumed, AppliedInSandbox: applied, BeforeReproduction: b, EvidenceClass: "NOT_DYNAMICALLY_VERIFIED", Timestamp: time.Now().UTC().Format(time.RFC3339)}
	a := s.approvals[tokenID]
	var p *MVPProposalResult
	if a != nil {
		p = s.proposals[a.ProposalID]
	}
	if p != nil {
		res.ProposalID = p.ProposalID
		res.TargetID = p.TargetID
		res.SandboxDir = p.SandboxDir
		res.CleanupOperation = "DISCARD_SANDBOX"
		res.RollbackAvailable = true
		res.VerificationPlanDigest = p.VerificationPlanDigest
		if snap := s.snapshots[p.SnapshotID]; snap != nil {
			res.OriginalRepoPreDigest = snap.RootDigest
			if dir, e := s.resolveWorkspaceDirLocked(p.WorkspaceID); e == nil {
				res.OriginalRepoPostDigest, _ = ComputeRepoDigest(dir)
				res.OriginalRepoImmutable = res.OriginalRepoPreDigest != "" && res.OriginalRepoPreDigest == res.OriginalRepoPostDigest
			}
		}
		res.CandidateDigest, _ = ComputeRepoDigest(p.SandboxDir)
	}
	s.executions[res.ExecutionID] = res
	s.auditTrail = append(s.auditTrail, MVPAuditRecord{RecordID: res.ExecutionID, TargetID: res.TargetID, Proposal: p, Approval: a, Execution: res, FinalStatus: v, Timestamp: res.Timestamp})
	return res
}

// proofRunner rejects successful commands that did not actually run the named test.
// It records every invocation, including each regression command, without replacing
// the hardened runtime. Doubles can exercise this gate but cannot certify a repair.
type proofRunner struct {
	ContainerRunner
	Records []*ContainerRunResult
}

func (r *proofRunner) RunCommand(ctx context.Context, dir, cmd string) (*ContainerRunResult, error) {
	result, err := r.ContainerRunner.RunCommand(ctx, dir, cmd)
	r.Records = append(r.Records, result)
	if result == nil {
		return nil, fmt.Errorf("EXECUTION_FAILED: missing execution result")
	}
	if err != nil {
		return result, err
	}
	if result.ExitCode == 0 && strings.HasPrefix(cmd, "go test ") {
		fields := strings.Fields(cmd)
		test := ""
		for i, f := range fields {
			if f == "-run" && i+1 < len(fields) {
				test = strings.Trim(fields[i+1], "^$")
			}
		}
		if test == "" || !strings.Contains(result.Output, "--- PASS: "+test+" (") || strings.Contains(result.Output, "--- SKIP: "+test+" (") {
			result.ExitCode = -1
			return result, fmt.Errorf("EXECUTION_FAILED: named test did not pass")
		}
	}
	if !validContainerEvidence(result.Evidence, result.ExitCode) {
		result.ExitCode = -1
		return result, fmt.Errorf("EXECUTION_FAILED: incomplete isolation or cleanup evidence")
	}
	return result, nil
}

func (p VerificationPlan) modulePath(root string) (string, error) {
	if p.ModuleDir == "" || p.ModuleDir == "." {
		if _, err := readWorkspaceFile(root, "go.mod"); err != nil {
			return "", fmt.Errorf("SECURITY_REJECTION: approved go.mod missing")
		}
		return root, nil
	}
	if !regexp.MustCompile(`^[A-Za-z0-9_/-]+$`).MatchString(p.ModuleDir) || filepath.IsAbs(p.ModuleDir) || strings.Contains(p.ModuleDir, "..") {
		return "", fmt.Errorf("SECURITY_REJECTION: unsafe module path")
	}
	if _, err := readWorkspaceFile(root, filepath.Join(p.ModuleDir, "go.mod")); err != nil {
		return "", err
	}
	return filepath.Join(root, filepath.FromSlash(p.ModuleDir)), nil
}
