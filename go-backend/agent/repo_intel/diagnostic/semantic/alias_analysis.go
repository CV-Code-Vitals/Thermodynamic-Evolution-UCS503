package semantic

import (
	"fmt"
	"go/ast"
	"go/token"
)

// AliasRelationType represents the certainty of an alias relationship.
type AliasRelationType string

const (
	AliasMustAlias AliasRelationType = "MUST_ALIAS"
	AliasMayAlias  AliasRelationType = "MAY_ALIAS"
	AliasNoAlias   AliasRelationType = "NO_ALIAS"
)

// AliasEdge describes a directed alias relationship from a source to a target variable.
type AliasEdge struct {
	TargetVar    string            `json:"target_var"`
	Relation     AliasRelationType `json:"relation"`
	HopCount     int               `json:"hop_count"`
	AssignedLine int               `json:"assigned_line"`
	IsPointer    bool              `json:"is_pointer"`
}

// AliasGraph tracks bounded multi-hop variable aliases and reassignments within a function.
type AliasGraph struct {
	MaxHops      int                       // Configured budget (default: 5)
	MaxSetSize   int                       // Maximum aliases per variable (default: 32)
	DirectEdges  map[string][]AliasEdge    // sourceVar -> list of direct targets assigned from it
	AllAliases   map[string]map[string]AliasEdge // rootVar -> { aliasVar -> edge }
	Invalidated  map[string][]int          // varName -> lines where reassigned/invalidated
	Truncated    map[string]bool           // rootVar -> true if chain exceeded MaxHops
}

// NewAliasGraph initializes an AliasGraph with default bounds.
func NewAliasGraph(maxHops, maxSetSize int) *AliasGraph {
	if maxHops <= 0 {
		maxHops = 5
	}
	if maxSetSize <= 0 {
		maxSetSize = 32
	}
	return &AliasGraph{
		MaxHops:     maxHops,
		MaxSetSize:  maxSetSize,
		DirectEdges: make(map[string][]AliasEdge),
		AllAliases:  make(map[string]map[string]AliasEdge),
		Invalidated: make(map[string][]int),
		Truncated:   make(map[string]bool),
	}
}

// AddAssignment records an assignment `lhs = rhs` or `lhs := rhs`.
func (g *AliasGraph) AddAssignment(lhs, rhs string, isPointer bool, relation AliasRelationType, line int) {
	if lhs == "" || rhs == "" || lhs == "_" || rhs == "_" || lhs == rhs {
		return
	}

	// Record potential invalidation of lhs at this line if it was already assigned
	g.Invalidated[lhs] = append(g.Invalidated[lhs], line)

	edge := AliasEdge{
		TargetVar:    lhs,
		Relation:     relation,
		HopCount:     1,
		AssignedLine: line,
		IsPointer:    isPointer,
	}

	g.DirectEdges[rhs] = append(g.DirectEdges[rhs], edge)
}

// ComputeTransitiveClosure builds bounded multi-hop alias sets for a root variable.
func (g *AliasGraph) ComputeTransitiveClosure(rootVar string) map[string]AliasEdge {
	if res, ok := g.AllAliases[rootVar]; ok {
		return res
	}

	resultSet := make(map[string]AliasEdge)
	visited := make(map[string]bool)

	type workItem struct {
		currVar  string
		hopCount int
		relation AliasRelationType
	}

	worklist := []workItem{{currVar: rootVar, hopCount: 0, relation: AliasMustAlias}}
	visited[rootVar] = true

	for len(worklist) > 0 {
		item := worklist[0]
		worklist = worklist[1:]

		if item.hopCount >= g.MaxHops {
			// Budget reached: mark rootVar as truncated and do not traverse further
			g.Truncated[rootVar] = true
			continue
		}

		if len(resultSet) >= g.MaxSetSize {
			g.Truncated[rootVar] = true
			break
		}

		edges := g.DirectEdges[item.currVar]
		for _, edge := range edges {
			nextVar := edge.TargetVar
			nextHop := item.hopCount + 1

			// Combine relations: MUST + MUST = MUST; any MAY = MAY
			effRelation := AliasMustAlias
			if item.relation == AliasMayAlias || edge.Relation == AliasMayAlias {
				effRelation = AliasMayAlias
			}

			// Cycle detection
			if visited[nextVar] {
				continue
			}
			visited[nextVar] = true

			resolvedEdge := AliasEdge{
				TargetVar:    nextVar,
				Relation:     effRelation,
				HopCount:     nextHop,
				AssignedLine: edge.AssignedLine,
				IsPointer:    edge.IsPointer,
			}
			resultSet[nextVar] = resolvedEdge

			worklist = append(worklist, workItem{
				currVar:  nextVar,
				hopCount: nextHop,
				relation: effRelation,
			})
		}
	}

	g.AllAliases[rootVar] = resultSet
	return resultSet
}

