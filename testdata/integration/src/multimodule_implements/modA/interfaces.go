package modA // want package:"package modA"

// Blank import to ensure the package is loaded for cross-module analysis
import _ "io"

// Reader is an interface from modA
type Reader interface {
	Read(p []byte) (n int, err error)
}

// Writer is an interface from modA
type Writer interface {
	Write(p []byte) (n int, err error)
}

// MyReader implements io.Reader with pointer receiver
// @implements &io.Reader
type MyReader struct {
	data []byte
}

func (r *MyReader) Read(p []byte) (n int, err error) {
	return 0, nil
}
