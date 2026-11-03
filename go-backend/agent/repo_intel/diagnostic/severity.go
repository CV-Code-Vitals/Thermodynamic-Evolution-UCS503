package diagnostic

import (
	"fmt"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent/repo_intel"
)

// SeverityAssessor evaluates findings independently of complexity and assigns
// defensible severity classifications with clear impact rationales.
type SeverityAssessor struct{}

// NewSeverityAssessor creates a new SeverityAssessor.
func NewSeverityAssessor() *SeverityAssessor {
	return &SeverityAssessor{}
}

// AssessSeverityAndImpact evaluates functional, security, and architectural blast radius.
func (sa *SeverityAssessor) AssessSeverityAndImpact(f *ValidatedFinding) (repo_intel.Severity, string, string, []string) {
	var sev repo_intel.Severity
	var rationale string
	var impactScope string
	var potentialConsequences []string

	switch f.Category {
	case repo_intel.CategorySecurity:
		if strings.Contains(f.FindingID, "SECRET") || strings.Contains(f.FindingID, "KEY") {
			sev = repo_intel.SeverityCritical
			rationale = "Hardcoded credentials or private keys risk immediate unauthorized account access"
			impactScope = "Cloud Infrastructure & Auth Boundaries"
			potentialConsequences = []string{"Data breach", "Lateral cloud privilege escalation", "Credential leakage in git history"}
		} else if strings.Contains(f.FindingID, "CMD-INJECT") {
			sev = repo_intel.SeverityHigh
			rationale = "Dynamic argument formatting into process execution enables arbitrary command execution"
			impactScope = "Host Operating System & Container Sandbox"
			potentialConsequences = []string{"Remote code execution", "Sandbox escape", "Host filesystem compromise"}
		} else if strings.Contains(f.FindingID, "SQL") {
			sev = repo_intel.SeverityHigh
			rationale = "Dynamic SQL query string concatenation risks unauthenticated database extraction"
			impactScope = "Database Layer & Data Confidentiality"
			potentialConsequences = []string{"Unauthorized data access", "SQL injection exploitation", "Database corruption"}
		} else if strings.Contains(f.FindingID, "DOCKER-ROOT") {
			sev = repo_intel.SeverityMedium
			rationale = "Absence of non-root USER instruction allows container process to execute as root"
			impactScope = "Container Isolation & Host Boundary"
			potentialConsequences = []string{"Privilege escalation in container break-out scenarios"}
		} else {
			sev = repo_intel.SeverityLow
			rationale = "Minor security hygiene finding without immediate exploitability"
			impactScope = "Component Configuration"
			potentialConsequences = []string{"Sub-optimal container reproducibility"}
		}

	case repo_intel.CategoryArchitecture:
		if strings.Contains(f.FindingID, "CYCLE") {
			sev = repo_intel.SeverityHigh
			rationale = "Circular dependency graph creates tight coupling, prevents modular builds, and risks runtime init deadlocks"
			impactScope = "Cross-Module Build & Runtime Architecture"
			potentialConsequences = []string{"Package initialization deadlock", "Inability to decouple or refactor modules independently", "Unbounded compilation overhead"}
		} else {
			sev = repo_intel.SeverityMedium
			rationale = "Service boundary violation impairs architectural separation of concerns"
			impactScope = "Inter-Service Interface Boundaries"
			potentialConsequences = []string{"Leaky module abstractions", "Hidden architectural coupling"}
		}

	case repo_intel.CategoryStructural:
		if strings.Contains(f.FindingID, "PARSE-ERROR") {
			sev = repo_intel.SeverityHigh
			rationale = "Syntax parse errors prevent compiler analysis and may break downstream builds"
			impactScope = "Source Compilation Unit"
			potentialConsequences = []string{"Compilation failure", "Incomplete static analysis coverage"}
		} else if strings.Contains(f.FindingID, "HIGH-COMPLEXITY") {
			sev = repo_intel.SeverityMedium
			rationale = "High cyclomatic complexity increases cognitive load and branch testing surface"
			impactScope = "Unit Testability & Local Function Maintenance"
			potentialConsequences = []string{"High defect density", "Uncovered edge-case branches", "Difficult regression verification"}
		} else if strings.Contains(f.FindingID, "HIGH-NESTING") {
			sev = repo_intel.SeverityMedium
			rationale = "Excessive control-flow nesting impairs code comprehension and refactoring safety"
			impactScope = "Local Control Flow Structure"
			potentialConsequences = []string{"Cognitive overload", "Missed error handling pathways"}
		} else {
			sev = repo_intel.SeverityLow
			rationale = "Monolithic AST node volume indicates multi-responsibility implementation"
			impactScope = "Function Modularity"
			potentialConsequences = []string{"Reduced code reusability"}
		}

	case repo_intel.CategoryReliability:
		if strings.Contains(f.FindingID, "UNHANDLED-ERR") {
			sev = repo_intel.SeverityMedium
			rationale = "Explicitly discarded error return suppresses failure signals at runtime"
			impactScope = "Runtime Resilience & Fault Detection"
			potentialConsequences = []string{"Silent runtime failures", "Resource leakage", "Corrupted downstream application state"}
		} else {
			sev = repo_intel.SeverityLow
			rationale = "Repository documentation or licensing omission"
			impactScope = "Repository Engineering Hygiene"
			potentialConsequences = []string{"Legal ambiguity", "Developer onboarding confusion"}
		}

	case repo_intel.CategoryTestQuality:
		if strings.Contains(f.FindingID, "NOCI") {
			sev = repo_intel.SeverityHigh
			rationale = "Absence of continuous integration workflows leaves commits unvalidated before merge"
			impactScope = "Repository Release Gate & CI/CD"
			potentialConsequences = []string{"Uncaught regressions reaching staging/production"}
		} else {
			sev = repo_intel.SeverityLow
			rationale = "Directory lacks co-located unit test files"
			impactScope = "Package Regression Safety"
			potentialConsequences = []string{"Lack of unit test coverage during refactoring"}
		}

	case repo_intel.CategoryDependency:
		sev = repo_intel.SeverityMedium
		rationale = "Missing lockfile compromises deterministic and tamper-proof builds"
		impactScope = "Build Pipeline & Supply Chain"
		potentialConsequences = []string{"Non-reproducible CI builds", "Vulnerability to upstream dependency drift"}

	default:
		sev = repo_intel.SeverityLow
		rationale = "Informational observation"
		impactScope = "General"
		potentialConsequences = []string{"Minor maintenance overhead"}
	}

	return sev, fmt.Sprintf("%s: %s", sev, rationale), impactScope, potentialConsequences
}
