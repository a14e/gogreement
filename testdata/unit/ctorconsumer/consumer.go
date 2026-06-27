package ctorconsumer

import "github.com/a14e/gogreement/testdata/unit/ctorsource"

// NewWidget collides in name with ctorsource's Widget constructor, but it lives
// in a different package, so instantiating the external @constructor type here
// must still be flagged (the constructor lives in ctorsource, not here).
func NewWidget() ctorsource.Widget {
	return ctorsource.Widget{Value: 1} // ❌ VIOLATION: cross-package instantiation
}
