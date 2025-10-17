package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests for valid annotations
func TestParseImplementsAnnotation_Simple(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_WithPointer(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     true,
		PackageName:   "",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements &Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_WithPackage(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "pkg",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements pkg.Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_WithPointerAndPackage(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     true,
		PackageName:   "pkg",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements &pkg.Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_WithPackageLongName(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     true,
		PackageName:   "mypackage123",
		InterfaceName: "MyInterface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements &mypackage123.MyInterface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

// Tests for whitespace
func TestParseImplementsAnnotation_ExtraSpacesAfterSlashes(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("//   @implements Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_LeadingSpaces(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("  // @implements Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_TrailingSpaces(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements Interface  ", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_MultipleSpacesInAnnotation(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     true,
		PackageName:   "pkg",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("//  @implements  &pkg.Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_TabsAndSpaces(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("//\t@implements\tInterface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

// Tests for invalid cases
func TestParseImplementsAnnotation_NoCommentPrefix(t *testing.T) {
	result := parseImplementsAnnotation("@implements Interface", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_TextBeforeAnnotation(t *testing.T) {
	result := parseImplementsAnnotation("// some text @implements Interface", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_TextAfterAnnotation(t *testing.T) {
	result := parseImplementsAnnotation("// @implements Interface text", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_BlockComment(t *testing.T) {
	result := parseImplementsAnnotation("/* @implements Interface */", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_EmptyString(t *testing.T) {
	result := parseImplementsAnnotation("", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_JustComment(t *testing.T) {
	result := parseImplementsAnnotation("// some comment", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_AnnotationWithoutInterface(t *testing.T) {
	result := parseImplementsAnnotation("// @implements", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_AnnotationWithOnlyPointer(t *testing.T) {
	result := parseImplementsAnnotation("// @implements &", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_AnnotationWithOnlyPackage(t *testing.T) {
	result := parseImplementsAnnotation("// @implements pkg.", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_ExampleInDescription(t *testing.T) {
	result := parseImplementsAnnotation(`// parse result of "@implements MyStruct" annotation`, "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_MultipleAnnotationsInOneLine(t *testing.T) {
	result := parseImplementsAnnotation("// @implements Interface @implements Another", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_WrongKeyword(t *testing.T) {
	result := parseImplementsAnnotation("// @implement Interface", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_SpecialCharactersInInterface(t *testing.T) {
	result := parseImplementsAnnotation("// @implements My-Interface", "MyStruct", 100)
	assert.Nil(t, result)
}

func TestParseImplementsAnnotation_DotWithoutPackage(t *testing.T) {
	result := parseImplementsAnnotation("// @implements .Interface", "MyStruct", 100)
	assert.Nil(t, result)
}

// Edge cases
func TestParseImplementsAnnotation_SingleLetterInterface(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "",
		InterfaceName: "I",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements I", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_SingleLetterPackage(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "a",
		InterfaceName: "Interface",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements a.Interface", "MyStruct", 100)

	assert.Equal(t, expected, result)
}

func TestParseImplementsAnnotation_NumbersInNames(t *testing.T) {
	expected := &ImplementsAnnotation{
		IsPointer:     false,
		PackageName:   "pkg123",
		InterfaceName: "Interface456",
		OnType:        "MyStruct",
		OnTypePos:     100,
	}

	result := parseImplementsAnnotation("// @implements pkg123.Interface456", "MyStruct", 100)

	assert.Equal(t, expected, result)
}
