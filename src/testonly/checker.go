package testonly

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/config"
	"goagreement/src/indexing"
	"goagreement/src/util"
)

// CheckTestOnly checks that @testonly annotated items are only used in test files
func CheckTestOnly(pass *analysis.Pass, packageAnnotations annotations.PackageAnnotations) []TestOnlyViolation {
	var violations []TestOnlyViolation

	// Build indices for @testonly items (including imported packages)
	testOnlyTypes := indexing.BuildTestOnlyTypesIndex(pass, packageAnnotations)
	testOnlyFuncs := indexing.BuildTestOnlyFuncsIndex(pass, packageAnnotations)
	testOnlyMethods := indexing.BuildTestOnlyMethodsIndex(pass, packageAnnotations)

	// If no @testonly items at all (local + imported), nothing to check
	if testOnlyTypes.Len() == 0 && testOnlyFuncs.Len() == 0 && testOnlyMethods.Len() == 0 {
		return violations
	}

	currentPkgPath := pass.Pkg.Path()

	// Check all files (but skip test files as they can use @testonly items)
	filesToCheck := config.Global.FilterFiles(pass)

	context := testOnlyContext{
		pass:            pass,
		testOnlyFuncs:   &testOnlyFuncs,
		testOnlyMethods: &testOnlyMethods,
		currentPkgPath:  &currentPkgPath,
		testOnlyTypes:   &testOnlyTypes,
	}

	for _, file := range filesToCheck {
		fileName := pass.Fset.Position(file.Pos()).Filename
		context.fileName = &fileName

		// Check if this is a test file
		if isTestFile(fileName) {
			continue // Test files can use @testonly items
		}

		// Track reported type violations per file to avoid spam
		reportedTypes := make(map[string]bool)

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.FuncDecl:
				// Check if this function is @testonly - if so, skip checking its body
				if isInTestOnlyContext(&context, node) {
					return false // Don't inspect the body of @testonly functions
				}
				return true

			case *ast.CallExpr:
				// Check function and method calls
				if v := findFunctionCallViolation(&context, node); v != nil {
					violations = append(violations, *v)
				}

			case *ast.CompositeLit:
				// Check type instantiation: TestHelper{...}
				if v := findTypeLiteralViolation(&context, node); v != nil {
					if !reportedTypes[v.TestOnlyObj] {
						violations = append(violations, *v)
						reportedTypes[v.TestOnlyObj] = true
					}
				}

			case *ast.ValueSpec:
				// Check variable declarations: var x TestHelper
				if v := findTypeUsageViolation(&context, node.Type, node.Pos()); v != nil {
					if !reportedTypes[v.TestOnlyObj] {
						violations = append(violations, *v)
						reportedTypes[v.TestOnlyObj] = true
					}
				}

			case *ast.Field:
				// Check struct fields and function parameters
				if v := findTypeUsageViolation(&context, node.Type, node.Pos()); v != nil {
					if !reportedTypes[v.TestOnlyObj] {
						violations = append(violations, *v)
						reportedTypes[v.TestOnlyObj] = true
					}
				}
			}
			return true
		})
	}

	return violations
}

type testOnlyContext struct {
	pass            *analysis.Pass
	testOnlyFuncs   *util.TypeFuncRegistry
	testOnlyMethods *util.TypeFuncRegistry
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
		// Direct function call: CreateMockData()
		funcName := fun.Name
		if ctx.testOnlyFuncs.Match(*ctx.currentPkgPath, funcName, funcName) {
			return &TestOnlyViolation{
				Pos:         call.Pos(),
				TestOnlyObj: funcName,
				Kind:        annotations.TestOnlyOnFunc,
				UsedInFile:  *ctx.fileName,
				Reason:      fmt.Sprintf("function %s is marked @testonly and can only be called in test files", funcName),
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
				}
			}
		}
	}
	return nil
}

// findTypeLiteralViolation checks composite literals for @testonly types
// Returns violation or nil
func findTypeLiteralViolation(
	ctx *testOnlyContext,
	node *ast.CompositeLit,
) *TestOnlyViolation {
	typeInfo := util.ExtractTypeInfo(ctx.pass.TypesInfo.TypeOf(node))
	if typeInfo == nil {
		return nil
	}

	if ctx.testOnlyTypes.Contains(typeInfo.PkgPath, typeInfo.TypeName) {
		return &TestOnlyViolation{
			Pos:         node.Pos(),
			TestOnlyObj: typeInfo.TypeName,
			Kind:        annotations.TestOnlyOnType,
			UsedInFile:  *ctx.fileName,
			Reason:      fmt.Sprintf("type %s is marked @testonly and can only be used in test files", typeInfo.TypeName),
		}
	}
	return nil
}

// findTypeUsageViolation checks if a type expression uses @testonly type
// Returns violation or nil
func findTypeUsageViolation(
	ctx *testOnlyContext,
	typeExpr ast.Expr,
	pos token.Pos,
) *TestOnlyViolation {
	if typeExpr == nil {
		return nil
	}

	typeInfo := util.ExtractTypeInfo(ctx.pass.TypesInfo.TypeOf(typeExpr))
	if typeInfo == nil {
		return nil
	}

	if ctx.testOnlyTypes.Contains(typeInfo.PkgPath, typeInfo.TypeName) {
		return &TestOnlyViolation{
			Pos:         pos,
			TestOnlyObj: typeInfo.TypeName,
			Kind:        annotations.TestOnlyOnType,
			UsedInFile:  *ctx.fileName,
			Reason:      fmt.Sprintf("type %s is marked @testonly and can only be used in test files", typeInfo.TypeName),
		}
	}
	return nil
}

// isTestFile checks if a file is a test file (ends with _test.go)
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}
