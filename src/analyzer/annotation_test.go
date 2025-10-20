package analyzer

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"goagreement/src/util"
)

func TestParseImplementsAnnotation(t *testing.T) {
	// Create mock import map
	imports := &importmap.ImportMap{}
	imports.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"io"`},
	})
	imports.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"context"`},
	})

	currentPkgPath := "mypackage/path"

	tests := []struct {
		name          string
		comment       string
		typeName      string
		expectNil     bool
		expectedAnnot *ImplementsAnnotation
	}{
		{
			name:      "simple interface",
			comment:   "// @implements MyInterface",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:          "MyStruct",
				InterfaceName:   "MyInterface",
				PackageName:     "",
				IsPointer:       false,
				PackageFullPath: currentPkgPath,
				PackageNotFound: false,
			},
		},
		{
			name:      "pointer interface",
			comment:   "// @implements &MyInterface",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:          "MyStruct",
				InterfaceName:   "MyInterface",
				PackageName:     "",
				IsPointer:       true,
				PackageFullPath: currentPkgPath,
				PackageNotFound: false,
			},
		},
		{
			name:      "package qualified",
			comment:   "// @implements io.Reader",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:          "MyStruct",
				InterfaceName:   "Reader",
				PackageName:     "io",
				IsPointer:       false,
				PackageFullPath: "io",
				PackageNotFound: false,
			},
		},
		{
			name:      "pointer with package",
			comment:   "// @implements &io.Reader",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:          "MyStruct",
				InterfaceName:   "Reader",
				PackageName:     "io",
				IsPointer:       true,
				PackageFullPath: "io",
				PackageNotFound: false,
			},
		},
		{
			name:      "package not found",
			comment:   "// @implements http.Handler",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:          "MyStruct",
				InterfaceName:   "Handler",
				PackageName:     "http",
				IsPointer:       false,
				PackageFullPath: "",
				PackageNotFound: true,
			},
		},
		{
			name:      "with extra text before",
			comment:   "// text before @implements &io.Reader",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "with extra text after before",
			comment:   "//  @implements &io.Reader text after",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "with extra spaces",
			comment:   "//   @implements   &io.Reader   ",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:          "MyStruct",
				InterfaceName:   "Reader",
				PackageName:     "io",
				IsPointer:       true,
				PackageFullPath: "io",
				PackageNotFound: false,
			},
		},
		{
			name:      "invalid format - no interface",
			comment:   "// @implements",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "not an annotation",
			comment:   "// This is a regular comment",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "wrong annotation",
			comment:   "// @deprecated",
			typeName:  "MyStruct",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseImplementsAnnotation(tt.comment, tt.typeName, 0, imports, currentPkgPath)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedAnnot.OnType, result.OnType)
				assert.Equal(t, tt.expectedAnnot.InterfaceName, result.InterfaceName)
				assert.Equal(t, tt.expectedAnnot.PackageName, result.PackageName)
				assert.Equal(t, tt.expectedAnnot.IsPointer, result.IsPointer)
				assert.Equal(t, tt.expectedAnnot.PackageFullPath, result.PackageFullPath)
				assert.Equal(t, tt.expectedAnnot.PackageNotFound, result.PackageNotFound)
			}
		})
	}
}

func TestParseConstructorAnnotation(t *testing.T) {
	tests := []struct {
		name          string
		comment       string
		typeName      string
		expectNil     bool
		expectedNames []string
	}{
		{
			name:          "single constructor",
			comment:       "// @constructor New",
			typeName:      "MyStruct",
			expectNil:     false,
			expectedNames: []string{"New"},
		},
		{
			name:          "multiple constructors",
			comment:       "// @constructor New, Create",
			typeName:      "MyStruct",
			expectNil:     false,
			expectedNames: []string{"New", "Create"},
		},
		{
			name:          "three constructors",
			comment:       "// @constructor New, Create, Build",
			typeName:      "MyStruct",
			expectNil:     false,
			expectedNames: []string{"New", "Create", "Build"},
		},
		{
			name:      "no constructor name - should return nil",
			comment:   "// @constructor",
			typeName:  "User",
			expectNil: true,
		},
		{
			name:          "with extra spaces",
			comment:       "//   @constructor   New  ,  Create   ",
			typeName:      "MyStruct",
			expectNil:     false,
			expectedNames: []string{"New", "Create"},
		},
		{
			name:          "with trailing comma",
			comment:       "// @constructor New, Create,",
			typeName:      "MyStruct",
			expectNil:     false,
			expectedNames: []string{"New", "Create"},
		},
		{
			name:      "only commas - should return nil",
			comment:   "// @constructor , , ,",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "not an annotation",
			comment:   "// This is a regular comment",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "wrong annotation",
			comment:   "// @implements Something",
			typeName:  "MyStruct",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConstructorAnnotation(tt.comment, tt.typeName, 0)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.typeName, result.OnType)
				assert.Equal(t, tt.expectedNames, result.ConstructorNames)
			}
		})
	}
}

