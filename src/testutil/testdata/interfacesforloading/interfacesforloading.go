package interfacesforloading

// Reader is a simple interface for reading
// @immutable
type Reader interface {
	Read(p []byte) (n int, err error)
	Close() error
}

// Writer is a simple interface for writing
// @immutable
type Writer interface {
	Write(data []byte) (int, error)
}

// Processor has various parameter types
// @immutable
type Processor interface {
	Process(input string) string
	ProcessMany(items ...string) []string
	ProcessPointer(ptr *int) *string
}

// Empty interface with no methods
// @immutable
type Empty interface{}

// FileReader implements Reader
// @implements &Reader
// @immutable
type FileReader struct {
	data []byte
}

func (f *FileReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (f *FileReader) Close() error {
	return nil
}

// BufferWriter implements Writer
// @implements &Writer
// @immutable
type BufferWriter struct {
	buffer []byte
}

func (b *BufferWriter) Write(data []byte) (int, error) {
	b.buffer = append(b.buffer, data...)
	return len(data), nil
}

// StringProcessor implements Processor
// @implements &Processor
// @immutable
type StringProcessor struct{}

func (s *StringProcessor) Process(input string) string {
	return input
}

func (s *StringProcessor) ProcessMany(items ...string) []string {
	return items
}

func (s *StringProcessor) ProcessPointer(ptr *int) *string {
	result := "processed"
	return &result
}

// EmptyImpl implements Empty interface
// @implements &Empty
// @immutable
type EmptyImpl struct{}

// Config is a configuration without annotations (for negative tests)
type Config struct {
	name string
}

// MutableType has no @immutable annotation
type MutableType struct {
	counter int
}
