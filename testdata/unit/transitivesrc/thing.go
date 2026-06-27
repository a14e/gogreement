package transitivesrc

// Thing is immutable and lives in the deepest package of the chain.
// @immutable
// @constructor NewThing
type Thing struct {
	Field int
}

func NewThing() Thing {
	return Thing{}
}
