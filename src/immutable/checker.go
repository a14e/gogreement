package immutable

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/util"
)

// ImmutableViolation represents a mutation of an immutable type
// @immutable
type ImmutableViolation struct {
	TypeName string
	Reason   string
	Pos      token.Pos
}

// CheckImmutable validates that immutable types are not mutated outside constructors
func CheckImmutable(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) []ImmutableViolation {
	var violations []ImmutableViolation

	// Build index of immutable types
	immutableTypes := buildImmutableTypesIndex(pass, packageAnnotations)

	if immutableTypes.Len() == 0 {
		return violations
	}

	// Build index of constructors
	constructors := buildConstructorIndex(pass, packageAnnotations)

	// Analyze each file
	for _, file := range pass.Files {
		// Find all function declarations to identify constructors
		currentFunction := ""

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				currentFunction = node.Name.Name
				return true

			case *ast.AssignStmt:
				// Check assignments: t.field = value or t.Items[0] = value
				v := checkAssignment(pass, node, immutableTypes, constructors, currentFunction)
				violations = append(violations, v...)
				return true

			case *ast.IncDecStmt:
				// Check ++ and --: t.value++
				v := checkIncDec(pass, node, immutableTypes, constructors, currentFunction)
				violations = append(violations, v...)
				return true
			}
			return true
		})

		// Second pass for compound assignments (+=, -=, etc.)
		// We need a separate pass because AssignStmt handles both = and +=
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				currentFunction = node.Name.Name
				return true

			case *ast.AssignStmt:
				// Check compound assignments: t.value += 1
				if node.Tok != token.ASSIGN {
					v := checkCompoundAssignment(pass, node, immutableTypes, constructors, currentFunction)
					violations = append(violations, v...)
				}
				return true
			}
			return true
		})
	}

	return violations
}

// buildImmutableTypesIndex creates an index of immutable types
func buildImmutableTypesIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.TypesMap {
	result := util.NewTypesMap()

	// Add types from current package - use full path
	for _, annot := range packageAnnotations.ImmutableAnnotations {
		result.Add(pass.Pkg.Path(), annot.OnType)
	}

	// Load facts from imported packages
	for _, imp := range pass.Pkg.Imports() {
		var importedAnnotations annotations.PackageAnnotations
		if pass.ImportPackageFact(imp, &importedAnnotations) {
			for _, annot := range importedAnnotations.ImmutableAnnotations {
				result.Add(imp.Path(), annot.OnType)
			}
		}
	}

	return result
}

// buildConstructorIndex creates an index of constructors
func buildConstructorIndex(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) util.FuncMap {
	result := util.NewFuncMap()

	// Add constructors from current package - use full path
	for _, annot := range packageAnnotations.ConstructorAnnotations {
		for _, constructorName := range annot.ConstructorNames {
			result.Add(pass.Pkg.Path(), constructorName, annot.OnType)
		}
	}

	// Load facts from imported packages
	for _, imp := range pass.Pkg.Imports() {
		var importedAnnotations annotations.PackageAnnotations
		if pass.ImportPackageFact(imp, &importedAnnotations) {
			for _, annot := range importedAnnotations.ConstructorAnnotations {
				for _, constructorName := range annot.ConstructorNames {
					result.Add(imp.Path(), constructorName, annot.OnType)
				}
			}
		}
	}

	return result
}

