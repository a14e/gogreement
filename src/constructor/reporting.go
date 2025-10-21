package constructor

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type ConstructorViolation struct {
	TypeName string
	Reason   string
	Pos      token.Pos
	Node     ast.Node
}

// ReportViolations reports constructor violations via analysis.Pass
func ReportViolations(pass *analysis.Pass, violations []ConstructorViolation) {
	for _, v := range violations {
		pass.Report(analysis.Diagnostic{
			Pos:     v.Pos,
			End:     0,
			Message: v.Reason,
			Related: nil,
		})
	}
}
