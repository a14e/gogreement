package testonly

import (
	"go/token"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
)

// TestOnlyViolation represents a violation of @testonly usage
// @immutable
type TestOnlyViolation struct {
	Pos         token.Pos
	TestOnlyObj string // Name of the @testonly object being used
	Kind        annotations.TestOnlyKind
	UsedInFile  string // File where @testonly object is used
	Reason      string
}

func ReportViolations(pass *analysis.Pass, violations []TestOnlyViolation) {
	for _, v := range violations {
		pass.Report(analysis.Diagnostic{
			Pos:     v.Pos,
			Message: v.Reason,
		})
	}
}
