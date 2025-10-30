package packageonly

import (
	"go/token"
	"testing"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/testutil/testfacts"

	"github.com/stretchr/testify/assert"
)

func TestCheckPackageOnly_ForbiddenUsage(t *testing.T) {
	// Test violations in forbidden package with source package facts
	pass := testfacts.CreateTestPassWithFacts(t, "packageonlyviolations", "packageonlysource")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	// Debug: print what annotations were read
	t.Logf("Read %d PackageOnlyAnnotations", len(packageAnnotations.PackageOnlyAnnotations))
	for i, ann := range packageAnnotations.PackageOnlyAnnotations {
		t.Logf("  Annotation %d: Kind=%v, ObjectName=%s, AllowedPackages=%v", i, ann.Kind, ann.ObjectName, ann.AllowedPackages)
	}

	violations := CheckPackageOnly(cfg, pass, &packageAnnotations, nil)

	t.Run("Should detect violations", func(t *testing.T) {
		assert.NotEmpty(t, violations, "expected to find violations")
	})

	// Check that we have the expected violation types
	violationTypes := make(map[string]bool)
	for _, v := range violations {
		violationTypes[v.ViolationType] = true
		t.Logf("Violation: %s", v.GetMessage())
	}

	expectedTypes := []string{"type", "function", "method"}
	for _, expectedType := range expectedTypes {
		if !violationTypes[expectedType] {
			t.Errorf("Expected violation type %s not found", expectedType)
		}
	}
}

func TestCheckPackageOnly_AllowedUsage(t *testing.T) {
	// Test violations in allowed package with source package facts
	pass := testfacts.CreateTestPassWithFacts(t, "packageonlyallowed", "packageonlysource")
	cfg := config.Empty()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	violations := CheckPackageOnly(cfg, pass, &packageAnnotations, nil)

	// Debug: print what violations were found
	t.Logf("Found %d violations", len(violations))
	for i, v := range violations {
		t.Logf("  Violation %d: %s", i, v.GetMessage())
	}

	t.Run("Should have no violations", func(t *testing.T) {
		assert.Empty(t, violations, "expected no violations for allowed usage")
	})
}

func TestPackageOnlyViolation_GetCode(t *testing.T) {
	tests := []struct {
		name          string
		violationType string
		expectedCode  string
	}{
		{"Type violation", "type", "PKGO01"},
		{"Function violation", "function", "PKGO02"},
		{"Method violation", "method", "PKGO03"},
		{"Unknown violation", "unknown", "PKGO00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violation := PackageOnlyViolation{
				ViolationType: tt.violationType,
				Pos:           token.Pos(1),
			}

			code := violation.GetCode()
			if code != tt.expectedCode {
				t.Errorf("GetCode() = %s, want %s", code, tt.expectedCode)
			}
		})
	}
}

func TestPackageOnlyViolation_GetMessage(t *testing.T) {
	tests := []struct {
		name           string
		violation      PackageOnlyViolation
		expectedSubstr string
	}{
		{
			name: "Type violation",
			violation: PackageOnlyViolation{
				ItemName:        "MyType",
				CurrentPkgPath:  "github.com/example/current",
				AllowedPackages: []string{"github.com/example/allowed"},
				ViolationType:   "type",
			},
			expectedSubstr: "MyType type is @packageonly",
		},
		{
			name: "Function violation",
			violation: PackageOnlyViolation{
				ItemName:        "MyFunction",
				CurrentPkgPath:  "github.com/example/current",
				AllowedPackages: []string{"github.com/example/allowed"},
				ViolationType:   "function",
			},
			expectedSubstr: "MyFunction function is @packageonly",
		},
		{
			name: "Method violation",
			violation: PackageOnlyViolation{
				ItemName:        "MyMethod",
				ReceiverType:    "MyStruct",
				CurrentPkgPath:  "github.com/example/current",
				AllowedPackages: []string{"github.com/example/allowed"},
				ViolationType:   "method",
			},
			expectedSubstr: "MyStruct.MyMethod method is @packageonly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := tt.violation.GetMessage()
			if !contains(message, tt.expectedSubstr) {
				t.Errorf("GetMessage() = %s, want to contain %s", message, tt.expectedSubstr)
			}
		})
	}
}

func TestPackageOnlyViolation_GetPos(t *testing.T) {
	expectedPos := token.Pos(123)
	violation := PackageOnlyViolation{
		Pos: expectedPos,
	}

	if violation.GetPos() != expectedPos {
		t.Errorf("GetPos() = %v, want %v", violation.GetPos(), expectedPos)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
