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

	"gogreement/src/config"
	"gogreement/src/testutil/testfacts"
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
	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)
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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)
	require.Equal(t, 3, ignoreSet.Len(), "expected 3 ignore annotations")
}

// TestReadIgnoreAnnotationsFromTestData tests reading ignore annotations from testdata
func TestReadIgnoreAnnotationsFromTestData(t *testing.T) {

	pass := testfacts.CreateTestPassWithFacts(t, "ignoretests")

	// Read ignore annotations
	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

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
	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

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

func TestReadIgnoreAnnotations_InlineComments(t *testing.T) {
	testCode := `package testpkg

func TestFunction() {
	x := 1 // @ignore CODE1
	y := 2 // @ignore CODE2, CODE3
	z := 3
	_ = x
	_ = y
	_ = z
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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

	// We expect 2 inline ignore annotations
	require.Equal(t, 2, ignoreSet.Len(), "expected 2 inline ignore annotations")

	// Get the function body statements
	funcDecl := file.Decls[0].(*ast.FuncDecl)
	stmts := funcDecl.Body.List

	// First statement: x := 1 // @ignore CODE1
	xAssign := stmts[0].(*ast.AssignStmt)
	assert.True(t, ignoreSet.Contains("CODE1", xAssign.Pos()),
		"CODE1 should cover x := 1")

	// Second statement: y := 2 // @ignore CODE2, CODE3
	yAssign := stmts[1].(*ast.AssignStmt)
	assert.True(t, ignoreSet.Contains("CODE2", yAssign.Pos()),
		"CODE2 should cover y := 2")
	assert.True(t, ignoreSet.Contains("CODE3", yAssign.Pos()),
		"CODE3 should cover y := 2")

	// Third statement: z := 3 (no @ignore)
	zAssign := stmts[2].(*ast.AssignStmt)
	assert.False(t, ignoreSet.Contains("CODE1", zAssign.Pos()),
		"CODE1 should NOT cover z := 3")
	assert.False(t, ignoreSet.Contains("CODE2", zAssign.Pos()),
		"CODE2 should NOT cover z := 3")
}

func TestReadIgnoreAnnotations_InlineVsBlock(t *testing.T) {
	testCode := `package testpkg

func TestFunction() {
	// @ignore BLOCK
	x := 1

	y := 2 // @ignore INLINE

	z := 3

	_ = x
	_ = y
	_ = z
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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

	// We expect 2 annotations: 1 block, 1 inline
	require.Equal(t, 2, ignoreSet.Len(), "expected 2 ignore annotations")

	// Get the function body statements
	funcDecl := file.Decls[0].(*ast.FuncDecl)
	stmts := funcDecl.Body.List

	// First statement: x := 1 (covered by BLOCK comment above)
	xAssign := stmts[0].(*ast.AssignStmt)
	assert.True(t, ignoreSet.Contains("BLOCK", xAssign.Pos()),
		"BLOCK should cover x := 1")

	// Second statement: y := 2 (covered by INLINE comment on same line)
	yAssign := stmts[1].(*ast.AssignStmt)
	assert.True(t, ignoreSet.Contains("INLINE", yAssign.Pos()),
		"INLINE should cover y := 2")

	// Third statement: z := 3 (not covered)
	zAssign := stmts[2].(*ast.AssignStmt)
	assert.False(t, ignoreSet.Contains("BLOCK", zAssign.Pos()),
		"BLOCK should NOT cover z := 3")
	assert.False(t, ignoreSet.Contains("INLINE", zAssign.Pos()),
		"INLINE should NOT cover z := 3")
}

func TestReadIgnoreAnnotations_InlineOnAssignment(t *testing.T) {
	testCode := `package testpkg

func TestFunction(u *User) {
	u.Name = "modified" // @ignore IMM01
	u.Age = 30          // @ignore IMM01
}

type User struct {
	Name string
	Age  int
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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

	// We expect 2 inline annotations
	require.Equal(t, 2, ignoreSet.Len(), "expected 2 inline ignore annotations")

	// Get the function body statements
	funcDecl := file.Decls[0].(*ast.FuncDecl)
	stmts := funcDecl.Body.List

	// First statement: u.Name = "modified"
	nameAssign := stmts[0].(*ast.AssignStmt)
	assert.True(t, ignoreSet.Contains("IMM01", nameAssign.Pos()),
		"IMM01 should cover u.Name assignment")

	// Second statement: u.Age = 30
	ageAssign := stmts[1].(*ast.AssignStmt)
	assert.True(t, ignoreSet.Contains("IMM01", ageAssign.Pos()),
		"IMM01 should cover u.Age assignment")
}

func TestReadIgnoreAnnotations_InlineDoesNotAffectNextLine(t *testing.T) {
	testCode := `package testpkg

func TestFunction(u *User) {
	var h1 User // @ignore CODE1
	_ = h1

	var h2 User // This should NOT be covered by CODE1
	_ = h2
}

type User struct {
	Name string
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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

	// We expect 1 inline annotation
	require.Equal(t, 1, ignoreSet.Len(), "expected 1 inline ignore annotation")

	// Get the function body statements
	funcDecl := file.Decls[0].(*ast.FuncDecl)
	stmts := funcDecl.Body.List

	// First statement: var h1 User // @ignore CODE1
	h1Decl := stmts[0].(*ast.DeclStmt)
	assert.True(t, ignoreSet.Contains("CODE1", h1Decl.Pos()),
		"CODE1 should cover h1 declaration")

	// Third statement: var h2 User (should NOT be covered)
	h2Decl := stmts[2].(*ast.DeclStmt)
	assert.False(t, ignoreSet.Contains("CODE1", h2Decl.Pos()),
		"CODE1 should NOT cover h2 declaration")
}

func TestReadIgnoreAnnotations_InlineForValueSpec(t *testing.T) {
	testCode := `package testpkg

import "somepkg"

func TestFunction() {
	var h1 somepkg.TestHelper // @ignore CODE1
	_ = h1

	var h2 somepkg.TestHelper // This should NOT be covered
	_ = h2
}

type TestHelper struct {
	Name string
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

	cfg := config.Empty()
	ignoreSet := ReadIgnoreAnnotations(cfg, pass)

	// We expect 1 inline annotation
	require.Equal(t, 1, ignoreSet.Len(), "expected 1 inline ignore annotation")

	// Get the function body statements - find the FuncDecl
	var funcDecl *ast.FuncDecl
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fd
			break
		}
	}
	require.NotNil(t, funcDecl, "should find function declaration")
	stmts := funcDecl.Body.List

	// First statement: var h1 somepkg.TestHelper // @ignore CODE1
	h1Decl := stmts[0].(*ast.DeclStmt)
	h1Spec := h1Decl.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	assert.True(t, ignoreSet.Contains("CODE1", h1Spec.Pos()),
		"CODE1 should cover h1 declaration at Pos()")

	// Third statement: var h2 somepkg.TestHelper (should NOT be covered)
	h2Decl := stmts[2].(*ast.DeclStmt)
	h2Spec := h2Decl.Decl.(*ast.GenDecl).Specs[0].(*ast.ValueSpec)
	assert.False(t, ignoreSet.Contains("CODE1", h2Spec.Pos()),
		"CODE1 should NOT cover h2 declaration")
}
