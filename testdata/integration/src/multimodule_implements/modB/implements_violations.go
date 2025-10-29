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

// === @ignore examples ===

// IgnoredPackageNotFoundType should not report IMPL01 due to @ignore
// @ignore IMPL01
// @implements nonexistent.Reader
type IgnoredPackageNotFoundType struct {
	data string
}

// IgnoredInterfaceNotFoundType should not report IMPL02 due to @ignore
// @ignore IMPL02
// @implements io.NonExistentInterface
type IgnoredInterfaceNotFoundType struct {
	value int
}

// IgnoredBadReader should not report IMPL03 due to @ignore
// @ignore IMPL03
// @implements io.Reader
type IgnoredBadReader struct {
	data string
}

// IgnoredBadModAReader should not report IMPL03 due to @ignore
// @ignore IMPL03
// @implements modA.Reader
type IgnoredBadModAReader struct {
	value int
}

// MultipleIgnoredType should not report any IMPL codes due to @ignore IMPL01
// @ignore IMPL01
// @implements nonexistent.Reader
type MultipleIgnoredType struct {
	data string
}

// AllIgnoredType should not report any errors due to @ignore ALL
// @ignore ALL
// @implements nonexistent.Reader
// @implements io.NonExistentInterface
type AllIgnoredType struct {
	data string
}
