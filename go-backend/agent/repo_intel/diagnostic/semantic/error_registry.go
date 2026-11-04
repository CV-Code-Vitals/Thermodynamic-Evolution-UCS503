package semantic

import (
	"go/ast"
	"go/types"
	"strings"
)

// StandardErrorSignature specifies a well-known error-producing function or method contract.
type StandardErrorSignature struct {
	NumReturns int
	ErrorIndex int
	Category   ErrorOperationCategory
}

// ErrorRegistry provides signature resolution for error-returning calls across standard library, local, and type-checked functions.
type ErrorRegistry struct {
	stdAPIs map[string]StandardErrorSignature
}

// NewErrorRegistry initializes standard library contracts and classification rules.
func NewErrorRegistry() *ErrorRegistry {
	r := &ErrorRegistry{
		stdAPIs: make(map[string]StandardErrorSignature),
	}
	r.registerStandardAPIs()
	return r
}

func (r *ErrorRegistry) registerStandardAPIs() {
	// encoding/json
	r.stdAPIs["json.Marshal"] = StandardErrorSignature{2, 1, ErrorCatSerialization}
	r.stdAPIs["json.Unmarshal"] = StandardErrorSignature{1, 0, ErrorCatSerialization}
	r.stdAPIs["json.MarshalIndent"] = StandardErrorSignature{2, 1, ErrorCatSerialization}
	r.stdAPIs["json.NewEncoder.Encode"] = StandardErrorSignature{1, 0, ErrorCatSerialization}
	r.stdAPIs["json.NewDecoder.Decode"] = StandardErrorSignature{1, 0, ErrorCatSerialization}

	// os
	r.stdAPIs["os.Open"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
	r.stdAPIs["os.OpenFile"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
	r.stdAPIs["os.Create"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
	r.stdAPIs["os.ReadFile"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
	r.stdAPIs["os.WriteFile"] = StandardErrorSignature{1, 0, ErrorCatFileIO}
	r.stdAPIs["os.Remove"] = StandardErrorSignature{1, 0, ErrorCatFileIO}
	r.stdAPIs["os.RemoveAll"] = StandardErrorSignature{1, 0, ErrorCatFileIO}
	r.stdAPIs["os.Mkdir"] = StandardErrorSignature{1, 0, ErrorCatFileIO}
	r.stdAPIs["os.MkdirAll"] = StandardErrorSignature{1, 0, ErrorCatFileIO}
	r.stdAPIs["os.Chmod"] = StandardErrorSignature{1, 0, ErrorCatFileIO}
	r.stdAPIs["os.Rename"] = StandardErrorSignature{1, 0, ErrorCatFileIO}

	// net
	r.stdAPIs["net.Dial"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["net.DialTimeout"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["net.DialTCP"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["net.Listen"] = StandardErrorSignature{2, 1, ErrorCatNetwork}

	// net/http
	r.stdAPIs["http.Get"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["http.Post"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["http.PostForm"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["http.Head"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["http.NewRequest"] = StandardErrorSignature{2, 1, ErrorCatNetwork}
	r.stdAPIs["http.NewRequestWithContext"] = StandardErrorSignature{2, 1, ErrorCatNetwork}

	// database/sql
	r.stdAPIs["sql.Open"] = StandardErrorSignature{2, 1, ErrorCatDatabaseQuery}
	r.stdAPIs["sql.OpenDB"] = StandardErrorSignature{1, 0, ErrorCatDatabaseQuery}

	// io
	r.stdAPIs["io.ReadAll"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
	r.stdAPIs["io.Copy"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
	r.stdAPIs["io.ReadFull"] = StandardErrorSignature{2, 1, ErrorCatFileIO}
}

// ResolveCallSignature attempts to determine if a call returns an error, its return position, and its category.
func (r *ErrorRegistry) ResolveCallSignature(
	call *ast.CallExpr,
	typesInfo *types.Info,
	localFuncs map[string]*ast.FuncDecl,
) (ErrorReturnSignature, ErrorOperationCategory) {
	// 1. If go/types information is available, use type-checked signature
	if typesInfo != nil {
		if sig := r.resolveWithTypes(call, typesInfo); sig.IsErrorReturning {
			cat := r.categorizeFromCall(call)
			return sig, cat
		}
	}

	// 2. Check standard library contracts
	callName := callExprToString(call.Fun)
	if stdSig, ok := r.stdAPIs[callName]; ok {
		return ErrorReturnSignature{
			IsErrorReturning: true,
			ErrorIndex:       stdSig.ErrorIndex,
			NumReturns:       stdSig.NumReturns,
			Resolution:       SignatureASTInferred,
		}, stdSig.Category
	}

	// Handle chained calls like json.NewEncoder(w).Encode(data)
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		methodName := sel.Sel.Name
		if innerCall, ok := sel.X.(*ast.CallExpr); ok {
			innerName := callExprToString(innerCall.Fun)
			chainedKey := innerName + "." + methodName
			if stdSig, ok := r.stdAPIs[chainedKey]; ok {
				return ErrorReturnSignature{
					IsErrorReturning: true,
					ErrorIndex:       stdSig.ErrorIndex,
					NumReturns:       stdSig.NumReturns,
					Resolution:       SignatureASTInferred,
				}, stdSig.Category
			}
		}

		// Method calls on known receivers or common idioms
		switch methodName {
		case "Close":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       0,
				NumReturns:       1,
				Resolution:       SignatureASTInferred,
			}, ErrorCatCleanup

		case "Rollback":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       0,
				NumReturns:       1,
				Resolution:       SignatureASTInferred,
			}, ErrorCatCleanup

		case "Commit":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       0,
				NumReturns:       1,
				Resolution:       SignatureASTInferred,
			}, ErrorCatDatabaseTx

		case "Encode", "Decode":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       0,
				NumReturns:       1,
				Resolution:       SignatureASTInferred,
			}, ErrorCatSerialization

		case "Scan":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       0,
				NumReturns:       1,
				Resolution:       SignatureASTInferred,
			}, ErrorCatDatabaseQuery

		case "Query", "QueryContext":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       1,
				NumReturns:       2,
				Resolution:       SignatureASTInferred,
			}, ErrorCatDatabaseQuery

		case "Exec", "ExecContext":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       1,
				NumReturns:       2,
				Resolution:       SignatureASTInferred,
			}, ErrorCatDatabaseQuery

		case "Begin", "BeginTx":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       1,
				NumReturns:       2,
				Resolution:       SignatureASTInferred,
			}, ErrorCatDatabaseTx

		case "Ping", "PingContext":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       0,
				NumReturns:       1,
				Resolution:       SignatureASTInferred,
			}, ErrorCatDatabaseQuery

		case "Write", "Read":
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       1,
				NumReturns:       2,
				Resolution:       SignatureASTInferred,
			}, ErrorCatFileIO
		}
	}

	// 3. Check locally defined functions in the same file or package AST
	if id, ok := call.Fun.(*ast.Ident); ok {
		if fnDecl, ok := localFuncs[id.Name]; ok && fnDecl.Type != nil {
			if sig := r.resolveFromFuncDecl(fnDecl); sig.IsErrorReturning {
				return sig, ErrorCatGeneral
			}
		}
	}

	return ErrorReturnSignature{
		IsErrorReturning: false,
		Resolution:       SignatureUnresolved,
	}, ErrorCatGeneral
}

// resolveWithTypes inspects the go/types type of a call expression.
func (r *ErrorRegistry) resolveWithTypes(call *ast.CallExpr, typesInfo *types.Info) ErrorReturnSignature {
	tv, ok := typesInfo.Types[call.Fun]
	if !ok || tv.Type == nil {
		return ErrorReturnSignature{IsErrorReturning: false, Resolution: SignatureUnresolved}
	}

	sig, ok := tv.Type.(*types.Signature)
	if !ok {
		return ErrorReturnSignature{IsErrorReturning: false, Resolution: SignatureUnresolved}
	}

	results := sig.Results()
	if results == nil || results.Len() == 0 {
		return ErrorReturnSignature{IsErrorReturning: false, Resolution: SignatureTypeResolved}
	}

	numReturns := results.Len()
	for i := 0; i < numReturns; i++ {
		v := results.At(i)
		if v.Type().String() == "error" {
			namedVar := v.Name()
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       i,
				NumReturns:       numReturns,
				Resolution:       SignatureTypeResolved,
				HasNamedReturns:  namedVar != "",
				NamedErrorVar:    namedVar,
			}
		}
	}

	return ErrorReturnSignature{
		IsErrorReturning: false,
		NumReturns:       numReturns,
		Resolution:       SignatureTypeResolved,
	}
}

// resolveFromFuncDecl analyzes an AST FuncDecl to inspect its return results for an error type.
func (r *ErrorRegistry) resolveFromFuncDecl(fnDecl *ast.FuncDecl) ErrorReturnSignature {
	if fnDecl.Type == nil || fnDecl.Type.Results == nil {
		return ErrorReturnSignature{IsErrorReturning: false, Resolution: SignatureASTInferred}
	}

	fields := fnDecl.Type.Results.List
	var retTypes []string
	var namedVars []string

	for _, f := range fields {
		typeStr := typeExprToString(f.Type)
		if len(f.Names) > 0 {
			for _, name := range f.Names {
				retTypes = append(retTypes, typeStr)
				namedVars = append(namedVars, name.Name)
			}
		} else {
			retTypes = append(retTypes, typeStr)
			namedVars = append(namedVars, "")
		}
	}

	numReturns := len(retTypes)
	for i, t := range retTypes {
		if t == "error" {
			return ErrorReturnSignature{
				IsErrorReturning: true,
				ErrorIndex:       i,
				NumReturns:       numReturns,
				Resolution:       SignatureASTInferred,
				HasNamedReturns:  namedVars[i] != "",
				NamedErrorVar:    namedVars[i],
			}
		}
	}

	return ErrorReturnSignature{
		IsErrorReturning: false,
		NumReturns:       numReturns,
		Resolution:       SignatureASTInferred,
	}
}

func (r *ErrorRegistry) categorizeFromCall(call *ast.CallExpr) ErrorOperationCategory {
	name := callExprToString(call.Fun)
	switch {
	case strings.Contains(name, "json.") || strings.Contains(name, "xml.") || strings.Contains(name, "yaml.") || strings.Contains(name, "Unmarshal") || strings.Contains(name, "Marshal"):
		return ErrorCatSerialization
	case strings.Contains(name, "os.") || strings.Contains(name, "io.") || strings.Contains(name, "File"):
		return ErrorCatFileIO
	case strings.Contains(name, "net.") || strings.Contains(name, "http.") || strings.Contains(name, "Dial") || strings.Contains(name, "Client"):
		return ErrorCatNetwork
	case strings.Contains(name, "Commit"):
		return ErrorCatDatabaseTx
	case strings.Contains(name, "Rollback") || strings.Contains(name, "Close"):
		return ErrorCatCleanup
	case strings.Contains(name, "sql.") || strings.Contains(name, "Query") || strings.Contains(name, "Exec") || strings.Contains(name, "Scan"):
		return ErrorCatDatabaseQuery
	default:
		return ErrorCatGeneral
	}
}

func callExprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return callExprToString(e.X) + "." + e.Sel.Name
	case *ast.CallExpr:
		return callExprToString(e.Fun)
	default:
		return ""
	}
}

func typeExprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return typeExprToString(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + typeExprToString(e.X)
	case *ast.ArrayType:
		return "[]" + typeExprToString(e.Elt)
	default:
		return ""
	}
}

