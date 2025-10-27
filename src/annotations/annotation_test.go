package annotations

import (
	"go/ast"
	"gogreement/src/config"
	"gogreement/src/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gogreement/src/util"
)

func TestParseImplementsAnnotation(t *testing.T) {
	// Create mock import map
	imports := &util.ImportMap{}
	imports.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"io"`},
	}, nil)
	imports.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"context"`},
	}, nil)

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
			name:      "with extra text after should work now",
			comment:   "//  @implements &io.Reader text after",
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

func TestParseImmutableAnnotation(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		typeName  string
		expectNil bool
	}{
		{
			name:      "simple immutable",
			comment:   "// @immutable",
			typeName:  "MyStruct",
			expectNil: false,
		},
		{
			name:      "immutable with spaces",
			comment:   "//   @immutable   ",
			typeName:  "MyStruct",
			expectNil: false,
		},
		{
			name:      "immutable with tabs",
			comment:   "//\t@immutable\t",
			typeName:  "AnotherStruct",
			expectNil: false,
		},
		{
			name:      "extra text before - should fail",
			comment:   "// text before @immutable",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "extra text after - should work now",
			comment:   "// @immutable text after",
			typeName:  "MyStruct",
			expectNil: false,
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
		{
			name:      "wrong annotation - constructor",
			comment:   "// @constructor New",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "partial match - immutability",
			comment:   "// @immutability",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "empty comment",
			comment:   "//",
			typeName:  "MyStruct",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseImmutableAnnotation(tt.comment, tt.typeName, 0)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.typeName, result.OnType)
			}
		})
	}
}

func TestReadAllAnnotations(t *testing.T) {

	pass := testutil.CreateTestPass(t, "withimports")

	cfg := config.Empty()
	annotations := ReadAllAnnotations(cfg, pass)

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

	// Test @immutable annotations
	t.Run("ImmutableAnnotations", func(t *testing.T) {
		// Helper to find immutable annotation
		findImmutable := func(onType string) *ImmutableAnnotation {
			for _, a := range annotations.ImmutableAnnotations {
				if a.OnType == onType {
					return &a
				}
			}
			return nil
		}

		t.Run("MyReader is immutable", func(t *testing.T) {
			annot := findImmutable("MyReader")
			require.NotNil(t, annot, "annotation not found")
			assert.Equal(t, "MyReader", annot.OnType)
		})

		t.Run("MyWriteCloser is immutable", func(t *testing.T) {
			annot := findImmutable("MyWriteCloser")
			require.NotNil(t, annot, "annotation not found")
			assert.Equal(t, "MyWriteCloser", annot.OnType)
		})

		t.Run("MyContext is immutable", func(t *testing.T) {
			annot := findImmutable("MyContext")
			require.NotNil(t, annot, "annotation not found")
			assert.Equal(t, "MyContext", annot.OnType)
		})

		t.Run("Duration is immutable", func(t *testing.T) {
			annot := findImmutable("Duration")
			require.NotNil(t, annot, "annotation not found")
			assert.Equal(t, "Duration", annot.OnType)
		})

		t.Run("HandlerFunc is immutable", func(t *testing.T) {
			annot := findImmutable("HandlerFunc")
			require.NotNil(t, annot, "annotation not found")
			assert.Equal(t, "HandlerFunc", annot.OnType)
		})

		t.Run("MyString should not be immutable", func(t *testing.T) {
			annot := findImmutable("MyString")
			assert.Nil(t, annot, "MyString should not have @immutable annotation")
		})

		t.Run("ByteSlice should not be immutable", func(t *testing.T) {
			annot := findImmutable("ByteSlice")
			assert.Nil(t, annot, "ByteSlice should not have @immutable annotation")
		})
	})
}

func TestReadAllAnnotationsWithImmutable(t *testing.T) {
	pass := testutil.CreateTestPass(t, "interfacesforloading")

	cfg := config.Empty()
	annotations := ReadAllAnnotations(cfg, pass)

	// Helper to find immutable annotation
	findImmutable := func(onType string) *ImmutableAnnotation {
		for _, a := range annotations.ImmutableAnnotations {
			if a.OnType == onType {
				return &a
			}
		}
		return nil
	}

	t.Run("FileReader is immutable", func(t *testing.T) {
		annot := findImmutable("FileReader")
		require.NotNil(t, annot, "FileReader should have @immutable annotation")
		assert.Equal(t, "FileReader", annot.OnType)
	})

	t.Run("BufferWriter is immutable", func(t *testing.T) {
		annot := findImmutable("BufferWriter")
		require.NotNil(t, annot, "BufferWriter should have @immutable annotation")
		assert.Equal(t, "BufferWriter", annot.OnType)
	})

	t.Run("StringProcessor is immutable", func(t *testing.T) {
		annot := findImmutable("StringProcessor")
		require.NotNil(t, annot, "StringProcessor should have @immutable annotation")
		assert.Equal(t, "StringProcessor", annot.OnType)
	})

	t.Run("EmptyImpl is immutable", func(t *testing.T) {
		annot := findImmutable("EmptyImpl")
		require.NotNil(t, annot, "EmptyImpl should have @immutable annotation")
		assert.Equal(t, "EmptyImpl", annot.OnType)
	})

	t.Run("Interfaces should be immutable", func(t *testing.T) {
		interfaces := []string{"Reader", "Writer", "Processor", "Empty"}
		for _, ifaceName := range interfaces {
			annot := findImmutable(ifaceName)
			require.NotNil(t, annot, "%s interface should have @immutable annotation", ifaceName)
			assert.Equal(t, ifaceName, annot.OnType)
		}
	})

	t.Run("Config should not be immutable", func(t *testing.T) {
		annot := findImmutable("Config")
		assert.Nil(t, annot, "Config should not have @immutable annotation")
	})

	t.Run("MutableType should not be immutable", func(t *testing.T) {
		annot := findImmutable("MutableType")
		assert.Nil(t, annot, "MutableType should not have @immutable annotation")
	})
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

	queries := packageAnnotations.ToTypeQuery()

	require.Len(t, queries, 3, "should deduplicate")

	typeNames := make(map[string]bool)
	for _, q := range queries {
		typeNames[q.TypeName] = true
	}

	assert.True(t, typeNames["MyReader"])
	assert.True(t, typeNames["MyWriter"])
	assert.True(t, typeNames["MyContext"])
}

