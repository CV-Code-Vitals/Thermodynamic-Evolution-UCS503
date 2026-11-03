package semantic

import (
	"go/ast"
	"go/token"
)

// SymbolKind represents the classification of a declared program entity.
type SymbolKind string

const (
	SymbolFunc      SymbolKind = "FUNC"
	SymbolMethod    SymbolKind = "METHOD"
	SymbolType      SymbolKind = "TYPE"
	SymbolVar       SymbolKind = "VAR"
	SymbolConst     SymbolKind = "CONST"
	SymbolParam     SymbolKind = "PARAM"
	SymbolInterface SymbolKind = "INTERFACE"
)

// ParamInfo holds the name and type string for a function parameter.
type ParamInfo struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// ReturnInfo holds the type and error classification for a function return value.
type ReturnInfo struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type"`
	IsError bool   `json:"is_error"`
}

// FunctionSignature represents the parsed parameter and result contract of a function or method.
type FunctionSignature struct {
	Parameters []ParamInfo  `json:"parameters"`
	Results    []ReturnInfo `json:"results"`
	Variadic   bool         `json:"variadic"`
}

// HasErrorReturn reports whether the function signature includes an error return type.
func (s *FunctionSignature) HasErrorReturn() bool {
	if s == nil {
		return false
	}
	for _, r := range s.Results {
		if r.IsError || r.Type == "error" {
			return true
		}
	}
	return false
}

// Symbol represents an extracted code entity with metadata, location, and type information.
type Symbol struct {
	Name         string             `json:"name"`
	Kind         SymbolKind         `json:"kind"`
	Package      string             `json:"package"`
	File         string             `json:"file"`
	Line         int                `json:"line"`
	Exported     bool               `json:"exported"`
	ReceiverType string             `json:"receiver_type,omitempty"`
	ReceiverName string             `json:"receiver_name,omitempty"`
	TypeString   string             `json:"type_string"`
	Signature    *FunctionSignature `json:"signature,omitempty"`
	Doc          string             `json:"doc,omitempty"`
	ASTNode      ast.Node           `json:"-"`
}

// BlockKind represents the role of a basic block in the Control Flow Graph.
type BlockKind string

const (
	BlockEntry       BlockKind = "ENTRY"
	BlockExit        BlockKind = "EXIT"
	BlockNormal      BlockKind = "NORMAL"
	BlockCondition   BlockKind = "CONDITION"
	BlockLoopHeader  BlockKind = "LOOP_HEADER"
	BlockLoopBody    BlockKind = "LOOP_BODY"
	BlockDefer       BlockKind = "DEFER"
	BlockReturn      BlockKind = "RETURN"
)

// DeferSite captures the invocation details of a deferred function call.
type DeferSite struct {
	CallExpr    *ast.CallExpr `json:"-"`
	TargetFunc  string        `json:"target_func"`  // e.g. "Close", "Rollback", "Stop"
	TargetVar   string        `json:"target_var"`   // e.g. "f", "resp.Body", "rows"
	ReceiverPkg string        `json:"receiver_pkg"` // e.g. "os", "net"
	Line        int           `json:"line"`
	InLoop      bool          `json:"in_loop"`
}

// Block represents a single basic block in the intraprocedural CFG.
type Block struct {
	ID           int          `json:"id"`
	Kind         BlockKind    `json:"kind"`
	Statements   []ast.Stmt   `json:"-"`
	Terminator   ast.Expr     `json:"-"`
	Predecessors []*Block     `json:"-"`
	Successors   []*Block     `json:"-"`
	TrueBranch   *Block       `json:"-"`
	FalseBranch  *Block       `json:"-"`
	Defers       []*DeferSite `json:"-"`
	LineStart    int          `json:"line_start"`
	LineEnd      int          `json:"line_end"`
}

// AddStatement appends a statement to the block and expands line bounds.
func (b *Block) AddStatement(stmt ast.Stmt, fset *token.FileSet) {
	b.Statements = append(b.Statements, stmt)
	if fset != nil {
		pos := fset.Position(stmt.Pos())
		end := fset.Position(stmt.End())
		if b.LineStart == 0 || pos.Line < b.LineStart {
			b.LineStart = pos.Line
		}
		if end.Line > b.LineEnd {
			b.LineEnd = end.Line
		}
	}
}

