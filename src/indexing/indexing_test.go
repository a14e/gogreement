package indexing

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/testutil"
	"goagreement/src/testutil/testfacts"
)

func TestBuildImmutableTypesIndex(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")

	index := BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass)

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
	pass := testfacts.CreateTestPassWithFacts(t, "withimports")

	// Setup ImportPackageFact to return empty annotations
	pass.ImportPackageFact = func(pkg *types.Package, fact analysis.Fact) bool {
		if targetAnnotations, ok := fact.(*annotations.ImmutableCheckerFact); ok {
			*targetAnnotations = annotations.ImmutableCheckerFact{}
			return true
		}
		return false
	}

	index := BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass)

	assert.Equal(t, 0, index.Len(), "should be empty when no immutable annotations")
}

func TestBuildImmutableTypesIndexWithImports(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")

	index := BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass)

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
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildConstructorIndex[*annotations.ConstructorCheckerFact](pass, &packageAnnotations)

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
	pass := testfacts.CreateTestPassWithFacts(t, "withimports")

	emptyAnnotations := annotations.PackageAnnotations{
		ConstructorAnnotations: []annotations.ConstructorAnnotation{},
	}

	index := BuildConstructorIndex[*annotations.ConstructorCheckerFact](pass, &emptyAnnotations)

	assert.Equal(t, 0, index.Len(), "should be empty when no constructor annotations")
}

func TestBuildConstructorIndexWithImports(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildConstructorIndex[*annotations.ConstructorCheckerFact](pass, &packageAnnotations)

	// Check local constructors
	localPkgPath := pass.Pkg.Path()
	assert.True(t, index.Match(localPkgPath, "NewPerson", "Person"))
	assert.True(t, index.Match(localPkgPath, "NewConfig", "Config"))

	// Note: interfacesforloading doesn't have @constructor annotations currently
	// This test verifies the indexing mechanism works for imports
	assert.Greater(t, index.Len(), 0, "should have constructors indexed")
}

func TestBuildTestOnlyTypesIndex(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildTestOnlyTypesIndex[*annotations.TestOnlyCheckerFact](pass, &packageAnnotations)

	t.Run("Contains TestHelper type", func(t *testing.T) {
		contains := index.Contains(pass.Pkg.Path(), "TestHelper")
		assert.True(t, contains, "TestHelper should be in index")
	})

	t.Run("Does not contain MyService type", func(t *testing.T) {
		contains := index.Contains(pass.Pkg.Path(), "MyService")
		assert.False(t, contains, "MyService should not be in index")
	})

	t.Run("Index has correct size", func(t *testing.T) {
		// Only TestHelper type should be indexed
		assert.Equal(t, 1, index.Len())
	})
}

func TestBuildTestOnlyFuncsIndex(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildTestOnlyFuncsIndex[*annotations.TestOnlyCheckerFact](pass, &packageAnnotations)

	t.Run("Contains CreateMockData function", func(t *testing.T) {
		matches := index.Match(pass.Pkg.Path(), "CreateMockData", "CreateMockData")
		assert.True(t, matches, "CreateMockData should be in index")
	})

	t.Run("Does not contain ProcessData function", func(t *testing.T) {
		matches := index.Match(pass.Pkg.Path(), "ProcessData", "ProcessData")
		assert.False(t, matches, "ProcessData should not be in index")
	})

	t.Run("Index has correct size", func(t *testing.T) {
		// Only CreateMockData function should be indexed
		assert.Equal(t, 1, index.Len())
	})
}

func TestBuildTestOnlyMethodsIndex(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	index := BuildTestOnlyMethodsIndex[*annotations.TestOnlyCheckerFact](pass, &packageAnnotations)

	t.Run("Contains Reset method", func(t *testing.T) {
		matches := index.Match(pass.Pkg.Path(), "Reset", "MyService")
		assert.True(t, matches, "Reset method should be in index")
	})

	t.Run("Contains GetTestData method", func(t *testing.T) {
		matches := index.Match(pass.Pkg.Path(), "GetTestData", "MyService")
		assert.True(t, matches, "GetTestData method should be in index")
	})

	t.Run("Does not contain Process method", func(t *testing.T) {
		matches := index.Match(pass.Pkg.Path(), "Process", "MyService")
		assert.False(t, matches, "Process method should not be in index")
	})

	t.Run("Index has correct size", func(t *testing.T) {
		// Reset and GetTestData methods should be indexed
		assert.Equal(t, 2, index.Len())
	})
}
