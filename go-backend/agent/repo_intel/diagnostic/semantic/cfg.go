package semantic

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

// CFGBuilder constructs an intraprocedural Control Flow Graph from an AST FuncDecl.
type CFGBuilder struct {
	fset         *token.FileSet
	filePath     string
	cfg          *ControlFlowGraph
	blockCounter int
	loopStack    []loopContext
}

type loopContext struct {
	header *Block
	exit   *Block
}

// BuildCFG constructs the complete Control Flow Graph for a function or method declaration.
func BuildCFG(fn *ast.FuncDecl, fset *token.FileSet, filePath string) *ControlFlowGraph {
	if fn == nil || fn.Body == nil {
		return nil
	}

	fnName := ""
	if fn.Name != nil {
		fnName = fn.Name.Name
	}

	recvType := ""
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvType = formatTypeExpr(fn.Recv.List[0].Type, fset)
	}

	cfg := &ControlFlowGraph{
		FunctionName: fnName,
		ReceiverType: recvType,
		File:         filePath,
		Blocks:       make([]*Block, 0),
		ReturnBlocks: make([]*Block, 0),
		AllDefers:    make([]*DeferSite, 0),
		FuncDecl:     fn,
	}

	builder := &CFGBuilder{
		fset:      fset,
		filePath:  filePath,
		cfg:       cfg,
		loopStack: make([]loopContext, 0),
	}

	// 1. Create Entry and Exit blocks
	entry := builder.newBlock(BlockEntry)
	exit := builder.newBlock(BlockExit)
	cfg.Entry = entry
	cfg.Exit = exit

	if fset != nil {
		entryPos := fset.Position(fn.Pos())
		entry.LineStart = entryPos.Line
		entry.LineEnd = entryPos.Line

		exitPos := fset.Position(fn.End())
		exit.LineStart = exitPos.Line
		exit.LineEnd = exitPos.Line
	}

	// 2. Build CFG through the function body
	lastBlock := builder.buildBlockStmt(fn.Body, entry)

	// 3. Connect fallthrough exit to Exit block if not already returned
	if lastBlock != nil && len(lastBlock.Successors) == 0 && lastBlock.Kind != BlockReturn {
		builder.addEdge(lastBlock, exit)
		cfg.ReturnBlocks = append(cfg.ReturnBlocks, lastBlock)
	}

	// 4. Connect all Return blocks to Exit
	for _, retBlock := range cfg.ReturnBlocks {
		if !hasEdge(retBlock, exit) {
			builder.addEdge(retBlock, exit)
		}
	}

	return cfg
}

func (b *CFGBuilder) newBlock(kind BlockKind) *Block {
	blk := &Block{
		ID:           b.blockCounter,
		Kind:         kind,
		Statements:   make([]ast.Stmt, 0),
		Predecessors: make([]*Block, 0),
		Successors:   make([]*Block, 0),
		Defers:       make([]*DeferSite, 0),
	}
	b.blockCounter++
	b.cfg.Blocks = append(b.cfg.Blocks, blk)
	return blk
}

func (b *CFGBuilder) addEdge(from, to *Block) {
	if from == nil || to == nil {
		return
	}
	if !hasEdge(from, to) {
		from.Successors = append(from.Successors, to)
		to.Predecessors = append(to.Predecessors, from)
	}
}

func hasEdge(from, to *Block) bool {
	for _, s := range from.Successors {
		if s == to {
			return true
		}
	}
	return false
}

