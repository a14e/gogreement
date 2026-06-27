package immutable

import (
	"fmt"
	"github.com/a14e/gogreement/src/indexing"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/util"
)

func CheckImmutable(
	cfg *config.Config,
	pass *analysis.Pass,
	packageAnnotations *annotations.PackageAnnotations,
) []ImmutableViolation {
	var violations []ImmutableViolation

	// Build indices for efficient lookup during AST traversal
	immutableTypes := indexing.BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass, packageAnnotations)
	if immutableTypes.Empty() {
		return violations // No immutable types to check
	}

	constructors := indexing.BuildConstructorIndex[*annotations.ImmutableCheckerFact](pass, packageAnnotations)
	mutableFields := indexing.BuildMutableFieldsIndex[*annotations.ImmutableCheckerFact](pass, packageAnnotations)

	// Filter files based on configuration (skip test files by default)
	filesToCheck := cfg.FilterFiles(pass)

	ctx := &checkerContext{
		pass:           pass,
		immutableTypes: immutableTypes,
		constructors:   constructors,
		mutableFields:  mutableFields,
	}

	// inspectNode handles assignment / inc-dec nodes. It reads the enclosing
	// function from ctx, which is set per top-level declaration below.
	// Compound assignments (+=, -=, ...) are processed separately from plain
	// assignments so the same node is never reported twice.
	inspectNode := func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Tok != token.ASSIGN {
				violations = append(violations, checkCompoundAssignment(ctx, node)...)
				return true
			}
			violations = append(violations, checkAssignment(ctx, node)...)
			return true

		case *ast.IncDecStmt:
			violations = append(violations, checkIncDec(ctx, node)...)
			return true
		}
		return true
	}

	for file := range filesToCheck {
		for _, decl := range file.Decls {
			// Establish the enclosing-function context per top-level declaration
			// so it never leaks across siblings and is never unset (a mutation
			// inside a package-level func literal must not deref a nil function).
			// Mutations inside a named function (including its nested func
			// literals) are evaluated against that function; package-level
			// declarations are not inside any constructor.
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				ctx.currentFunction = funcDecl.Name.Name
				ctx.currentReceiver = extractReceiverInfo(ctx.pass, funcDecl)
			} else {
				ctx.currentFunction = ""
				ctx.currentReceiver = nil
			}
			ast.Inspect(decl, inspectNode)
		}
	}

	return violations
}

type checkerContext struct {
	pass            *analysis.Pass
	immutableTypes  util.TypesMap
	constructors    util.TypeAssociationRegistry
	mutableFields   util.TypeAssociationRegistry
	currentFunction string
	currentReceiver *receiverInfo
}

// receiverInfo contains information about a method's receiver
// @immutable
type receiverInfo struct {
	typeName string
	pkgPath  string
	// obj is the receiver variable's object, used to confirm an identifier
	// actually refers to the receiver (not a shadowing local of the same name).
	obj types.Object
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

	recvIdent := recvField.Names[0]
	recvType := pass.TypesInfo.TypeOf(recvField.Type)
	if recvType == nil {
		return nil
	}

	typeInfo := util.ExtractTypeInfo(recvType)
	if typeInfo == nil {
		return nil
	}

	return &receiverInfo{
		typeName: typeInfo.TypeName,
		pkgPath:  typeInfo.PkgPath,
		obj:      pass.TypesInfo.Defs[recvIdent],
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
	typeName, pkgPath, ok := immutableReceiverOfField(ctx, selector)
	if !ok {
		return nil
	}

	if ctx.constructors.Match(pkgPath, ctx.currentFunction, typeName) {
		return nil
	}

	// Check if the field is marked as @mutable
	if ctx.mutableFields.Match(pkgPath, selector.Sel.Name, typeName) {
		return nil
	}

	return &ImmutableViolation{
		TypeName: typeName,
		Code:     codes.ImmutableFieldAssignment,
		Pos:      selector.Pos(),
		Reason:   fmt.Sprintf("cannot assign to field %q of immutable type", selector.Sel.Name),
		Node:     stmt,
	}
}

// immutableReceiverOfField resolves the immutable type whose field is written by
// selector. It first checks the immediately-selected receiver (t.field), then,
// if that type is not immutable, walks an explicit embedded-field access path
// (o.Inner.field) so a write through the embedded path is treated the same as
// the promoted form (o.field) that field promotion would expose.
func immutableReceiverOfField(ctx *checkerContext, selector *ast.SelectorExpr) (string, string, bool) {
	receiverType := ctx.pass.TypesInfo.TypeOf(selector.X)
	if receiverType == nil {
		return "", "", false
	}

	if ptr, ok := receiverType.(*types.Pointer); ok {
		receiverType = ptr.Elem()
	}

	if named, ok := receiverType.(*types.Named); ok && named.Obj().Pkg() != nil {
		typeName := named.Obj().Name()
		pkgPath := named.Obj().Pkg().Path()
		if ctx.immutableTypes.Contains(pkgPath, typeName) {
			return typeName, pkgPath, true
		}
	}

	return immutableViaEmbedded(ctx, selector.X)
}

// immutableViaEmbedded reports the immutable type reachable from expr through one
// or more embedded-field hops (o.Inner, o.Inner.Deeper, ...). Only embedded
// (anonymous) fields are followed, matching Go's field promotion; a named field
// hop stops the walk so shallow immutability is preserved for named fields.
func immutableViaEmbedded(ctx *checkerContext, expr ast.Expr) (string, string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}

	selection := ctx.pass.TypesInfo.Selections[sel]
	if selection == nil || selection.Kind() != types.FieldVal {
		return "", "", false
	}
	field, ok := selection.Obj().(*types.Var)
	if !ok || !field.Embedded() {
		return "", "", false
	}

	baseType := ctx.pass.TypesInfo.TypeOf(sel.X)
	if baseType != nil {
		if ptr, ok := baseType.(*types.Pointer); ok {
			baseType = ptr.Elem()
		}
		if named, ok := baseType.(*types.Named); ok && named.Obj().Pkg() != nil {
			typeName := named.Obj().Name()
			pkgPath := named.Obj().Pkg().Path()
			if ctx.immutableTypes.Contains(pkgPath, typeName) {
				return typeName, pkgPath, true
			}
		}
	}

	// The base may itself be an embedded access on an immutable type.
	return immutableViaEmbedded(ctx, sel.X)
}

