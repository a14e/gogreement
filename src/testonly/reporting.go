package testonly

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
)

// TestOnlyViolation represents a violation of @testonly usage
// @immutable
type TestOnlyViolation struct {
	Pos         token.Pos
	TestOnlyObj string // Name of the @testonly object being used
	Kind        annotations.TestOnlyKind
	UsedInFile  string // File where @testonly object is used
	Reason      string
	Code        string // Error code from codes package
}

// ReportViolations reports testonly violations via analysis.Pass
// NOTE: violations should already be filtered by @ignore directives in CheckTestOnly
func ReportViolations(pass *analysis.Pass, violations []TestOnlyViolation) {
	for _, v := range violations {
		msg := fmt.Sprintf("[%s] %s", v.Code, v.Reason)

		pass.Report(analysis.Diagnostic{
			Pos:     v.Pos,
			Message: msg,
		})
	}
}
