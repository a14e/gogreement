package interfacesforloading

// Reader is a simple interface for reading
type Reader interface {
	Read(p []byte) (n int, err error)
	Close() error
}

// Writer is a simple interface for writing
type Writer interface {
	Write(data []byte) (int, error)
}

// Processor has various parameter types
type Processor interface {
	Process(input string) string
	ProcessMany(items ...string) []string
	ProcessPointer(ptr *int) *string
}

// Empty interface with no methods
type Empty interface{}

// FileReader implements Reader
// @implements &Reader
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
type BufferWriter struct {
	buffer []byte
}

func (b *BufferWriter) Write(data []byte) (int, error) {
	b.buffer = append(b.buffer, data...)
	return len(data), nil
}

// StringProcessor implements Processor
// @implements &Processor
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
type EmptyImpl struct{}
