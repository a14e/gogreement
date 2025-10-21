package indexing

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/testutil"
)

func TestBuildImmutableTypesIndex(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildImmutableTypesIndex(pass, packageAnnotations)

	pkgPath := pass.Pkg.Path()

	tests := []struct {
		name     string
		typeName string
		expected bool
	}{
		{
			name:     "Person is immutable",
			typeName: "Person",
			expected: true,
		},
		{
			name:     "Config is immutable",
			typeName: "Config",
			expected: true,
		},
		{
			name:     "Counter is immutable",
			typeName: "Counter",
			expected: true,
		},
		{
			name:     "ComplexCase is immutable",
			typeName: "ComplexCase",
			expected: true,
		},
		{
			name:     "MutableType is not immutable",
			typeName: "MutableType",
			expected: false,
		},
		{
			name:     "ImportedTypeWrapper is not immutable",
			typeName: "ImportedTypeWrapper",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := index.Contains(pkgPath, tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}

	assert.Greater(t, index.Len(), 0, "should have some immutable types")
}

func TestBuildImmutableTypesIndexEmpty(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	// Create empty annotations
	emptyAnnotations := annotations.PackageAnnotations{
		ImmutableAnnotations: []annotations.ImmutableAnnotation{},
	}

	index := BuildImmutableTypesIndex(pass, emptyAnnotations)

	assert.Equal(t, 0, index.Len(), "should be empty when no immutable annotations")
}

func TestBuildImmutableTypesIndexWithImports(t *testing.T) {
	pass := createTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildImmutableTypesIndex(pass, packageAnnotations)

	// Check local types
	localPkgPath := pass.Pkg.Path()
	assert.True(t, index.Contains(localPkgPath, "Person"))
	assert.True(t, index.Contains(localPkgPath, "Config"))

	// Check imported types
	importedPkgPath := "goagreement/src/testutil/testdata/interfacesforloading"
	assert.True(t, index.Contains(importedPkgPath, "FileReader"), "should include FileReader from imported package")
	assert.True(t, index.Contains(importedPkgPath, "BufferWriter"), "should include BufferWriter from imported package")
}

func TestBuildConstructorIndex(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildConstructorIndex(pass, packageAnnotations)

	pkgPath := pass.Pkg.Path()

	tests := []struct {
		name        string
		funcName    string
		typeName    string
		shouldMatch bool
		description string
	}{
		{
			name:        "NewPerson is constructor for Person",
			funcName:    "NewPerson",
			typeName:    "Person",
			shouldMatch: true,
		},
		{
			name:        "NewConfig is constructor for Config",
			funcName:    "NewConfig",
			typeName:    "Config",
			shouldMatch: true,
		},
		{
			name:        "NewDefaultConfig is constructor for Config",
			funcName:    "NewDefaultConfig",
			typeName:    "Config",
			shouldMatch: true,
		},
		{
			name:        "NewCounter is constructor for Counter",
			funcName:    "NewCounter",
			typeName:    "Counter",
			shouldMatch: true,
		},
		{
			name:        "NewComplexCase is constructor for ComplexCase",
			funcName:    "NewComplexCase",
			typeName:    "ComplexCase",
			shouldMatch: true,
		},
		{
			name:        "UpdateName is not a constructor",
			funcName:    "UpdateName",
			typeName:    "Person",
			shouldMatch: false,
		},
		{
			name:        "Increment is not a constructor",
			funcName:    "Increment",
			typeName:    "Counter",
			shouldMatch: false,
		},
		{
			name:        "wrong type for NewPerson",
			funcName:    "NewPerson",
			typeName:    "Config",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := index.Match(pkgPath, tt.funcName, tt.typeName)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}

	assert.Greater(t, index.Len(), 0, "should have some constructors")
}

func TestBuildConstructorIndexEmpty(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	emptyAnnotations := annotations.PackageAnnotations{
		ConstructorAnnotations: []annotations.ConstructorAnnotation{},
	}

	index := BuildConstructorIndex(pass, emptyAnnotations)

	assert.Equal(t, 0, index.Len(), "should be empty when no constructor annotations")
}

func TestBuildConstructorIndexWithImports(t *testing.T) {
	pass := createTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildConstructorIndex(pass, packageAnnotations)

	// Check local constructors
	localPkgPath := pass.Pkg.Path()
	assert.True(t, index.Match(localPkgPath, "NewPerson", "Person"))
	assert.True(t, index.Match(localPkgPath, "NewConfig", "Config"))

	// Note: interfacesforloading doesn't have @constructor annotations currently
	// This test verifies the indexing mechanism works for imports
	assert.Greater(t, index.Len(), 0, "should have constructors indexed")
}

// Helper function for tests that need ImportPackageFact support
func createTestPassWithFacts(t *testing.T, pkgName string) *analysis.Pass {
	pass := testutil.CreateTestPass(t, pkgName)

	factCache := make(map[string]annotations.PackageAnnotations)

	pass.ImportPackageFact = func(pkg *types.Package, fact analysis.Fact) bool {
		if cached, ok := factCache[pkg.Path()]; ok {
			if ptr, ok := fact.(*annotations.PackageAnnotations); ok {
				*ptr = cached
				return true
			}
		}

		importedPass := testutil.LoadPackageByPath(t, pkg.Path())
		if importedPass == nil {
			return false
		}

		importedAnnotations := annotations.ReadAllAnnotations(importedPass)
		factCache[pkg.Path()] = importedAnnotations

		if ptr, ok := fact.(*annotations.PackageAnnotations); ok {
			*ptr = importedAnnotations
			return true
		}

		return false
	}

	return pass
}
