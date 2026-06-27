package transitivemid

import "github.com/a14e/gogreement/testdata/unit/transitivesrc"

// Get returns a transitivesrc.Thing, exposing the immutable type to callers
// that import transitivemid without importing transitivesrc directly.
func Get() transitivesrc.Thing {
	return transitivesrc.NewThing()
}
