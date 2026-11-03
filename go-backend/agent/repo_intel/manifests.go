package repo_intel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ManifestAnalyzer analyzes dependency manifests and builds service boundaries.
type ManifestAnalyzer struct {
	repoRoot string
}

func NewManifestAnalyzer(repoRoot string) *ManifestAnalyzer {
	return &ManifestAnalyzer{repoRoot: repoRoot}
}

// AnalyzeManifests scans the discovered manifest files and extracts dependency metadata.
func (ma *ManifestAnalyzer) AnalyzeManifests(manifestPaths []string) ([]ManifestInfo, []ServiceBoundary, []Finding) {
	var infos []ManifestInfo
	var findings []Finding
	servicesMap := make(map[string]*ServiceBoundary)

	for _, relPath := range manifestPaths {
		fullPath := filepath.Join(ma.repoRoot, filepath.FromSlash(relPath))
		baseName := filepath.Base(fullPath)
		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = ""
		}

		info := ManifestInfo{
			FilePath:     relPath,
			ManifestType: baseName,
		}

		// Register service boundary by directory
		svcName := dir
		if svcName == "" {
			svcName = "root"
		}
		if _, exists := servicesMap[svcName]; !exists {
			servicesMap[svcName] = &ServiceBoundary{
				Name:     svcName,
				RootPath: dir,
				Type:     "module",
			}
		}
		servicesMap[svcName].Manifests = append(servicesMap[svcName].Manifests, relPath)

		switch {
		case baseName == "go.mod":
			info.ManifestType = "go.mod"
			deps, f := parseGoMod(fullPath, relPath)
			info.Dependencies = deps
			findings = append(findings, f...)
			// Check for go.sum
			lockRel := filepath.ToSlash(filepath.Join(dir, "go.sum"))
			if fileExists(filepath.Join(ma.repoRoot, filepath.FromSlash(lockRel))) {
				info.HasLockfile = true
				info.LockfilePath = lockRel
			} else {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-DEP-GOSUM-%s", sanitizeID(dir)),
					Category:    CategoryDependency,
					Severity:    SeverityMedium,
					Confidence:  0.95,
					Title:       "Missing go.sum lockfile",
					Description: fmt.Sprintf("Go module at %s lacks a go.sum lockfile, preventing cryptographic build verification.", relPath),
					RootCause:   "go.sum file not committed or generated",
					AffectedFiles: []string{relPath},
					Evidence: []EvidenceLocation{
						{File: relPath, Description: "go.mod without adjacent go.sum"},
					},
					RecommendedAction:      "Run 'go mod download' or 'go mod tidy' and commit go.sum.",
					ExpectedBenefit:        "Ensures reproducible and tamper-proof Go dependency builds.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"go.sum exists adjacent to go.mod"},
					DetectionMethod:        "MANIFEST_ANALYSIS",
					VerificationStatus:     StatusConfirmed,
				})
			}

		case baseName == "Cargo.toml":
			info.ManifestType = "Cargo.toml"
			deps, f := parseCargoToml(fullPath, relPath)
			info.Dependencies = deps
			findings = append(findings, f...)
			lockRel := filepath.ToSlash(filepath.Join(dir, "Cargo.lock"))
			if fileExists(filepath.Join(ma.repoRoot, filepath.FromSlash(lockRel))) {
				info.HasLockfile = true
				info.LockfilePath = lockRel
			}

		case baseName == "package.json":
			info.ManifestType = "package.json"
			deps, f := parsePackageJSON(fullPath, relPath)
			info.Dependencies = deps
			findings = append(findings, f...)
			// Check package-lock.json / pnpm-lock.yaml / yarn.lock
			for _, lockName := range []string{"package-lock.json", "pnpm-lock.yaml", "yarn.lock"} {
				lockRel := filepath.ToSlash(filepath.Join(dir, lockName))
				if fileExists(filepath.Join(ma.repoRoot, filepath.FromSlash(lockRel))) {
					info.HasLockfile = true
					info.LockfilePath = lockRel
					break
				}
			}
			if !info.HasLockfile {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-DEP-NOLOCK-%s", sanitizeID(dir)),
					Category:    CategoryDependency,
					Severity:    SeverityMedium,
					Confidence:  0.90,
					Title:       "Missing package.json lockfile",
					Description: fmt.Sprintf("Node.js module at %s lacks a lockfile (package-lock.json, pnpm-lock.yaml, or yarn.lock).", relPath),
					RootCause:   "Lockfile was not generated or was omitted from version control.",
					AffectedFiles: []string{relPath},
					Evidence: []EvidenceLocation{
						{File: relPath, Description: "package.json without lockfile"},
					},
					RecommendedAction:      "Run package manager install and commit the generated lockfile.",
					ExpectedBenefit:        "Enforces deterministic node_modules installation in CI and production.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"A valid lockfile exists adjacent to package.json"},
					DetectionMethod:        "MANIFEST_ANALYSIS",
					VerificationStatus:     StatusConfirmed,
				})
			}

		case strings.HasPrefix(baseName, "Dockerfile"):
			info.ManifestType = "Dockerfile"
			servicesMap[svcName].Dockerfiles = append(servicesMap[svcName].Dockerfiles, relPath)
			deps, f := parseDockerfile(fullPath, relPath)
			info.Dependencies = deps
			findings = append(findings, f...)
		}

		infos = append(infos, info)
	}

	var services []ServiceBoundary
	for _, s := range servicesMap {
		services = append(services, *s)
	}

	return infos, services, findings
}

