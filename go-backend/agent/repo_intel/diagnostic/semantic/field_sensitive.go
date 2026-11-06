package semantic

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

// StructOwnershipState classifies the ownership and lifecycle status of a struct-embedded resource.
type StructOwnershipState string

const (
	StateOwnedLocal       StructOwnershipState = "OWNED_LOCAL"
	StateOwnedTransferred StructOwnershipState = "OWNED_TRANSFERRED"
	StateFieldReleased    StructOwnershipState = "RELEASED"
	StateFieldMaybeRel    StructOwnershipState = "MAYBE_RELEASED"
	StateEscapedUnknown   StructOwnershipState = "ESCAPED_UNKNOWN"
)

// StructFieldInstance models a single access-path-tracked resource inside a struct.
type StructFieldInstance struct {
	AccessPath    string               `json:"access_path"`     // e.g. "h.file", "obj.inner.file"
	RootVar       string               `json:"root_var"`        // e.g. "h"
	FieldName     string               `json:"field_name"`      // e.g. "file"
	Depth         int                  `json:"depth"`           // e.g. 1, 2
	ResourceKind  ResourceKind         `json:"resource_kind"`   // e.g. ResourceKindFile
	OriginalVar   string               `json:"original_var"`    // local var from which it was assigned
	State         StructOwnershipState `json:"state"`
	AssignedLine  int                  `json:"assigned_line"`
	ReleasedLine  int                  `json:"released_line,omitempty"`
	IsPointer     bool                 `json:"is_pointer"`
	IsTruncated   bool                 `json:"is_truncated"`
	Evidence      string               `json:"evidence"`
}

// FieldSensitiveTracker manages bounded access-path resource tracking.
type FieldSensitiveTracker struct {
	MaxDepth     int                                // Default: 3
	Fields       map[string]*StructFieldInstance    // AccessPath -> FieldInstance
	StructAliases map[string]string                 // aliasVar -> rootStructVar (e.g. "p" -> "h")
	Truncated    map[string]bool
}

// NewFieldSensitiveTracker initializes a new access-path tracker.
func NewFieldSensitiveTracker(maxDepth int) *FieldSensitiveTracker {
	if maxDepth <= 0 {
		maxDepth = 3
	}
	return &FieldSensitiveTracker{
		MaxDepth:      maxDepth,
		Fields:        make(map[string]*StructFieldInstance),
		StructAliases: make(map[string]string),
		Truncated:     make(map[string]bool),
	}
}

// TrackAssignment processes struct field assignments: `h.file = f` or `h := &Holder{file: f}`.
func (t *FieldSensitiveTracker) TrackAssignment(
	lhsExpr ast.Expr,
	rhsExpr ast.Expr,
	fset *token.FileSet,
	knownResources map[string]ResourceKind,
) {
	line := fset.Position(lhsExpr.Pos()).Line

	// Case 1: Direct field assignment: h.file = f or h.conn = c
	if sel, ok := lhsExpr.(*ast.SelectorExpr); ok {
		accessPath, depth, isTruncated := t.resolveAccessPath(sel)
		if isTruncated {
			t.Truncated[accessPath] = true
			return
		}

		rootVar := strings.Split(accessPath, ".")[0]
		fieldName := sel.Sel.Name

		var origVar string
		var resKind ResourceKind

		if rhsId, ok := rhsExpr.(*ast.Ident); ok {
			origVar = rhsId.Name
			if k, found := knownResources[origVar]; found {
				resKind = k
			}
		}

		if resKind != "" || t.isLikelyResourceField(fieldName) {
			if resKind == "" {
				resKind = t.inferKindFromFieldName(fieldName)
			}
			t.Fields[accessPath] = &StructFieldInstance{
				AccessPath:   accessPath,
				RootVar:      rootVar,
				FieldName:    fieldName,
				Depth:        depth,
				ResourceKind: resKind,
				OriginalVar:  origVar,
				State:        StateOwnedLocal,
				AssignedLine: line,
				Evidence:     fmt.Sprintf("Assigned %s to struct field %s", origVar, accessPath),
			}
		}
		return
	}

	// Case 2: Composite literal instantiation: h := &Holder{file: f}
	if lhsId, ok := lhsExpr.(*ast.Ident); ok {
		rootVar := lhsId.Name
		var compLit *ast.CompositeLit
		if cl, ok := rhsExpr.(*ast.CompositeLit); ok {
			compLit = cl
		} else if unary, ok := rhsExpr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
			if cl, ok := unary.X.(*ast.CompositeLit); ok {
				compLit = cl
			}
		}

		if compLit != nil {
			for _, elt := range compLit.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if keyId, ok := kv.Key.(*ast.Ident); ok {
						fieldName := keyId.Name
						if valId, ok := kv.Value.(*ast.Ident); ok {
							valName := valId.Name
							if k, found := knownResources[valName]; found || t.isLikelyResourceField(fieldName) {
								if k == "" {
									k = t.inferKindFromFieldName(fieldName)
								}
								accessPath := fmt.Sprintf("%s.%s", rootVar, fieldName)
								t.Fields[accessPath] = &StructFieldInstance{
									AccessPath:   accessPath,
									RootVar:      rootVar,
									FieldName:    fieldName,
									Depth:        1,
									ResourceKind: k,
									OriginalVar:  valName,
									State:        StateOwnedLocal,
									AssignedLine: line,
									Evidence:     fmt.Sprintf("Initialized field %s via composite literal with %s", accessPath, valName),
								}
							}
						}
					}
				}
			}
		}
	}
}

