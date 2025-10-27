package constructor

import (
	"go/ast"
	"go/token"
	"gogreement/src/annotations"
	"gogreement/src/config"
	"gogreement/src/testutil/testfacts"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

func TestCheckConstructor(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	t.Logf("Found %d constructor annotations", len(packageAnnotations.ConstructorAnnotations))

	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	t.Logf("Found %d violations", len(violations))
	for _, v := range violations {
		position := pass.Fset.Position(v.Pos)
		t.Logf("Violation in %s at %s: %s", v.TypeName, position, v.Reason)
	}

	assert.NotEmpty(t, violations, "expected to find violations in test data")
}

func TestCompositeLiteralViolation(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	hasUserViolation := false
	for _, v := range violations {
		if v.TypeName == "User" {
			hasUserViolation = true
			t.Logf("Found User violation: %s", v.Reason)
		}
	}

	assert.True(t, hasUserViolation, "should detect User composite literal violation")
}

func TestNewCallViolation(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	hasNewViolation := false
	for _, v := range violations {
		if contains(v.Reason, "new()") {
			hasNewViolation = true
			t.Logf("Found new() violation: %s in %s", v.Reason, v.TypeName)
		}
	}

	assert.True(t, hasNewViolation, "should detect new() call violation")
}

func TestConstructorAllowed(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	for _, v := range violations {
		position := pass.Fset.Position(v.Pos)
		funcName := getFunctionNameFromPosition(pass, v.Pos)

		assert.NotEqual(t, "NewUser", funcName, "should not report violations in NewUser")
		assert.NotEqual(t, "NewConfig", funcName, "should not report violations in NewConfig")
		assert.NotEqual(t, "NewDefaultConfig", funcName, "should not report violations in NewDefaultConfig")
		assert.NotEqual(t, "NewDatabase", funcName, "should not report violations in NewDatabase")
		assert.NotEqual(t, "NewPoint", funcName, "should not report violations in NewPoint")

		t.Logf("Violation at %s: %s", position, v.Reason)
	}
}

func TestMultipleConstructors(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	var configAnnotation *annotations.ConstructorAnnotation
	for _, annot := range packageAnnotations.ConstructorAnnotations {
		if annot.OnType == "Config" {
			configAnnotation = &annot
			break
		}
	}

	assert.NotNil(t, configAnnotation, "Config should have constructor annotation")
	assert.Contains(t, configAnnotation.ConstructorNames, "NewConfig")
	assert.Contains(t, configAnnotation.ConstructorNames, "NewDefaultConfig")

	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	for _, v := range violations {
		if v.TypeName == "Config" {
			funcName := getFunctionNameFromPosition(pass, v.Pos)
			assert.NotEqual(t, "NewConfig", funcName, "NewConfig should be allowed")
			assert.NotEqual(t, "NewDefaultConfig", funcName, "NewDefaultConfig should be allowed")
		}
	}
}

func TestNoAnnotationAllowed(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	for _, v := range violations {
		assert.NotEqual(t, "Service", v.TypeName, "should not report violations for types without @constructor")
	}
}

func TestReportViolations(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")

	violations := []ConstructorViolation{
		{
			TypeName: "TestType",
			Pos:      0,
			Reason:   "instantiation outside constructor",
		},
	}

	ReportViolations(pass, violations, nil)
	t.Log("ReportViolations executed successfully")
}

func TestEmptyConstructorAnnotations(t *testing.T) {
	pass := testfacts.CreateTestPassWithFacts(t, "withimports")
	cfg := config.Empty()

	emptyAnnotations := annotations.PackageAnnotations{
		ConstructorAnnotations: []annotations.ConstructorAnnotation{},
	}

	violations := CheckConstructor(cfg, pass, &emptyAnnotations)

	assert.Empty(t, violations, "should have no violations when no @constructor annotations")
}

func TestValueReceiverConstructor(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	hasPointViolation := false
	for _, v := range violations {
		if v.TypeName == "Point" {
			funcName := getFunctionNameFromPosition(pass, v.Pos)
			if funcName != "NewPoint" {
				hasPointViolation = true
			}
		}
	}

	assert.True(t, hasPointViolation, "should detect Point violation outside NewPoint")
}

func TestNestedInstantiation(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "constructortests")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)
	violations := CheckConstructor(cfg, pass, &packageAnnotations)

	hasHelperViolation := false
	for _, v := range violations {
		funcName := getFunctionNameFromPosition(pass, v.Pos)
		if funcName == "HelperFunction" {
			hasHelperViolation = true
			t.Logf("Found violation in HelperFunction: %s", v.Reason)
		}
	}

	assert.True(t, hasHelperViolation, "should detect violations in helper functions")
}

func getFunctionNameFromPosition(pass *analysis.Pass, pos token.Pos) string {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok {
				if funcDecl.Pos() <= pos && pos <= funcDecl.End() {
					return funcDecl.Name.Name
				}
			}
		}
	}
	return ""
}

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
