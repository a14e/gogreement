package constructor

import (
	"fmt"
	"go/ast"
	"go/types"
	"goagreement/src/annotations"
	"goagreement/src/codes"
	"goagreement/src/config"
	"goagreement/src/indexing"
	"goagreement/src/util"

	"golang.org/x/tools/go/analysis"
)

func CheckConstructor(pass *analysis.Pass, packageAnnotations *annotations.PackageAnnotations) []ConstructorViolation {
	var violations []ConstructorViolation

	constructors := indexing.BuildConstructorIndex[*annotations.ConstructorCheckerFact](pass, packageAnnotations)
	if constructors.Empty() {
		return violations
	}

	// Filter files based on configuration (skip test files by default)
	filesToCheck := config.Global.FilterFiles(pass)

	for _, file := range filesToCheck {
		currentFunction := ""

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				currentFunction = node.Name.Name
				return true

			case *ast.CompositeLit:
				v := checkCompositeLiteral(pass, node, constructors, currentFunction)
				if v != nil {
					violations = append(violations, *v)
				}
				return true

			case *ast.CallExpr:
				v := checkNewCall(pass, node, constructors, currentFunction)
				if v != nil {
					violations = append(violations, *v)
				}
				return true
			}
			return true
		})
	}

	return violations
}

func checkCompositeLiteral(
	pass *analysis.Pass,
	lit *ast.CompositeLit,
	constructors util.TypeFuncRegistry,
	currentFunction string,
) *ConstructorViolation {
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return nil
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil
	}

	pkgPath := pkg.Path()

	// Check if this type has constructor annotations
	if !constructors.HasType(pkgPath, typeName) {
		return nil
	}

	// Check if we're in one of the allowed constructors
	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Get list of allowed constructors for error message
	constructorList := constructors.GetFuncs(pkgPath, typeName)
	reason := fmt.Sprintf("type instantiation must be in constructor (allowed: %v)", constructorList)

	return &ConstructorViolation{
		TypeName: typeName,
		Code:     codes.ConstructorCompositeLiteral,
		Pos:      lit.Pos(),
		Reason:   reason,
		Node:     lit,
	}
}

func checkNewCall(
	pass *analysis.Pass,
	call *ast.CallExpr,
	constructors util.TypeFuncRegistry,
	currentFunction string,
) *ConstructorViolation {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "new" {
		return nil
	}

	if len(call.Args) != 1 {
		return nil
	}

	t := pass.TypesInfo.TypeOf(call.Args[0])
	if t == nil {
		return nil
	}

	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil
	}

	pkgPath := pkg.Path()

	// Check if this type has constructor annotations
	if !constructors.HasType(pkgPath, typeName) {
		return nil
	}

	// Check if we're in one of the allowed constructors
	if constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Get list of allowed constructors for error message
	constructorList := constructors.GetFuncs(pkgPath, typeName)
	reason := fmt.Sprintf("type instantiation with new() must be in constructor (allowed: %v)", constructorList)

	return &ConstructorViolation{
		TypeName: typeName,
		Code:     codes.ConstructorNewCall,
		Pos:      call.Pos(),
		Reason:   reason,
		Node:     call,
	}
}
