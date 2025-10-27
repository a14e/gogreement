package immutabletests

import (
	"gogreement/testdata/unit/interfacesforloading"
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

// Methods that reassign receiver

// Reset tries to reassign the receiver (should be violation)
func (p *Person) Reset() {
	*p = Person{} // ❌ VIOLATION: cannot reassign immutable receiver
}

// UpdateCounter tries to reassign Counter receiver
func (c *Counter) UpdateCounter(value, step int) {
	*c = Counter{value: value, step: step} // ❌ VIOLATION
}

// MutableTypeReset is OK since MutableType is not immutable
func (m *MutableType) Reset() {
	*m = MutableType{} // ✅ OK: not immutable
}

// Test for primitive type aliases

// ImmutableInt is an immutable integer type
// @immutable
// @constructor NewImmutableInt
type ImmutableInt int

func NewImmutableInt(value int) ImmutableInt {
	var i ImmutableInt
	i = ImmutableInt(value) // ✅ OK: in constructor
	return i
}

// SetValue tries to reassign receiver
func (i *ImmutableInt) SetValue(value int) {
	*i = ImmutableInt(value) // ❌ VIOLATION: cannot reassign immutable receiver
}

// Increment tries to modify receiver
func (i *ImmutableInt) Increment() {
	*i++ // ❌ VIOLATION: cannot reassign immutable receiver
}

// ImmutableString is an immutable string type
// @immutable
// @constructor NewImmutableString
type ImmutableString string

func NewImmutableString(value string) ImmutableString {
	var s ImmutableString
	s = ImmutableString(value) // ✅ OK: in constructor
	return s
}

// Update tries to reassign receiver
func (s *ImmutableString) Update(value string) {
	*s = ImmutableString(value) // ❌ VIOLATION
}

// Test for map field modifications

// ConfigWithMap has a map field
// @immutable
// @constructor NewConfigWithMap
type ConfigWithMap struct {
	settings map[string]string
	values   map[int]int
}

func NewConfigWithMap() *ConfigWithMap {
	c := &ConfigWithMap{}
	c.settings = make(map[string]string) // ✅ OK: in constructor
	c.settings["key"] = "value"          // ✅ OK: in constructor
	return c
}

// ModifyMapString tries to modify map field
func ModifyMapString(c *ConfigWithMap, key, value string) {
	c.settings[key] = value // ❌ VIOLATION: modifying map element
}

// ModifyMapInt tries to modify map field
func ModifyMapInt(c *ConfigWithMap, key, value int) {
	c.values[key] = value // ❌ VIOLATION: modifying map element
}

// DeleteFromMap tries to delete from map
func DeleteFromMap(c *ConfigWithMap, key string) {
	delete(c.settings, key) // This is a CallExpr, not checked by current implementation
}
