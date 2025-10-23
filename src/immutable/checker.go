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

func CheckImmutable(pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) []ImmutableViolation {
	var violations []ImmutableViolation

	// Build indices for efficient lookup during AST traversal
	immutableTypes := indexing.BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass)
	if immutableTypes.Len() == 0 {
		return violations // No immutable types to check
	}

	constructors := indexing.BuildConstructorIndex[*annotations.ConstructorCheckerFact](pass, packageAnnotations)

	// Filter files based on configuration (skip test files by default)
	filesToCheck := config.Global.FilterFiles(pass)

	ctx := &checkerContext{
		pass:            pass,
		immutableTypes:  immutableTypes,
		constructors:    constructors,
		currentFunction: nil,
	}

	for _, file := range filesToCheck {

		// First pass: check simple assignments and inc/dec operations
		// We skip compound assignments (+=, -=, etc.) here to avoid duplicates
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				ctx.currentFunction = &node.Name.Name

				// Track receiver information for methods
				ctx.currentReceiver = extractReceiverInfo(ctx.pass, node)
				return true

			case *ast.AssignStmt:
				// Only process compound assignments here
				// Check: x.field += value, x.field *= value, etc.
				if node.Tok != token.ASSIGN {
					v := checkCompoundAssignment(ctx, node)
					violations = append(violations, v...)
					return true
				}

				// Check: x.field = value, x.items[0] = value
				v := checkAssignment(ctx, node)
				violations = append(violations, v...)
				return true

			case *ast.IncDecStmt:
				// Check: x.field++, x.field--
				v := checkIncDec(ctx, node)
				violations = append(violations, v...)
				return true
			}
			return true
		})
	}

	return violations
}

type checkerContext struct {
	pass            *analysis.Pass
	immutableTypes  util.TypesMap
	constructors    util.TypeFuncRegistry
	currentFunction *string
	currentReceiver *receiverInfo
}

// receiverInfo contains information about a method's receiver
// @immutable
type receiverInfo struct {
	name     string
	typeName string
	pkgPath  string
}

// extractReceiverInfo extracts receiver information from a method declaration
func extractReceiverInfo(pass *analysis.Pass, funcDecl *ast.FuncDecl) *receiverInfo {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return nil
	}

	recvField := funcDecl.Recv.List[0]
	if len(recvField.Names) == 0 {
		return nil
	}

	recvName := recvField.Names[0].Name
	recvType := pass.TypesInfo.TypeOf(recvField.Type)
	if recvType == nil {
		return nil
	}

	typeInfo := util.ExtractTypeInfo(recvType)
	if typeInfo == nil {
		return nil
	}

	return &receiverInfo{
		name:     recvName,
		typeName: typeInfo.TypeName,
		pkgPath:  typeInfo.PkgPath,
	}
}

