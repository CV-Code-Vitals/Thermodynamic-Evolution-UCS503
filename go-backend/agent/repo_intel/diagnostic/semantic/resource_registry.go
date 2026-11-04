package semantic

import (
	"go/ast"
	"go/token"
)

// AcquisitionContract defines the signature and semantics of a resource-allocating function or method.
type AcquisitionContract struct {
	PackageName    string       `json:"package_name"`
	MethodName     string       `json:"method_name"`
	ResourceKind   ResourceKind `json:"resource_kind"`
	ResultIndex    int          `json:"result_index"`
	ErrorIndex     int          `json:"error_index"`
	IsSubfield     bool         `json:"is_subfield"`
	SubfieldName   string       `json:"subfield_name,omitempty"`
	IdempotentClose bool        `json:"idempotent_close"`
}

// ReleaseContract defines valid release actions for a specific resource kind.
type ReleaseContract struct {
	ResourceKind   ResourceKind `json:"resource_kind"`
	ReleaseMethods []string     `json:"release_methods"`
}

// ResourceRegistry maintains the inventory of supported resource acquisition and release contracts.
type ResourceRegistry struct {
	Acquisitions []AcquisitionContract
	Releases     map[ResourceKind]ReleaseContract
}

// DefaultResourceRegistry instantiates the canonical resource contracts documented in Milestone 3.4.
func DefaultResourceRegistry() *ResourceRegistry {
	reg := &ResourceRegistry{
		Acquisitions: []AcquisitionContract{
			// os.File
			{PackageName: "os", MethodName: "Open", ResourceKind: ResourceKindFile, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},
			{PackageName: "os", MethodName: "OpenFile", ResourceKind: ResourceKindFile, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},
			{PackageName: "os", MethodName: "Create", ResourceKind: ResourceKindFile, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},

			// net.Conn
			{PackageName: "net", MethodName: "Dial", ResourceKind: ResourceKindNetConn, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},
			{PackageName: "net", MethodName: "DialTimeout", ResourceKind: ResourceKindNetConn, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},
			{PackageName: "", MethodName: "Accept", ResourceKind: ResourceKindNetConn, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},

			// http.Response.Body (via http.Get, http.Post, client.Do, client.Get)
			{PackageName: "http", MethodName: "Get", ResourceKind: ResourceKindHTTPBody, ResultIndex: 0, ErrorIndex: 1, IsSubfield: true, SubfieldName: "Body", IdempotentClose: true},
			{PackageName: "http", MethodName: "Post", ResourceKind: ResourceKindHTTPBody, ResultIndex: 0, ErrorIndex: 1, IsSubfield: true, SubfieldName: "Body", IdempotentClose: true},
			{PackageName: "", MethodName: "Do", ResourceKind: ResourceKindHTTPBody, ResultIndex: 0, ErrorIndex: 1, IsSubfield: true, SubfieldName: "Body", IdempotentClose: true},

			// sql.Rows
			{PackageName: "", MethodName: "Query", ResourceKind: ResourceKindSQLRows, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},
			{PackageName: "", MethodName: "QueryContext", ResourceKind: ResourceKindSQLRows, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: true},

			// sql.Tx
			{PackageName: "", MethodName: "Begin", ResourceKind: ResourceKindSQLTx, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: false},
			{PackageName: "", MethodName: "BeginTx", ResourceKind: ResourceKindSQLTx, ResultIndex: 0, ErrorIndex: 1, IdempotentClose: false},

			// time.Timer
			{PackageName: "time", MethodName: "NewTimer", ResourceKind: ResourceKindTimer, ResultIndex: 0, ErrorIndex: -1, IdempotentClose: true},
		},
		Releases: map[ResourceKind]ReleaseContract{
			ResourceKindFile:     {ResourceKind: ResourceKindFile, ReleaseMethods: []string{"Close"}},
			ResourceKindNetConn:  {ResourceKind: ResourceKindNetConn, ReleaseMethods: []string{"Close"}},
			ResourceKindHTTPBody: {ResourceKind: ResourceKindHTTPBody, ReleaseMethods: []string{"Close"}},
			ResourceKindSQLRows:  {ResourceKind: ResourceKindSQLRows, ReleaseMethods: []string{"Close"}},
			ResourceKindSQLTx:    {ResourceKind: ResourceKindSQLTx, ReleaseMethods: []string{"Commit", "Rollback"}},
			ResourceKindTimer:    {ResourceKind: ResourceKindTimer, ReleaseMethods: []string{"Stop"}},
		},
	}
	return reg
}

// MatchAcquisition checks whether a call expression matches an acquisition contract.
func (reg *ResourceRegistry) MatchAcquisition(call *ast.CallExpr, fset *token.FileSet) (*AcquisitionContract, string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, "", false
	}

	methodName := sel.Sel.Name
	pkgOrVarName := ""

	if id, isId := sel.X.(*ast.Ident); isId {
		pkgOrVarName = id.Name
	}

	for _, ac := range reg.Acquisitions {
		if ac.MethodName != methodName {
			continue
		}

		if ac.PackageName != "" {
			if pkgOrVarName == ac.PackageName {
				return &ac, pkgOrVarName, true
			}
		} else {
			// Method on instance (e.g. db.Query, r.db.QueryContext, client.Do, listener.Accept)
			// Exclude if package name matches non-target stdlib packages
			if pkgOrVarName != "fmt" && pkgOrVarName != "log" && pkgOrVarName != "os" {
				return &ac, pkgOrVarName, true
			}
		}
	}

	return nil, "", false
}

// IsReleaseCall reports whether a call expression represents a valid release operation on the target resource.
func (reg *ResourceRegistry) IsReleaseCall(call *ast.CallExpr, kind ResourceKind, varName string, aliases []string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	relContract, hasContract := reg.Releases[kind]
	if !hasContract {
		return false
	}

	methodName := sel.Sel.Name
	methodMatches := false
	for _, rm := range relContract.ReleaseMethods {
		if rm == methodName {
			methodMatches = true
			break
		}
	}
	if !methodMatches {
		return false
	}

	// Verify receiver target matches varName or an alias
	// Case 1: direct method call on variable: f.Close(), tx.Commit(), timer.Stop()
	if id, ok := sel.X.(*ast.Ident); ok {
		if id.Name == varName {
			return true
		}
		for _, al := range aliases {
			if id.Name == al {
				return true
			}
		}
	}

	// Case 2: subfield method call: resp.Body.Close()
	if subSel, ok := sel.X.(*ast.SelectorExpr); ok {
		if id, ok := subSel.X.(*ast.Ident); ok {
			if id.Name == varName && subSel.Sel.Name == "Body" {
				return true
			}
			for _, al := range aliases {
				if id.Name == al && subSel.Sel.Name == "Body" {
					return true
				}
			}
		}
	}

	return false
}