// TrackStructAlias records pointer aliases between structs: `p := h` or `p := &h`.
func (t *FieldSensitiveTracker) TrackStructAlias(lhs, rhs string) {
	if lhs == "" || rhs == "" || lhs == rhs {
		return
	}
	// If rhs is already an alias, resolve transitively
	root := rhs
	if r, ok := t.StructAliases[rhs]; ok {
		root = r
	}
	t.StructAliases[lhs] = root
}

// TrackReleaseCall checks if a call expression releases an access path: `h.file.Close()` or `p.file.Close()`.
func (t *FieldSensitiveTracker) TrackReleaseCall(call *ast.CallExpr, fset *token.FileSet) bool {
	if call == nil {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Close" {
		return false
	}

	line := fset.Position(call.Pos()).Line

	// Inspect receiver: h.file or p.file
	if innerSel, ok := sel.X.(*ast.SelectorExpr); ok {
		accessPath, _, _ := t.resolveAccessPath(innerSel)
		// Check direct match
		if field, exists := t.Fields[accessPath]; exists {
			field.State = StateFieldReleased
			field.ReleasedLine = line
			field.Evidence = fmt.Sprintf("Released field %s via explicit %s()", accessPath, sel.Sel.Name)
			return true
		}

		// Check alias match: p.file -> h.file
		parts := strings.Split(accessPath, ".")
		if len(parts) >= 2 {
			if root, isAlias := t.StructAliases[parts[0]]; isAlias {
				aliasedPath := root + "." + strings.Join(parts[1:], ".")
				if field, exists := t.Fields[aliasedPath]; exists {
					field.State = StateFieldReleased
					field.ReleasedLine = line
					field.Evidence = fmt.Sprintf("Released field %s via alias %s.%s()", aliasedPath, parts[0], sel.Sel.Name)
					return true
				}
			}
		}
	}

	return false
}

// TrackReturnTransfer checks whether a return statement returns a struct containing tracked fields (`return h`).
func (t *FieldSensitiveTracker) TrackReturnTransfer(retStmt *ast.ReturnStmt) {
	if retStmt == nil {
		return
	}
	for _, res := range retStmt.Results {
		var retVar string
		if id, ok := res.(*ast.Ident); ok {
			retVar = id.Name
		} else if unary, ok := res.(*ast.UnaryExpr); ok && unary.Op == token.AND {
			if id, ok := unary.X.(*ast.Ident); ok {
				retVar = id.Name
			}
		}

		if retVar == "" {
			continue
		}

		// Resolve alias
		root := retVar
		if r, ok := t.StructAliases[retVar]; ok {
			root = r
		}

		// Mark all fields of root as OWNED_TRANSFERRED
		prefix := root + "."
		for path, field := range t.Fields {
			if strings.HasPrefix(path, prefix) && field.State == StateOwnedLocal {
				field.State = StateOwnedTransferred
				field.Evidence = fmt.Sprintf("Field %s transferred to caller via return %s", path, retVar)
			}
		}
	}
}

// resolveAccessPath decomposes a nested selector expression up to MaxDepth.
func (t *FieldSensitiveTracker) resolveAccessPath(sel *ast.SelectorExpr) (string, int, bool) {
	var parts []string
	curr := sel
	depth := 0

	for curr != nil {
		depth++
		parts = append([]string{curr.Sel.Name}, parts...)
		if depth > t.MaxDepth {
			return strings.Join(parts, "."), depth, true // Truncated!
		}

		if nextSel, ok := curr.X.(*ast.SelectorExpr); ok {
			curr = nextSel
		} else if id, ok := curr.X.(*ast.Ident); ok {
			parts = append([]string{id.Name}, parts...)
			break
		} else {
			break
		}
	}

	return strings.Join(parts, "."), depth, false
}

func (t *FieldSensitiveTracker) isLikelyResourceField(name string) bool {
	lower := strings.ToLower(name)
	return lower == "file" || lower == "conn" || lower == "body" || lower == "rows" || lower == "tx"
}

func (t *FieldSensitiveTracker) inferKindFromFieldName(name string) ResourceKind {
	lower := strings.ToLower(name)
	switch lower {
	case "file":
		return ResourceKindFile
	case "conn":
		return ResourceKindNetConn
	case "body":
		return ResourceKindHTTPBody
	case "rows":
		return ResourceKindSQLRows
	case "tx":
		return ResourceKindSQLTx
	default:
		return ResourceKindFile
	}
}
