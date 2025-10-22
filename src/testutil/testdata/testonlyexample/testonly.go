package testonlyexample

// @testonly
type TestHelper struct {
	Name string
}

// @testonly
func CreateMockData() []string {
	return []string{"test1", "test2"}
}

type MyService struct {
	data []string
}

// @testonly
func (s *MyService) Reset() {
	s.data = nil
}

// @testonly
func (s MyService) GetTestData() []string {
	return s.data
}

// Regular function without annotation
func ProcessData(data []string) int {
	return len(data)
}

// Regular method without annotation
func (s *MyService) Process() {
	// process data
}
