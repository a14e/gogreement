package constructor

import (
	"fmt"
	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/indexing"
	"github.com/a14e/gogreement/src/util"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func CheckConstructor(
	config *config.Config,
	pass *analysis.Pass,
	packageAnnotations *annotations.PackageAnnotations,
) []ConstructorViolation {
	var violations []ConstructorViolation

	constructors := indexing.BuildConstructorIndex[*annotations.ConstructorCheckerFact](pass, packageAnnotations)
	if constructors.Empty() {
		return violations
	}

	// Filter files based on configuration (skip test files by default)
	filesToCheck := config.FilterFiles(pass)

	for file := range filesToCheck {
		for _, decl := range file.Decls {
			// Determine the enclosing function per top-level declaration so the
			// value never leaks across siblings. Only a receiverless (free)
			// function can be a constructor; methods and package-level
			// declarations are evaluated with an empty function name so they are
			// never wrongly exempted.
			currentFunction := ""
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
				currentFunction = fn.Name.Name
			}

			ast.Inspect(decl, func(n ast.Node) bool {
				switch node := n.(type) {
				case *ast.CompositeLit:
					v := checkCompositeLiteral(pass, node, constructors, currentFunction)
					if v != nil {
						violations = append(violations, *v)
					}
					return true

				case *ast.CallExpr:
					if v := checkNewCall(pass, node, constructors, currentFunction); v != nil {
						violations = append(violations, *v)
					} else if v := checkConversionCall(pass, node, constructors, currentFunction); v != nil {
						violations = append(violations, *v)
					}
					return true

				case *ast.GenDecl:
					if node.Tok == token.VAR {
						vs := checkVarDeclaration(pass, node, constructors, currentFunction)
						violations = append(violations, vs...)
					}
					return true
				}
				return true
			})
		}
	}

	return violations
}

func checkCompositeLiteral(
	pass *analysis.Pass,
	lit *ast.CompositeLit,
	constructors util.TypeAssociationRegistry,
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

	// Constructors live in the type's own package; only exempt when the type is
	// declared in the package being analyzed and the enclosing function is one
	// of its declared constructors.
	if pass.Pkg.Path() == pkgPath && constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Get list of allowed constructors for error message
	constructorList := constructors.GetAssociated(pkgPath, typeName)
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
	constructors util.TypeAssociationRegistry,
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

	// Do not strip a pointer here: new(*T) allocates a **T pointing at a nil
	// *T and never instantiates a T, so it must not be flagged.
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

	// Constructors live in the type's own package; only exempt when the type is
	// declared in the package being analyzed and the enclosing function is one
	// of its declared constructors.
	if pass.Pkg.Path() == pkgPath && constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Get list of allowed constructors for error message
	constructorList := constructors.GetAssociated(pkgPath, typeName)
	reason := fmt.Sprintf("type instantiation with new() must be in constructor (allowed: %v)", constructorList)

	return &ConstructorViolation{
		TypeName: typeName,
		Code:     codes.ConstructorNewCall,
		Pos:      call.Pos(),
		Reason:   reason,
		Node:     call,
	}
}

// checkConversionCall reports a violation when a value is built via a type
// conversion (e.g. Email(input)) of a @constructor type outside its allowed
// constructors. A call expression is a conversion when its function position
// denotes a type rather than a value.
func checkConversionCall(
	pass *analysis.Pass,
	call *ast.CallExpr,
	constructors util.TypeAssociationRegistry,
	currentFunction string,
) *ConstructorViolation {
	if len(call.Args) != 1 {
		return nil
	}

	// Distinguish a conversion T(x) from a regular call f(x): in a conversion
	// the function expression is a type.
	tv, ok := pass.TypesInfo.Types[call.Fun]
	if !ok || !tv.IsType() {
		return nil
	}

	// A conversion to *T does not instantiate a T value, so only direct
	// conversions to the named type are constructor-controlled.
	named, ok := tv.Type.(*types.Named)
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

	// Constructors live in the type's own package; only exempt when the type is
	// declared in the package being analyzed and the enclosing function is one
	// of its declared constructors.
	if pass.Pkg.Path() == pkgPath && constructors.Match(pkgPath, currentFunction, typeName) {
		return nil
	}

	// Get list of allowed constructors for error message
	constructorList := constructors.GetAssociated(pkgPath, typeName)
	reason := fmt.Sprintf("type conversion must be in constructor (allowed: %v)", constructorList)

	return &ConstructorViolation{
		TypeName: typeName,
		Code:     codes.ConstructorConversion,
		Pos:      call.Pos(),
		Reason:   reason,
		Node:     call,
	}
}

func checkVarDeclaration(
	pass *analysis.Pass,
	decl *ast.GenDecl,
	constructors util.TypeAssociationRegistry,
	currentFunction string,
) []ConstructorViolation {
	var violations []ConstructorViolation

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		// Skip if there's an assignment (this is not a zero-initialized declaration)
		if len(valueSpec.Values) > 0 {
			continue
		}

		// Check each variable name in the declaration
		for _, name := range valueSpec.Names {
			if name.Name == "_" {
				continue // Skip blank identifier
			}

			// Get the type of the variable
			t := pass.TypesInfo.TypeOf(name)
			if t == nil {
				continue
			}

			// Skip pointer types - var p *Struct just creates a nil pointer, not an instance
			if _, ok := t.(*types.Pointer); ok {
				continue
			}

			named, ok := t.(*types.Named)
			if !ok {
				continue
			}

			typeName := named.Obj().Name()
			pkg := named.Obj().Pkg()
			if pkg == nil {
				continue
			}

			pkgPath := pkg.Path()

			// Check if this type has constructor annotations
			if !constructors.HasType(pkgPath, typeName) {
				continue
			}

			// Constructors live in the type's own package; only exempt when the
			// type is declared in the package being analyzed and the enclosing
			// function is one of its declared constructors.
			if pass.Pkg.Path() == pkgPath && constructors.Match(pkgPath, currentFunction, typeName) {
				continue
			}

			// Get list of allowed constructors for error message
			constructorList := constructors.GetAssociated(pkgPath, typeName)
			reason := fmt.Sprintf("zero-initialized variable declaration must be in constructor (allowed: %v)", constructorList)

			violations = append(violations, ConstructorViolation{
				TypeName: typeName,
				Code:     codes.ConstructorVarDeclaration,
				Pos:      name.Pos(),
				Reason:   reason,
				Node:     name,
			})
		}
	}

	return violations
}
