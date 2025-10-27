package immutable

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"gogreement/src/util"
)

// ImmutableViolation represents a mutation of an immutable type
// @immutable
type ImmutableViolation struct {
	TypeName string
	Reason   string
	Code     string // Error code from codes package
	Pos      token.Pos
	Node     ast.Node
}

func ReportViolations(pass *analysis.Pass, violations []ImmutableViolation, ignoreSet *util.IgnoreSet) {
	for _, v := range violations {
		// Check if this violation should be ignored
		if ignoreSet != nil && ignoreSet.Contains(v.Code, v.Pos) {
			continue
		}

		msg := fmt.Sprintf("[%s] immutability violation in type %q: %s", v.Code, v.TypeName, v.Reason)

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
