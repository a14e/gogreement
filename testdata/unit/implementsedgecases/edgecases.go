package implementsedgecases

// Reader is satisfied by a single Foo() method.
type Reader interface {
	Foo()
}

type inner struct{}

func (i *inner) Foo() {}

// Outer embeds *inner, so the value method set of Outer includes Foo (promoted
// through the embedded pointer). The value-form @implements must be satisfied,
// i.e. this must NOT be reported as missing.
// @implements Reader
type Outer struct {
	*inner
}

// Box is a generic type.
type Box[T any] struct {
	v T
}

// IntBoxer requires Get returning Box[int].
type IntBoxer interface {
	Get() Box[int]
}

// StringBoxImpl returns Box[string], so it does NOT implement IntBoxer. The
// generic type argument must be taken into account.
// @implements IntBoxer
type StringBoxImpl struct{}

func (StringBoxImpl) Get() Box[string] {
	return Box[string]{}
}

// DoublePtr requires a parameter of type **int.
type DoublePtr interface {
	Take(**int)
}

// SinglePtrImpl takes *int, so it does NOT implement DoublePtr — pointer depth
// must be significant.
// @implements DoublePtr
type SinglePtrImpl struct{}

func (SinglePtrImpl) Take(*int) {}