// ControlFlowGraph models intraprocedural control flow for a single function or method.
type ControlFlowGraph struct {
	FunctionName string       `json:"function_name"`
	ReceiverType string       `json:"receiver_type,omitempty"`
	File         string       `json:"file"`
	Entry        *Block       `json:"-"`
	Exit         *Block       `json:"-"`
	Blocks       []*Block     `json:"blocks"`
	ReturnBlocks []*Block     `json:"-"`
	AllDefers    []*DeferSite `json:"defers"`
	FuncDecl     *ast.FuncDecl `json:"-"`
}

// Definition represents a variable or parameter definition site in the CFG.
type Definition struct {
	ID          int      `json:"id"`
	VarName     string   `json:"var_name"`
	TypeName    string   `json:"type_name"`
	BlockID     int      `json:"block_id"`
	Line        int      `json:"line"`
	IsParameter bool     `json:"is_parameter"`
	Expr        ast.Expr `json:"-"`
	Stmt        ast.Stmt `json:"-"`
}

// VariableUse represents a variable read or evaluation site.
type VariableUse struct {
	VarName string   `json:"var_name"`
	BlockID int      `json:"block_id"`
	Line    int      `json:"line"`
	Expr    ast.Expr `json:"-"`
}

// ReachingDefinitionsAnalysis captures the results of intraprocedural reaching definitions analysis.
type ReachingDefinitionsAnalysis struct {
	CFG          *ControlFlowGraph        `json:"-"`
	Definitions  []*Definition            `json:"definitions"`
	Gen          map[int]map[int]bool     `json:"-"` // BlockID -> set of Def IDs
	Kill         map[int]map[int]bool     `json:"-"` // BlockID -> set of Def IDs
	In           map[int]map[int]bool     `json:"-"` // BlockID -> set of Def IDs
	Out          map[int]map[int]bool     `json:"-"` // BlockID -> set of Def IDs
	DefUseChains map[int][]VariableUse    `json:"-"` // Def ID -> uses
}

// ResourceAction represents a lifecycle transition on an allocated system resource.
type ResourceAction string

const (
	ActionAcquire  ResourceAction = "ACQUIRE"
	ActionRelease  ResourceAction = "RELEASE"
	ActionTransfer ResourceAction = "TRANSFER"
	ActionEscape   ResourceAction = "ESCAPE"
)

// ResourceEvent records a state-changing event on an external resource handle.
type ResourceEvent struct {
	Action      ResourceAction `json:"action"`
	VarName     string         `json:"var_name"`
	ResourceType string        `json:"resource_type"` // e.g. "*os.File", "*sql.Rows", "net.Conn", "http.Response.Body"
	Allocator   string         `json:"allocator"`     // e.g. "os.Open", "db.Query", "http.Get"
	Releaser    string         `json:"releaser"`      // e.g. "Close", "Rollback", "Stop"
	BlockID     int            `json:"block_id"`
	Line        int            `json:"line"`
	IsDeferred  bool           `json:"is_deferred"`
}

// ErrorHandlingState classifies how a returned error is processed by caller code.
type ErrorHandlingState string

const (
	ErrorCheckedInCondition ErrorHandlingState = "CHECKED_IN_CONDITION"
	ErrorReturnedDirectly   ErrorHandlingState = "RETURNED_DIRECTLY"
	ErrorExplicitlyDiscard  ErrorHandlingState = "EXPLICITLY_DISCARDED" // e.g. _ = call()
	ErrorIgnoredBareExpr    ErrorHandlingState = "IGNORED_IN_EXPR_STMT"  // e.g. call()
	ErrorWrapped            ErrorHandlingState = "WRAPPED"
)

// ErrorEvent records an error observation or handling site.
type ErrorEvent struct {
	CallExpression string             `json:"call_expression"`
	FunctionName   string             `json:"function_name"`
	TargetVar      string             `json:"target_var,omitempty"`
	State          ErrorHandlingState `json:"state"`
	BlockID        int                `json:"block_id"`
	Line           int                `json:"line"`
	IsDeferred     bool               `json:"is_deferred"`
}
