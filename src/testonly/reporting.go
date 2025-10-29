package testonly

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/reporting"
)

// TestOnlyViolation represents a violation of @testonly usage
// @immutable
// implements reporting.Violation
type TestOnlyViolation struct {
	Pos         token.Pos
	TestOnlyObj string // Name of the @testonly object being used
	Kind        annotations.TestOnlyKind
	UsedInFile  string // File where @testonly object is used
	Reason      string
	Code        string // Error code from codes package
}

// GetCode returns the error code for this violation
func (v TestOnlyViolation) GetCode() string {
	return v.Code
}

// GetPos returns the position of the violation
func (v TestOnlyViolation) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v TestOnlyViolation) GetMessage() string {
	return fmt.Sprintf("[%s] %s", v.Code, v.Reason)
}

// ReportViolations reports testonly violations using the new pretty formatter
// NOTE: violations should already be filtered by @ignore directives in CheckTestOnly
func ReportViolations(pass *analysis.Pass, violations []TestOnlyViolation) {
	reporter := reporting.NewReporter(pass, nil) // No ignore set needed, already filtered

	// Convert to generic violations and report
	for _, violation := range violations {
		reporter.ReportViolation(violation)
	}
}