func parseGoMod(fullPath, relPath string) ([]ManifestDependency, []Finding) {
	var deps []ManifestDependency
	var findings []Finding

	f, err := os.Open(fullPath)
	if err != nil {
		return nil, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inRequire := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}

		if inRequire || strings.HasPrefix(line, "require ") {
			parts := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(parts) >= 2 {
				depType := "direct"
				if strings.Contains(line, "// indirect") {
					depType = "indirect"
				}
				deps = append(deps, ManifestDependency{
					Name:    parts[0],
					Version: parts[1],
					Type:    depType,
				})
			}
		}

		// Check for local replace directives
		if strings.HasPrefix(line, "replace ") && strings.Contains(line, "=>") {
			findings = append(findings, Finding{
				FindingID:   fmt.Sprintf("RI-DEP-GOREPLACE-%s-L%d", sanitizeID(relPath), lineNum),
				Category:    CategoryDependency,
				Severity:    SeverityLow,
				Confidence:  0.85,
				Title:       "Local module replace directive detected",
				Description: fmt.Sprintf("go.mod at %s defines a local replace directive: %s", relPath, line),
				RootCause:   "Replace directive used for local development override.",
				AffectedFiles: []string{relPath},
				Evidence: []EvidenceLocation{
					{File: relPath, LineStart: lineNum, LineEnd: lineNum, Description: line},
				},
				RecommendedAction:      "Verify if replace directive is intended for production or development only.",
				ExpectedBenefit:        "Avoids broken builds when module is published or built remotely.",
				RegressionRisk:         SeverityLow,
				ValidationRequirements: []string{"Module resolves cleanly without broken relative paths"},
				DetectionMethod:        "MANIFEST_ANALYSIS",
				VerificationStatus:     StatusConfirmed,
			})
		}
	}
	return deps, findings
}

func parseCargoToml(fullPath, relPath string) ([]ManifestDependency, []Finding) {
	var deps []ManifestDependency
	var findings []Finding

	f, err := os.Open(fullPath)
	if err != nil {
		return nil, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	currentSection := ""
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			continue
		}

		if strings.Contains(currentSection, "dependencies") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				pkgName := strings.TrimSpace(parts[0])
				rawVer := strings.TrimSpace(parts[1])
				depType := "direct"
				if strings.Contains(currentSection, "dev-dependencies") {
					depType = "dev"
				} else if strings.Contains(currentSection, "build-dependencies") {
					depType = "build"
				}

				deps = append(deps, ManifestDependency{
					Name:    pkgName,
					Version: strings.Trim(rawVer, "\"'{ }"),
					Type:    depType,
				})
			}
		}
	}
	return deps, findings
}

