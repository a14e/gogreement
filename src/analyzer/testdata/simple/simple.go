package simple

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
