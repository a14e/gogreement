package constructor

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/util"
)

type ConstructorViolation struct {
	TypeName string
	Reason   string
	Code     string // Error code from codes package
	Pos      token.Pos
	Node     ast.Node
}

// ReportViolations reports constructor violations via analysis.Pass
func ReportViolations(pass *analysis.Pass, violations []ConstructorViolation, ignoreSet *util.IgnoreSet) {
	for _, v := range violations {
		// Check if this violation should be ignored
		if ignoreSet.Contains(v.Code, v.Pos) {
			continue
		}

		msg := fmt.Sprintf("[%s] %s", v.Code, v.Reason)

		pass.Report(analysis.Diagnostic{
			Pos:     v.Pos,
			End:     0,
			Message: msg,
			Related: nil,
		})
	}
}
