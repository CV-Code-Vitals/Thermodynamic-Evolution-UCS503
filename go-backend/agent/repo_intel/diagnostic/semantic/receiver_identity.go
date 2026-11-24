package semantic

import "go/ast"

// declaredReceiverType returns only an explicit source declaration. A method
// name alone is insufficient to establish a standard-library contract.
func declaredReceiverType(expr ast.Expr) (string, string) {
	id, ok := expr.(*ast.Ident)
	if !ok || id.Obj == nil {
		return "", ""
	}
	field, ok := id.Obj.Decl.(*ast.Field)
	if !ok {
		return "", ""
	}
	typ := field.Type
	if ptr, ok := typ.(*ast.StarExpr); ok {
		typ = ptr.X
	}
	sel, ok := typ.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", ""
	}
	return pkg.Name, sel.Sel.Name
}

func isExplicitNonSQLReceiver(expr ast.Expr) bool {
	pkg, name := declaredReceiverType(expr)
	return name != "" && !(pkg == "sql" && (name == "DB" || name == "Tx" || name == "Stmt" || name == "Conn"))
}