func TestReadAllAnnotations(t *testing.T) {
	pass := createTestPass(t, "withimports")

	annotations := ReadAllAnnotations(pass)

	require.NotEmpty(t, annotations.ImplementsAnnotations, "expected to find implements annotations")

	// Test @implements annotations
	t.Run("ImplementsAnnotations", func(t *testing.T) {
		// Helper to find annotation
		findImplements := func(onType, interfaceName string) *ImplementsAnnotation {
			for _, a := range annotations.ImplementsAnnotations {
				if a.OnType == onType && a.InterfaceName == interfaceName {
					return &a
				}
			}
			return nil
		}

		t.Run("MyReader implements io.Reader", func(t *testing.T) {
			annot := findImplements("MyReader", "Reader")
			require.NotNil(t, annot, "annotation not found")

			assert.Equal(t, "MyReader", annot.OnType)
			assert.Equal(t, "Reader", annot.InterfaceName)
			assert.Equal(t, "io", annot.PackageName)
			assert.True(t, annot.IsPointer)
			assert.Equal(t, "io", annot.PackageFullPath)
			assert.False(t, annot.PackageNotFound)
		})

		t.Run("MyWriteCloser implements io.Writer", func(t *testing.T) {
			annot := findImplements("MyWriteCloser", "Writer")
			require.NotNil(t, annot, "annotation not found")

			assert.Equal(t, "MyWriteCloser", annot.OnType)
			assert.Equal(t, "Writer", annot.InterfaceName)
			assert.Equal(t, "io", annot.PackageName)
			assert.True(t, annot.IsPointer)
			assert.Equal(t, "io", annot.PackageFullPath)
			assert.False(t, annot.PackageNotFound)
		})

		t.Run("MyWriteCloser implements io.Closer", func(t *testing.T) {
			annot := findImplements("MyWriteCloser", "Closer")
			require.NotNil(t, annot, "annotation not found")

			assert.Equal(t, "MyWriteCloser", annot.OnType)
			assert.Equal(t, "Closer", annot.InterfaceName)
			assert.Equal(t, "io", annot.PackageName)
			assert.True(t, annot.IsPointer)
			assert.Equal(t, "io", annot.PackageFullPath)
			assert.False(t, annot.PackageNotFound)
		})

		t.Run("MyContext implements context.Context", func(t *testing.T) {
			annot := findImplements("MyContext", "Context")
			require.NotNil(t, annot, "annotation not found")

			assert.Equal(t, "MyContext", annot.OnType)
			assert.Equal(t, "Context", annot.InterfaceName)
			assert.Equal(t, "context", annot.PackageName)
			assert.True(t, annot.IsPointer)
			assert.Equal(t, "context", annot.PackageFullPath)
			assert.False(t, annot.PackageNotFound)
		})
	})
}

func TestImportMapAdd(t *testing.T) {
	importMap := &ImportMap{}

	// Add simple import
	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"io"`},
	})

	// Add import with alias
	importMap.Add(&ast.ImportSpec{
		Name: &ast.Ident{Name: "foo"},
		Path: &ast.BasicLit{Value: `"github.com/example/bar"`},
	})

	// Add another import
	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"context"`},
	})

	assert.Len(t, *importMap, 3)

	assert.Equal(t, "io", (*importMap)[0].FullPath)
	assert.Equal(t, "", (*importMap)[0].Alias)

	assert.Equal(t, "github.com/example/bar", (*importMap)[1].FullPath)
	assert.Equal(t, "foo", (*importMap)[1].Alias)

	assert.Equal(t, "context", (*importMap)[2].FullPath)
	assert.Equal(t, "", (*importMap)[2].Alias)
}

