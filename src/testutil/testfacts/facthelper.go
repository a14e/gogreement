package testfacts

import (
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/testutil"
)

// CreateTestPassWithFacts creates a test pass with ImportPackageFact support for annotations
// @testonly
func CreateTestPassWithFacts(t *testing.T, pkgName string) *analysis.Pass {
	pass := testutil.CreateTestPass(t, pkgName)

	factCache := make(map[string]annotations.PackageAnnotations)

	pass.ImportPackageFact = func(pkg *types.Package, fact analysis.Fact) bool {
		// Handle all fact types
		var targetAnnotations *annotations.PackageAnnotations

		switch ptr := fact.(type) {
		case *annotations.ImmutableCheckerFact:
			targetAnnotations = (*annotations.PackageAnnotations)(ptr)
		case *annotations.ConstructorCheckerFact:
			targetAnnotations = (*annotations.PackageAnnotations)(ptr)
		case *annotations.TestOnlyCheckerFact:
			targetAnnotations = (*annotations.PackageAnnotations)(ptr)
		case *annotations.AnnotationReaderFact:
			targetAnnotations = (*annotations.PackageAnnotations)(ptr)
		case *annotations.ImplementsCheckerFact:
			targetAnnotations = (*annotations.PackageAnnotations)(ptr)
		case *annotations.PackageAnnotations:
			targetAnnotations = ptr
		default:
			return false
		}

		// Check cache first
		if cached, ok := factCache[pkg.Path()]; ok {
			*targetAnnotations = cached
			return true
		}

		// Load annotations from the imported package
		importedPass := testutil.LoadPackageByPath(t, pkg.Path())
		if importedPass == nil {
			return false
		}

		importedAnnotations := annotations.ReadAllAnnotations(config.Empty(), importedPass)
		factCache[pkg.Path()] = importedAnnotations
		*targetAnnotations = importedAnnotations
		return true
	}

	return pass
}
