package implements

import (
	"gogreement/src/annotations"
	"gogreement/src/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadInterfaces(t *testing.T) {
	pass := testutil.CreateTestPass(t, "interfacesforloading")
	expectedPkgPath := pass.Pkg.Path()

	tests := []struct {
		name              string
		queries           []annotations.InterfaceQuery
		expectedCount     int
		expectedInterface string
		expectedMethods   []string
	}{
		{
			name: "load single interface from current package",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Reader", PackageName: ""},
			},
			expectedCount:     1,
			expectedInterface: "Reader",
			expectedMethods:   []string{"Read", "Close"},
		},
		{
			name: "load single interface with explicit package path",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Writer", PackageName: expectedPkgPath},
			},
			expectedCount:     1,
			expectedInterface: "Writer",
			expectedMethods:   []string{"Write"},
		},
		{
			name: "load multiple interfaces",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Reader", PackageName: ""},
				{InterfaceName: "Writer", PackageName: ""},
			},
			expectedCount: 2,
		},
		{
			name: "load Processor interface",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Processor", PackageName: ""},
			},
			expectedCount:     1,
			expectedInterface: "Processor",
			expectedMethods:   []string{"Process", "ProcessMany", "ProcessPointer"},
		},
		{
			name: "load Empty interface",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Empty", PackageName: ""},
			},
			expectedCount:     1,
			expectedInterface: "Empty",
			expectedMethods:   []string{},
		},
		{
			name: "interface not found",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "NonExistent", PackageName: ""},
			},
			expectedCount: 0,
		},
		{
			name: "load all test interfaces",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Reader", PackageName: ""},
				{InterfaceName: "Writer", PackageName: ""},
				{InterfaceName: "Processor", PackageName: ""},
				{InterfaceName: "Empty", PackageName: ""},
			},
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LoadInterfaces(pass, tt.queries)

			assert.Len(t, result, tt.expectedCount)

			if tt.expectedInterface != "" && len(result) > 0 {
				iface := result[0]

				assert.Equal(t, tt.expectedInterface, iface.Name)
				assert.Equal(t, expectedPkgPath, iface.Package)
				assert.Len(t, iface.Methods, len(tt.expectedMethods))

				// Check method names
				methodNames := make([]string, len(iface.Methods))
				for i, method := range iface.Methods {
					methodNames[i] = method.Name
				}

				for _, expectedMethod := range tt.expectedMethods {
					assert.Contains(t, methodNames, expectedMethod)
				}
			}
		})
	}
}

func TestLoadInterfacesMethodSignatures(t *testing.T) {
	pass := testutil.CreateTestPass(t, "interfacesforloading")

	queries := []annotations.InterfaceQuery{
		{InterfaceName: "Reader", PackageName: ""},
		{InterfaceName: "Processor", PackageName: ""},
		{InterfaceName: "Writer", PackageName: ""},
	}

	result := LoadInterfaces(pass, queries)
	require.NotEmpty(t, result, "expected to find interfaces")

	// Helper to find interface by name
	findInterface := func(name string) *InterfaceModel {
		for _, iface := range result {
			if iface.Name == name {
				return iface
			}
		}
		return nil
	}

	// Helper to find method by name
	findMethod := func(iface *InterfaceModel, methodName string) *InterfaceMethod {
		for _, method := range iface.Methods {
			if method.Name == methodName {
				return &method
			}
		}
		return nil
	}

	t.Run("Reader.Read signature", func(t *testing.T) {
		reader := findInterface("Reader")
		require.NotNil(t, reader, "Reader interface not found")

		readMethod := findMethod(reader, "Read")
		require.NotNil(t, readMethod, "Read method not found")

		// Check inputs: (p []byte)
		assert.Len(t, readMethod.Inputs, 1)

		// Check outputs: (n int, err error)
		require.Len(t, readMethod.Outputs, 2)
		assert.Equal(t, "int", readMethod.Outputs[0].TypeName)
		assert.Equal(t, "error", readMethod.Outputs[1].TypeName)
	})

	t.Run("Processor.ProcessPointer signature", func(t *testing.T) {
		processor := findInterface("Processor")
		require.NotNil(t, processor, "Processor interface not found")

		method := findMethod(processor, "ProcessPointer")
		require.NotNil(t, method, "ProcessPointer method not found")

		// Check input: (ptr *int)
		require.Len(t, method.Inputs, 1)
		assert.True(t, method.Inputs[0].IsPointer, "expected input to be pointer")
		assert.Equal(t, "int", method.Inputs[0].TypeName)

		// Check output: *string
		require.Len(t, method.Outputs, 1)
		assert.True(t, method.Outputs[0].IsPointer, "expected output to be pointer")
		assert.Equal(t, "string", method.Outputs[0].TypeName)
	})

	t.Run("Processor.ProcessMany variadic", func(t *testing.T) {
		processor := findInterface("Processor")
		require.NotNil(t, processor, "Processor interface not found")

		method := findMethod(processor, "ProcessMany")
		require.NotNil(t, method, "ProcessMany method not found")

		// Check variadic input: (items ...string)
		require.Len(t, method.Inputs, 1)
		assert.True(t, method.Inputs[0].IsVariadic, "expected input to be variadic")
		assert.Equal(t, "string", method.Inputs[0].TypeName)
	})

	t.Run("Reader.Close signature", func(t *testing.T) {
		reader := findInterface("Reader")
		require.NotNil(t, reader, "Reader interface not found")

		closeMethod := findMethod(reader, "Close")
		require.NotNil(t, closeMethod, "Close method not found")

		// Check no inputs
		assert.Empty(t, closeMethod.Inputs)

		// Check outputs: error
		require.Len(t, closeMethod.Outputs, 1)
		assert.Equal(t, "error", closeMethod.Outputs[0].TypeName)
		assert.False(t, closeMethod.Outputs[0].IsPointer)
	})

	t.Run("Writer.Write signature", func(t *testing.T) {
		writer := findInterface("Writer")
		require.NotNil(t, writer, "Writer interface not found")

		writeMethod := findMethod(writer, "Write")
		require.NotNil(t, writeMethod, "Write method not found")

		// Check inputs: (data []byte)
		assert.Len(t, writeMethod.Inputs, 1)

		// Check outputs: (int, error)
		require.Len(t, writeMethod.Outputs, 2)
		assert.Equal(t, "int", writeMethod.Outputs[0].TypeName)
		assert.Equal(t, "error", writeMethod.Outputs[1].TypeName)
	})
}

