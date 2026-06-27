package transitivetop

import "github.com/a14e/gogreement/testdata/unit/transitivemid"

// Use mutates a transitivesrc.Thing obtained via transitivemid. transitivetop
// imports only transitivemid (not transitivesrc directly), so detecting this
// requires loading the @immutable facts of a transitive dependency.
func Use() {
	x := transitivemid.Get()
	x.Field = 5 // VIOLATION: mutating an @immutable type reached transitively
}
