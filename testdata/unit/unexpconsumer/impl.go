package unexpconsumer

import "github.com/a14e/gogreement/testdata/unit/unexpsrc"

var _ = func(r unexpsrc.Reader) {} // keeps the import used

// Impl declares its own read() in this package, which does NOT satisfy
// unexpsrc.Reader: the interface's unexported read() belongs to unexpsrc, so
// only same-package types can implement it. This must be reported as missing.
// @implements unexpsrc.Reader
type Impl struct{}

func (Impl) read() int { return 0 }
