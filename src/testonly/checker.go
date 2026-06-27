package testonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/indexing"
	"github.com/a14e/gogreement/src/util"
)

// CheckTestOnly checks that @testonly annotated items are only used in test files
func CheckTestOnly(
	cfg *config.Config,
	pass *analysis.Pass,
	packageAnnotations *annotations.PackageAnnotations,
	ignoreSet *util.IgnoreSet,
) []TestOnlyViolation {
	var violations []TestOnlyViolation

	// Build indices for @testonly items (including imported packages)
	testOnlyTypes := indexing.BuildTestOnlyTypesIndex[*annotations.TestOnlyCheckerFact](pass, packageAnnotations)
	testOnlyFuncs := indexing.BuildTestOnlyFuncsIndex[*annotations.TestOnlyCheckerFact](pass, packageAnnotations)
	testOnlyMethods := indexing.BuildTestOnlyMethodsIndex[*annotations.TestOnlyCheckerFact](pass, packageAnnotations)

	// If no @testonly items at all (local + imported), nothing to check
	if testOnlyTypes.Empty() && testOnlyFuncs.Empty() && testOnlyMethods.Empty() {
		return violations
	}

	currentPkgPath := pass.Pkg.Path()

	// Check all files (but skip test files as they can use @testonly items)
	filesToCheck := cfg.FilterFiles(pass)

	context := testOnlyContext{
		pass:            pass,
		testOnlyFuncs:   &testOnlyFuncs,
		testOnlyMethods: &testOnlyMethods,
		currentPkgPath:  &currentPkgPath,
		testOnlyTypes:   &testOnlyTypes,
	}

	for file := range filesToCheck {
		fileName := pass.Fset.Position(file.Pos()).Filename
		context.fileName = &fileName

		// Check if this is a test file
		if isTestFile(fileName) {
			continue // Test files can use @testonly items
		}

		// Track reported type violations per file to avoid spam. The key is the
		// package-qualified type identity so equally named @testonly types from
		// different packages do not collide. ignoreSet is checked BEFORE marking
		// a type reported so an ignored violation does not suppress a later one.
		reportedTypes := make(map[string]bool)

		// Receiver fields of method declarations: declaring a method on a
		// @testonly type is legitimate and must not be reported as type usage.
		receiverFields := make(map[*ast.Field]bool)

		// reportTypeUsage applies the ignore filter and per-file dedup for
		// type-usage (TONL01) violations.
		reportTypeUsage := func(v *TestOnlyViolation) {
			if v == nil || ignoreSet.Contains(v.Code, v.Pos) || reportedTypes[v.TypeKey] {
				return
			}
			violations = append(violations, *v)
			reportedTypes[v.TypeKey] = true
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				// Record receiver fields so a method declared on a @testonly type
				// is not mistaken for escaping type usage.
				if node.Recv != nil {
					for _, f := range node.Recv.List {
						receiverFields[f] = true
					}
				}
				// Check if this function is @testonly - if so, skip checking its body
				if isInTestOnlyContext(&context, node) {
					return false // Don't inspect the body of @testonly functions
				}
				return true

			case *ast.CallExpr:
				// Function and method calls (TONL02/TONL03), reported per occurrence.
				if v := findFunctionCallViolation(&context, node); v != nil {
					if !ignoreSet.Contains(v.Code, v.Pos) {
						violations = append(violations, *v)
					}
				}
				// Type construction via conversion T(x), new(T) or make([]T, ...)
				// is type usage (TONL01) and is deduplicated.
				reportTypeUsage(findTypeConstructionViolation(&context, node))

			case *ast.CompositeLit:
				// Type instantiation: TestHelper{...}, []TestHelper{...}
				reportTypeUsage(findTypeLiteralViolation(&context, node))

			case *ast.ValueSpec:
				// Variable declarations: var x TestHelper
				reportTypeUsage(findTypeUsageViolation(&context, node.Type, node.Pos()))

			case *ast.Field:
				// Struct fields and function parameters (but not method receivers).
				if receiverFields[node] {
					return true
				}
				reportTypeUsage(findTypeUsageViolation(&context, node.Type, node.Pos()))

			case *ast.TypeAssertExpr:
				// Type assertions: x.(MockHelper) / x.(*MockHelper).
				// The x.(type) form used in type switches has a nil Type and is skipped.
				if node.Type != nil {
					reportTypeUsage(findTypeUsageViolation(&context, node.Type, node.Type.Pos()))
				}
			}
			return true
		})
	}

	return violations
}