func TestLoadInterfacesEmptyQueries(t *testing.T) {
	pass := testutil.CreateTestPass(t, "interfacesforloading")

	result := LoadInterfaces(pass, []annotations.InterfaceQuery{})

	assert.Empty(t, result, "expected no interfaces when queries are empty")
}

func TestLoadInterfacesDuplicateQueries(t *testing.T) {
	pass := testutil.CreateTestPass(t, "interfacesforloading")

	queries := []annotations.InterfaceQuery{
		{InterfaceName: "Reader", PackageName: ""},
		{InterfaceName: "Reader", PackageName: ""}, // duplicate
	}

	result := LoadInterfaces(pass, queries)

	// Should return only one instance despite duplicate query
	assert.Len(t, result, 1)
	assert.Equal(t, "Reader", result[0].Name)
}

func TestLoadInterfacesFromStdlib(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	tests := []struct {
		name              string
		queries           []annotations.InterfaceQuery
		expectedCount     int
		expectedInterface string
		expectedPackage   string
		expectedMethods   []string
	}{
		{
			name: "load io.Reader from stdlib",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Reader", PackageName: "io"},
			},
			expectedCount:     1,
			expectedInterface: "Reader",
			expectedPackage:   "io",
			expectedMethods:   []string{"Read"},
		},
		{
			name: "load io.Writer from stdlib",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Writer", PackageName: "io"},
			},
			expectedCount:     1,
			expectedInterface: "Writer",
			expectedPackage:   "io",
			expectedMethods:   []string{"Write"},
		},
		{
			name: "load io.Closer from stdlib",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Closer", PackageName: "io"},
			},
			expectedCount:     1,
			expectedInterface: "Closer",
			expectedPackage:   "io",
			expectedMethods:   []string{"Close"},
		},
		{
			name: "load context.Context from stdlib",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Context", PackageName: "context"},
			},
			expectedCount:     1,
			expectedInterface: "Context",
			expectedPackage:   "context",
			expectedMethods:   []string{"Deadline", "Done", "Err", "Value"},
		},
		{
			name: "load multiple stdlib interfaces",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Reader", PackageName: "io"},
				{InterfaceName: "Writer", PackageName: "io"},
				{InterfaceName: "Closer", PackageName: "io"},
			},
			expectedCount: 3,
		},
		{
			name: "mix local and stdlib interfaces",
			queries: []annotations.InterfaceQuery{
				{InterfaceName: "Reader", PackageName: "io"}, // stdlib
				{InterfaceName: "Reader", PackageName: ""},   // local (if exists)
			},
			expectedCount: 1, // only io.Reader since local doesn't exist
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LoadInterfaces(pass, tt.queries)

			assert.Len(t, result, tt.expectedCount)

			if tt.expectedInterface != "" && len(result) > 0 {
				iface := result[0]

				assert.Equal(t, tt.expectedInterface, iface.Name)
				assert.Equal(t, tt.expectedPackage, iface.Package)
				assert.Len(t, iface.Methods, len(tt.expectedMethods))

				// Check method names
				methodNames := make([]string, len(iface.Methods))
				for i, method := range iface.Methods {
					methodNames[i] = method.Name
				}

				for _, expectedMethod := range tt.expectedMethods {
					assert.Contains(t, methodNames, expectedMethod)
				}
			}
		})
	}
}

