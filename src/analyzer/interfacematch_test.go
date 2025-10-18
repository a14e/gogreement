package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ========== Tests for findMissingPackages ==========

func TestFindMissingPackages(t *testing.T) {
	tests := []struct {
		name        string
		annotations []ImplementsAnnotation
		expected    []MissingPackageReport
	}{
		{
			name: "no missing packages",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					PackageName:     "io",
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			expected: []MissingPackageReport{},
		},
		{
			name: "single missing package",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyWriter",
					PackageName:     "http",
					PackageNotFound: true,
					OnTypePos:       200,
				},
			},
			expected: []MissingPackageReport{
				{
					PackageName: "http",
					TypeName:    "MyWriter",
					Pos:         200,
				},
			},
		},
		{
			name: "mixed - some found, some not",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					PackageName:     "io",
					PackageNotFound: false,
					OnTypePos:       100,
				},
				{
					OnType:          "MyWriter",
					PackageName:     "http",
					PackageNotFound: true,
					OnTypePos:       200,
				},
				{
					OnType:          "MyContext",
					PackageName:     "context",
					PackageNotFound: false,
					OnTypePos:       300,
				},
				{
					OnType:          "MyHandler",
					PackageName:     "net",
					PackageNotFound: true,
					OnTypePos:       400,
				},
			},
			expected: []MissingPackageReport{
				{
					PackageName: "http",
					TypeName:    "MyWriter",
					Pos:         200,
				},
				{
					PackageName: "net",
					TypeName:    "MyHandler",
					Pos:         400,
				},
			},
		},
		{
			name:        "empty annotations",
			annotations: []ImplementsAnnotation{},
			expected:    []MissingPackageReport{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMissingPackages(tt.annotations)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========== Tests for findMissingInterfaces ==========

func TestFindMissingInterfaces(t *testing.T) {
	tests := []struct {
		name        string
		annotations []ImplementsAnnotation
		interfaces  []*InterfaceModel
		expected    []MissingInterfaceReport
	}{
		{
			name: "all interfaces found",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "io",
					PackageFullPath: "io",
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
				},
			},
			expected: []MissingInterfaceReport{},
		},
		{
			name: "interface not found",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyWriter",
					InterfaceName:   "NonExistent",
					PackageName:     "io",
					PackageFullPath: "io",
					PackageNotFound: false,
					OnTypePos:       200,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
				},
			},
			expected: []MissingInterfaceReport{
				{
					InterfaceName: "NonExistent",
					PackageName:   "io",
					TypeName:      "MyWriter",
					Pos:           200,
				},
			},
		},
		{
			name: "skip annotations with package not found",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyWriter",
					InterfaceName:   "Writer",
					PackageName:     "http",
					PackageFullPath: "",
					PackageNotFound: true, // Should skip this
					OnTypePos:       200,
				},
			},
			interfaces: []*InterfaceModel{},
			expected:   []MissingInterfaceReport{}, // Empty because we skip PackageNotFound
		},
		{
			name: "mixed - some found, some not",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "io",
					PackageFullPath: "io",
					PackageNotFound: false,
					OnTypePos:       100,
				},
				{
					OnType:          "MyWriter",
					InterfaceName:   "Writer",
					PackageName:     "io",
					PackageFullPath: "io",
					PackageNotFound: false,
					OnTypePos:       200,
				},
				{
					OnType:          "MyCloser",
					InterfaceName:   "Closer",
					PackageName:     "io",
					PackageFullPath: "io",
					PackageNotFound: false,
					OnTypePos:       300,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
				},
				// Writer and Closer not loaded
			},
			expected: []MissingInterfaceReport{
				{
					InterfaceName: "Writer",
					PackageName:   "io",
					TypeName:      "MyWriter",
					Pos:           200,
				},
				{
					InterfaceName: "Closer",
					PackageName:   "io",
					TypeName:      "MyCloser",
					Pos:           300,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMissingInterfaces(tt.annotations, tt.interfaces)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========== Tests for findMissingMethods ==========

func TestFindMissingMethods(t *testing.T) {
	tests := []struct {
		name        string
		annotations []ImplementsAnnotation
		interfaces  []*InterfaceModel
		types       []*TypeModel
		expected    []MissingMethodsReport
	}{
		{
			name: "type implements interface fully",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "io",
					PackageFullPath: "io",
					IsPointer:       true,
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
					Methods: []InterfaceMethod{
						{
							Name: "Read",
							Inputs: []InterfaceType{
								{TypeName: "[]byte"},
							},
							Outputs: []InterfaceType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
				},
			},
			types: []*TypeModel{
				{
					Name: "MyReader",
					Methods: []TypeMethod{
						{
							Name:              "Read",
							ReceiverIsPointer: true,
							Inputs: []MethodType{
								{TypeName: "[]byte"},
							},
							Outputs: []MethodType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
				},
			},
			expected: []MissingMethodsReport{},
		},
		{
			name: "type missing method",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "io",
					PackageFullPath: "io",
					IsPointer:       true,
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
					Methods: []InterfaceMethod{
						{
							Name: "Read",
							Inputs: []InterfaceType{
								{TypeName: "[]byte"},
							},
							Outputs: []InterfaceType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
						{
							Name:    "Close",
							Inputs:  []InterfaceType{},
							Outputs: []InterfaceType{{TypeName: "error"}},
						},
					},
				},
			},
			types: []*TypeModel{
				{
					Name: "MyReader",
					Methods: []TypeMethod{
						{
							Name:              "Read",
							ReceiverIsPointer: true,
							Inputs: []MethodType{
								{TypeName: "[]byte"},
							},
							Outputs: []MethodType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
						// Missing Close method
					},
				},
			},
			expected: []MissingMethodsReport{
				{
					InterfaceName: "Reader",
					PackageName:   "io",
					TypeName:      "MyReader",
					Methods: []InterfaceMethod{
						{
							Name:    "Close",
							Inputs:  []InterfaceType{},
							Outputs: []InterfaceType{{TypeName: "error"}},
						},
					},
					Pos: 100,
				},
			},
		},
		{
			name: "type has wrong signature",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "io",
					PackageFullPath: "io",
					IsPointer:       true,
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
					Methods: []InterfaceMethod{
						{
							Name: "Read",
							Inputs: []InterfaceType{
								{TypeName: "[]byte"},
							},
							Outputs: []InterfaceType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
				},
			},
			types: []*TypeModel{
				{
					Name: "MyReader",
					Methods: []TypeMethod{
						{
							Name:              "Read",
							ReceiverIsPointer: true,
							Inputs: []MethodType{
								{TypeName: "string"}, // Wrong type!
							},
							Outputs: []MethodType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
				},
			},
			expected: []MissingMethodsReport{
				{
					InterfaceName: "Reader",
					PackageName:   "io",
					TypeName:      "MyReader",
					Methods: []InterfaceMethod{
						{
							Name: "Read",
							Inputs: []InterfaceType{
								{TypeName: "[]byte"},
							},
							Outputs: []InterfaceType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
					Pos: 100,
				},
			},
		},
		{
			name: "value receiver required but only pointer receiver exists",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "io",
					PackageFullPath: "io",
					IsPointer:       false, // Requires value receiver
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
					Methods: []InterfaceMethod{
						{
							Name: "Read",
							Inputs: []InterfaceType{
								{TypeName: "[]byte"},
							},
							Outputs: []InterfaceType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
				},
			},
			types: []*TypeModel{
				{
					Name: "MyReader",
					Methods: []TypeMethod{
						{
							Name:              "Read",
							ReceiverIsPointer: true, // Only pointer receiver
							Inputs: []MethodType{
								{TypeName: "[]byte"},
							},
							Outputs: []MethodType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
				},
			},
			expected: []MissingMethodsReport{
				{
					InterfaceName: "Reader",
					PackageName:   "io",
					TypeName:      "MyReader",
					Methods: []InterfaceMethod{
						{
							Name: "Read",
							Inputs: []InterfaceType{
								{TypeName: "[]byte"},
							},
							Outputs: []InterfaceType{
								{TypeName: "int"},
								{TypeName: "error"},
							},
						},
					},
					Pos: 100,
				},
			},
		},
		{
			name: "skip when package not found",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "Reader",
					PackageName:     "http",
					PackageFullPath: "",
					IsPointer:       true,
					PackageNotFound: true,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{},
			types:      []*TypeModel{},
			expected:   []MissingMethodsReport{},
		},
		{
			name: "skip when interface not found",
			annotations: []ImplementsAnnotation{
				{
					OnType:          "MyReader",
					InterfaceName:   "NonExistent",
					PackageName:     "io",
					PackageFullPath: "io",
					IsPointer:       true,
					PackageNotFound: false,
					OnTypePos:       100,
				},
			},
			interfaces: []*InterfaceModel{
				{
					Name:    "Reader",
					Package: "io",
					Methods: []InterfaceMethod{},
				},
			},
			types: []*TypeModel{
				{
					Name:    "MyReader",
					Methods: []TypeMethod{},
				},
			},
			expected: []MissingMethodsReport{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMissingMethods(tt.annotations, tt.interfaces, tt.types)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ========== Tests for helper functions ==========

func TestSignaturesMatch(t *testing.T) {
	tests := []struct {
		name        string
		typeMethod  TypeMethod
		ifaceMethod InterfaceMethod
		expected    bool
	}{
		{
			name: "exact match",
			typeMethod: TypeMethod{
				Name: "Read",
				Inputs: []MethodType{
					{TypeName: "[]byte"},
				},
				Outputs: []MethodType{
					{TypeName: "int"},
					{TypeName: "error"},
				},
			},
			ifaceMethod: InterfaceMethod{
				Name: "Read",
				Inputs: []InterfaceType{
					{TypeName: "[]byte"},
				},
				Outputs: []InterfaceType{
					{TypeName: "int"},
					{TypeName: "error"},
				},
			},
			expected: true,
		},
		{
			name: "different input count",
			typeMethod: TypeMethod{
				Name: "Process",
				Inputs: []MethodType{
					{TypeName: "int"},
				},
				Outputs: []MethodType{},
			},
			ifaceMethod: InterfaceMethod{
				Name: "Process",
				Inputs: []InterfaceType{
					{TypeName: "int"},
					{TypeName: "string"},
				},
				Outputs: []InterfaceType{},
			},
			expected: false,
		},
		{
			name: "different output count",
			typeMethod: TypeMethod{
				Name:   "Get",
				Inputs: []MethodType{},
				Outputs: []MethodType{
					{TypeName: "string"},
				},
			},
			ifaceMethod: InterfaceMethod{
				Name:   "Get",
				Inputs: []InterfaceType{},
				Outputs: []InterfaceType{
					{TypeName: "string"},
					{TypeName: "error"},
				},
			},
			expected: false,
		},
		{
			name: "different input type",
			typeMethod: TypeMethod{
				Name: "Write",
				Inputs: []MethodType{
					{TypeName: "string"},
				},
				Outputs: []MethodType{},
			},
			ifaceMethod: InterfaceMethod{
				Name: "Write",
				Inputs: []InterfaceType{
					{TypeName: "[]byte"},
				},
				Outputs: []InterfaceType{},
			},
			expected: false,
		},
		{
			name: "pointer mismatch",
			typeMethod: TypeMethod{
				Name: "Process",
				Inputs: []MethodType{
					{TypeName: "int", IsPointer: false},
				},
				Outputs: []MethodType{},
			},
			ifaceMethod: InterfaceMethod{
				Name: "Process",
				Inputs: []InterfaceType{
					{TypeName: "int", IsPointer: true},
				},
				Outputs: []InterfaceType{},
			},
			expected: false,
		},
		{
			name: "variadic match",
			typeMethod: TypeMethod{
				Name: "Printf",
				Inputs: []MethodType{
					{TypeName: "string"},
					{TypeName: "interface{}", IsVariadic: true},
				},
				Outputs: []MethodType{},
			},
			ifaceMethod: InterfaceMethod{
				Name: "Printf",
				Inputs: []InterfaceType{
					{TypeName: "string"},
					{TypeName: "interface{}", IsVariadic: true},
				},
				Outputs: []InterfaceType{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := signaturesMatch(tt.typeMethod, tt.ifaceMethod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypesMatch(t *testing.T) {
	tests := []struct {
		name     string
		t1       MethodType
		t2       InterfaceType
		expected bool
	}{
		{
			name:     "exact match",
			t1:       MethodType{TypeName: "int", TypePackage: "", IsPointer: false, IsVariadic: false},
			t2:       InterfaceType{TypeName: "int", TypePackage: "", IsPointer: false, IsVariadic: false},
			expected: true,
		},
		{
			name:     "different type name",
			t1:       MethodType{TypeName: "int", TypePackage: "", IsPointer: false, IsVariadic: false},
			t2:       InterfaceType{TypeName: "string", TypePackage: "", IsPointer: false, IsVariadic: false},
			expected: false,
		},
		{
			name:     "different package",
			t1:       MethodType{TypeName: "Reader", TypePackage: "io", IsPointer: false, IsVariadic: false},
			t2:       InterfaceType{TypeName: "Reader", TypePackage: "bufio", IsPointer: false, IsVariadic: false},
			expected: false,
		},
		{
			name:     "different pointer status",
			t1:       MethodType{TypeName: "int", TypePackage: "", IsPointer: true, IsVariadic: false},
			t2:       InterfaceType{TypeName: "int", TypePackage: "", IsPointer: false, IsVariadic: false},
			expected: false,
		},
		{
			name:     "different variadic status",
			t1:       MethodType{TypeName: "string", TypePackage: "", IsPointer: false, IsVariadic: true},
			t2:       InterfaceType{TypeName: "string", TypePackage: "", IsPointer: false, IsVariadic: false},
			expected: false,
		},
		{
			name:     "match with package",
			t1:       MethodType{TypeName: "Context", TypePackage: "context", IsPointer: false, IsVariadic: false},
			t2:       InterfaceType{TypeName: "Context", TypePackage: "context", IsPointer: false, IsVariadic: false},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := typesMatch(&tt.t1, &tt.t2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