type testOnlyContext struct {
	pass            *analysis.Pass
	testOnlyFuncs   *util.TypeAssociationRegistry
	testOnlyMethods *util.TypeAssociationRegistry
	testOnlyTypes   *util.TypesMap
	currentPkgPath  *string
	fileName        *string
}

// isInTestOnlyContext checks if we're currently inside a @testonly function or method
func isInTestOnlyContext(
	ctx *testOnlyContext,
	currentFunc *ast.FuncDecl,
) bool {
	if currentFunc == nil {
		return false
	}

	// Check if it's a method
	if currentFunc.Recv != nil && len(currentFunc.Recv.List) > 0 {
		// Extract receiver type
		receiverType := ""
		if len(currentFunc.Recv.List) > 0 {
			receiverType = annotations.ExtractReceiverType(currentFunc.Recv.List[0].Type)
		}
		methodName := currentFunc.Name.Name
		return ctx.testOnlyMethods.Match(*ctx.currentPkgPath, methodName, receiverType)
	}

	// Check if it's a function
	funcName := currentFunc.Name.Name
	return ctx.testOnlyFuncs.Match(*ctx.currentPkgPath, funcName, funcName)
}

// findFunctionCallViolation checks if a function call uses @testonly function or method
// Returns violation or nil
func findFunctionCallViolation(
	ctx *testOnlyContext,
	call *ast.CallExpr,
) *TestOnlyViolation {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		// Direct function call: CreateMockData(). Resolve the identifier to its
		// object so a local variable shadowing a @testonly function name (and a
		// type conversion T(x), whose Fun is a type) is not treated as the call.
		fn, ok := ctx.pass.TypesInfo.Uses[fun].(*types.Func)
		if !ok || fn.Pkg() == nil {
			return nil
		}
		if ctx.testOnlyFuncs.Match(fn.Pkg().Path(), fn.Name(), fn.Name()) {
			return &TestOnlyViolation{
				Pos:         call.Pos(),
				TestOnlyObj: fn.Name(),
				Kind:        annotations.TestOnlyOnFunc,
				UsedInFile:  *ctx.fileName,
				Reason:      fmt.Sprintf("function %s is marked @testonly and can only be called in test files", fn.Name()),
				Code:        codes.TestOnlyFunctionCall,
			}
		}

	case *ast.SelectorExpr:
		// Could be: pkg.Func() or obj.Method()
		funcName := fun.Sel.Name

		// Check if it's a package-qualified function call (pkg.Func)
		if pkgIdent, ok := fun.X.(*ast.Ident); ok {
			if obj := ctx.pass.TypesInfo.Uses[pkgIdent]; obj != nil {
				if pkgName, ok := obj.(*types.PkgName); ok {
					// It's a package function call like testonlyexample.CreateMockData()
					pkgPath := pkgName.Imported().Path()
					if ctx.testOnlyFuncs.Match(pkgPath, funcName, funcName) {
						return &TestOnlyViolation{
							Pos:         call.Pos(),
							TestOnlyObj: funcName,
							Kind:        annotations.TestOnlyOnFunc,
							UsedInFile:  *ctx.fileName,
							Reason:      fmt.Sprintf("function %s is marked @testonly and can only be called in test files", funcName),
							Code:        codes.TestOnlyFunctionCall,
						}
					}
					return nil // Not a testonly func, but also not a method
				}
			}
		}

		// Check if it's a method call (obj.Method)
		typeInfo := util.ExtractTypeInfo(ctx.pass.TypesInfo.TypeOf(fun.X))
		if typeInfo != nil {
			methodName := fun.Sel.Name
			if ctx.testOnlyMethods.Match(typeInfo.PkgPath, methodName, typeInfo.TypeName) {
				return &TestOnlyViolation{
					Pos:         call.Pos(),
					TestOnlyObj: fmt.Sprintf("%s.%s", typeInfo.TypeName, methodName),
					Kind:        annotations.TestOnlyOnMethod,
					UsedInFile:  *ctx.fileName,
					Reason:      fmt.Sprintf("method %s on %s is marked @testonly and can only be called in test files", methodName, typeInfo.TypeName),
					Code:        codes.TestOnlyMethodCall,
				}
			}
		}
	}
	return nil
}