func checkAssignment(
	ctx *checkerContext,
	node *ast.AssignStmt,
) []ImmutableViolation {
	var violations []ImmutableViolation

	for _, lhs := range node.Lhs {
		violation := checkLHS(ctx, node, lhs)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

func checkLHS(
	ctx *checkerContext,
	stmt *ast.AssignStmt,
	expr ast.Expr,
) *ImmutableViolation {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return checkFieldAssignment(ctx, stmt, e)
	case *ast.IndexExpr:
		return checkIndexAssignment(ctx, stmt, e)
	case *ast.StarExpr:
		// Check for receiver reassignment: *receiver = value
		return checkReceiverReassignment(ctx, stmt, e)
	}

	return nil
}

func checkFieldAssignment(
	ctx *checkerContext,
	stmt *ast.AssignStmt,
	selector *ast.SelectorExpr,
) *ImmutableViolation {
	// Get type of the receiver (t in t.field)
	receiverType := ctx.pass.TypesInfo.TypeOf(selector.X)
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

	if !ctx.immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if ctx.constructors.Match(pkgPath, *ctx.currentFunction, typeName) {
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
	ctx *checkerContext,
	stmt *ast.AssignStmt,
	index *ast.IndexExpr,
) *ImmutableViolation {
	selector, ok := index.X.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	receiverType := ctx.pass.TypesInfo.TypeOf(selector.X)
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

	if !ctx.immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if ctx.constructors.Match(pkgPath, *ctx.currentFunction, typeName) {
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
	ctx *checkerContext,
	node *ast.IncDecStmt,
) []ImmutableViolation {
	var violations []ImmutableViolation

	// Check for field increment/decrement: x.field++
	if selector, ok := node.X.(*ast.SelectorExpr); ok {
		violation := checkFieldIncDec(ctx, node, selector)
		if violation != nil {
			violations = append(violations, *violation)
		}
		return violations
	}

	// Check for receiver increment/decrement: *receiver++
	if star, ok := node.X.(*ast.StarExpr); ok {
		violation := checkReceiverIncDec(ctx, node, star)
		if violation != nil {
			violations = append(violations, *violation)
		}
		return violations
	}

	return violations
}

func checkFieldIncDec(
	ctx *checkerContext,
	node *ast.IncDecStmt,
	selector *ast.SelectorExpr,
) *ImmutableViolation {
	receiverType := ctx.pass.TypesInfo.TypeOf(selector.X)
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

	if !ctx.immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if ctx.constructors.Match(pkgPath, *ctx.currentFunction, typeName) {
		return nil
	}

	op := "++"
	if node.Tok == token.DEC {
		op = "--"
	}

	return &ImmutableViolation{
		TypeName: typeName,
		Pos:      node.Pos(),
		Reason:   fmt.Sprintf("cannot use %s on field %q of immutable type (outside constructor)", op, selector.Sel.Name),
		Node:     node,
	}
}

func checkReceiverIncDec(
	ctx *checkerContext,
	node *ast.IncDecStmt,
	star *ast.StarExpr,
) *ImmutableViolation {
	// Check if we're in a method with a receiver
	if ctx.currentReceiver == nil {
		return nil
	}

	// Check if the increment/decrement is on the receiver: *receiver++
	ident, ok := star.X.(*ast.Ident)
	if !ok {
		return nil
	}

	// Check if the identifier is the receiver
	if ident.Name != ctx.currentReceiver.name {
		return nil
	}

	// Check if the receiver type is immutable
	if !ctx.immutableTypes.Contains(ctx.currentReceiver.pkgPath, ctx.currentReceiver.typeName) {
		return nil
	}

	// Allow in constructors
	if ctx.constructors.Match(ctx.currentReceiver.pkgPath, *ctx.currentFunction, ctx.currentReceiver.typeName) {
		return nil
	}

	op := "++"
	if node.Tok == token.DEC {
		op = "--"
	}

	return &ImmutableViolation{
		TypeName: ctx.currentReceiver.typeName,
		Pos:      star.Pos(),
		Reason:   fmt.Sprintf("cannot use %s on immutable receiver (outside constructor)", op),
		Node:     node,
	}
}

func checkCompoundAssignment(
	ctx *checkerContext,
	node *ast.AssignStmt,
) []ImmutableViolation {
	var violations []ImmutableViolation

	for _, lhs := range node.Lhs {
		violation := checkCompoundLHS(ctx, node, lhs, node.Tok)
		if violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

func checkCompoundLHS(
	ctx *checkerContext,
	stmt *ast.AssignStmt,
	expr ast.Expr,
	tok token.Token,
) *ImmutableViolation {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	receiverType := ctx.pass.TypesInfo.TypeOf(selector.X)
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

	if !ctx.immutableTypes.Contains(pkgPath, typeName) {
		return nil
	}

	if ctx.constructors.Match(pkgPath, *ctx.currentFunction, typeName) {
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

// checkReceiverReassignment checks if a method reassigns its receiver (*receiver = value)
// This is only checked for methods in the same package where the type is declared
func checkReceiverReassignment(
	ctx *checkerContext,
	stmt *ast.AssignStmt,
	star *ast.StarExpr,
) *ImmutableViolation {
	// Check if we're in a method with a receiver
	if ctx.currentReceiver == nil {
		return nil
	}

	// Check if the assignment is to the receiver: *r = value
	ident, ok := star.X.(*ast.Ident)
	if !ok {
		return nil
	}

	// Check if the identifier is the receiver
	if ident.Name != ctx.currentReceiver.name {
		return nil
	}

	// Check if the receiver type is immutable
	if !ctx.immutableTypes.Contains(ctx.currentReceiver.pkgPath, ctx.currentReceiver.typeName) {
		return nil
	}

	// Allow reassignment in constructors
	if ctx.constructors.Match(ctx.currentReceiver.pkgPath, *ctx.currentFunction, ctx.currentReceiver.typeName) {
		return nil
	}

	return &ImmutableViolation{
		TypeName: ctx.currentReceiver.typeName,
		Pos:      star.Pos(),
		Reason:   "cannot reassign immutable receiver (outside constructor)",
		Node:     stmt,
	}
}