// checkAssignment checks if assignment violates immutability
func checkAssignment(
	pass *analysis.Pass,
	node *ast.AssignStmt,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) []ImmutableViolation {
	var violations []ImmutableViolation

	for _, lhs := range node.Lhs {
		violation := checkLHS(pass, lhs, immutableTypes, constructors, currentFunction)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// checkLHS checks left-hand side of assignment
func checkLHS(
	pass *analysis.Pass,
	expr ast.Expr,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		// t.field = value
		return checkFieldAssignment(pass, e, immutableTypes, constructors, currentFunction)

	case *ast.IndexExpr:
		// t.Items[0] = value or arr[i] = value
		return checkIndexAssignment(pass, e, immutableTypes, constructors, currentFunction)
	}

	return nil
}

// checkFieldAssignment checks field assignment: t.field = value
func checkFieldAssignment(
	pass *analysis.Pass,
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

	// Remove pointer
	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	// Get named type
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

	// Check if this type is immutable
	if !immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	// Check if we're in a constructor for this type
	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Violation: mutating immutable type outside constructor
	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      selector.Pos(),
		Reason:   fmt.Sprintf("cannot assign to field %q of immutable type (outside constructor)", selector.Sel.Name),
	}
}

// checkIndexAssignment checks index assignment: t.Items[0] = value
func checkIndexAssignment(
	pass *analysis.Pass,
	index *ast.IndexExpr,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	// Check if the indexed expression is a field of immutable type
	// t.Items[0] = value
	selector, ok := index.X.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	// Get type of receiver
	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return nil
	}

	// Remove pointer
	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	// Get named type
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

	// Check if this type is immutable
	if !immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	// Check if we're in a constructor
	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Violation
	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      index.Pos(),
		Reason:   fmt.Sprintf("cannot modify element of field %q of immutable type (outside constructor)", selector.Sel.Name),
	}
}

// checkIncDec checks increment/decrement: t.value++
func checkIncDec(
	pass *analysis.Pass,
	node *ast.IncDecStmt,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) []ImmutableViolation {
	var violations []ImmutableViolation

	// Check if it's a field access
	selector, ok := node.X.(*ast.SelectorExpr)
	if !ok {
		return violations
	}

	// Get type of receiver
	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return violations
	}

	// Remove pointer
	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	// Get named type
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

	// Check if immutable
	if !immutableTypes.Contains(pkgPath, typeName) {
		return violations
	}

	// Check if in constructor
	if constructors.Match(pkgPath, currentFunction, typeName) {
		return violations
	}

	// Violation
	op := "++"
	if node.Tok == token.DEC {
		op = "--"
	}

	violations = append(violations, ImmutableViolation{
		TypeName: typeName,
		Pos:      node.Pos(),
		Reason:   fmt.Sprintf("cannot use %s on field %q of immutable type (outside constructor)", op, selector.Sel.Name),
	})

	return violations
}

// checkCompoundAssignment checks compound assignments: t.value += 1
func checkCompoundAssignment(
	pass *analysis.Pass,
	node *ast.AssignStmt,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) []ImmutableViolation {
	var violations []ImmutableViolation

	for _, lhs := range node.Lhs {
		violation := checkCompoundLHS(pass, lhs, node.Tok, immutableTypes, constructors, currentFunction)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// checkCompoundLHS checks left-hand side of compound assignment
func checkCompoundLHS(
	pass *analysis.Pass,
	expr ast.Expr,
	tok token.Token,
	immutableTypes util.TypesMap,
	constructors util.FuncMap,
	currentFunction string,
) *ImmutableViolation {
	// Only check selector expressions: t.field += value
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	// Get type of receiver
	receiverType := pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return nil
	}

	// Remove pointer
	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	// Get named type
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

	// Check if immutable
	if !immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	// Check if in constructor
	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Violation
	op := tok.String()
	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      selector.Pos(),
		Reason:   fmt.Sprintf("cannot use %s on field %q of immutable type (outside constructor)", op, selector.Sel.Name),
	}
}

// getPkgPath safely extracts package path from types.Package
func getPkgPath(pkg *types.Package) string {
	if pkg == nil {
		return ""
	}
	return pkg.Path()
}

// ReportViolations reports all immutability violations
func ReportViolations(pass *analysis.Pass, violations []ImmutableViolation) {
	for _, v := range violations {
		msg := fmt.Sprintf("immutability violation in type %q: %s", v.TypeName, v.Reason)

		pass.Report(analysis.Diagnostic{
			Pos:     v.Pos,
			Message: msg,
		})
	}
}
