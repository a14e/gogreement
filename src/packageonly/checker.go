package packageonly

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/indexing"
	"github.com/a14e/gogreement/src/util"
)

// CheckPackageOnly checks that @packageonly annotated items are only used in allowed packages
func CheckPackageOnly(
	cfg *config.Config,
	pass *analysis.Pass,
	packageAnnotations *annotations.PackageAnnotations,
	ignoreSet *util.IgnoreSet,
) []PackageOnlyViolation {
	var violations []PackageOnlyViolation

	// Build index for @packageonly items (including imported packages)
	packageOnlyIndex := indexing.BuildPackageOnlyIndex[*annotations.PackageOnlyCheckerFact](pass, packageAnnotations)

	// If no @packageonly items at all (local + imported), nothing to check
	if packageOnlyIndex.Empty() {
		return violations
	}

	// Check all files
	filesToCheck := cfg.FilterFiles(pass)

	context := packageOnlyContext{
		pass:             pass,
		packageOnlyIndex: packageOnlyIndex,
		currentPkgPath:   pass.Pkg.Path(),
		currentPkgName:   pass.Pkg.Name(),
		ignoreSet:        ignoreSet,
	}

	for file := range filesToCheck {
		// Track reported type violations per file to avoid spam
		// NOTE: We check ignoreSet BEFORE adding to reportedTypes to ensure that
		// ignored violations don't prevent subsequent non-ignored violations of the
		// same type from being detected. See case statements below for implementation.
		reportedTypes := make(map[string]bool)
		context.reportedTypes = &reportedTypes

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.SelectorExpr:
				// Check selector expressions like "pkg.Type" or "pkg.Function"
				if v := findSelectorExprViolation(&context, node); v != nil {
					violations = append(violations, *v)
				}

			case *ast.Ident:
				// Check identifier usage for local package objects
				if v := findIdentViolation(&context, node); v != nil {
					violations = append(violations, *v)
				}
			}
			return true
		})
	}

	return violations
}

// packageOnlyContext holds the context for checking package-only violations
type packageOnlyContext struct {
	pass             *analysis.Pass
	packageOnlyIndex *util.AttachmentsMap
	currentPkgPath   string
	currentPkgName   string
	ignoreSet        *util.IgnoreSet
	reportedTypes    *map[string]bool
}

// findSelectorExprViolation checks selector expressions like "pkg.Type" or "pkg.Function"
// Returns violation or nil
func findSelectorExprViolation(
	ctx *packageOnlyContext,
	expr *ast.SelectorExpr,
) *PackageOnlyViolation {
	// Get the type information
	obj := ctx.pass.TypesInfo.ObjectOf(expr.Sel)
	if obj == nil {
		return nil
	}

	// Get package information
	pkg := obj.Pkg()
	if pkg == nil {
		return nil
	}

	pkgPath := pkg.Path()
	if pkgPath == ctx.currentPkgPath {
		return nil // Usage within the same package is always allowed
	}

	// Check different types of objects
	switch obj := obj.(type) {
	case *types.TypeName:
		return findTypeViolation(ctx, pkgPath, obj.Name(), expr.Pos())

	case *types.Func:
		if obj.Type() != nil && obj.Type().(*types.Signature).Recv() != nil {
			// Method
			recvType := util.ExtractTypeName(obj.Type().(*types.Signature).Recv().Type())
			return findMethodViolation(ctx, pkgPath, recvType, obj.Name(), expr.Pos())
		} else {
			// Function
			return findFunctionViolation(ctx, pkgPath, obj.Name(), expr.Pos())
		}
	}

	return nil
}

// findIdentViolation checks identifier usage for local package objects
// Returns violation or nil
func findIdentViolation(
	ctx *packageOnlyContext,
	ident *ast.Ident,
) *PackageOnlyViolation {
	obj := ctx.pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return nil
	}

	// Only check local package objects (imports are handled by selector expressions)
	if obj.Pkg() == nil || obj.Pkg().Path() != ctx.currentPkgPath {
		return nil
	}

	switch obj := obj.(type) {
	case *types.TypeName:
		return findTypeViolation(ctx, ctx.currentPkgPath, obj.Name(), ident.Pos())

	case *types.Func:
		if obj.Type() != nil && obj.Type().(*types.Signature).Recv() != nil {
			// Method
			recvType := util.ExtractTypeName(obj.Type().(*types.Signature).Recv().Type())
			return findMethodViolation(ctx, ctx.currentPkgPath, recvType, obj.Name(), ident.Pos())
		} else {
			// Function
			return findFunctionViolation(ctx, ctx.currentPkgPath, obj.Name(), ident.Pos())
		}
	}

	return nil
}