func TestParseTestOnlyAnnotation(t *testing.T) {
	tests := []struct {
		name      string
		comment   string
		typeName  string
		expectNil bool
	}{
		{
			name:      "simple testonly",
			comment:   "// @testonly",
			typeName:  "MyStruct",
			expectNil: false,
		},
		{
			name:      "testonly with spaces",
			comment:   "//   @testonly   ",
			typeName:  "MyStruct",
			expectNil: false,
		},
		{
			name:      "testonly with tabs",
			comment:   "//\t@testonly\t",
			typeName:  "TestHelper",
			expectNil: false,
		},
		{
			name:      "extra text before - should fail",
			comment:   "// text before @testonly",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "extra text after - should work now",
			comment:   "// @testonly text after",
			typeName:  "MyStruct",
			expectNil: false,
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
		{
			name:      "wrong annotation - constructor",
			comment:   "// @constructor New",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "wrong annotation - immutable",
			comment:   "// @immutable",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "partial match - testonlymode",
			comment:   "// @testonlymode",
			typeName:  "MyStruct",
			expectNil: true,
		},
		{
			name:      "empty comment",
			comment:   "//",
			typeName:  "MyStruct",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTestOnlyAnnotation(tt.comment, tt.typeName, 0, TestOnlyOnType, "")

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, TestOnlyOnType, result.Kind)
				assert.Equal(t, tt.typeName, result.ObjectName)
				assert.Equal(t, "", result.ReceiverType)
			}
		})
	}
}

func TestReadTestOnlyAnnotations(t *testing.T) {
	pass := testutil.CreateTestPass(t, "testonlyexample")

	cfg := config.Empty()
	annotations := ReadAllAnnotations(cfg, pass)

	// Helper to find testonly annotation
	findTestOnly := func(objectName string, kind TestOnlyKind) *TestOnlyAnnotation {
		for _, a := range annotations.TestonlyAnnotations {
			if a.ObjectName == objectName && a.Kind == kind {
				return &a
			}
		}
		return nil
	}

	t.Run("TestHelper type has @testonly", func(t *testing.T) {
		annot := findTestOnly("TestHelper", TestOnlyOnType)
		require.NotNil(t, annot, "@testonly annotation not found on TestHelper type")
		assert.Equal(t, TestOnlyOnType, annot.Kind)
		assert.Equal(t, "TestHelper", annot.ObjectName)
		assert.Equal(t, "", annot.ReceiverType)
	})

	t.Run("CreateMockData function has @testonly", func(t *testing.T) {
		annot := findTestOnly("CreateMockData", TestOnlyOnFunc)
		require.NotNil(t, annot, "@testonly annotation not found on CreateMockData function")
		assert.Equal(t, TestOnlyOnFunc, annot.Kind)
		assert.Equal(t, "CreateMockData", annot.ObjectName)
		assert.Equal(t, "", annot.ReceiverType)
	})

	t.Run("Reset method has @testonly", func(t *testing.T) {
		annot := findTestOnly("Reset", TestOnlyOnMethod)
		require.NotNil(t, annot, "@testonly annotation not found on Reset method")
		assert.Equal(t, TestOnlyOnMethod, annot.Kind)
		assert.Equal(t, "Reset", annot.ObjectName)
		assert.Equal(t, "MyService", annot.ReceiverType)
	})

	t.Run("GetTestData method has @testonly", func(t *testing.T) {
		annot := findTestOnly("GetTestData", TestOnlyOnMethod)
		require.NotNil(t, annot, "@testonly annotation not found on GetTestData method")
		assert.Equal(t, TestOnlyOnMethod, annot.Kind)
		assert.Equal(t, "GetTestData", annot.ObjectName)
		assert.Equal(t, "MyService", annot.ReceiverType)
	})

	t.Run("Regular items should not have @testonly", func(t *testing.T) {
		// MyService type should not have annotation
		typeAnnot := findTestOnly("MyService", TestOnlyOnType)
		assert.Nil(t, typeAnnot, "MyService type should not have @testonly annotation")

		// ProcessData function should not have annotation
		funcAnnot := findTestOnly("ProcessData", TestOnlyOnFunc)
		assert.Nil(t, funcAnnot, "ProcessData function should not have @testonly annotation")

		// Process method should not have annotation
		methodAnnot := findTestOnly("Process", TestOnlyOnMethod)
		assert.Nil(t, methodAnnot, "Process method should not have @testonly annotation")
	})

	t.Run("Total count of @testonly annotations", func(t *testing.T) {
		assert.Equal(t, 4, len(annotations.TestonlyAnnotations), "should have exactly 4 @testonly annotations")
	})
}