// buildBlockStmt traverses statements in a block, chaining them into basic blocks.
func (b *CFGBuilder) buildBlockStmt(body *ast.BlockStmt, current *Block) *Block {
	if body == nil {
		return current
	}

	curr := current
	for _, stmt := range body.List {
		if curr == nil {
			// Unreachable code after an unconditional return/jump; create dead-code block
			curr = b.newBlock(BlockNormal)
		}

		switch s := stmt.(type) {
		case *ast.IfStmt:
			curr = b.buildIfStmt(s, curr)

		case *ast.ForStmt:
			curr = b.buildForStmt(s, curr)

		case *ast.RangeStmt:
			curr = b.buildRangeStmt(s, curr)

		case *ast.SwitchStmt:
			curr = b.buildSwitchStmt(s, curr)

		case *ast.TypeSwitchStmt:
			curr = b.buildTypeSwitchStmt(s, curr)

		case *ast.ReturnStmt:
			curr.AddStatement(s, b.fset)
			curr.Kind = BlockReturn
			b.cfg.ReturnBlocks = append(b.cfg.ReturnBlocks, curr)
			curr = nil // Control terminates here

		case *ast.BranchStmt:
			curr.AddStatement(s, b.fset)
			b.handleBranchStmt(s, curr)
			curr = nil // Control transferred via jump

		case *ast.DeferStmt:
			curr.AddStatement(s, b.fset)
			deferSite := b.extractDeferSite(s)
			curr.Defers = append(curr.Defers, deferSite)
			b.cfg.AllDefers = append(b.cfg.AllDefers, deferSite)

		default:
			curr.AddStatement(s, b.fset)
		}
	}

	return curr
}

// buildIfStmt constructs conditional branches and joins.
func (b *CFGBuilder) buildIfStmt(stmt *ast.IfStmt, current *Block) *Block {
	if stmt.Init != nil {
		current.AddStatement(stmt.Init, b.fset)
	}

	condBlock := current
	if len(condBlock.Statements) > 0 {
		// Split into a dedicated condition block if current already has statements
		condBlock = b.newBlock(BlockCondition)
		b.addEdge(current, condBlock)
	} else {
		condBlock.Kind = BlockCondition
	}
	condBlock.Terminator = stmt.Cond

	thenEntry := b.newBlock(BlockNormal)
	b.addEdge(condBlock, thenEntry)
	condBlock.TrueBranch = thenEntry

	thenExit := b.buildBlockStmt(stmt.Body, thenEntry)

	var elseExit *Block
	joinBlock := b.newBlock(BlockNormal)

	if stmt.Else != nil {
		elseEntry := b.newBlock(BlockNormal)
		b.addEdge(condBlock, elseEntry)
		condBlock.FalseBranch = elseEntry

		switch e := stmt.Else.(type) {
		case *ast.BlockStmt:
			elseExit = b.buildBlockStmt(e, elseEntry)
		case *ast.IfStmt:
			elseExit = b.buildIfStmt(e, elseEntry)
		default:
			elseEntry.AddStatement(e, b.fset)
			elseExit = elseEntry
		}
	} else {
		// Direct fallthrough to join when false
		b.addEdge(condBlock, joinBlock)
		condBlock.FalseBranch = joinBlock
	}

	if thenExit != nil {
		b.addEdge(thenExit, joinBlock)
	}
	if elseExit != nil {
		b.addEdge(elseExit, joinBlock)
	}

	// If both branches terminated in returns, joinBlock is unreachable
	if len(joinBlock.Predecessors) == 0 {
		return nil
	}

	return joinBlock
}

// buildForStmt handles standard for loops.
func (b *CFGBuilder) buildForStmt(stmt *ast.ForStmt, current *Block) *Block {
	if stmt.Init != nil {
		current.AddStatement(stmt.Init, b.fset)
	}

	header := b.newBlock(BlockLoopHeader)
	header.Terminator = stmt.Cond
	b.addEdge(current, header)

	bodyEntry := b.newBlock(BlockLoopBody)
	b.addEdge(header, bodyEntry)
	header.TrueBranch = bodyEntry

	loopExit := b.newBlock(BlockNormal)
	b.addEdge(header, loopExit)
	header.FalseBranch = loopExit

	b.loopStack = append(b.loopStack, loopContext{header: header, exit: loopExit})
	bodyExit := b.buildBlockStmt(stmt.Body, bodyEntry)
	b.loopStack = b.loopStack[:len(b.loopStack)-1]

	if bodyExit != nil {
		if stmt.Post != nil {
			bodyExit.AddStatement(stmt.Post, b.fset)
		}
		b.addEdge(bodyExit, header) // Loop back edge
	}

	return loopExit
}