// findTypeLiteralViolation checks composite literals for @testonly types,
// including slice/array/map element types (e.g. []TestHelper{...}).
func findTypeLiteralViolation(
	ctx *testOnlyContext,
	node *ast.CompositeLit,
) *TestOnlyViolation {
	return ctx.typeViolation(ctx.pass.TypesInfo.TypeOf(node), node.Pos())
}

// findTypeUsageViolation checks if a type expression uses a @testonly type,
// unwrapping pointer/slice/array/map/chan layers.
func findTypeUsageViolation(
	ctx *testOnlyContext,
	typeExpr ast.Expr,
	pos token.Pos,
) *TestOnlyViolation {
	if typeExpr == nil {
		return nil
	}
	return ctx.typeViolation(ctx.pass.TypesInfo.TypeOf(typeExpr), pos)
}

// findTypeConstructionViolation detects @testonly type usage inside a call
// expression: a type conversion T(x), or the builtins new(T) / make([]T, ...).
func findTypeConstructionViolation(
	ctx *testOnlyContext,
	call *ast.CallExpr,
) *TestOnlyViolation {
	// Type conversion T(x): the function position denotes a type, not a value.
	if tv, ok := ctx.pass.TypesInfo.Types[call.Fun]; ok && tv.IsType() {
		return ctx.typeViolation(tv.Type, call.Pos())
	}

	// Builtins new(T) and make(T, ...): the first argument is a type expression.
	if ident, ok := call.Fun.(*ast.Ident); ok && (ident.Name == "new" || ident.Name == "make") {
		if len(call.Args) > 0 {
			return ctx.typeViolation(ctx.pass.TypesInfo.TypeOf(call.Args[0]), call.Args[0].Pos())
		}
	}
	return nil
}

// typeViolation builds a TONL01 violation if t references a @testonly type
// (after unwrapping pointer/slice/array/map/chan layers).
func (ctx *testOnlyContext) typeViolation(t types.Type, pos token.Pos) *TestOnlyViolation {
	if t == nil {
		return nil
	}
	info := ctx.firstTestOnlyType(t, make(map[types.Type]bool))
	if info == nil {
		return nil
	}
	return &TestOnlyViolation{
		Pos:         pos,
		TestOnlyObj: info.TypeName,
		TypeKey:     info.PkgPath + "." + info.TypeName,
		Kind:        annotations.TestOnlyOnType,
		UsedInFile:  *ctx.fileName,
		Reason:      fmt.Sprintf("type %s is marked @testonly and can only be used in test files", info.TypeName),
		Code:        codes.TestOnlyTypeUsage,
	}
}

// firstTestOnlyType unwraps pointer/slice/array/map/chan layers and returns the
// info of the first @testonly named type found, or nil. The seen set guards
// against cycles in recursive type definitions.
func (ctx *testOnlyContext) firstTestOnlyType(t types.Type, seen map[types.Type]bool) *util.TypeInfo {
	if t == nil || seen[t] {
		return nil
	}
	seen[t] = true

	switch tt := t.(type) {
	case *types.Pointer:
		return ctx.firstTestOnlyType(tt.Elem(), seen)
	case *types.Slice:
		return ctx.firstTestOnlyType(tt.Elem(), seen)
	case *types.Array:
		return ctx.firstTestOnlyType(tt.Elem(), seen)
	case *types.Chan:
		return ctx.firstTestOnlyType(tt.Elem(), seen)
	case *types.Map:
		if info := ctx.firstTestOnlyType(tt.Key(), seen); info != nil {
			return info
		}
		return ctx.firstTestOnlyType(tt.Elem(), seen)
	case *types.Named:
		if info := util.ExtractTypeInfo(tt); info != nil &&
			ctx.testOnlyTypes.Contains(info.PkgPath, info.TypeName) {
			return info
		}
	}
	return nil
}

// isTestFile checks if a file is a test file (ends with _test.go)
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
