package testonly

import (
	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/testutil/testfacts"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTestOnly(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyviolations")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	violations := CheckTestOnly(cfg, pass, &packageAnnotations, nil)

	t.Run("Should detect violations", func(t *testing.T) {
		assert.NotEmpty(t, violations, "expected to find violations")
	})

	t.Run("Should detect type usage violation", func(t *testing.T) {
		found := false
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnType && v.TestOnlyObj == "TestHelper" {
				found = true
				assert.Contains(t, v.Reason, "TestHelper")
				assert.Contains(t, v.Reason, "@testonly")
				break
			}
		}
		assert.True(t, found, "should detect TestHelper usage violation")
	})

	t.Run("Should detect type assertion violation", func(t *testing.T) {
		found := false
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnType && v.TestOnlyObj == "MockCache" {
				found = true
				assert.Contains(t, v.Reason, "MockCache")
				assert.Contains(t, v.Reason, "@testonly")
				break
			}
		}
		assert.True(t, found, "should detect MockCache type assertion violation")
	})

	t.Run("Should detect function call violation", func(t *testing.T) {
		found := false
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnFunc && v.TestOnlyObj == "CreateMockData" {
				found = true
				assert.Contains(t, v.Reason, "CreateMockData")
				assert.Contains(t, v.Reason, "@testonly")
				break
			}
		}
		assert.True(t, found, "should detect CreateMockData call violation")
	})

	t.Run("Should detect method call violation", func(t *testing.T) {
		found := false
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnMethod && v.TestOnlyObj == "Worker.Reset" {
				found = true
				assert.Contains(t, v.Reason, "Reset")
				assert.Contains(t, v.Reason, "@testonly")
				break
			}
		}
		assert.True(t, found, "should detect Worker.Reset method call violation")
	})

	t.Run("Should NOT detect recursive @testonly function calls", func(t *testing.T) {
		// CreateMockDataWrapper is @testonly and calls CreateMockData which is also @testonly
		// This should NOT create a violation
		for _, v := range violations {
			if v.TestOnlyObj == "CreateMockData" {
				// Make sure this violation is from ProcessData, not from CreateMockDataWrapper
				assert.NotContains(t, v.UsedInFile, "CreateMockDataWrapper",
					"should not report violation for @testonly calling @testonly")
			}
		}
	})

	t.Run("Should NOT detect recursive @testonly method calls", func(t *testing.T) {
		// Worker.ResetAll is @testonly and calls Worker.Reset which is also @testonly
		// This should NOT create a violation
		for _, v := range violations {
			if v.TestOnlyObj == "Worker.Reset" {
				// The violation should be from UseWorker, not from ResetAll
				position := pass.Fset.Position(v.Pos)
				assert.NotContains(t, position.String(), "ResetAll",
					"should not report violation for @testonly method calling @testonly method")
			}
		}
	})

	t.Run("Violations should be in non-test files", func(t *testing.T) {
		for _, v := range violations {
			assert.NotContains(t, v.UsedInFile, "_test.go", "violations should not be in test files")
		}
	})
}

func TestCheckTestOnlyExtended(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "testonlyviolations")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckTestOnly(cfg, pass, &packageAnnotations, nil)

	hasType := func(obj string) bool {
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnType && v.TestOnlyObj == obj {
				return true
			}
		}
		return false
	}

	t.Run("method receiver on a @testonly type is not flagged", func(t *testing.T) {
		for _, v := range violations {
			assert.NotEqual(t, "ReceiverOnly", v.TestOnlyObj,
				"declaring a method on a @testonly type must not be reported")
		}
	})

	t.Run("type conversion is detected", func(t *testing.T) {
		assert.True(t, hasType("TestID"), "conversion TestID(x) should be flagged")
	})

	t.Run("new(T) is detected", func(t *testing.T) {
		assert.True(t, hasType("MockNew"), "new(MockNew) should be flagged")
	})

	t.Run("slice element type is detected", func(t *testing.T) {
		assert.True(t, hasType("MockElem"), "[]MockElem{} should be flagged")
	})

	t.Run("make element type is detected", func(t *testing.T) {
		assert.True(t, hasType("MockMake"), "make([]MockMake, 0) should be flagged")
	})

	t.Run("shadowed function name is not flagged", func(t *testing.T) {
		count := 0
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnFunc && v.TestOnlyObj == "CreateMockData" {
				count++
			}
		}
		assert.Equal(t, 1, count,
			"only the real CreateMockData call should be flagged, not the shadowing local")
	})

	t.Run("@testonly method on a generic type is detected", func(t *testing.T) {
		found := false
		for _, v := range violations {
			if v.Kind == annotations.TestOnlyOnMethod && v.TestOnlyObj == "Container.DebugContainer" {
				found = true
			}
		}
		assert.True(t, found, "c.DebugContainer() on a generic type should be flagged")
	})

	t.Run("@testonly generic type itself is detected", func(t *testing.T) {
		assert.True(t, hasType("GenericMock"),
			"using a @testonly generic type (GenericMock[int]{}) should be flagged")
	})
}

func TestCrossPackageSameNameNotDeduped(t *testing.T) {
	// Two @testonly types named "Dup" from different packages, used in one file,
	// must both be reported (dedup is keyed by package-qualified identity).
	pass := testfacts.CreateTestPassWithFacts(t, "dupconsumer", "dupsrca", "dupsrcb")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckTestOnly(cfg, pass, &packageAnnotations, nil)

	dupCount := 0
	for _, v := range violations {
		if v.Kind == annotations.TestOnlyOnType && v.TestOnlyObj == "Dup" {
			dupCount++
		}
	}

	assert.Equal(t, 2, dupCount,
		"both same-named @testonly types from different packages must be reported")
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"test file", "foo_test.go", true},
		{"test file with path", "/path/to/foo_test.go", true},
		{"regular file", "foo.go", false},
		{"regular file with test in name", "testing.go", false},
		{"test in directory name", "/test/foo.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTestFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetTypeName(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "testonlyexample")

	t.Run("Extract type name from types", func(t *testing.T) {
		// We'll test this indirectly through the checker
		// as we need actual types.Type objects
		cfg := config.Empty()
		packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
		require.NotEmpty(t, packageAnnotations.TestonlyAnnotations)
	})
}

func TestCheckTestOnlyWithNoAnnotations(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "immutabletests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	// Filter out testonly annotations for this test
	packageAnnotations.TestonlyAnnotations = nil

	violations := CheckTestOnly(cfg, pass, &packageAnnotations, nil)

	assert.Empty(t, violations, "should have no violations when no @testonly annotations")
}