// buildRangeStmt handles for range loops.
func (b *CFGBuilder) buildRangeStmt(stmt *ast.RangeStmt, current *Block) *Block {
	header := b.newBlock(BlockLoopHeader)
	header.Terminator = stmt.X
	b.addEdge(current, header)

	bodyEntry := b.newBlock(BlockLoopBody)
	b.addEdge(header, bodyEntry)
	header.TrueBranch = bodyEntry

	loopExit := b.newBlock(BlockNormal)
	b.addEdge(header, loopExit)
	header.FalseBranch = loopExit

	b.loopStack = append(b.loopStack, loopContext{header: header, exit: loopExit})
	bodyExit := b.buildBlockStmt(stmt.Body, bodyEntry)
	b.loopStack = b.loopStack[:len(b.loopStack)-1]

	if bodyExit != nil {
		b.addEdge(bodyExit, header) // Loop back edge
	}

	return loopExit
}

// buildSwitchStmt handles switch statements.
func (b *CFGBuilder) buildSwitchStmt(stmt *ast.SwitchStmt, current *Block) *Block {
	if stmt.Init != nil {
		current.AddStatement(stmt.Init, b.fset)
	}

	switchExit := b.newBlock(BlockNormal)
	hasDefault := false

	for _, clause := range stmt.Body.List {
		cc, ok := clause.(*ast.CaseClause)
		if !ok {
			continue
		}

		if cc.List == nil {
			hasDefault = true
		}

		caseBlock := b.newBlock(BlockNormal)
		b.addEdge(current, caseBlock)

		caseExit := b.buildBlockStmt(&ast.BlockStmt{List: cc.Body}, caseBlock)
		if caseExit != nil {
			b.addEdge(caseExit, switchExit)
		}
	}

	if !hasDefault {
		b.addEdge(current, switchExit)
	}

	if len(switchExit.Predecessors) == 0 {
		return nil
	}

	return switchExit
}

// buildTypeSwitchStmt handles type switches.
func (b *CFGBuilder) buildTypeSwitchStmt(stmt *ast.TypeSwitchStmt, current *Block) *Block {
	if stmt.Init != nil {
		current.AddStatement(stmt.Init, b.fset)
	}
	if stmt.Assign != nil {
		current.AddStatement(stmt.Assign, b.fset)
	}

	switchExit := b.newBlock(BlockNormal)
	hasDefault := false

	for _, clause := range stmt.Body.List {
		cc, ok := clause.(*ast.CaseClause)
		if !ok {
			continue
		}

		if cc.List == nil {
			hasDefault = true
		}

		caseBlock := b.newBlock(BlockNormal)
		b.addEdge(current, caseBlock)

		caseExit := b.buildBlockStmt(&ast.BlockStmt{List: cc.Body}, caseBlock)
		if caseExit != nil {
			b.addEdge(caseExit, switchExit)
		}
	}

	if !hasDefault {
		b.addEdge(current, switchExit)
	}

	if len(switchExit.Predecessors) == 0 {
		return nil
	}

	return switchExit
}

// handleBranchStmt connects break and continue jumps to their targets.
func (b *CFGBuilder) handleBranchStmt(stmt *ast.BranchStmt, current *Block) {
	if len(b.loopStack) == 0 {
		return
	}
	innermost := b.loopStack[len(b.loopStack)-1]

	switch stmt.Tok {
	case token.BREAK:
		b.addEdge(current, innermost.exit)
	case token.CONTINUE:
		b.addEdge(current, innermost.header)
	}
}

// extractDeferSite inspects a defer statement for target function, variable, and loop nesting.
func (b *CFGBuilder) extractDeferSite(stmt *ast.DeferStmt) *DeferSite {
	pos := b.fset.Position(stmt.Pos())
	site := &DeferSite{
		CallExpr: stmt.Call,
		Line:     pos.Line,
		InLoop:   len(b.loopStack) > 0,
	}

	if stmt.Call == nil {
		return site
	}

	switch fn := stmt.Call.Fun.(type) {
	case *ast.SelectorExpr:
		site.TargetFunc = fn.Sel.Name
		var buf bytes.Buffer
		_ = printer.Fprint(&buf, b.fset, fn.X)
		site.TargetVar = strings.TrimSpace(buf.String())

		if id, ok := fn.X.(*ast.Ident); ok {
			site.ReceiverPkg = id.Name
		}
	case *ast.Ident:
		site.TargetFunc = fn.Name
	case *ast.FuncLit:
		site.TargetFunc = "anonymous"
	}

	return site
}
