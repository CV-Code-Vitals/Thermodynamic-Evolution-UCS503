package semantic

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"
	"unicode"
)

// SymbolTable maintains an indexed registry of declarations across a package or file set.
type SymbolTable struct {
	PackageName string               `json:"package_name"`
	PackagePath string               `json:"package_path"`
	Symbols     map[string]*Symbol   `json:"symbols"`      // Name/QualifiedName -> Symbol
	Methods     map[string][]*Symbol `json:"methods"`      // ReceiverType -> list of Methods
	Types       map[string]*Symbol   `json:"types"`        // TypeName -> Symbol
	Funcs       map[string]*Symbol   `json:"funcs"`        // FuncName -> Symbol
	Files       []string             `json:"files"`
}

// NewSymbolTable initializes an empty SymbolTable.
func NewSymbolTable(pkgName, pkgPath string) *SymbolTable {
	return &SymbolTable{
		PackageName: pkgName,
		PackagePath: pkgPath,
		Symbols:     make(map[string]*Symbol),
		Methods:     make(map[string][]*Symbol),
		Types:       make(map[string]*Symbol),
		Funcs:       make(map[string]*Symbol),
		Files:       make([]string, 0),
	}
}

// ExtractPackageSymbols extracts and indexes all symbols from a set of AST files.
func ExtractPackageSymbols(fset *token.FileSet, files []*ast.File, pkgPath string) (*SymbolTable, error) {
	if len(files) == 0 {
		return NewSymbolTable("", pkgPath), nil
	}

	pkgName := files[0].Name.Name
	st := NewSymbolTable(pkgName, pkgPath)

	for _, file := range files {
		fileName := ""
		if fset != nil {
			pos := fset.Position(file.Package)
			fileName = pos.Filename
		}
		st.Files = append(st.Files, fileName)
		st.extractFileSymbols(fset, file, fileName)
	}

	// Graceful type-checking augmentation using go/types
	st.tryTypeCheck(fset, files, pkgPath)

	return st, nil
}

// extractFileSymbols traverses declarations in a single file AST.
func (st *SymbolTable) extractFileSymbols(fset *token.FileSet, file *ast.File, fileName string) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			st.extractFuncDecl(fset, d, fileName)
		case *ast.GenDecl:
			st.extractGenDecl(fset, d, fileName)
		}
	}
}

// extractFuncDecl extracts function and method declarations.
func (st *SymbolTable) extractFuncDecl(fset *token.FileSet, fn *ast.FuncDecl, fileName string) {
	if fn.Name == nil {
		return
	}

	pos := fset.Position(fn.Pos())
	name := fn.Name.Name
	isExported := fn.Name.IsExported()

	sig := &FunctionSignature{
		Parameters: make([]ParamInfo, 0),
		Results:    make([]ReturnInfo, 0),
	}

	// Extract parameters
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			typeStr := formatTypeExpr(field.Type, fset)
			if len(field.Names) == 0 {
				sig.Parameters = append(sig.Parameters, ParamInfo{
					Name: "",
					Type: typeStr,
				})
			} else {
				for _, paramName := range field.Names {
					sig.Parameters = append(sig.Parameters, ParamInfo{
						Name: paramName.Name,
						Type: typeStr,
					})
				}
			}
		}
	}

	// Extract return results
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			typeStr := formatTypeExpr(field.Type, fset)
			isErr := typeStr == "error" || strings.HasSuffix(typeStr, ".error")
			if len(field.Names) == 0 {
				sig.Results = append(sig.Results, ReturnInfo{
					Type:    typeStr,
					IsError: isErr,
				})
			} else {
				for _, retName := range field.Names {
					sig.Results = append(sig.Results, ReturnInfo{
						Name:    retName.Name,
						Type:    typeStr,
						IsError: isErr,
					})
				}
			}
		}
	}

	doc := ""
	if fn.Doc != nil {
		doc = strings.TrimSpace(fn.Doc.Text())
	}

	// Receiver determination (Method vs Standalone Func)
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recvField := fn.Recv.List[0]
		recvType := formatTypeExpr(recvField.Type, fset)
		recvName := ""
		if len(recvField.Names) > 0 {
			recvName = recvField.Names[0].Name
		}

		cleanRecvType := strings.TrimPrefix(recvType, "*")
		qualName := fmt.Sprintf("%s.%s", cleanRecvType, name)

		sym := &Symbol{
			Name:         name,
			Kind:         SymbolMethod,
			Package:      st.PackageName,
			File:         fileName,
			Line:         pos.Line,
			Exported:     isExported,
			ReceiverType: recvType,
			ReceiverName: recvName,
			TypeString:   fmt.Sprintf("func(%s) %s", recvType, qualName),
			Signature:    sig,
			Doc:          doc,
			ASTNode:      fn,
		}

		st.Symbols[qualName] = sym
		st.Methods[cleanRecvType] = append(st.Methods[cleanRecvType], sym)
	} else {
		sym := &Symbol{
			Name:       name,
			Kind:       SymbolFunc,
			Package:    st.PackageName,
			File:       fileName,
			Line:       pos.Line,
			Exported:   isExported,
			TypeString: fmt.Sprintf("func %s", name),
			Signature:  sig,
			Doc:        doc,
			ASTNode:    fn,
		}

		st.Symbols[name] = sym
		st.Funcs[name] = sym
	}
}

