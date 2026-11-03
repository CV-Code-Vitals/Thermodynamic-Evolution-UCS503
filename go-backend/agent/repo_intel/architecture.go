package repo_intel

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/saksham/self-healing-pipeline/agent"
)

// ArchitecturalAnalyzer analyzes cross-file dependencies, detects circular dependencies,
// maps layers, and identifies architectural coupling across services.
type ArchitecturalAnalyzer struct {
	repoRoot string
}

func NewArchitecturalAnalyzer(repoRoot string) *ArchitecturalAnalyzer {
	return &ArchitecturalAnalyzer{repoRoot: repoRoot}
}

// AnalyzeArchitecture constructs the dependency graph and evaluates architectural constraints.
func (aa *ArchitecturalAnalyzer) AnalyzeArchitecture(
	report *agent.ThermodynamicReport,
	services []ServiceBoundary,
) (DependencyGraph, []Finding) {
	var findings []Finding
	graph := aa.buildDependencyGraph(report)

	// 1. Detect circular dependencies at file and package levels
	cycleFindings := aa.detectCycles(graph)
	findings = append(findings, cycleFindings...)

	// 2. Cross-service interface & boundary analysis
	serviceFindings := aa.analyzeServiceBoundaries(services, graph)
	findings = append(findings, serviceFindings...)

	// 3. Layering analysis (e.g. HTTP handler importing database pool directly or backward dependency)
	layerFindings := aa.analyzeLayeringRules(graph)
	findings = append(findings, layerFindings...)

	return graph, findings
}

