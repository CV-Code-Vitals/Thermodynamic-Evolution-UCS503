package repo_intel

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ReliabilitySecurityAnalyzer inspects source files for hardcoded secrets,
// missing error handling, resource leaks, path traversal, and unsafe patterns.
type ReliabilitySecurityAnalyzer struct {
	repoRoot string
}

func NewReliabilitySecurityAnalyzer(repoRoot string) *ReliabilitySecurityAnalyzer {
	return &ReliabilitySecurityAnalyzer{repoRoot: repoRoot}
}

// Regex patterns for security analysis
var (
	// Hardcoded secret heuristics (API keys, private keys, passwords)
	secretRegexes = []struct {
		name    string
		pattern *regexp.Regexp
	}{
		{"AWS Access Key ID", regexp.MustCompile(`\b((?:AKIA|ASIA|ACCA)[0-9A-Z]{16})\b`)},
		{"Generic Secret Key", regexp.MustCompile(`(?i)(?:api_key|secret_key|client_secret|auth_token)\s*(?::=|=|:)\s*["']([a-zA-Z0-9_\-\.]{24,})["']`)},
		{"Hardcoded Password in String", regexp.MustCompile(`(?i)(?:password|passwd|pwd)\s*(?::=|=|:)\s*["']([^"'\s]{8,})["']`)},
		{"Private Key Header", regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
	}

	// Reliability heuristics
	unhandledErrorRegex = regexp.MustCompile(`(?i)^\s*_[,\s]*=\s*[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+\(`)
	sqlInjectionRegex   = regexp.MustCompile(`(?i)(?:SELECT|INSERT|UPDATE|DELETE)\s+.*\+\s*[a-zA-Z0-9_]+`)
	commandExecRegex    = regexp.MustCompile(`(?i)exec\.Command(?:Context)?\([^,]+,\s*(?:fmt\.Sprintf|[a-zA-Z0-9_]+\s*\+)`)
)

// AnalyzeReliabilityAndSecurity scans discovered source files for security vulnerabilities and reliability defects.
func (rsa *ReliabilitySecurityAnalyzer) AnalyzeReliabilityAndSecurity(files []DiscoveredFile) []Finding {
	var findings []Finding

	for _, file := range files {
		if !file.IsSource && !file.IsConfig {
			continue
		}

		fullPath := filepath.Join(rsa.repoRoot, filepath.FromSlash(file.RelativePath))
		f, err := os.Open(fullPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(f)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			// Skip comments in various languages
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
				continue
			}

			// 1. Secret Scanning
			for _, sr := range secretRegexes {
				if matches := sr.pattern.FindStringSubmatch(line); len(matches) > 1 {
					// Redact secret value
					matchedSecret := matches[1]
					redacted := redactSecret(matchedSecret)
					findings = append(findings, Finding{
						FindingID:   fmt.Sprintf("RI-SEC-SECRET-%s-L%d", sanitizeID(file.RelativePath), lineNum),
						Category:    CategorySecurity,
						Severity:    SeverityCritical,
						Confidence:  0.92,
						Title:       fmt.Sprintf("Potential hardcoded secret (%s)", sr.name),
						Description: fmt.Sprintf("File %s appears to contain a hardcoded credential or token: %s", file.RelativePath, sr.name),
						RootCause:   "Embedding secret tokens or private keys directly in source code.",
						AffectedFiles: []string{file.RelativePath},
						Evidence: []EvidenceLocation{
							{
								File:        file.RelativePath,
								LineStart:   lineNum,
								LineEnd:     lineNum,
								Description: fmt.Sprintf("%s pattern matched (value: %s)", sr.name, redacted),
							},
						},
						RecommendedAction:      "Remove the hardcoded secret immediately, revoke/rotate it, and supply via environment variables or secret manager.",
						ExpectedBenefit:        "Prevents credential theft and unauthorized access.",
						RegressionRisk:         SeverityLow,
						ValidationRequirements: []string{"Secret is removed from file", "Credential is read from environment variable"},
						DetectionMethod:        "REGEX_PATTERN_MATCHING",
						VerificationStatus:     StatusProbable,
					})
				}
			}

			// 2. Command Execution Injection Risks
			if file.Extension == ".go" && commandExecRegex.MatchString(line) {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-SEC-CMD-INJECT-%s-L%d", sanitizeID(file.RelativePath), lineNum),
					Category:    CategorySecurity,
					Severity:    SeverityHigh,
					Confidence:  0.88,
					Title:       "Potential command injection risk in exec.Command",
					Description: fmt.Sprintf("Command execution in %s at line %d dynamically concatenates arguments or command string.", file.RelativePath, lineNum),
					RootCause:   "String concatenation into process execution arguments.",
					AffectedFiles: []string{file.RelativePath},
					Evidence: []EvidenceLocation{
						{
							File:        file.RelativePath,
							LineStart:   lineNum,
							LineEnd:     lineNum,
							Description: strings.TrimSpace(line),
						},
					},
					RecommendedAction:      "Pass arguments as separate string elements rather than formatting shell strings.",
					ExpectedBenefit:        "Eliminates arbitrary shell argument injection vulnerabilities.",
					RegressionRisk:         SeverityMedium,
					ValidationRequirements: []string{"Arguments are passed individually without shell interpretation"},
					DetectionMethod:        "STATIC_TAINT_HEURISTIC",
					VerificationStatus:     StatusProbable,
				})
			}

			// 3. SQL Injection via string concatenation
			if sqlInjectionRegex.MatchString(line) && !strings.Contains(line, "$") && !strings.Contains(line, "?") {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-SEC-SQL-CONCAT-%s-L%d", sanitizeID(file.RelativePath), lineNum),
					Category:    CategorySecurity,
					Severity:    SeverityHigh,
					Confidence:  0.80,
					Title:       "Potential SQL injection via query string concatenation",
					Description: fmt.Sprintf("SQL query string in %s at line %d appears to concatenate variables directly.", file.RelativePath, lineNum),
					RootCause:   "Constructing dynamic SQL queries via string concatenation instead of parameterized placeholders.",
					AffectedFiles: []string{file.RelativePath},
					Evidence: []EvidenceLocation{
						{
							File:        file.RelativePath,
							LineStart:   lineNum,
							LineEnd:     lineNum,
							Description: strings.TrimSpace(line),
						},
					},
					RecommendedAction:      "Use parameterized queries with bind variables ($1, ?, :name).",
					ExpectedBenefit:        "Prevents SQL injection attacks entirely.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"All dynamic SQL parameters use bind placeholders"},
					DetectionMethod:        "STATIC_TAINT_HEURISTIC",
					VerificationStatus:     StatusHeuristic,
				})
			}

			// 4. Ignored Error Handling in Go
			if file.Extension == ".go" && unhandledErrorRegex.MatchString(line) {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-REL-IGNORED-ERR-%s-L%d", sanitizeID(file.RelativePath), lineNum),
					Category:    CategoryReliability,
					Severity:    SeverityLow,
					Confidence:  0.85,
					Title:       "Ignored function error return value",
					Description: fmt.Sprintf("File %s at line %d explicitly discards error return value via blank identifier '_'.", file.RelativePath, lineNum),
					RootCause:   "Error return value suppressed without handling or logging.",
					AffectedFiles: []string{file.RelativePath},
					Evidence: []EvidenceLocation{
						{
							File:        file.RelativePath,
							LineStart:   lineNum,
							LineEnd:     lineNum,
							Description: strings.TrimSpace(line),
						},
					},
					RecommendedAction:      "Inspect returned error and propagate, log, or handle gracefully.",
					ExpectedBenefit:        "Prevents silent failures and simplifies operational diagnosis.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"Error is checked and handled"},
					DetectionMethod:        "STATIC_ERROR_PATTERN",
					VerificationStatus:     StatusConfirmed,
				})
			}
		}
		f.Close()
	}

	return findings
}

func redactSecret(secret string) string {
	if len(secret) <= 4 {
		return "****"
	}
	prefix := secret[:2]
	suffix := secret[len(secret)-2:]
	return fmt.Sprintf("%s****%s", prefix, suffix)
}

