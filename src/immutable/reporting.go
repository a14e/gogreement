package immutable

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/reporting"
	"github.com/a14e/gogreement/src/util"
)

// ImmutableViolation represents a mutation of an immutable type
// @immutable
// implements reporting.Violation
type ImmutableViolation struct {
	TypeName string
	Reason   string
	Code     string // Error code from codes package
	Pos      token.Pos
	Node     ast.Node
}

// GetCode returns the error code for this violation
func (v ImmutableViolation) GetCode() string {
	return v.Code
}

// GetPos returns the position of the violation
func (v ImmutableViolation) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v ImmutableViolation) GetMessage() string {
	return fmt.Sprintf("immutability violation in type %q: %s", v.TypeName, v.Reason)
}

// ReportViolations reports immutable violations using the new pretty formatter
func ReportViolations(pass *analysis.Pass, violations []ImmutableViolation, ignoreSet *util.IgnoreSet) {
	reporter := reporting.NewReporter(pass, ignoreSet)

	// Convert to generic violations and report
	for _, violation := range violations {
		reporter.ReportViolation(violation)
	}
}