// extractGenDecl extracts types, variables, and constants.
func (st *SymbolTable) extractGenDecl(fset *token.FileSet, gen *ast.GenDecl, fileName string) {
	for _, spec := range gen.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			pos := fset.Position(s.Pos())
			name := s.Name.Name
			kind := SymbolType
			if _, isInterface := s.Type.(*ast.InterfaceType); isInterface {
				kind = SymbolInterface
			}

			doc := ""
			if s.Doc != nil {
				doc = strings.TrimSpace(s.Doc.Text())
			} else if gen.Doc != nil {
				doc = strings.TrimSpace(gen.Doc.Text())
			}

			sym := &Symbol{
				Name:       name,
				Kind:       kind,
				Package:    st.PackageName,
				File:       fileName,
				Line:       pos.Line,
				Exported:   s.Name.IsExported(),
				TypeString: formatTypeExpr(s.Type, fset),
				Doc:        doc,
				ASTNode:    s,
			}
			st.Symbols[name] = sym
			st.Types[name] = sym

		case *ast.ValueSpec:
			kind := SymbolVar
			if gen.Tok == token.CONST {
				kind = SymbolConst
			}
			pos := fset.Position(s.Pos())
			typeStr := ""
			if s.Type != nil {
				typeStr = formatTypeExpr(s.Type, fset)
			}
			for _, id := range s.Names {
				sym := &Symbol{
					Name:       id.Name,
					Kind:       kind,
					Package:    st.PackageName,
					File:       fileName,
					Line:       pos.Line,
					Exported:   id.IsExported(),
					TypeString: typeStr,
					ASTNode:    s,
				}
				st.Symbols[id.Name] = sym
			}
		}
	}
}

// tryTypeCheck attempts standard type-checking with graceful degradation.
func (st *SymbolTable) tryTypeCheck(fset *token.FileSet, files []*ast.File, pkgPath string) {
	config := &types.Config{
		Error: func(err error) {
			// Suppress errors to allow partial/graceful resolution
		},
		FakeImportC: true,
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	pkg, err := config.Check(pkgPath, fset, files, info)
	if err != nil && pkg == nil {
		// Type-checking failed; graceful degradation ensures AST symbols remain intact
		return
	}

	// If type-checking produced definitions, enrich existing symbols
	for id, obj := range info.Defs {
		if obj == nil {
			continue
		}
		if sym, ok := st.Symbols[id.Name]; ok {
			sym.TypeString = obj.Type().String()
		}
	}
}

// LookupFunc finds a package-level function by name.
func (st *SymbolTable) LookupFunc(name string) *Symbol {
	return st.Funcs[name]
}

// LookupMethod finds a method given its clean receiver type and method name.
func (st *SymbolTable) LookupMethod(receiverType, methodName string) *Symbol {
	cleanType := strings.TrimPrefix(receiverType, "*")
	qualName := fmt.Sprintf("%s.%s", cleanType, methodName)
	if sym, ok := st.Symbols[qualName]; ok {
		return sym
	}
	for _, m := range st.Methods[cleanType] {
		if m.Name == methodName {
			return m
		}
	}
	return nil
}

// LookupType finds a declared type by name.
func (st *SymbolTable) LookupType(name string) *Symbol {
	return st.Types[name]
}

// formatTypeExpr converts an AST type expression to its string representation.
func formatTypeExpr(expr ast.Expr, fset *token.FileSet) string {
	if expr == nil {
		return ""
	}
	var buf bytes.Buffer
	if fset != nil {
		_ = printer.Fprint(&buf, fset, expr)
		return buf.String()
	}
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatTypeExpr(t.X, nil)
	case *ast.SelectorExpr:
		return formatTypeExpr(t.X, nil) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[" + formatTypeExpr(t.Len, nil) + "]" + formatTypeExpr(t.Elt, nil)
	case *ast.MapType:
		return "map[" + formatTypeExpr(t.Key, nil) + "]" + formatTypeExpr(t.Value, nil)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + formatTypeExpr(t.Elt, nil)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

// IsExportedIdentifier reports whether the identifier starts with an uppercase Unicode letter.
func IsExportedIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	return unicode.IsUpper([]rune(name)[0])
}
