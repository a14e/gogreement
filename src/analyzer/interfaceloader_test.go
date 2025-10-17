package analyzer

import (
	"go/types"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

// getTestdataPath returns absolute path to testdata directory
func getTestdataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "testdata")
}

// loadTestPackage loads a test package and returns types.Package
func loadTestPackage(t *testing.T, pkgName string) *types.Package {
	testdataPath := getTestdataPath()
	pkgPath := filepath.Join(testdataPath, pkgName)

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedTypes |
			packages.NeedImports | packages.NeedDeps | packages.NeedSyntax,
		Dir: testdataPath,
	}

	// Use relative path pattern "./simple" instead of just "simple"
	pattern := "./" + pkgName
	pkgs, err := packages.Load(cfg, pattern)
	require.NoError(t, err, "failed to load package")
	require.NotEmpty(t, pkgs, "no packages loaded")

	if len(pkgs[0].Errors) > 0 {
		t.Logf("Package path: %s", pkgPath)
		t.Logf("Pattern: %s", pattern)
		for _, e := range pkgs[0].Errors {
			t.Logf("Error: %v", e)
		}
	}
	require.Empty(t, pkgs[0].Errors, "package has errors")

	return pkgs[0].Types
}

func TestFindInterfacesInPackage(t *testing.T) {
	pkg := loadTestPackage(t, "simple")

	tests := []struct {
		name              string
		targetInterfaces  map[string]bool
		expectedCount     int
		expectedInterface string
		expectedMethods   []string
	}{
		{
			name: "find Reader interface",
			targetInterfaces: map[string]bool{
				"Reader": true,
			},
			expectedCount:     1,
			expectedInterface: "Reader",
			expectedMethods:   []string{"Read", "Close"},
		},
		{
			name: "find Writer interface",
			targetInterfaces: map[string]bool{
				"Writer": true,
			},
			expectedCount:     1,
			expectedInterface: "Writer",
			expectedMethods:   []string{"Write"},
		},
		{
			name: "find multiple interfaces",
			targetInterfaces: map[string]bool{
				"Reader": true,
				"Writer": true,
			},
			expectedCount: 2,
		},
		{
			name: "find Processor with various params",
			targetInterfaces: map[string]bool{
				"Processor": true,
			},
			expectedCount:     1,
			expectedInterface: "Processor",
			expectedMethods:   []string{"Process", "ProcessMany", "ProcessPointer"},
		},
		{
			name: "find Empty interface",
			targetInterfaces: map[string]bool{
				"Empty": true,
			},
			expectedCount:     1,
			expectedInterface: "Empty",
			expectedMethods:   []string{}, // no methods
		},
		{
			name: "interface not found",
			targetInterfaces: map[string]bool{
				"NonExistent": true,
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findInterfacesInPackage(pkg, tt.targetInterfaces)

			assert.Len(t, result, tt.expectedCount)

			if tt.expectedInterface != "" && len(result) > 0 {
				iface := result[0]

				assert.Equal(t, tt.expectedInterface, iface.Name)
				assert.Equal(t, "simple", iface.Package)
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

func TestInterfaceMethodDetails(t *testing.T) {
	pkg := loadTestPackage(t, "simple")

	targetInterfaces := map[string]bool{
		"Reader":    true,
		"Processor": true,
	}

	result := findInterfacesInPackage(pkg, targetInterfaces)
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

	t.Run("Empty interface has no methods", func(t *testing.T) {
		targetEmpty := map[string]bool{"Empty": true}
		emptyResult := findInterfacesInPackage(pkg, targetEmpty)

		require.Len(t, emptyResult, 1)
		assert.Equal(t, "Empty", emptyResult[0].Name)
		assert.Empty(t, emptyResult[0].Methods, "Empty interface should have no methods")
	})
}
