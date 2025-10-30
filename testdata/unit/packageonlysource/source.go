package packageonlysource

// Source package with @packageonly annotations

// PackageOnlyType can only be used in allowed packages
// @packageonly allowedpkg, packageonlyallowed
type PackageOnlyType struct {
	value int
}

func (t *PackageOnlyType) Method() {
	t.value++
}

// PackageOnlyFunction can only be called from allowed packages
// @packageonly allowedpkg, packageonlyallowed
func PackageOnlyFunction() int {
	return 42
}

// PackageOnlyStruct with packageonly method
// @packageonly allowedpkg, packageonlyallowed
type PackageOnlyStruct struct {
	data string
}

// PackageOnlyMethod can only be called from allowed packages
// @packageonly allowedpkg, packageonlyallowed
func (s *PackageOnlyStruct) PackageOnlyMethod() string {
	return s.data
}

// Regular items without @packageonly restrictions
type RegularType struct {
	value int
}

func (t *RegularType) Method() {
	t.value++
}

func RegularFunction() int {
	return 100
}

type RegularStruct struct {
	data string
}

func (s *RegularStruct) RegularMethod() string {
	return s.data
}
