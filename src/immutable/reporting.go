package immutable

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ImmutableViolation represents a mutation of an immutable type
// @immutable
type ImmutableViolation struct {
	TypeName string
	Reason   string
	Pos      token.Pos
	Node     ast.Node
}

func ReportViolations(pass *analysis.Pass, violations []ImmutableViolation) {
	for _, v := range violations {
		msg := fmt.Sprintf("immutability violation in type %q: %s", v.TypeName, v.Reason)

		if v.Node != nil {
			var buf bytes.Buffer
			if err := format.Node(&buf, pass.Fset, v.Node); err == nil {
				code := strings.TrimSpace(buf.String())
				msg = fmt.Sprintf("%s\n\t%s", msg, code)
			}
		}

		pass.Report(analysis.Diagnostic{
			Pos:     v.Pos,
			Message: msg,
		})
	}
}