func TestLoadInterfacesStdlibMethodSignatures(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	queries := []annotations.InterfaceQuery{
		{InterfaceName: "Reader", PackageName: "io"},
		{InterfaceName: "Writer", PackageName: "io"},
		{InterfaceName: "Context", PackageName: "context"},
	}

	result := LoadInterfaces(pass, queries)
	require.Len(t, result, 3, "expected to find 3 interfaces")

	// Helper to find interface by name and package
	findInterface := func(name, pkg string) *InterfaceModel {
		for _, iface := range result {
			if iface.Name == name && iface.Package == pkg {
				return iface
			}
		}
		return nil
	}

	// Helper to find method by name
	findMethod := func(iface *InterfaceModel, methodName string) *InterfaceMethod {
		for _, method := range iface.Methods {
			if method.Name == methodName {
				return &method
			}
		}
		return nil
	}

	t.Run("io.Reader.Read signature", func(t *testing.T) {
		reader := findInterface("Reader", "io")
		require.NotNil(t, reader, "io.Reader not found")

		readMethod := findMethod(reader, "Read")
		require.NotNil(t, readMethod, "Read method not found")

		// Check signature: Read(p []byte) (n int, err error)
		require.Len(t, readMethod.Inputs, 1)
		assert.Equal(t, "[]byte", readMethod.Inputs[0].TypeName) // Changed from "byte" to "[]byte"

		require.Len(t, readMethod.Outputs, 2)
		assert.Equal(t, "int", readMethod.Outputs[0].TypeName)
		assert.Equal(t, "error", readMethod.Outputs[1].TypeName)
	})

	t.Run("io.Writer.Write signature", func(t *testing.T) {
		writer := findInterface("Writer", "io")
		require.NotNil(t, writer, "io.Writer not found")

		writeMethod := findMethod(writer, "Write")
		require.NotNil(t, writeMethod, "Write method not found")

		// Check signature: Write(p []byte) (n int, err error)
		require.Len(t, writeMethod.Inputs, 1)

		require.Len(t, writeMethod.Outputs, 2)
		assert.Equal(t, "int", writeMethod.Outputs[0].TypeName)
		assert.Equal(t, "error", writeMethod.Outputs[1].TypeName)
	})

	t.Run("context.Context methods", func(t *testing.T) {
		ctx := findInterface("Context", "context")
		require.NotNil(t, ctx, "context.Context not found")

		// Check all methods exist
		assert.NotNil(t, findMethod(ctx, "Deadline"))
		assert.NotNil(t, findMethod(ctx, "Done"))
		assert.NotNil(t, findMethod(ctx, "Err"))
		assert.NotNil(t, findMethod(ctx, "Value"))

		// Check Done() signature: Done() <-chan struct{}
		doneMethod := findMethod(ctx, "Done")
		require.NotNil(t, doneMethod)
		assert.Empty(t, doneMethod.Inputs)
		assert.Len(t, doneMethod.Outputs, 1)
	})
}

func TestLoadInterfacesNonExistentStdlib(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	queries := []annotations.InterfaceQuery{
		{InterfaceName: "NonExistentInterface", PackageName: "io"},
	}

	result := LoadInterfaces(pass, queries)

	assert.Empty(t, result, "should not find non-existent interface")
}

func TestLoadInterfacesUnimportedPackage(t *testing.T) {
	pass := testutil.CreateTestPass(t, "withimports")

	// Try to load from a package that's not imported
	queries := []annotations.InterfaceQuery{
		{InterfaceName: "ResponseWriter", PackageName: "net/http"},
	}

	result := LoadInterfaces(pass, queries)

	// Should be empty because net/http is not imported in withimports package
	assert.Empty(t, result, "should not find interface from unimported package")
}
