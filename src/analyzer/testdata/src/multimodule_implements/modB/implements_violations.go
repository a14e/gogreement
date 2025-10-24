package modB // want package:"package modB"

import (
	// Blank imports to ensure packages are loaded for cross-module analysis
	_ "io"
	_ "multimodule_implements/modA"
)

// PackageNotFoundType references non-imported package
// @implements nonexistent.Reader
type PackageNotFoundType struct { // want "\\[IMPL01\\].*package \"nonexistent\" referenced in @implements.*not imported"
	data string
}

// InterfaceNotFoundType references non-existent interface from io
// @implements io.NonExistentInterface
type InterfaceNotFoundType struct { // want "\\[IMPL02\\].*interface \"io.NonExistentInterface\" not found"
	value int
}

// BadReader claims to implement io.Reader but doesn't
// @implements io.Reader
type BadReader struct { // want "\\[IMPL03\\].*does not implement interface.*io.Reader"
	data string
}

// BadModAReader claims to implement modA.Reader but doesn't
// @implements modA.Reader
type BadModAReader struct { // want "\\[IMPL03\\].*does not implement interface.*modA.Reader"
	value int
}

// GoodModAReader properly implements modA.Reader with pointer receiver
// @implements &modA.Reader
type GoodModAReader struct {
	buffer []byte
}

func (r *GoodModAReader) Read(p []byte) (n int, err error) {
	return 0, nil
}

// GoodReader properly implements io.Reader with pointer receiver
// @implements &io.Reader
type GoodReader struct {
	data []byte
}

func (r *GoodReader) Read(p []byte) (n int, err error) {
	return 0, nil
}