func (aa *ArchitecturalAnalyzer) buildDependencyGraph(report *agent.ThermodynamicReport) DependencyGraph {
	nodeSet := make(map[string]bool)
	var edges []RelationshipEdge

	if report == nil {
		return DependencyGraph{}
	}

	for _, rel := range report.Relationships {
		src := filepath.ToSlash(rel.SourceFile)
		tgt := filepath.ToSlash(rel.TargetFile)
		if src == "" || tgt == "" || src == tgt {
			continue
		}

		nodeSet[src] = true
		nodeSet[tgt] = true

		edges = append(edges, RelationshipEdge{
			Source:     src,
			Target:     tgt,
			Type:       rel.Kind,
			Confidence: 0.95,
			Evidence:   fmt.Sprintf("Source: %s -> Target: %s (%s)", src, tgt, rel.Kind),
		})
	}

	for _, fr := range report.FileReports {
		src := filepath.ToSlash(fr.FilePath)
		nodeSet[src] = true
		for _, imp := range fr.Imports {
			cleanImp := filepath.ToSlash(imp.Path)
			if cleanImp != "" && cleanImp != src {
				nodeSet[cleanImp] = true
				edges = append(edges, RelationshipEdge{
					Source:     src,
					Target:     cleanImp,
					Type:       "import",
					Confidence: 0.90,
					Evidence:   fmt.Sprintf("Import '%s' in %s", imp.Path, src),
				})
			}
		}
	}

	var nodes []string
	for n := range nodeSet {
		nodes = append(nodes, n)
	}
	sort.Strings(nodes)

	return DependencyGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

// detectCycles detects cycles in the dependency graph using Tarjan's strongly connected components algorithm.
func (aa *ArchitecturalAnalyzer) detectCycles(graph DependencyGraph) []Finding {
	var findings []Finding

	// Build adjacency list for file-level and package-level graph
	adj := make(map[string][]string)
	for _, edge := range graph.Edges {
		adj[edge.Source] = append(adj[edge.Source], edge.Target)
	}

	// Simple cycle detection via DFS coloring (white=0, gray=1, black=2)
	color := make(map[string]int)
	parent := make(map[string]string)
	var detectedCycles [][]string

	var dfs func(u string, path []string)
	dfs = func(u string, path []string) {
		color[u] = 1 // in progress
		path = append(path, u)

		for _, v := range adj[u] {
			if color[v] == 1 {
				// Found a back-edge (cycle)
				cycleStart := -1
				for i, node := range path {
					if node == v {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := append([]string{}, path[cycleStart:]...)
					cycle = append(cycle, v)
					detectedCycles = append(detectedCycles, cycle)
				}
			} else if color[v] == 0 {
				parent[v] = u
				dfs(v, path)
			}
		}

		color[u] = 2 // finished
	}

	for _, n := range graph.Nodes {
		if color[n] == 0 {
			dfs(n, nil)
		}
	}

	// Deduplicate cycles by canonical string representation
	seenCycles := make(map[string]bool)
	for _, c := range detectedCycles {
		if len(c) < 2 {
			continue
		}
		cycleKey := strings.Join(c, "->")
		if seenCycles[cycleKey] {
			continue
		}
		seenCycles[cycleKey] = true

		affectedFiles := make([]string, 0, len(c)-1)
		for i := 0; i < len(c)-1; i++ {
			affectedFiles = append(affectedFiles, c[i])
		}

		findings = append(findings, Finding{
			FindingID:   fmt.Sprintf("RI-ARCH-CYCLE-%s", sanitizeID(c[0])),
			Category:    CategoryArchitecture,
			Severity:    SeverityHigh,
			Confidence:  0.95,
			Title:       "Circular dependency cycle detected",
			Description: fmt.Sprintf("Dependency cycle detected across files/modules: %s", strings.Join(c, " ➔ ")),
			RootCause:   "Circular import or mutual reference between components.",
			AffectedFiles: affectedFiles,
			Evidence: []EvidenceLocation{
				{File: c[0], Description: fmt.Sprintf("Cycle path: %s", strings.Join(c, " -> "))},
			},
			RecommendedAction:      "Extract common types/interfaces into an independent leaf module or introduce dependency inversion.",
			ExpectedBenefit:        "Eliminates tight coupling, allows modular compilation, and prevents initialization deadlocks.",
			RegressionRisk:         SeverityMedium,
			ValidationRequirements: []string{"No cycle detected in directed dependency graph", "Build succeeds"},
			DetectionMethod:        "STATIC_DEPENDENCY_GRAPH",
			VerificationStatus:     StatusConfirmed,
		})
	}

	return findings
}

func (aa *ArchitecturalAnalyzer) analyzeServiceBoundaries(services []ServiceBoundary, graph DependencyGraph) []Finding {
	var findings []Finding

	// Identify tight coupling across distinct service roots
	serviceRoots := make(map[string]string) // file -> serviceName
	for _, svc := range services {
		if svc.RootPath != "" && svc.RootPath != "." {
			serviceRoots[svc.RootPath] = svc.Name
		}
	}

	for _, edge := range graph.Edges {
		srcSvc := findServiceForFile(edge.Source, serviceRoots)
		tgtSvc := findServiceForFile(edge.Target, serviceRoots)

		if srcSvc != "" && tgtSvc != "" && srcSvc != tgtSvc {
			// Direct file import between different microservices
			if edge.Type == "import" {
				findings = append(findings, Finding{
					FindingID:   fmt.Sprintf("RI-ARCH-SVC-COUPLING-%s-%s", sanitizeID(srcSvc), sanitizeID(tgtSvc)),
					Category:    CategoryArchitecture,
					Severity:    SeverityMedium,
					Confidence:  0.85,
					Title:       fmt.Sprintf("Direct code import across service boundary (%s ➔ %s)", srcSvc, tgtSvc),
					Description: fmt.Sprintf("Service %q directly imports file %q from service %q, violating service isolation.", srcSvc, edge.Target, tgtSvc),
					RootCause:   "Direct file import instead of calling via shared client library, API, or message bus.",
					AffectedFiles: []string{edge.Source, edge.Target},
					Evidence: []EvidenceLocation{
						{File: edge.Source, Description: fmt.Sprintf("Imports %s from service %s", edge.Target, tgtSvc)},
					},
					RecommendedAction:      "Decouple services using an explicit shared contract or independent library.",
					ExpectedBenefit:        "Allows independent deployment and prevents monorepo entanglement.",
					RegressionRisk:         SeverityMedium,
					ValidationRequirements: []string{"Services do not import each other's internal source files directly"},
					DetectionMethod:        "ARCHITECTURAL_BOUNDARY_SCAN",
					VerificationStatus:     StatusProbable,
				})
			}
		}
	}

	return findings
}

func (aa *ArchitecturalAnalyzer) analyzeLayeringRules(graph DependencyGraph) []Finding {
	var findings []Finding

	// Check for common architectural anti-patterns:
	// e.g., Domain or Core models importing Presentation / Web frameworks (gin, react, etc.)
	for _, edge := range graph.Edges {
		srcLower := strings.ToLower(edge.Source)
		tgtLower := strings.ToLower(edge.Target)

		// Rule 1: Internal/core domain importing web/router handlers
		isCore := strings.Contains(srcLower, "/domain/") || strings.Contains(srcLower, "/core/") || strings.Contains(srcLower, "/model/")
		isPresentation := strings.Contains(tgtLower, "/handler") || strings.Contains(tgtLower, "/router") || strings.Contains(tgtLower, "/api/") || strings.Contains(tgtLower, "gin-gonic/gin")

		if isCore && isPresentation {
			findings = append(findings, Finding{
				FindingID:   fmt.Sprintf("RI-ARCH-LAYER-REVERSE-%s", sanitizeID(edge.Source)),
				Category:    CategoryArchitecture,
				Severity:    SeverityHigh,
				Confidence:  0.90,
				Title:       "Reverse architectural layering dependency",
				Description: fmt.Sprintf("Core/domain file %s depends on presentation/transport component %s.", edge.Source, edge.Target),
				RootCause:   "Violation of Clean Architecture / Hexagonal dependency inversion principle.",
				AffectedFiles: []string{edge.Source, edge.Target},
				Evidence: []EvidenceLocation{
					{File: edge.Source, Description: fmt.Sprintf("Core layer imports presentation layer %s", edge.Target)},
				},
				RecommendedAction:      "Invert dependency by defining interface in domain and implementing in presentation layer.",
				ExpectedBenefit:        "Restores domain purity and enables independent domain testing.",
				RegressionRisk:         SeverityMedium,
				ValidationRequirements: []string{"Domain layer contains zero dependencies on presentation layer"},
				DetectionMethod:        "LAYER_DIRECTION_CHECK",
				VerificationStatus:     StatusConfirmed,
			})
		}
	}

	return findings
}

func findServiceForFile(filePath string, serviceRoots map[string]string) string {
	for root, svcName := range serviceRoots {
		if strings.HasPrefix(filePath, root+"/") || filePath == root {
			return svcName
		}
	}
	return ""
}

