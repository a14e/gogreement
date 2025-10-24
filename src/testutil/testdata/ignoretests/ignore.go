// @ignore FILELEVEL
package ignoretests

// @ignore CODE1
func FunctionWithIgnore() {

	// This function should be ignored for CODE1
}

// @ignore CODE2, CODE3
type StructWithIgnore struct {
	Field int
}

// Regular comment
func RegularFunction() {
	// @ignore CODE4
	someStatement := 1
	_ = someStatement

	// @ignore CODE5, CODE6, CODE7
	anotherStatement := 2
	_ = anotherStatement
}

// @ignore
func InvalidIgnoreNoCode() {
	// This should not be parsed as it has no codes
}
