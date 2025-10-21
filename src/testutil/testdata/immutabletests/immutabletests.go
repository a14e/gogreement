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
	p := &Person{}
	p.Name = Name // ✅ OK: in constructor
	p.Age = Age   // ✅ OK: in constructor
	return p
}

// UpdateName violates immutability - assigns to field
func UpdateName(p *Person, name string) {
	p.Name = name // ❌ VIOLATION
}

// IncrementAge violates immutability - uses ++
func IncrementAge(p *Person) {
	p.Age++ // ❌ VIOLATION
}

// ModifyItem violates immutability - modifies slice element
func ModifyItem(p *Person, index int, value string) {
	p.Items[index] = value // ❌ VIOLATION
}

// Config with multiple constructors
// @immutable
// @constructor NewConfig, NewDefaultConfig
type Config struct {
	host string
	port int
}

func NewConfig(host string, port int) *Config {
	c := &Config{}
	c.host = host // ✅ OK: in constructor
	c.port = port // ✅ OK: in constructor
	return c
}

func NewDefaultConfig() *Config {
	c := &Config{}
	c.host = "localhost" // ✅ OK: in constructor
	c.port = 8080        // ✅ OK: in constructor
	return c
}

// Counter with various operations
// @immutable
// @constructor NewCounter
type Counter struct {
	value int
	step  int
}

func NewCounter() *Counter {
	c := &Counter{}
	c.value = 0 // ✅ OK: in constructor
	c.step = 1  // ✅ OK: in constructor
	return c
}

func Increment(c *Counter) {
	c.value++ // ❌ VIOLATION
}

func Decrement(c *Counter) {
	c.value-- // ❌ VIOLATION
}

func ChangeStep(c *Counter, delta int) {
	c.step += delta // ❌ VIOLATION
}

func MultiplyStep(c *Counter, factor int) {
	c.step *= factor // ❌ VIOLATION
}

func DivideStep(c *Counter, divisor int) {
	c.step /= divisor // ❌ VIOLATION
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
}

// ImportedTypeWrapper uses immutable type from another package
type ImportedTypeWrapper struct {
	reader *interfacesforloading.FileReader
	value  int
}

func NewImportedTypeWrapper() *ImportedTypeWrapper {
	return &ImportedTypeWrapper{
		reader: &interfacesforloading.FileReader{},
		value:  0,
	}
}

// MutableType has no @immutable annotation - should not report violations
type MutableType struct {
	counter int
}

func MutateMutableType(m *MutableType) {
	m.counter++ // ✅ OK: not immutable
}

// TryToMutateImported tries to mutate imported immutable type
func TryToMutateImported(w *ImportedTypeWrapper) {
	w.reader.Data = []byte{1, 2, 3} // ❌ VIOLATION: FileReader is immutable
}
