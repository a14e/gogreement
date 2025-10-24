package ignore

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"

	"goagreement/src/testutil"
	"goagreement/src/testutil/testfacts"
)

func TestParseIgnoreAnnotation(t *testing.T) {
	tests := []struct {
		name          string
		comment       string
		expectNil     bool
		expectedCodes []string
	}{
		{
			name:          "single code",
			comment:       "// @ignore CODE1",
			expectNil:     false,
			expectedCodes: []string{"CODE1"},
		},
		{
			name:          "multiple codes",
			comment:       "// @ignore CODE1, CODE2",
			expectNil:     false,
			expectedCodes: []string{"CODE1", "CODE2"},
		},
		{
			name:          "multiple codes with spaces",
			comment:       "// @ignore CODE1 , CODE2 , CODE3",
			expectNil:     false,
			expectedCodes: []string{"CODE1", "CODE2", "CODE3"},
		},
		{
			name:          "no codes provided",
			comment:       "// @ignore",
			expectNil:     true,
			expectedCodes: nil,
		},
		{
			name:          "no codes with whitespace",
			comment:       "// @ignore   ",
			expectNil:     true,
			expectedCodes: nil,
		},
		{
			name:          "not an ignore comment",
			comment:       "// @implements MyInterface",
			expectNil:     true,
			expectedCodes: nil,
		},
		{
			name:          "regular comment",
			comment:       "// some comment",
			expectNil:     true,
			expectedCodes: nil,
		},
		{
			name:          "ignore with extra whitespace",
			comment:       "//   @ignore   CODE1,CODE2",
			expectNil:     false,
			expectedCodes: []string{"CODE1", "CODE2"},
		},
		{
			name:          "single code with trailing comma",
			comment:       "// @ignore CODE1,",
			expectNil:     false,
			expectedCodes: []string{"CODE1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseIgnoreAnnotation(tt.comment, token.Pos(1), token.Pos(10))

			if tt.expectNil {
				assert.Nil(t, result, "expected nil result")
			} else {
				require.NotNil(t, result, "expected non-nil result")
				assert.Equal(t, tt.expectedCodes, result.Codes, "codes mismatch")
				assert.Equal(t, token.Pos(1), result.StartPos, "start position mismatch")
				assert.Equal(t, token.Pos(10), result.EndPos, "end position mismatch")
			}
		})
	}
}

func TestReadIgnoreAnnotations(t *testing.T) {
	testCode := `package testpkg

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
}

// @ignore
func InvalidIgnoreNoCode() {
	// This should not be parsed
}
`

	// Create a temporary test pass
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	// Create a minimal analysis.Pass for testing
	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage("testpkg", "testpkg"),
	}

	// Read annotations
	ignoreSet := ReadIgnoreAnnotations(pass)

	// We expect 3 valid @ignore annotations (CODE1, CODE2+CODE3, CODE4)
	// The one without codes should be filtered out
	require.Equal(t, 3, ignoreSet.Len(), "expected 3 ignore annotations")

	// Verify annotations work through Contains() method
	// CODE1 should be found at the function position
	// CODE2, CODE3 should be found at the struct position
	// CODE4 should be found inside the function body
	// We can't verify exact positions without accessing markers field
}

func TestReadIgnoreAnnotations_EmptyFile(t *testing.T) {
	testCode := `package testpkg

func EmptyFunction() {
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage("testpkg", "testpkg"),
	}

	ignoreSet := ReadIgnoreAnnotations(pass)
	assert.Equal(t, 0, ignoreSet.Len(), "expected no annotations in empty file")
}

func TestReadIgnoreAnnotations_MultipleIgnoresInBlock(t *testing.T) {
	testCode := `package testpkg

func ComplexFunction() {
	// @ignore ERR1
	statement1 := 1

	// @ignore ERR2
	statement2 := 2

	// @ignore ERR3, ERR4
	statement3 := 3

	_ = statement1
	_ = statement2
	_ = statement3
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage("testpkg", "testpkg"),
	}

	ignoreSet := ReadIgnoreAnnotations(pass)
	require.Equal(t, 3, ignoreSet.Len(), "expected 3 ignore annotations")
}

