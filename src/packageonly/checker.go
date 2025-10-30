package packageonly

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
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

	// TODO: Check if index is empty and return early

	currentPkgPath := pass.Pkg.Path()

	// Check all files
	filesToCheck := cfg.FilterFiles(pass)

	context := packageOnlyContext{
		pass:             pass,
		packageOnlyIndex: packageOnlyIndex,
		currentPkgPath:   currentPkgPath,
		ignoreSet:        ignoreSet,
		violations:       &violations,
	}

	for file := range filesToCheck {
		context.checkFile(file)
	}

	return violations
}

// packageOnlyContext holds the context for checking package-only violations
type packageOnlyContext struct {
	pass             *analysis.Pass
	packageOnlyIndex *util.AttachmentsMap
	currentPkgPath   string
	ignoreSet        *util.IgnoreSet
	violations       *[]PackageOnlyViolation
}

// checkFile checks a single file for package-only violations
func (c *packageOnlyContext) checkFile(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr:
			c.checkSelectorExpr(node)
		case *ast.Ident:
			c.checkIdent(node)
		}
		return true
	})
}

// checkSelectorExpr checks selector expressions like "pkg.Type" or "pkg.Function"
func (c *packageOnlyContext) checkSelectorExpr(expr *ast.SelectorExpr) {
	// Get the type information
	obj := c.pass.TypesInfo.ObjectOf(expr.Sel)
	if obj == nil {
		return
	}

	// Get package information
	pkg := obj.Pkg()
	if pkg == nil {
		return
	}

	pkgPath := pkg.Path()
	if pkgPath == c.currentPkgPath {
		return // Usage within the same package is always allowed
	}

	// Check different types of objects
	switch obj := obj.(type) {
	case *types.TypeName:
		c.checkTypeUsage(pkgPath, obj.Name(), expr.Pos())
	case *types.Func:
		if obj.Type() != nil && obj.Type().(*types.Signature).Recv() != nil {
			// Method
			recvType := util.ExtractTypeName(obj.Type().(*types.Signature).Recv().Type())
			c.checkMethodUsage(pkgPath, recvType, obj.Name(), expr.Pos())
		} else {
			// Function
			c.checkFunctionUsage(pkgPath, obj.Name(), expr.Pos())
		}
	}
}

// checkIdent checks identifier usage
func (c *packageOnlyContext) checkIdent(ident *ast.Ident) {
	obj := c.pass.TypesInfo.ObjectOf(ident)
	if obj == nil {
		return
	}

	// Only check local package objects (imports are handled by selector expressions)
	if obj.Pkg() != nil && obj.Pkg().Path() == c.currentPkgPath {
		switch obj := obj.(type) {
		case *types.TypeName:
			c.checkTypeUsage(c.currentPkgPath, obj.Name(), ident.Pos())
		case *types.Func:
			if obj.Type() != nil && obj.Type().(*types.Signature).Recv() != nil {
				// Method
				recvType := util.ExtractTypeName(obj.Type().(*types.Signature).Recv().Type())
				c.checkMethodUsage(c.currentPkgPath, recvType, obj.Name(), ident.Pos())
			} else {
				// Function
				c.checkFunctionUsage(c.currentPkgPath, obj.Name(), ident.Pos())
			}
		}
	}
}

// checkTypeUsage checks if a type usage violates @packageonly restrictions
func (c *packageOnlyContext) checkTypeUsage(pkgPath string, typeName string, pos token.Pos) {
	if !c.packageOnlyIndex.HasAnyTypeAttachments(pkgPath, typeName) {
		return
	}

	// If not same package, check if current package is allowed
	// Check both full path and package name
	isAllowed := c.packageOnlyIndex.HasPkgTypeAttachment(pkgPath, typeName, c.currentPkgPath)

	// Also check by extracting package name from path
	if !isAllowed {
		currentPkgName := extractPackageName(c.currentPkgPath)
		isAllowed = c.packageOnlyIndex.HasPkgTypeAttachment(pkgPath, typeName, currentPkgName)
	}

	if pkgPath != c.currentPkgPath && !isAllowed {
		// Get all allowed packages for error message
		allowedPackages := c.packageOnlyIndex.GetAttachmentsForType(pkgPath, typeName)
		violation := PackageOnlyViolation{
			ItemName:        typeName,
			ItemPkgPath:     pkgPath,
			CurrentPkgPath:  c.currentPkgPath,
			AllowedPackages: allowedPackages,
			ViolationType:   "type",
			Pos:             pos,
		}
		c.addViolation(violation)
	}
}