func checkIndexAssignment(
	ctx *checkerContext,
	stmt *ast.AssignStmt,
	index *ast.IndexExpr,
) *ImmutableViolation {
	return checkImmutableIndex(ctx, index, stmt)
}

// checkImmutableIndex reports IMM04 when an element of an immutable type's field
// is modified through an index expression, e.g. x.items[0] = v, x.items[0] += v,
// or x.items[0]++. Shared by the plain-assignment, compound-assignment, and
// inc/dec paths so the same gap is closed for all of them.
func checkImmutableIndex(
	ctx *checkerContext,
	index *ast.IndexExpr,
	node ast.Node,
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

	if ctx.constructors.Match(pkgPath, ctx.currentFunction, typeName) {
		return nil
	}

	// Check if the field is marked as @mutable
	if ctx.mutableFields.Match(pkgPath, selector.Sel.Name, typeName) {
		return nil
	}

	return &ImmutableViolation{
		TypeName: typeName,
		Code:     codes.ImmutableIndexAssignment,
		Pos:      index.Pos(),
		Reason:   fmt.Sprintf("cannot modify element of field %q of immutable type", selector.Sel.Name),
		Node:     node,
	}
}

func checkIncDec(
	ctx *checkerContext,
	node *ast.IncDecStmt,
) []ImmutableViolation {
	var violations []ImmutableViolation

	// Unwrap parentheses so forms like (*c)-- and (x.field)++ are handled
	// the same as their unparenthesized counterparts.
	target := ast.Unparen(node.X)

	// Check for field increment/decrement: x.field++
	if selector, ok := target.(*ast.SelectorExpr); ok {
		violation := checkFieldIncDec(ctx, node, selector)
		if violation != nil {
			violations = append(violations, *violation)
		}
		return violations
	}

	// Check for indexed element increment/decrement: x.items[0]++
	if index, ok := target.(*ast.IndexExpr); ok {
		if violation := checkImmutableIndex(ctx, index, node); violation != nil {
			violations = append(violations, *violation)
		}
		return violations
	}

	// Check for receiver increment/decrement: *receiver++
	if star, ok := target.(*ast.StarExpr); ok {
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

	if ctx.constructors.Match(pkgPath, ctx.currentFunction, typeName) {
		return nil
	}

	// Check if the field is marked as @mutable
	if ctx.mutableFields.Match(pkgPath, selector.Sel.Name, typeName) {
		return nil
	}

	op := "++"
	if node.Tok == token.DEC {
		op = "--"
	}

	return &ImmutableViolation{
		TypeName: typeName,
		Code:     codes.ImmutableFieldIncDec,
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

	// Confirm the identifier actually refers to the receiver and not a
	// shadowing local variable of the same name.
	if ctx.currentReceiver.obj == nil || ctx.pass.TypesInfo.ObjectOf(ident) != ctx.currentReceiver.obj {
		return nil
	}

	// Check if the receiver type is immutable
	if !ctx.immutableTypes.Contains(ctx.currentReceiver.pkgPath, ctx.currentReceiver.typeName) {
		return nil
	}

	// Allow in constructors
	if ctx.constructors.Match(ctx.currentReceiver.pkgPath, ctx.currentFunction, ctx.currentReceiver.typeName) {
		return nil
	}

	op := "++"
	if node.Tok == token.DEC {
		op = "--"
	}

	return &ImmutableViolation{
		TypeName: ctx.currentReceiver.typeName,
		Code:     codes.ImmutableFieldIncDec,
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
	// Unwrap parentheses so (x.field) += v and (x.items[0]) += v are handled.
	expr = ast.Unparen(expr)

	// Compound assignment to an indexed element: x.items[0] += v
	if index, ok := expr.(*ast.IndexExpr); ok {
		return checkImmutableIndex(ctx, index, stmt)
	}

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

	if ctx.constructors.Match(pkgPath, ctx.currentFunction, typeName) {
		return nil
	}

	// Check if the field is marked as @mutable
	if ctx.mutableFields.Match(pkgPath, selector.Sel.Name, typeName) {
		return nil
	}

	op := tok.String()
	return &ImmutableViolation{
		TypeName: typeName,
		Code:     codes.ImmutableFieldCompoundAssign,
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

	// Confirm the identifier actually refers to the receiver and not a
	// shadowing local variable of the same name.
	if ctx.currentReceiver.obj == nil || ctx.pass.TypesInfo.ObjectOf(ident) != ctx.currentReceiver.obj {
		return nil
	}

	// Check if the receiver type is immutable
	if !ctx.immutableTypes.Contains(ctx.currentReceiver.pkgPath, ctx.currentReceiver.typeName) {
		return nil
	}

	// Allow reassignment in constructors
	if ctx.constructors.Match(ctx.currentReceiver.pkgPath, ctx.currentFunction, ctx.currentReceiver.typeName) {
		return nil
	}

	return &ImmutableViolation{
		TypeName: ctx.currentReceiver.typeName,
		Code:     codes.ImmutableFieldAssignment,
		Pos:      star.Pos(),
		Reason:   "cannot reassign immutable receiver (outside constructor)",
		Node:     stmt,
	}
}
