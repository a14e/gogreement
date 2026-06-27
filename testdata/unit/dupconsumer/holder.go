package dupconsumer

import (
	"github.com/a14e/gogreement/testdata/unit/dupsrca"
	"github.com/a14e/gogreement/testdata/unit/dupsrcb"
)

// Holder uses two equally named @testonly types from different packages. Both
// must be reported — per-file dedup must not collapse them by bare name.
type Holder struct {
	a dupsrca.Dup // VIOLATION
	b dupsrcb.Dup // VIOLATION
}
