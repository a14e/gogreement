package annotationedgecases

// Grouped declaration: the group-level @immutable must apply to the spec even
// though the spec carries its own doc comment.
// @immutable
type (
	// GroupedWithDoc has its own doc comment in addition to the group annotation.
	GroupedWithDoc struct {
		Name string
	}
)

/* @immutable */
type BlockCommented struct {
	Value int
}

// The same annotation on both the group and the spec must be counted once.
// @immutable
type (
	// @immutable
	DoubleAnnotated struct {
		Name string
	}
)

// Stack is a generic type carrying a @testonly method, used to verify that
// receiver-type extraction unwraps the generic instantiation to the base name.
type Stack[T any] struct {
	items []T
}

// @testonly
func (s *Stack[T]) DebugDump() {}

// Pair is a generic type with multiple type parameters.
type Pair[K comparable, V any] struct {
	key   K
	value V
}

// @testonly
func (p *Pair[K, V]) DebugPair() {}
