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
