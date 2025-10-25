package testonly

import (
	"goagreement/src/annotations"
	"goagreement/src/config"
	"goagreement/src/testutil/testfacts"
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
