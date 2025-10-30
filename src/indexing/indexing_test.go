package indexing

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/testutil/testfacts"
)

func TestBuildImmutableTypesIndex(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

	index := BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass, &packageAnnotations)

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

	emptyAnnotations := annotations.PackageAnnotations{
		ImmutableAnnotations: []annotations.ImmutableAnnotation{},
	}

	index := BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass, &emptyAnnotations)

	assert.Equal(t, 0, index.Len(), "should be empty when no immutable annotations")
}

func TestBuildImmutableTypesIndexWithImports(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

	index := BuildImmutableTypesIndex[*annotations.ImmutableCheckerFact](pass, &packageAnnotations)

	// Check local types
	localPkgPath := pass.Pkg.Path()
	assert.True(t, index.Contains(localPkgPath, "Person"))
	assert.True(t, index.Contains(localPkgPath, "Config"))

	// Check imported types
	importedPkgPath := "github.com/a14e/gogreement/testdata/unit/interfacesforloading"
	assert.True(t, index.Contains(importedPkgPath, "FileReader"), "should include FileReader from imported package")
	assert.True(t, index.Contains(importedPkgPath, "BufferWriter"), "should include BufferWriter from imported package")
}

func TestBuildConstructorIndex(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

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

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

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

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

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

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

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

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

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

func TestBuildMutableFieldsIndex(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "withimports")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

	index := BuildMutableFieldsIndex[*annotations.AnnotationReaderFact](pass, &packageAnnotations)

	pkgPath := pass.Pkg.Path()

	t.Run("Contains cache field for MyReader", func(t *testing.T) {
		matches := index.Match(pkgPath, "cache", "MyReader")
		assert.True(t, matches, "cache field should be in index for MyReader")
	})

	t.Run("Get mutable fields for MyReader", func(t *testing.T) {
		fields := index.GetAssociated(pkgPath, "MyReader")
		require.NotNil(t, fields, "MyReader should have mutable fields")
		assert.Contains(t, fields, "cache", "cache should be in MyReader's mutable fields")
		assert.Equal(t, 1, len(fields), "MyReader should have exactly 1 mutable field")
	})

	t.Run("Index has correct size", func(t *testing.T) {
		// Only cache field should be indexed
		assert.Equal(t, 1, index.Len())
	})

	t.Run("HasType works correctly", func(t *testing.T) {
		assert.True(t, index.HasType(pkgPath, "MyReader"), "MyReader should have mutable fields")
		assert.False(t, index.HasType(pkgPath, "MyWriteCloser"), "MyWriteCloser should not have mutable fields")
		assert.False(t, index.HasType(pkgPath, "MyContext"), "MyContext should not have mutable fields")
	})
}

func TestBuildMutableFieldsIndexEmpty(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "withimports")

	emptyAnnotations := annotations.PackageAnnotations{
		MutableAnnotations: []annotations.MutableAnnotation{},
	}

	index := BuildMutableFieldsIndex[*annotations.AnnotationReaderFact](pass, &emptyAnnotations)

	assert.Equal(t, 0, index.Len(), "should be empty when no mutable annotations")
	assert.True(t, index.Empty(), "should be empty")
}

func TestBuildMutableFieldsIndexWithImports(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "withimports")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

	index := BuildMutableFieldsIndex[*annotations.AnnotationReaderFact](pass, &packageAnnotations)

	// Check local mutable fields
	localPkgPath := pass.Pkg.Path()
	assert.True(t, index.Match(localPkgPath, "cache", "MyReader"))

	// Verify the structure
	fields := index.GetAssociated(localPkgPath, "MyReader")
	require.NotNil(t, fields)
	assert.Contains(t, fields, "cache")

	assert.Greater(t, index.Len(), 0, "should have mutable fields indexed")
}

func TestBuildPackageOnlyIndex(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "packageonlysource")
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

	index := BuildPackageOnlyIndex[*annotations.AnnotationReaderFact](pass, &packageAnnotations)

	require.NotNil(t, index)

	pkgPath := pass.Pkg.Path()

	// Test type attachments
	t.Run("PackageOnlyType type attachments", func(t *testing.T) {
		// Check allowed packages
		assert.True(t, index.HasPkgTypeAttachment(pkgPath, "PackageOnlyType", "allowedpkg"),
			"PackageOnlyType should allow allowedpkg package")
		assert.True(t, index.HasPkgTypeAttachment(pkgPath, "PackageOnlyType", "packageonlyallowed"),
			"PackageOnlyType should allow packageonlyallowed package")
	})

	// Test function attachments
	t.Run("PackageOnlyFunction function attachments", func(t *testing.T) {
		// Check allowed packages
		assert.True(t, index.HasPkgFunctionAttachment(pkgPath, "PackageOnlyFunction", "allowedpkg"),
			"PackageOnlyFunction should allow allowedpkg package")
		assert.True(t, index.HasPkgFunctionAttachment(pkgPath, "PackageOnlyFunction", "packageonlyallowed"),
			"PackageOnlyFunction should allow packageonlyallowed package")
	})

	// Test method attachments
	t.Run("PackageOnlyStruct method attachments", func(t *testing.T) {
		// PackageOnlyMethod method
		assert.True(t, index.HasPkgTypeMethodAttachment(pkgPath, "PackageOnlyStruct", "PackageOnlyMethod", "allowedpkg"),
			"PackageOnlyMethod should allow allowedpkg package")
		assert.True(t, index.HasPkgTypeMethodAttachment(pkgPath, "PackageOnlyStruct", "PackageOnlyMethod", "packageonlyallowed"),
			"PackageOnlyMethod should allow packageonlyallowed package")
	})

	// Test that non-annotated items don't have packageonly markers
	t.Run("Non-annotated items should not have packageonly", func(t *testing.T) {
		assert.False(t, index.HasPkgTypeAttachment(pkgPath, "RegularType", "allowedpkg"),
			"RegularType type should not have allowedpkg attachment")
		assert.False(t, index.HasPkgFunctionAttachment(pkgPath, "RegularFunction", "allowedpkg"),
			"RegularFunction function should not have allowedpkg attachment")
		assert.False(t, index.HasPkgTypeMethodAttachment(pkgPath, "RegularStruct", "RegularMethod", "allowedpkg"),
			"RegularMethod method should not have allowedpkg attachment")
	})
}

func TestBuildPackageOnlyIndex_EmptyAnnotations(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests") // Package without packageonly annotations
	packageAnnotations := annotations.ReadAllAnnotations(config.Empty(), pass)

	index := BuildPackageOnlyIndex[*annotations.AnnotationReaderFact](pass, &packageAnnotations)

	require.NotNil(t, index)

	pkgPath := pass.Pkg.Path()

	// Should not have any packageonly attachments
	assert.False(t, index.HasAnyTypeAttachments(pkgPath, "AnyType"))
	assert.False(t, index.HasAnyFunctionAttachments(pkgPath, "AnyFunction"))
	assert.False(t, index.HasAnyMethodAttachments(pkgPath, "AnyType", "AnyMethod"))
}
