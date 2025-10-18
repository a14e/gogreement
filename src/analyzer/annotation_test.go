package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseImplementsAnnotation(t *testing.T) {
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
				OnType:        "MyStruct",
				InterfaceName: "MyInterface",
				PackageName:   "",
				IsPointer:     false,
			},
		},
		{
			name:      "pointer interface",
			comment:   "// @implements &MyInterface",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:        "MyStruct",
				InterfaceName: "MyInterface",
				PackageName:   "",
				IsPointer:     true,
			},
		},
		{
			name:      "package qualified",
			comment:   "// @implements io.Reader",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:        "MyStruct",
				InterfaceName: "Reader",
				PackageName:   "io",
				IsPointer:     false,
			},
		},
		{
			name:      "pointer with package",
			comment:   "// @implements &io.Reader",
			typeName:  "MyStruct",
			expectNil: false,
			expectedAnnot: &ImplementsAnnotation{
				OnType:        "MyStruct",
				InterfaceName: "Reader",
				PackageName:   "io",
				IsPointer:     true,
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
				OnType:        "MyStruct",
				InterfaceName: "Reader",
				PackageName:   "io",
				IsPointer:     true,
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
			result := parseImplementsAnnotation(tt.comment, tt.typeName, 0)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedAnnot.OnType, result.OnType)
				assert.Equal(t, tt.expectedAnnot.InterfaceName, result.InterfaceName)
				assert.Equal(t, tt.expectedAnnot.PackageName, result.PackageName)
				assert.Equal(t, tt.expectedAnnot.IsPointer, result.IsPointer)
			}
		})
	}
}

func TestReadAllImplementsAnnotations(t *testing.T) {
	pass := createTestPass(t, "withimports")

	annotations := ReadAllImplementsAnnotations(pass)

	// We expect annotations from withimports.go:
	// - MyReader implements &io.Reader
	// - MyWriteCloser implements &io.Writer
	// - MyWriteCloser implements &io.Closer
	// - MyContext implements &context.Context

	require.NotEmpty(t, annotations, "expected to find annotations")

	// Helper to find annotation
	findAnnotation := func(onType, interfaceName string) *ImplementsAnnotation {
		for _, a := range annotations {
			if a.OnType == onType && a.InterfaceName == interfaceName {
				return &a
			}
		}
		return nil
	}

	t.Run("MyReader implements io.Reader", func(t *testing.T) {
		annot := findAnnotation("MyReader", "Reader")
		require.NotNil(t, annot, "annotation not found")

		assert.Equal(t, "MyReader", annot.OnType)
		assert.Equal(t, "Reader", annot.InterfaceName)
		assert.Equal(t, "io", annot.PackageName)
		assert.True(t, annot.IsPointer)
		assert.Equal(t, "io", annot.PackageFullPath)
		assert.False(t, annot.PackageNotFound)
	})

	t.Run("MyWriteCloser implements io.Writer", func(t *testing.T) {
		annot := findAnnotation("MyWriteCloser", "Writer")
		require.NotNil(t, annot, "annotation not found")

		assert.Equal(t, "MyWriteCloser", annot.OnType)
		assert.Equal(t, "Writer", annot.InterfaceName)
		assert.Equal(t, "io", annot.PackageName)
		assert.True(t, annot.IsPointer)
		assert.Equal(t, "io", annot.PackageFullPath)
		assert.False(t, annot.PackageNotFound)
	})

	t.Run("MyWriteCloser implements io.Closer", func(t *testing.T) {
		annot := findAnnotation("MyWriteCloser", "Closer")
		require.NotNil(t, annot, "annotation not found")

		assert.Equal(t, "MyWriteCloser", annot.OnType)
		assert.Equal(t, "Closer", annot.InterfaceName)
		assert.Equal(t, "io", annot.PackageName)
		assert.True(t, annot.IsPointer)
		assert.Equal(t, "io", annot.PackageFullPath)
		assert.False(t, annot.PackageNotFound)
	})

	t.Run("MyContext implements context.Context", func(t *testing.T) {
		annot := findAnnotation("MyContext", "Context")
		require.NotNil(t, annot, "annotation not found")

		assert.Equal(t, "MyContext", annot.OnType)
		assert.Equal(t, "Context", annot.InterfaceName)
		assert.Equal(t, "context", annot.PackageName)
		assert.True(t, annot.IsPointer)
		assert.Equal(t, "context", annot.PackageFullPath)
		assert.False(t, annot.PackageNotFound)
	})

	t.Run("count all annotations", func(t *testing.T) {
		// Should have at least 4 annotations
		assert.GreaterOrEqual(t, len(annotations), 4)
	})
}

func TestResolvePackagePath(t *testing.T) {
	pass := createTestPass(t, "withimports")

	tests := []struct {
		name             string
		annotation       ImplementsAnnotation
		expectedFullPath string
		expectedNotFound bool
	}{
		{
			name: "current package (empty)",
			annotation: ImplementsAnnotation{
				PackageName: "",
			},
			expectedFullPath: pass.Pkg.Path(),
			expectedNotFound: false,
		},
		{
			name: "stdlib io package",
			annotation: ImplementsAnnotation{
				PackageName: "io",
			},
			expectedFullPath: "io",
			expectedNotFound: false,
		},
		{
			name: "stdlib context package",
			annotation: ImplementsAnnotation{
				PackageName: "context",
			},
			expectedFullPath: "context",
			expectedNotFound: false,
		},
		{
			name: "non-imported package",
			annotation: ImplementsAnnotation{
				PackageName: "http", // not imported in withimports
			},
			expectedFullPath: "",
			expectedNotFound: true,
		},
		{
			name: "non-existent package",
			annotation: ImplementsAnnotation{
				PackageName: "nonexistent",
			},
			expectedFullPath: "",
			expectedNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annot := tt.annotation
			resolvePackagePath(&annot, pass.Pkg)

			assert.Equal(t, tt.expectedFullPath, annot.PackageFullPath)
			assert.Equal(t, tt.expectedNotFound, annot.PackageNotFound)
		})
	}
}

func TestToInterfaceQuery(t *testing.T) {
	annotations := []ImplementsAnnotation{
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
	}

	queries := toInterfaceQuery(annotations)

	require.Len(t, queries, 3, "should skip unresolved packages")

	assert.Equal(t, "Reader", queries[0].InterfaceName)
	assert.Equal(t, "io", queries[0].PackageName)

	assert.Equal(t, "Writer", queries[1].InterfaceName)
	assert.Equal(t, "io", queries[1].PackageName)

	assert.Equal(t, "MyInterface", queries[2].InterfaceName)
	assert.Equal(t, "current/pkg/path", queries[2].PackageName)
}

func TestToTypeQuery(t *testing.T) {
	annotations := []ImplementsAnnotation{
		{OnType: "MyReader"},
		{OnType: "MyReader"}, // duplicate
		{OnType: "MyWriter"},
		{OnType: "MyContext"},
		{OnType: "MyWriter"}, // duplicate
	}

	queries := toTypeQuery(annotations)

	require.Len(t, queries, 3, "should deduplicate")

	typeNames := make(map[string]bool)
	for _, q := range queries {
		typeNames[q.TypeName] = true
	}

	assert.True(t, typeNames["MyReader"])
	assert.True(t, typeNames["MyWriter"])
	assert.True(t, typeNames["MyContext"])
}
