package withimports

import (
	_ "context"
	_ "io"
	"time"
)

// Existing struct types...
// @implements &io.Reader
type MyReader struct{}

func (m *MyReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

// @implements &io.Writer
// @implements &io.Closer
type MyWriteCloser struct{}

func (m *MyWriteCloser) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (m *MyWriteCloser) Close() error {
	return nil
}

// @implements &context.Context
type MyContext struct{}

func (m *MyContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (m *MyContext) Done() <-chan struct{} {
	return nil
}

func (m *MyContext) Err() error {
	return nil
}

func (m *MyContext) Value(key interface{}) interface{} {
	return nil
}

// NEW: Named types (not structs)

// Duration is like time.Duration
type Duration int64

func (d Duration) Seconds() float64 {
	return float64(d) / 1e9
}

func (d Duration) String() string {
	return "duration"
}

// MyString demonstrates methods on string alias
type MyString string

func (s MyString) Upper() MyString {
	// simplified
	return s
}

func (s *MyString) Append(suffix string) {
	*s = MyString(string(*s) + suffix)
}

// HandlerFunc demonstrates methods on function type
type HandlerFunc func(string) error

func (f HandlerFunc) ServeHTTP(path string) error {
	return f(path)
}

// ByteSlice demonstrates methods on slice type
type ByteSlice []byte

func (b ByteSlice) Len() int {
	return len(b)
}

func (b *ByteSlice) Append(data byte) {
	*b = append(*b, data)
}
