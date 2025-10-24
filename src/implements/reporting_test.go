package implements

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMethodSignature(t *testing.T) {
	tests := []struct {
		name     string
		method   InterfaceMethod
		expected string
	}{
		{
			name: "no params no return",
			method: InterfaceMethod{
				Name:    "Close",
				Inputs:  []InterfaceType{},
				Outputs: []InterfaceType{},
			},
			expected: "Close()",
		},
		{
			name: "single param no return",
			method: InterfaceMethod{
				Name: "Write",
				Inputs: []InterfaceType{
					{TypeName: "[]byte", TypePackage: ""},
				},
				Outputs: []InterfaceType{},
			},
			expected: "Write([]byte)",
		},
		{
			name: "single param single return",
			method: InterfaceMethod{
				Name:   "String",
				Inputs: []InterfaceType{},
				Outputs: []InterfaceType{
					{TypeName: "string", TypePackage: ""},
				},
			},
			expected: "String() string",
		},
		{
			name: "multiple params multiple returns",
			method: InterfaceMethod{
				Name: "Read",
				Inputs: []InterfaceType{
					{TypeName: "[]byte", TypePackage: ""},
				},
				Outputs: []InterfaceType{
					{TypeName: "int", TypePackage: ""},
					{TypeName: "error", TypePackage: ""},
				},
			},
			expected: "Read([]byte) (int, error)",
		},
		{
			name: "pointer types",
			method: InterfaceMethod{
				Name: "Process",
				Inputs: []InterfaceType{
					{TypeName: "Request", TypePackage: "", IsPointer: true},
				},
				Outputs: []InterfaceType{
					{TypeName: "Response", TypePackage: "", IsPointer: true},
					{TypeName: "error", TypePackage: ""},
				},
			},
			expected: "Process(*Request) (*Response, error)",
		},
		{
			name: "variadic params",
			method: InterfaceMethod{
				Name: "Printf",
				Inputs: []InterfaceType{
					{TypeName: "string", TypePackage: ""},
					{TypeName: "interface{}", TypePackage: "", IsVariadic: true},
				},
				Outputs: []InterfaceType{
					{TypeName: "int", TypePackage: ""},
					{TypeName: "error", TypePackage: ""},
				},
			},
			expected: "Printf(string, ...interface{}) (int, error)",
		},
		{
			name: "package qualified types",
			method: InterfaceMethod{
				Name: "ReadFile",
				Inputs: []InterfaceType{
					{TypeName: "string", TypePackage: ""},
				},
				Outputs: []InterfaceType{
					{TypeName: "File", TypePackage: "github.com/user/pkg/fs", IsPointer: true},
					{TypeName: "error", TypePackage: ""},
				},
			},
			expected: "ReadFile(string) (*fs.File, error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMethodSignature(tt.method)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatType(t *testing.T) {
	tests := []struct {
		name     string
		typ      InterfaceType
		expected string
	}{
		{
			name:     "simple type",
			typ:      InterfaceType{TypeName: "int", TypePackage: ""},
			expected: "int",
		},
		{
			name:     "pointer type",
			typ:      InterfaceType{TypeName: "string", TypePackage: "", IsPointer: true},
			expected: "*string",
		},
		{
			name:     "slice type",
			typ:      InterfaceType{TypeName: "[]byte", TypePackage: ""},
			expected: "[]byte",
		},
		{
			name:     "variadic type",
			typ:      InterfaceType{TypeName: "interface{}", TypePackage: "", IsVariadic: true},
			expected: "...interface{}",
		},
		{
			name:     "package qualified type",
			typ:      InterfaceType{TypeName: "Reader", TypePackage: "io"},
			expected: "io.Reader",
		},
		{
			name:     "package qualified type with path",
			typ:      InterfaceType{TypeName: "Context", TypePackage: "context"},
			expected: "context.Context",
		},
		{
			name:     "pointer to package qualified type",
			typ:      InterfaceType{TypeName: "File", TypePackage: "github.com/user/pkg/fs", IsPointer: true},
			expected: "*fs.File",
		},
		{
			name:     "variadic package qualified type",
			typ:      InterfaceType{TypeName: "Option", TypePackage: "github.com/user/pkg/opts", IsVariadic: true},
			expected: "...opts.Option",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatType(tt.typ)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTypeList(t *testing.T) {
	tests := []struct {
		name     string
		types    []InterfaceType
		expected string
	}{
		{
			name:     "empty list",
			types:    []InterfaceType{},
			expected: "",
		},
		{
			name: "single type",
			types: []InterfaceType{
				{TypeName: "int", TypePackage: ""},
			},
			expected: "int",
		},
		{
			name: "multiple types",
			types: []InterfaceType{
				{TypeName: "int", TypePackage: ""},
				{TypeName: "string", TypePackage: ""},
				{TypeName: "error", TypePackage: ""},
			},
			expected: "int, string, error",
		},
		{
			name: "mixed types with pointers and packages",
			types: []InterfaceType{
				{TypeName: "Context", TypePackage: "context"},
				{TypeName: "Request", TypePackage: "", IsPointer: true},
				{TypeName: "[]byte", TypePackage: ""},
			},
			expected: "context.Context, *Request, []byte",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTypeList(tt.types)
			assert.Equal(t, tt.expected, result)
		})
	}
}