// TestReadIgnoreAnnotationsFromTestData tests reading ignore annotations from testdata
func TestReadIgnoreAnnotationsFromTestData(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	pass := testfacts.CreateTestPassWithFacts(t, "ignoretests")

	// Read ignore annotations
	ignoreSet := ReadIgnoreAnnotations(pass)

	t.Logf("Found %d ignore annotations", ignoreSet.Len())

	// We expect 5 valid @ignore annotations:
	// 0. FILELEVEL (file-level)
	// 1. CODE1 (on function)
	// 2. CODE2, CODE3 (on struct)
	// 3. CODE4 (in function body)
	// 4. CODE5, CODE6, CODE7 (in function body)
	require.Equal(t, 5, ignoreSet.Len(), "expected 5 ignore annotations")

	// Verify we have the markers
	require.Len(t, ignoreSet.Markers, 5, "expected 5 markers")

	// Check file-level annotation (first one)
	fileLevelMarker := ignoreSet.Markers[0]
	assert.Equal(t, []string{"FILELEVEL"}, fileLevelMarker.Codes, "first annotation should be FILELEVEL")

	// Check other annotations
	assert.Equal(t, []string{"CODE1"}, ignoreSet.Markers[1].Codes)
	assert.Equal(t, []string{"CODE2", "CODE3"}, ignoreSet.Markers[2].Codes)
	assert.Equal(t, []string{"CODE4"}, ignoreSet.Markers[3].Codes)
	assert.Equal(t, []string{"CODE5", "CODE6", "CODE7"}, ignoreSet.Markers[4].Codes)

	// Get file from pass to test positions
	require.Len(t, pass.Files, 1, "expected 1 file")
	file := pass.Files[0]

	// Test that FILELEVEL covers all declarations
	for i, decl := range file.Decls {
		assert.True(t, ignoreSet.Contains("FILELEVEL", decl.Pos()),
			"FILELEVEL should cover declaration %d", i)
	}

	// Test that FILELEVEL covers end of file
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.End()),
		"FILELEVEL should cover end of file")

	// Test that specific codes still work at their locations
	// CODE1 should be at function position
	funcDecl := file.Decls[0].(*ast.FuncDecl)
	assert.True(t, ignoreSet.Contains("CODE1", funcDecl.Pos()),
		"CODE1 should be ignored at FunctionWithIgnore")

	// But CODE1 should NOT be ignored at other declarations (only FILELEVEL should)
	structDecl := file.Decls[1]
	assert.False(t, ignoreSet.Contains("CODE1", structDecl.Pos()),
		"CODE1 should NOT be ignored at StructWithIgnore")
	assert.True(t, ignoreSet.Contains("FILELEVEL", structDecl.Pos()),
		"FILELEVEL should be ignored at StructWithIgnore")
}

func TestReadIgnoreAnnotations_FileLevel(t *testing.T) {
	testCode := `// @ignore FILELEVEL
package testpkg

func Function1() {
	x := 1
	_ = x
}

type MyStruct struct {
	Field int
}

func Function2() {
	y := 2
	_ = y
}
`

	// Create a temporary test pass
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	// Create a minimal analysis.Pass for testing
	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage("testpkg", "testpkg"),
	}

	// Read annotations
	ignoreSet := ReadIgnoreAnnotations(pass)

	// We expect 1 file-level @ignore annotation
	require.Equal(t, 1, ignoreSet.Len(), "expected 1 file-level ignore annotation")

	// Get the marker
	require.Len(t, ignoreSet.Markers, 1, "expected 1 marker")
	marker := ignoreSet.Markers[0]

	// Verify codes
	assert.Equal(t, []string{"FILELEVEL"}, marker.Codes)

	// File-level annotation should cover the entire file
	// Let's test that FILELEVEL code is ignored at various positions in the file

	// Test at the beginning of Function1
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.Decls[0].Pos()),
		"FILELEVEL should be ignored at Function1")

	// Test at MyStruct
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.Decls[1].Pos()),
		"FILELEVEL should be ignored at MyStruct")

	// Test at Function2
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.Decls[2].Pos()),
		"FILELEVEL should be ignored at Function2")

	// Test at the end of the file
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.End()),
		"FILELEVEL should be ignored at end of file")
}

func TestReadIgnoreAnnotations_FileLevelWithOtherAnnotations(t *testing.T) {
	testCode := `// @ignore FILELEVEL
package testpkg

// @ignore SPECIFIC
func SpecificFunction() {
	x := 1
	_ = x
}

func RegularFunction() {
	y := 2
	_ = y
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage("testpkg", "testpkg"),
	}

	ignoreSet := ReadIgnoreAnnotations(pass)

	// We expect 2 annotations: file-level + specific function
	require.Equal(t, 2, ignoreSet.Len(), "expected 2 ignore annotations")

	// Test that FILELEVEL covers everything
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.Decls[0].Pos()))
	assert.True(t, ignoreSet.Contains("FILELEVEL", file.Decls[1].Pos()))

	// Test that SPECIFIC only covers SpecificFunction
	specificFuncPos := file.Decls[0].Pos()
	regularFuncPos := file.Decls[1].Pos()

	assert.True(t, ignoreSet.Contains("SPECIFIC", specificFuncPos),
		"SPECIFIC should be ignored at SpecificFunction")

	// SPECIFIC should NOT cover RegularFunction (but FILELEVEL should)
	assert.False(t, ignoreSet.Contains("SPECIFIC", regularFuncPos),
		"SPECIFIC should NOT be ignored at RegularFunction")
	assert.True(t, ignoreSet.Contains("FILELEVEL", regularFuncPos),
		"FILELEVEL should be ignored at RegularFunction")
}

func TestReadIgnoreAnnotations_NotFileLevel(t *testing.T) {
	testCode := `package testpkg

// @ignore NOTFILELEVEL
func Function1() {
	x := 1
	_ = x
}

func Function2() {
	y := 2
	_ = y
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	pass := &analysis.Pass{
		Fset:  fset,
		Files: []*ast.File{file},
		Pkg:   types.NewPackage("testpkg", "testpkg"),
	}

	ignoreSet := ReadIgnoreAnnotations(pass)

	// We expect 1 function-level annotation (not file-level)
	require.Equal(t, 1, ignoreSet.Len(), "expected 1 function-level ignore annotation")

	// NOTFILELEVEL should only cover Function1, not Function2
	func1Pos := file.Decls[0].Pos()
	func2Pos := file.Decls[1].Pos()

	assert.True(t, ignoreSet.Contains("NOTFILELEVEL", func1Pos),
		"NOTFILELEVEL should be ignored at Function1")
	assert.False(t, ignoreSet.Contains("NOTFILELEVEL", func2Pos),
		"NOTFILELEVEL should NOT be ignored at Function2")
}
