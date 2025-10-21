package immutabletests

import (
	"goagreement/src/testutil/testdata/interfacesforloading"
)

// Person is an immutable type
// @immutable
// @constructor NewPerson
type Person struct {
	Name  string
	Age   int
	Items []string
}

func NewPerson(Name string, Age int) *Person {
	return &Person{}
}

// ComplexCase with nested operations
// @immutable
// @constructor NewComplexCase
type ComplexCase struct {
	nested *Person
	count  int
}

func NewComplexCase() *ComplexCase {
	c := &ComplexCase{}
	c.nested = NewPerson("test", 30) // ✅ OK: in constructor
	c.count = 0                      // ✅ OK: in constructor
	return c
}

func ModifyNested(c *ComplexCase) {
	c.count++ // ❌ VIOLATION: modifying ComplexCase
	// Note: c.nested.Name = "x" would be caught as violation on Person, not ComplexCase
}

// ImportedTypeWrapper uses immutable type from another package
type ImportedTypeWrapper struct {
	reader *interfacesforloading.FileReader
	value  int
}

// NewImportedTypeWrapper creates wrapper
func NewImportedTypeWrapper() *ImportedTypeWrapper {
	return &ImportedTypeWrapper{
		reader: &interfacesforloading.FileReader{},
		value:  0,
	}
}

// TryToMutateImported tries to mutate imported immutable type - should fail
func TryToMutateImported(w *ImportedTypeWrapper) {
	// This should be caught as violation on FileReader
	// w.reader.data = []byte{} // ❌ VIOLATION: FileReader is immutable
}
