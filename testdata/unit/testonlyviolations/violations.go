package testonlyviolations

// @testonly
type TestHelper struct {
	Data string
}

// @testonly
func CreateMockData() string {
	return "mock"
}

// @testonly - this function calls another @testonly function, which is allowed
func CreateMockDataWrapper() string {
	return CreateMockData() // NOT a violation - recursive @testonly call
}

type Service struct {
	helper TestHelper // VIOLATION: using @testonly type in non-test file
}

func ProcessData() {
	data := CreateMockData() // VIOLATION: calling @testonly function in non-test file
	_ = data
}

func UseTestHelper() {
	h := TestHelper{Data: "test"} // VIOLATION: instantiating @testonly type
	_ = h
}

type Worker struct{}

// @testonly
func (w *Worker) Reset() {
	// test-only method
}

// @testonly - this method calls another @testonly method, which is allowed
func (w *Worker) ResetAll() {
	w.Reset() // NOT a violation - recursive @testonly call
}

func UseWorker() {
	w := &Worker{}
	w.Reset() // VIOLATION: calling @testonly method in non-test file
}

// @testonly
type MockCache struct {
	data map[string]string
}

func UseTypeAssertion(x any) {
	if c, ok := x.(*MockCache); ok { // VIOLATION: type assertion on @testonly type
		_ = c
	}
}

// ReceiverOnly is only ever used as a method receiver in this file. Declaring a
// method on a @testonly type is legitimate and must NOT be reported.
// @testonly
type ReceiverOnly struct {
	id int
}

func (r *ReceiverOnly) Helper() {} // NOT a violation: method declaration on a @testonly type

// TestID is a @testonly named type used to exercise conversion detection.
// @testonly
type TestID int

func ConvertToTestID(x int) {
	_ = TestID(x) // VIOLATION: conversion to a @testonly type (TONL01)
}

// MockNew exercises new(T) detection.
// @testonly
type MockNew struct{}

func BuildViaNew() {
	_ = new(MockNew) // VIOLATION: new() of a @testonly type (TONL01)
}

// MockElem exercises slice/make element-type detection.
// @testonly
type MockElem struct{}

func BuildSlice() {
	_ = []MockElem{} // VIOLATION: slice of a @testonly element type (TONL01)
}

// MockMake exercises make() element-type detection.
// @testonly
type MockMake struct{}

func BuildMake() {
	_ = make([]MockMake, 0) // VIOLATION: make() of a @testonly element type (TONL01)
}

func ShadowedFuncCall() {
	CreateMockData := func() string { return "local" } // shadows the @testonly function name
	_ = CreateMockData()                               // NOT a violation: resolves to the local
}

// Container is a generic type with a @testonly method.
type Container[T any] struct {
	items []T
}

// @testonly
func (c *Container[T]) DebugContainer() {}

func UseGenericTestOnlyMethod() {
	c := &Container[int]{}
	c.DebugContainer() // VIOLATION: @testonly method on a generic type (TONL03)
}

// GenericMock is a @testonly generic type (the type itself, not just a method).
// @testonly
type GenericMock[T any] struct {
	val T
}

func UseGenericTestOnlyType() {
	_ = GenericMock[int]{} // VIOLATION: @testonly generic type used (TONL01)
}
