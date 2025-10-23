package modB // want package:"package modB"

import (
	// Blank import for io package
	_ "io"

	// Import modA to access its types
	"multimodule_implements/modA"
)

// BadReader claims to implement io.Reader but doesn't
// @implements io.Reader
type BadReader struct { // want "does not implement interface.*io.Reader"
	data string
}

// GoodReader properly implements io.Reader with pointer receiver
// @implements &io.Reader
type GoodReader struct {
	data []byte
}

func (r *GoodReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

// UseModA uses modA.Reader to avoid "imported and not used" error
var _ modA.Reader