// checkFunctionUsage checks if a function usage violates @packageonly restrictions
func (c *packageOnlyContext) checkFunctionUsage(pkgPath string, funcName string, pos token.Pos) {
	if !c.packageOnlyIndex.HasAnyFunctionAttachments(pkgPath, funcName) {
		return
	}

	// If not same package, check if current package is allowed
	// Check both full path and package name
	isAllowed := c.packageOnlyIndex.HasPkgFunctionAttachment(pkgPath, funcName, c.currentPkgPath)

	// Also check by extracting package name from path
	if !isAllowed {
		currentPkgName := extractPackageName(c.currentPkgPath)
		isAllowed = c.packageOnlyIndex.HasPkgFunctionAttachment(pkgPath, funcName, currentPkgName)
	}

	if pkgPath != c.currentPkgPath && !isAllowed {
		// Get all allowed packages for error message
		allowedPackages := c.packageOnlyIndex.GetAttachmentsForFunction(pkgPath, funcName)
		violation := PackageOnlyViolation{
			ItemName:        funcName,
			ItemPkgPath:     pkgPath,
			CurrentPkgPath:  c.currentPkgPath,
			AllowedPackages: allowedPackages,
			ViolationType:   "function",
			Pos:             pos,
		}
		c.addViolation(violation)
	}
}

// checkMethodUsage checks if a method usage violates @packageonly restrictions
func (c *packageOnlyContext) checkMethodUsage(pkgPath string, typeName string, methodName string, pos token.Pos) {
	if !c.packageOnlyIndex.HasAnyMethodAttachments(pkgPath, typeName, methodName) {
		return
	}

	// If not same package, check if current package is allowed
	// Check both full path and package name
	isAllowed := c.packageOnlyIndex.HasPkgTypeMethodAttachment(pkgPath, typeName, methodName, c.currentPkgPath)

	// Also check by extracting package name from path
	if !isAllowed {
		currentPkgName := extractPackageName(c.currentPkgPath)
		isAllowed = c.packageOnlyIndex.HasPkgTypeMethodAttachment(pkgPath, typeName, methodName, currentPkgName)
	}

	if pkgPath != c.currentPkgPath && !isAllowed {
		// Get all allowed packages for error message
		allowedPackages := c.packageOnlyIndex.GetAttachmentsForMethod(pkgPath, typeName, methodName)
		violation := PackageOnlyViolation{
			ItemName:        methodName,
			ItemPkgPath:     pkgPath,
			CurrentPkgPath:  c.currentPkgPath,
			AllowedPackages: allowedPackages,
			ViolationType:   "method",
			ReceiverType:    typeName,
			Pos:             pos,
		}
		c.addViolation(violation)
	}
}

// extractPackageName extracts the package name from a full package path
// Example: "github.com/user/project/pkg/name" -> "name"
func extractPackageName(pkgPath string) string {
	// Split by slash and take the last part
	parts := strings.Split(pkgPath, "/")
	if len(parts) == 0 {
		return pkgPath
	}
	return parts[len(parts)-1]
}

// addViolation adds a violation if it's not ignored
func (c *packageOnlyContext) addViolation(violation PackageOnlyViolation) {
	// Check if this violation should be ignored
	if c.ignoreSet != nil && c.ignoreSet.Contains(violation.GetCode(), violation.Pos) {
		return
	}

	*c.violations = append(*c.violations, violation)
}
