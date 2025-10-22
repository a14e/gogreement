package immutable

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"goagreement/src/indexing"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/config"
	"goagreement/src/util"
)

func CheckImmutable(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) []ImmutableViolation {
	var violations []ImmutableViolation

	// Build indices for efficient lookup during AST traversal
	immutableTypes := indexing.BuildImmutableTypesIndex(pass, packageAnnotations)
	if immutableTypes.Len() == 0 {
		return violations // No immutable types to check
	}

	constructors := indexing.BuildConstructorIndex(pass, packageAnnotations)

	// Filter files based on configuration (skip test files by default)
	filesToCheck := config.Global.FilterFiles(pass)

	for _, file := range filesToCheck {
		currentFunction := ""

		// First pass: check simple assignments and inc/dec operations
		// We skip compound assignments (+=, -=, etc.) here to avoid duplicates
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				currentFunction = node.Name.Name
				return true

			case *ast.AssignStmt:
				// Only process compound assignments here
				// Check: x.field += value, x.field *= value, etc.
				if node.Tok != token.ASSIGN {
					v := checkCompoundAssignment(pass, node, immutableTypes, constructors, currentFunction)
					violations = append(violations, v...)
					return true
				}

				// Check: x.field = value, x.items[0] = value
				v := checkAssignment(pass, node, immutableTypes, constructors, currentFunction)
				violations = append(violations, v...)
				return true

			case *ast.IncDecStmt:
				// Check: x.field++, x.field--
				v := checkIncDec(pass, node, immutableTypes, constructors, currentFunction)
				violations = append(violations, v...)
				return true
			}
			return true
		})
	}

	return violations
}
func checkAssignment(
	pass *analysis.Pass,
	node *ast.AssignStmt,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) []ImmutableViolation {
	var violations []ImmutableViolation

	for _, lhs := range node.Lhs {
		violation := checkLHS(pass, node, lhs, immutableTypes, constructors, currentFunction)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

func checkLHS(
	pass *analysis.Pass,
	stmt *ast.AssignStmt,
	expr ast.Expr,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return checkFieldAssignment(pass, stmt, e, immutableTypes, constructors, currentFunction)
	case *ast.IndexExpr:
		return checkIndexAssignment(pass, stmt, e, immutableTypes, constructors, currentFunction)
	}

	return nil
}

func checkFieldAssignment(
	pass *analysis.Pass,
	stmt *ast.AssignStmt,
	selector *ast.SelectorExpr,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	// Get type of the receiver (t in t.field)
	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return nil
	}

	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	named, ok := receiverType.(*types.Named)
	if !ok {
		return nil
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil
	}

	pkgPath := pkg.Path()

	if !immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      selector.Pos(),
		Reason:   fmt.Sprintf("cannot assign to field %q of immutable type", selector.Sel.Name),
		Node:     stmt,
	}
}

func checkIndexAssignment(
	pass *analysis.Pass,
	stmt *ast.AssignStmt,
	index *ast.IndexExpr,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	selector, ok := index.X.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return nil
	}

	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	named, ok := receiverType.(*types.Named)
	if !ok {
		return nil
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil
	}

	pkgPath := pkg.Path()

	if !immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      index.Pos(),
		Reason:   fmt.Sprintf("cannot modify element of field %q of immutable type", selector.Sel.Name),
		Node:     stmt,
	}
}

func checkIncDec(
	pass *analysis.Pass,
	node *ast.IncDecStmt,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) []ImmutableViolation {
	var violations []ImmutableViolation

	selector, ok := node.X.(*ast.SelectorExpr)
	if !ok {
		return violations
	}

	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return violations
	}

	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	named, ok := receiverType.(*types.Named)
	if !ok {
		return violations
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return violations
	}

	pkgPath := pkg.Path()

	if !immutableTypes.Contains(pkgPath, typeName) {
		return violations
	}

	if constructors.Match(pkgPath, currentFunction, typeName) {
		return violations
	}

	op := "++"
	if node.Tok == token.DEC {
		op = "--"
	}

	violations = append(violations, ImmutableViolation{
		TypeName: typeName,
		Pos:      node.Pos(),
		Reason:   fmt.Sprintf("cannot use %s on field %q of immutable type (outside constructor)", op, selector.Sel.Name),
		Node:     node,
	})

	return violations
}

func checkCompoundAssignment(
	pass *analysis.Pass,
	node *ast.AssignStmt,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) []ImmutableViolation {
	var violations []ImmutableViolation

	for _, lhs := range node.Lhs {
		violation := checkCompoundLHS(pass, node, lhs, node.Tok, immutableTypes, constructors, currentFunction)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

func checkCompoundLHS(
	pass *analysis.Pass,
	stmt *ast.AssignStmt,
	expr ast.Expr,
	tok token.Token,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return nil
	}

	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	named, ok := receiverType.(*types.Named)
	if !ok {
		return nil
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil
	}

	pkgPath := pkg.Path()

	if !immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	op := tok.String()
	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      selector.Pos(),
		Reason:   fmt.Sprintf("cannot use %s on field %q of immutable type (outside constructor)", op, selector.Sel.Name),
		Node:     stmt,
	}
}