func parsePackageJSON(fullPath, relPath string) ([]ManifestDependency, []Finding) {
	var deps []ManifestDependency
	var findings []Finding

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, nil
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, nil
	}

	for k, v := range pkg.Dependencies {
		deps = append(deps, ManifestDependency{
			Name:       k,
			Version:    v,
			Type:       "direct",
			Constraint: v,
		})
	}
	for k, v := range pkg.DevDependencies {
		deps = append(deps, ManifestDependency{
			Name:       k,
			Version:    v,
			Type:       "dev",
			Constraint: v,
		})
	}

	return deps, findings
}

var dockerFromRegex = regexp.MustCompile(`(?i)^FROM\s+([^\s]+)(?:\s+AS\s+([^\s]+))?`)

func parseDockerfile(fullPath, relPath string) ([]ManifestDependency, []Finding) {
	var deps []ManifestDependency
	var findings []Finding

	f, err := os.Open(fullPath)
	if err != nil {
		return nil, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNum := 0
	hasUserDirective := false

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if matches := dockerFromRegex.FindStringSubmatch(line); len(matches) > 1 {
			baseImg := matches[1]
			deps = append(deps, ManifestDependency{
				Name:    baseImg,
				Type:    "base_image",
				Version: baseImg,
			})
			if strings.HasSuffix(baseImg, ":latest") || !strings.Contains(baseImg, ":") {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-SEC-DOCKER-LATEST-%s-L%d", sanitizeID(relPath), lineNum),
					Category:    CategorySecurity,
					Severity:    SeverityLow,
					Confidence:  0.90,
					Title:       "Dockerfile uses unpinned or :latest base image",
					Description: fmt.Sprintf("Dockerfile at %s references base image %q without an explicit version pin.", relPath, baseImg),
					RootCause:   "Use of mutable :latest tag in container image specification.",
					AffectedFiles: []string{relPath},
					Evidence: []EvidenceLocation{
						{File: relPath, LineStart: lineNum, LineEnd: lineNum, Description: line},
					},
					RecommendedAction:      "Pin base image to an immutable tag or SHA256 digest.",
					ExpectedBenefit:        "Guarantees build reproducibility and prevents surprise upstream base image vulnerabilities.",
					RegressionRisk:         SeverityLow,
					ValidationRequirements: []string{"Base image has explicit version or digest tag"},
					DetectionMethod:        "MANIFEST_ANALYSIS",
					VerificationStatus:     StatusConfirmed,
				})
			}
		}

		if strings.HasPrefix(strings.ToUpper(line), "USER ") {
			hasUserDirective = true
		}
	}

	if !hasUserDirective {
		findings = append(findings, Finding{
			FindingID:   fmt.Sprintf("RI-SEC-DOCKER-ROOT-%s", sanitizeID(relPath)),
			Category:    CategorySecurity,
			Severity:    SeverityMedium,
			Confidence:  0.85,
			Title:       "Dockerfile lacks unprivileged USER directive",
			Description: fmt.Sprintf("Dockerfile at %s does not specify a non-root USER directive; containers may default to root execution.", relPath),
			RootCause:   "Absence of USER instruction before runtime entrypoint.",
			AffectedFiles: []string{relPath},
			Evidence: []EvidenceLocation{
				{File: relPath, Description: "No USER directive found in Dockerfile"},
			},
			RecommendedAction:      "Create a dedicated unprivileged user and group and switch using 'USER appuser:appgroup'.",
			ExpectedBenefit:        "Enforces principle of least privilege in containerized runtime.",
			RegressionRisk:         SeverityMedium,
			ValidationRequirements: []string{"Container runs as non-root user (UID > 0)"},
			DetectionMethod:        "MANIFEST_ANALYSIS",
			VerificationStatus:     StatusProbable,
		})
	}

	return deps, findings
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func sanitizeID(s string) string {
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if s == "" {
		return "root"
	}
	return s
}