// findTypeViolation checks if a type usage violates @packageonly restrictions
// Returns violation or nil
func findTypeViolation(
	ctx *packageOnlyContext,
	pkgPath string,
	typeName string,
	pos token.Pos,
) *PackageOnlyViolation {
	if !ctx.packageOnlyIndex.HasAnyTypeAttachments(pkgPath, typeName) {
		return nil
	}

	// If not same package, check if current package is allowed
	// Check both full path and package name
	isAllowed := ctx.packageOnlyIndex.HasPkgTypeAttachment(pkgPath, typeName, ctx.currentPkgPath)

	// Also check by extracting package name from path
	if !isAllowed {
		isAllowed = ctx.packageOnlyIndex.HasPkgTypeAttachment(pkgPath, typeName, ctx.currentPkgName)
	}

	if pkgPath != ctx.currentPkgPath && !isAllowed {
		// Check if this violation should be ignored before adding to reportedTypes
		key := pkgPath + "." + typeName
		if !ctx.ignoreSet.Contains(codes.PackageOnlyTypeUsage, pos) {
			// Deduplicate only type violations
			if (*ctx.reportedTypes)[key] {
				return nil
			}
			(*ctx.reportedTypes)[key] = true

			// Get all allowed packages for error message
			allowedPackages := ctx.packageOnlyIndex.GetAttachmentsForType(pkgPath, typeName)
			return &PackageOnlyViolation{
				ItemName:        typeName,
				ItemPkgPath:     pkgPath,
				CurrentPkgPath:  ctx.currentPkgPath,
				AllowedPackages: allowedPackages,
				Pos:             pos,
				Code:            codes.PackageOnlyTypeUsage,
			}
		}
	}

	return nil
}

// findFunctionViolation checks if a function usage violates @packageonly restrictions
// Returns violation or nil
func findFunctionViolation(
	ctx *packageOnlyContext,
	pkgPath string,
	funcName string,
	pos token.Pos,
) *PackageOnlyViolation {
	if !ctx.packageOnlyIndex.HasAnyFunctionAttachments(pkgPath, funcName) {
		return nil
	}

	// If not same package, check if current package is allowed
	// Check both full path and package name
	isAllowed := ctx.packageOnlyIndex.HasPkgFunctionAttachment(pkgPath, funcName, ctx.currentPkgPath)

	// Also check by extracting package name from path
	if !isAllowed {
		isAllowed = ctx.packageOnlyIndex.HasPkgFunctionAttachment(pkgPath, funcName, ctx.currentPkgName)
	}

	if pkgPath != ctx.currentPkgPath && !isAllowed {
		// Check if this violation should be ignored (no deduplication for functions)
		if !ctx.ignoreSet.Contains(codes.PackageOnlyFunctionCall, pos) {
			// Get all allowed packages for error message
			allowedPackages := ctx.packageOnlyIndex.GetAttachmentsForFunction(pkgPath, funcName)
			return &PackageOnlyViolation{
				ItemName:        funcName,
				ItemPkgPath:     pkgPath,
				CurrentPkgPath:  ctx.currentPkgPath,
				AllowedPackages: allowedPackages,
				Pos:             pos,
				Code:            codes.PackageOnlyFunctionCall,
			}
		}
	}

	return nil
}

// findMethodViolation checks if a method usage violates @packageonly restrictions
// Returns violation or nil
func findMethodViolation(
	ctx *packageOnlyContext,
	pkgPath string,
	typeName string,
	methodName string,
	pos token.Pos,
) *PackageOnlyViolation {
	if !ctx.packageOnlyIndex.HasAnyMethodAttachments(pkgPath, typeName, methodName) {
		return nil
	}

	// If not same package, check if current package is allowed
	// Check both full path and package name
	isAllowed := ctx.packageOnlyIndex.HasPkgTypeMethodAttachment(pkgPath, typeName, methodName, ctx.currentPkgPath)

	// Also check by extracting package name from path
	if !isAllowed {
		isAllowed = ctx.packageOnlyIndex.HasPkgTypeMethodAttachment(pkgPath, typeName, methodName, ctx.currentPkgName)
	}

	if pkgPath != ctx.currentPkgPath && !isAllowed {
		// Check if this violation should be ignored (no deduplication for methods)
		if !ctx.ignoreSet.Contains(codes.PackageOnlyMethodCall, pos) {
			// Get all allowed packages for error message
			allowedPackages := ctx.packageOnlyIndex.GetAttachmentsForMethod(pkgPath, typeName, methodName)
			return &PackageOnlyViolation{
				ItemName:        methodName,
				ItemPkgPath:     pkgPath,
				CurrentPkgPath:  ctx.currentPkgPath,
				AllowedPackages: allowedPackages,
				ReceiverType:    typeName,
				Pos:             pos,
				Code:            codes.PackageOnlyMethodCall,
			}
		}
	}

	return nil
}
