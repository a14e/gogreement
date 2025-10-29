package constructor

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/reporting"
	"github.com/a14e/gogreement/src/util"
)

// ConstructorViolation represents a constructor violation
// implements reporting.Violation
type ConstructorViolation struct {
	TypeName string
	Reason   string
	Code     string // Error code from codes package
	Pos      token.Pos
	Node     ast.Node
}

// GetCode returns the error code for this violation
func (v ConstructorViolation) GetCode() string {
	return v.Code
}

// GetPos returns the position of the violation
func (v ConstructorViolation) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v ConstructorViolation) GetMessage() string {
	return fmt.Sprintf("[%s] %s", v.Code, v.Reason)
}

// ReportViolations reports constructor violations using the new pretty formatter
func ReportViolations(pass *analysis.Pass, violations []ConstructorViolation, ignoreSet *util.IgnoreSet) {
	reporter := reporting.NewReporter(pass, ignoreSet)

	// Convert to generic violations and report
	for _, violation := range violations {
		reporter.ReportViolation(violation)
	}
}
