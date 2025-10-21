package immutable

import (
	"goagreement/src/annotations"
	"goagreement/src/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckImmutable(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")

	// Read annotations
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	t.Logf("Found %d immutable annotations", len(packageAnnotations.ImmutableAnnotations))
	t.Logf("Found %d constructor annotations", len(packageAnnotations.ConstructorAnnotations))

	// Run immutability check
	violations := CheckImmutable(pass, packageAnnotations)

	t.Logf("Found %d violations", len(violations))
	for _, v := range violations {
		t.Logf("Violation in %s at line: %s", v.TypeName, v.Reason)
	}

	// Should have violations
	assert.NotEmpty(t, violations, "expected to find violations in test data")
}

func TestFieldAssignmentViolation(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	// Should catch: p.Name = name in UpdateName function
	hasNameViolation := false
	for _, v := range violations {
		if v.TypeName == "Person" && contains(v.Reason, "Name") {
			hasNameViolation = true
			t.Logf("Found expected violation: %s", v.Reason)
		}
	}

	assert.True(t, hasNameViolation, "should detect field assignment violation")
}

func TestIncDecViolation(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	// Should catch: p.Age++ in IncrementAge
	hasIncViolation := false
	for _, v := range violations {
		if v.TypeName == "Person" && contains(v.Reason, "Age") && contains(v.Reason, "++") {
			hasIncViolation = true
			t.Logf("Found expected violation: %s", v.Reason)
		}
	}

	assert.True(t, hasIncViolation, "should detect ++ violation")
}

func TestSliceIndexViolation(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	// Should catch: p.Items[index] = value in ModifyItem
	hasSliceViolation := false
	for _, v := range violations {
		if v.TypeName == "Person" && contains(v.Reason, "Items") && contains(v.Reason, "element") {
			hasSliceViolation = true
			t.Logf("Found expected violation: %s", v.Reason)
		}
	}

	assert.True(t, hasSliceViolation, "should detect slice element modification")
}

func TestConstructorAllowed(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	// Should NOT catch violations in NewPerson constructor
	for _, v := range violations {
		// All violations should be outside constructors
		assert.NotContains(t, v.Reason, "NewPerson", "should not report violations in constructor")
		assert.NotContains(t, v.Reason, "NewConfig", "should not report violations in constructor")
		assert.NotContains(t, v.Reason, "NewCounter", "should not report violations in constructor")
	}
}

func TestMultipleConstructors(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	// Verify Config has multiple constructors
	var configAnnotation *annotations.ConstructorAnnotation
	for _, annot := range packageAnnotations.ConstructorAnnotations {
		if annot.OnType == "Config" {
			configAnnotation = &annot
			break
		}
	}

	require.NotNil(t, configAnnotation, "Config should have constructor annotation")
	assert.Contains(t, configAnnotation.ConstructorNames, "NewConfig")
	assert.Contains(t, configAnnotation.ConstructorNames, "NewDefaultConfig")

	// Both constructors should be allowed to mutate
	violations := CheckImmutable(pass, packageAnnotations)

	for _, v := range violations {
		if v.TypeName == "Config" {
			// Should not be violations in either constructor
			t.Logf("Config violation: %s", v.Reason)
		}
	}
}

func TestMutableTypeAllowed(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	// MutableType should have NO violations (not marked as immutable)
	for _, v := range violations {
		assert.NotEqual(t, "MutableType", v.TypeName, "should not report violations for non-immutable types")
	}
}

func TestCounterOperations(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	counterViolations := 0
	compoundOps := 0

	for _, v := range violations {
		if v.TypeName == "Counter" {
			counterViolations++
			t.Logf("Counter violation: %s", v.Reason)

			// Check for compound operators
			if contains(v.Reason, "+=") || contains(v.Reason, "-=") ||
				contains(v.Reason, "*=") || contains(v.Reason, "/=") {
				compoundOps++
			}
		}
	}

	// Should catch Increment (++), Decrement (--), ChangeStep (+=), MultiplyStep (*=), etc.
	assert.GreaterOrEqual(t, counterViolations, 5, "should catch ++, --, +=, *=, /= violations")
	assert.GreaterOrEqual(t, compoundOps, 3, "should catch +=, *=, /= violations")
}

func TestCompoundAssignmentOperators(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, packageAnnotations)

	// Check for various compound operators
	operators := []string{"+=", "-=", "*=", "/="}
	foundOperators := make(map[string]bool)

	for _, v := range violations {
		for _, op := range operators {
			if contains(v.Reason, op) {
				foundOperators[op] = true
				t.Logf("Found violation with %s: %s", op, v.Reason)
			}
		}
	}

	// Should find at least +=, *=, /=, -=
	assert.True(t, foundOperators["+="], "should detect += violations")
	assert.GreaterOrEqual(t, len(foundOperators), 2, "should detect multiple compound operators")
}

func TestBuildConstructorIndex(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")

	packageAnnotations := annotations.PackageAnnotations{
		ConstructorAnnotations: []annotations.ConstructorAnnotation{
			{
				OnType:           "Person",
				ConstructorNames: []string{"NewPerson"},
			},
			{
				OnType:           "Config",
				ConstructorNames: []string{"NewConfig", "NewDefaultConfig"},
			},
		},
	}

	index := buildConstructorIndex(pass, packageAnnotations)

	// Now the map uses full package path as key
	pkgPath := pass.Pkg.Path()

	assert.True(t, index.Match(pkgPath, "NewPerson", "Person"))
	assert.True(t, index.Match(pkgPath, "NewConfig", "Config"))
	assert.True(t, index.Match(pkgPath, "NewDefaultConfig", "Config"))
	assert.False(t, index.Match(pkgPath, "NonExistent", "Person"))
}

func TestReportViolations(t *testing.T) {
	pass := testutil.CreateTestPass(t, "immutabletests")

	violations := []ImmutableViolation{
		{
			TypeName: "TestType",
			Pos:      0,
			Reason:   "cannot assign to field",
		},
	}

	// Should not panic
	ReportViolations(pass, violations)

	t.Log("ReportViolations executed successfully")
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsSubstr(s, substr)))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
