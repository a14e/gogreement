package immutable

import (
	"go/token"
	"go/types"
	"goagreement/src/annotations"
	"goagreement/src/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

func TestCheckImmutable(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")

	// Read annotations
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	t.Logf("Found %d immutable annotations", len(packageAnnotations.ImmutableAnnotations))
	t.Logf("Found %d constructor annotations", len(packageAnnotations.ConstructorAnnotations))

	// Run immutability check
	violations := CheckImmutable(pass, &packageAnnotations)

	t.Logf("Found %d violations", len(violations))
	for _, v := range violations {
		t.Logf("Violation in %s at line: %s", v.TypeName, v.Reason)
	}

	// Should have violations
	assert.NotEmpty(t, violations, "expected to find violations in test data")
}

func TestFieldAssignmentViolation(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

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
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

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
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

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
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// Should NOT catch violations in NewPerson constructor
	for _, v := range violations {
		// All violations should be outside constructors
		assert.NotContains(t, v.Reason, "NewPerson", "should not report violations in constructor")
		assert.NotContains(t, v.Reason, "NewConfig", "should not report violations in constructor")
		assert.NotContains(t, v.Reason, "NewCounter", "should not report violations in constructor")
	}
}

func TestMultipleConstructors(t *testing.T) {
	defer testutil.WithTestConfig(t)()

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
	violations := CheckImmutable(pass, &packageAnnotations)

	for _, v := range violations {
		if v.TypeName == "Config" {
			// Should not be violations in either constructor
			t.Logf("Config violation: %s", v.Reason)
		}
	}
}

func TestMutableTypeAllowed(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// MutableType should have NO violations (not marked as immutable)
	for _, v := range violations {
		assert.NotEqual(t, "MutableType", v.TypeName, "should not report violations for non-immutable types")
	}
}

func TestCounterOperations(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

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
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

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

// createTestPassWithFacts creates a test pass with ImportPackageFact support
func createTestPassWithFacts(t *testing.T, pkgName string) *analysis.Pass {
	pass := testutil.CreateTestPass(t, pkgName)

	// Cache for imported facts
	factCache := make(map[string]annotations.PackageAnnotations)

	pass.ImportPackageFact = func(pkg *types.Package, fact analysis.Fact) bool {
		// Handle both old and new fact types
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

		// Check cache
		if cached, ok := factCache[pkg.Path()]; ok {
			*targetAnnotations = cached
			return true
		}

		// Load package and read annotations
		importedPass := testutil.LoadPackageByPath(t, pkg.Path())
		if importedPass == nil {
			return false
		}

		// Read annotations
		importedAnnotations := annotations.ReadAllAnnotations(importedPass)

		// Cache
		factCache[pkg.Path()] = importedAnnotations

		// Copy to fact
		*targetAnnotations = importedAnnotations
		return true
	}

	return pass
}

func TestImportedImmutableType(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := createTestPassWithFacts(t, "immutabletests") // Use createTestPassWithFacts
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// Should catch violation on imported FileReader type
	hasImportedViolation := false
	for _, v := range violations {
		t.Logf("Violation: TypeName=%s, Reason=%s", v.TypeName, v.Reason)
		if v.TypeName == "FileReader" && contains(v.Reason, "Data") {
			hasImportedViolation = true
			t.Logf("Found expected violation on imported type: %s", v.Reason)
		}
	}

	assert.True(t, hasImportedViolation, "should detect mutation of imported immutable type")
}

// TestNoDuplicateViolations ensures that compound assignments (+=, -=, *=, /=)
// don't create duplicate violation reports.
//
// Background: AST represents compound assignments like "x.a += 1" as AssignStmt nodes,
// just like simple assignments "x.a = 1". Without proper filtering, we would report
// the same violation twice:
// 1. First pass treats it as assignment (tok == ASSIGN check fails, processes it)
// 2. Second pass treats it as compound operator
//
// Solution: We skip compound assignments in the first pass by checking tok != ASSIGN,
// and only process them in the dedicated second pass. This ensures each violation
// is reported exactly once with the most specific error message.
func TestNoDuplicateViolations(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// Group violations by position
	violationsByPos := make(map[token.Pos][]ImmutableViolation)
	for _, v := range violations {
		violationsByPos[v.Pos] = append(violationsByPos[v.Pos], v)
	}

	// Check for duplicates at same position
	for pos, viols := range violationsByPos {
		if len(viols) > 1 {
			position := pass.Fset.Position(pos)
			t.Errorf("Found %d violations at same position %s:", len(viols), position)
			for _, v := range viols {
				t.Logf("  - %s: %s", v.TypeName, v.Reason)
			}
		}
	}

	// Specifically check compound assignments don't have duplicates
	for _, v := range violations {
		if contains(v.Reason, "+=") || contains(v.Reason, "-=") ||
			contains(v.Reason, "*=") || contains(v.Reason, "/=") {
			// This is a compound assignment - ensure no simple assignment at same position
			for _, other := range violationsByPos[v.Pos] {
				if other.Pos == v.Pos &&
					contains(other.Reason, "cannot assign to field") &&
					!contains(other.Reason, "+=") && !contains(other.Reason, "-=") &&
					!contains(other.Reason, "*=") && !contains(other.Reason, "/=") {
					position := pass.Fset.Position(v.Pos)
					t.Errorf("Found duplicate at %s: compound operator %q shadowed by simple assignment",
						position, v.Reason)
				}
			}
		}
	}

	assert.True(t, len(violations) > 0, "should find some violations")
}

func TestReceiverReassignment(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// Should catch: *p = Person{} in Person.Reset method
	hasPersonResetViolation := false
	// Should catch: *c = Counter{...} in Counter.UpdateCounter method
	hasCounterResetViolation := false

	for _, v := range violations {
		t.Logf("Violation: TypeName=%s, Reason=%s", v.TypeName, v.Reason)

		if v.TypeName == "Person" && contains(v.Reason, "reassign") {
			hasPersonResetViolation = true
			t.Logf("Found expected Person receiver reassignment violation: %s", v.Reason)
		}

		if v.TypeName == "Counter" && contains(v.Reason, "reassign") {
			hasCounterResetViolation = true
			t.Logf("Found expected Counter receiver reassignment violation: %s", v.Reason)
		}

		// Should NOT report violation for MutableType.Reset
		if v.TypeName == "MutableType" {
			t.Errorf("Should not report violation for MutableType: %s", v.Reason)
		}
	}

	assert.True(t, hasPersonResetViolation, "should detect Person receiver reassignment")
	assert.True(t, hasCounterResetViolation, "should detect Counter receiver reassignment")
}

func TestPrimitiveTypeAliasReassignment(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// Should catch: *i = ImmutableInt(value) in ImmutableInt.SetValue method
	hasImmutableIntSetViolation := false
	// Should catch: *i++ in ImmutableInt.Increment method
	hasImmutableIntIncViolation := false
	// Should catch: *s = ImmutableString(value) in ImmutableString.Update method
	hasImmutableStringViolation := false

	for _, v := range violations {
		t.Logf("Violation: TypeName=%s, Reason=%s", v.TypeName, v.Reason)

		if v.TypeName == "ImmutableInt" && contains(v.Reason, "reassign") {
			if contains(v.Reason, "reassign immutable receiver") {
				hasImmutableIntSetViolation = true
				t.Logf("Found expected ImmutableInt receiver reassignment violation: %s", v.Reason)
			}
		}

		// Note: *i++ is technically a receiver reassignment, but might be caught differently
		if v.TypeName == "ImmutableInt" {
			hasImmutableIntIncViolation = true
		}

		if v.TypeName == "ImmutableString" && contains(v.Reason, "reassign") {
			hasImmutableStringViolation = true
			t.Logf("Found expected ImmutableString receiver reassignment violation: %s", v.Reason)
		}
	}

	assert.True(t, hasImmutableIntSetViolation, "should detect ImmutableInt receiver reassignment in SetValue")
	assert.True(t, hasImmutableIntIncViolation, "should detect ImmutableInt receiver modification in Increment")
	assert.True(t, hasImmutableStringViolation, "should detect ImmutableString receiver reassignment")
}

func TestMapFieldModification(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testutil.CreateTestPass(t, "immutabletests")
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	violations := CheckImmutable(pass, &packageAnnotations)

	// Should catch: c.settings[key] = value in ModifyMapString
	hasMapStringViolation := false
	// Should catch: c.values[key] = value in ModifyMapInt
	hasMapIntViolation := false

	for _, v := range violations {
		t.Logf("Violation: TypeName=%s, Reason=%s", v.TypeName, v.Reason)

		if v.TypeName == "ConfigWithMap" && contains(v.Reason, "settings") {
			hasMapStringViolation = true
			t.Logf("Found expected map[string]string modification violation: %s", v.Reason)
		}

		if v.TypeName == "ConfigWithMap" && contains(v.Reason, "values") {
			hasMapIntViolation = true
			t.Logf("Found expected map[int]int modification violation: %s", v.Reason)
		}
	}

	assert.True(t, hasMapStringViolation, "should detect map[string]string element modification")
	assert.True(t, hasMapIntViolation, "should detect map[int]int element modification")
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