func TestImportMapFind(t *testing.T) {
	importMap := &ImportMap{}

	// Add imports
	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"io"`},
	})
	importMap.Add(&ast.ImportSpec{
		Name: &ast.Ident{Name: "foo"},
		Path: &ast.BasicLit{Value: `"github.com/example/bar"`},
	})
	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"github.com/example/baz"`},
	})
	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"context"`},
	})

	tests := []struct {
		name         string
		shortName    string
		expectNil    bool
		expectedPath string
	}{
		{
			name:         "find by alias first",
			shortName:    "foo",
			expectNil:    false,
			expectedPath: "github.com/example/bar",
		},
		{
			name:         "find by suffix when no alias",
			shortName:    "io",
			expectNil:    false,
			expectedPath: "io",
		},
		{
			name:         "find by last path component",
			shortName:    "baz",
			expectNil:    false,
			expectedPath: "github.com/example/baz",
		},
		{
			name:         "find context",
			shortName:    "context",
			expectNil:    false,
			expectedPath: "context",
		},
		{
			name:      "not found",
			shortName: "nonexistent",
			expectNil: true,
		},
		{
			name:      "empty string",
			shortName: "",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := importMap.Find(tt.shortName)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedPath, result.FullPath)
			}
		})
	}
}

func TestImportMapFindPriority(t *testing.T) {
	// Test that alias has priority over suffix match
	importMap := &ImportMap{}

	// Add import with path "github.com/example/bar"
	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"github.com/example/bar"`},
	})

	// Add import with alias "bar" pointing to different package
	importMap.Add(&ast.ImportSpec{
		Name: &ast.Ident{Name: "bar"},
		Path: &ast.BasicLit{Value: `"github.com/other/package"`},
	})

	// When searching for "bar", should find the aliased one first
	result := importMap.Find("bar")
	require.NotNil(t, result)
	assert.Equal(t, "github.com/other/package", result.FullPath)
	assert.Equal(t, "bar", result.Alias)
}

func TestToInterfaceQuery(t *testing.T) {
	packageAnnotations := PackageAnnotations{
		ImplementsAnnotations: []ImplementsAnnotation{
			{
				InterfaceName:   "Reader",
				PackageName:     "io",
				PackageFullPath: "io",
				PackageNotFound: false,
			},
			{
				InterfaceName:   "Writer",
				PackageName:     "io",
				PackageFullPath: "io",
				PackageNotFound: false,
			},
			{
				InterfaceName:   "ResponseWriter",
				PackageName:     "http",
				PackageFullPath: "",
				PackageNotFound: true, // Should be skipped
			},
			{
				InterfaceName:   "MyInterface",
				PackageName:     "",
				PackageFullPath: "current/pkg/path",
				PackageNotFound: false,
			},
		},
	}

	queries := packageAnnotations.toInterfaceQuery()

	require.Len(t, queries, 3, "should skip unresolved packages")

	assert.Equal(t, "Reader", queries[0].InterfaceName)
	assert.Equal(t, "io", queries[0].PackageName)

	assert.Equal(t, "Writer", queries[1].InterfaceName)
	assert.Equal(t, "io", queries[1].PackageName)

	assert.Equal(t, "MyInterface", queries[2].InterfaceName)
	assert.Equal(t, "current/pkg/path", queries[2].PackageName)
}

func TestToTypeQuery(t *testing.T) {
	packageAnnotations := PackageAnnotations{
		ImplementsAnnotations: []ImplementsAnnotation{
			{OnType: "MyReader"},
			{OnType: "MyReader"}, // duplicate
			{OnType: "MyWriter"},
			{OnType: "MyContext"},
			{OnType: "MyWriter"}, // duplicate
		},
	}

	queries := packageAnnotations.toTypeQuery()

	require.Len(t, queries, 3, "should deduplicate")

	typeNames := make(map[string]bool)
	for _, q := range queries {
		typeNames[q.TypeName] = true
	}

	assert.True(t, typeNames["MyReader"])
	assert.True(t, typeNames["MyWriter"])
	assert.True(t, typeNames["MyContext"])
}