// IsAliasAtLine checks whether candidateVar is a valid alias of rootVar at a specific source line.
// Respects reassignment and invalidation.
func (g *AliasGraph) IsAliasAtLine(rootVar, candidateVar string, line int) (bool, AliasRelationType) {
	if rootVar == candidateVar {
		return true, AliasMustAlias
	}

	aliases := g.ComputeTransitiveClosure(rootVar)
	edge, exists := aliases[candidateVar]
	if !exists {
		return false, AliasNoAlias
	}

	// Check if candidateVar was assigned AFTER the call line
	if edge.AssignedLine > line && line > 0 {
		return false, AliasNoAlias
	}

	// Check if candidateVar was invalidated / reassigned between its assignment and the call line
	if invLines, ok := g.Invalidated[candidateVar]; ok {
		for _, invLine := range invLines {
			if invLine > edge.AssignedLine && invLine <= line {
				// Reassigned before this line: alias is no longer valid!
				return false, AliasNoAlias
			}
		}
	}

	return true, edge.Relation
}

// GetAliases returns the list of all active variable names in the transitive alias set.
func (g *AliasGraph) GetAliases(rootVar string) []string {
	aliasMap := g.ComputeTransitiveClosure(rootVar)
	res := make([]string, 0, len(aliasMap))
	for name := range aliasMap {
		res = append(res, name)
	}
	return res
}

// BuildAliasGraphFromAST inspects function body AST and constructs the AliasGraph.
func BuildAliasGraphFromAST(fnDecl *ast.FuncDecl, fset *token.FileSet, maxHops, maxSetSize int) *AliasGraph {
	graph := NewAliasGraph(maxHops, maxSetSize)
	if fnDecl == nil || fnDecl.Body == nil {
		return graph
	}

	ast.Inspect(fnDecl.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for i := 0; i < len(stmt.Lhs) && i < len(stmt.Rhs); i++ {
				lhsIdent, isLhs := stmt.Lhs[i].(*ast.Ident)
				if !isLhs || lhsIdent.Name == "_" {
					continue
				}

				rhsExpr := stmt.Rhs[i]
				line := fset.Position(stmt.Pos()).Line

				// Direct assignment: a = b or a := b
				if rhsIdent, isRhs := rhsExpr.(*ast.Ident); isRhs {
					graph.AddAssignment(lhsIdent.Name, rhsIdent.Name, false, AliasMustAlias, line)
					continue
				}

				// Pointer assignment: p := &b
				if unary, isUnary := rhsExpr.(*ast.UnaryExpr); isUnary && unary.Op == token.AND {
					if targetIdent, isTarget := unary.X.(*ast.Ident); isTarget {
						graph.AddAssignment(lhsIdent.Name, targetIdent.Name, true, AliasMustAlias, line)
						continue
					}
				}

				// Pointer dereference assignment: a = *p
				if star, isStar := rhsExpr.(*ast.StarExpr); isStar {
					if ptrIdent, isPtr := star.X.(*ast.Ident); isPtr {
						graph.AddAssignment(lhsIdent.Name, ptrIdent.Name, true, AliasMustAlias, line)
						continue
					}
				}
			}

		case *ast.IfStmt:
			// Within IfStmt blocks, assignments are conditional (MAY_ALIAS)
			if stmt.Body != nil {
				for _, s := range stmt.Body.List {
					if assign, ok := s.(*ast.AssignStmt); ok {
						for i := 0; i < len(assign.Lhs) && i < len(assign.Rhs); i++ {
							if lhsId, isLhs := assign.Lhs[i].(*ast.Ident); isLhs && lhsId.Name != "_" {
								if rhsId, isRhs := assign.Rhs[i].(*ast.Ident); isRhs {
									line := fset.Position(assign.Pos()).Line
									graph.AddAssignment(lhsId.Name, rhsId.Name, false, AliasMayAlias, line)
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	return graph
}

// Suppress unused imports
var (
	_ = fmt.Sprintf
)
