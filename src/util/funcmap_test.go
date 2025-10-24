package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncMap_Add(t *testing.T) {
	fm := NewTypeFuncRegistry()

	fm.Add("", "NewMyType", "MyType")
	fm.Add("io", "NewReader", "Reader")
	fm.Add("io", "NewWriter", "Writer")

	assert.True(t, fm.Match("", "NewMyType", "MyType"))
	assert.True(t, fm.Match("io", "NewReader", "Reader"))
	assert.True(t, fm.Match("io", "NewWriter", "Writer"))
}

func TestFuncMap_Match(t *testing.T) {
	fm := NewTypeFuncRegistry()
	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "NewConfig", "Config")

	tests := []struct {
		name          string
		pkgPath       string
		funcName      string
		expectedValue string
		expected      bool
	}{
		{"exact match", "pkg", "NewUser", "User", true},
		{"wrong value", "pkg", "NewUser", "Config", false},
		{"wrong func", "pkg", "NewAdmin", "User", false},
		{"wrong package", "other", "NewUser", "User", false},
		{"non-existent", "pkg", "NonExistent", "User", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.Match(tt.pkgPath, tt.funcName, tt.expectedValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncMap_MultipleConstructors(t *testing.T) {
	fm := NewTypeFuncRegistry()

	// Add multiple constructors for same type
	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "CreateUser", "User")
	fm.Add("pkg", "MakeUser", "User")

	// All should match
	assert.True(t, fm.Match("pkg", "NewUser", "User"))
	assert.True(t, fm.Match("pkg", "CreateUser", "User"))
	assert.True(t, fm.Match("pkg", "MakeUser", "User"))

	// Should have 3 constructors
	constructors := fm.GetFuncs("pkg", "User")
	assert.Equal(t, 3, len(constructors))
	assert.Contains(t, constructors, "NewUser")
	assert.Contains(t, constructors, "CreateUser")
	assert.Contains(t, constructors, "MakeUser")
}

func TestFuncMap_GetConstructors(t *testing.T) {
	fm := NewTypeFuncRegistry()
	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "NewConfig", "Config")
	fm.Add("pkg", "NewDefaultConfig", "Config")

	tests := []struct {
		name     string
		pkgPath  string
		typeName string
		expected []string
	}{
		{
			name:     "single constructor",
			pkgPath:  "pkg",
			typeName: "User",
			expected: []string{"NewUser"},
		},
		{
			name:     "multiple constructors",
			pkgPath:  "pkg",
			typeName: "Config",
			expected: []string{"NewConfig", "NewDefaultConfig"},
		},
		{
			name:     "non-existent type",
			pkgPath:  "pkg",
			typeName: "NonExistent",
			expected: nil,
		},
		{
			name:     "non-existent package",
			pkgPath:  "other",
			typeName: "User",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.GetFuncs(tt.pkgPath, tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncMap_HasType(t *testing.T) {
	fm := NewTypeFuncRegistry()
	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "NewConfig", "Config")

	tests := []struct {
		name     string
		pkgPath  string
		typeName string
		expected bool
	}{
		{"has type User", "pkg", "User", true},
		{"has type Config", "pkg", "Config", true},
		{"no type Admin", "pkg", "Admin", false},
		{"wrong package", "other", "User", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.HasType(tt.pkgPath, tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFuncMap_Len(t *testing.T) {
	fm := NewTypeFuncRegistry()

	assert.Equal(t, 0, fm.Len())

	fm.Add("", "Func1", "Type1")
	assert.Equal(t, 1, fm.Len())

	fm.Add("", "Func2", "Type2")
	assert.Equal(t, 2, fm.Len())

	fm.Add("pkg", "Func3", "Type3")
	assert.Equal(t, 3, fm.Len())

	// Add second constructor for Type1
	fm.Add("", "Func4", "Type1")
	assert.Equal(t, 4, fm.Len())
}

func TestFuncMap_Empty(t *testing.T) {
	fm := NewTypeFuncRegistry()

	assert.True(t, fm.Empty())

	fm.Add("", "Func1", "Type1")
	assert.False(t, fm.Empty())

	fm.Add("pkg", "Func2", "Type2")
	assert.False(t, fm.Empty())
}

func TestFuncMap_DuplicateAdd(t *testing.T) {
	fm := NewTypeFuncRegistry()

	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "NewUser", "User") // Duplicate

	// Should have both entries
	constructors := fm.GetFuncs("pkg", "User")
	assert.Equal(t, 2, len(constructors))

	// Match should still work
	assert.True(t, fm.Match("pkg", "NewUser", "User"))
}

func TestFuncMap_CurrentPackage(t *testing.T) {
	fm := NewTypeFuncRegistry()

	// Add func to current package (empty string key)
	fm.Add("", "NewMyType", "MyType")

	// Should be found with empty string
	assert.True(t, fm.Match("", "NewMyType", "MyType"))
	assert.Equal(t, []string{"NewMyType"}, fm.GetFuncs("", "MyType"))

	// Should NOT be found with actual package path
	assert.False(t, fm.Match("myapp/pkg", "NewMyType", "MyType"))
	assert.Nil(t, fm.GetFuncs("myapp/pkg", "MyType"))
}

func TestFuncMap_MultiplePackages(t *testing.T) {
	fm := NewTypeFuncRegistry()

	fm.Add("", "NewLocal", "Local")
	fm.Add("io", "NewReader", "Reader")
	fm.Add("context", "NewContext", "Context")

	assert.Equal(t, 3, fm.Len())

	assert.True(t, fm.Match("", "NewLocal", "Local"))
	assert.True(t, fm.Match("io", "NewReader", "Reader"))
	assert.True(t, fm.Match("context", "NewContext", "Context"))
}

func TestFuncMap_EmptyMap(t *testing.T) {
	fm := NewTypeFuncRegistry()

	assert.Equal(t, 0, fm.Len())
	assert.False(t, fm.Match("", "AnyFunc", "AnyType"))
	assert.False(t, fm.Match("pkg", "AnyFunc", "AnyType"))
	assert.False(t, fm.HasType("", "AnyType"))
	assert.Nil(t, fm.GetFuncs("", "AnyType"))
}

func TestFuncMap_BackwardCompatibility(t *testing.T) {
	// This test ensures the API is backward compatible
	// Old usage: fm.Add(pkgPath, funcName, typeName) -> fm.Match(pkgPath, funcName, typeName)

	fm := NewTypeFuncRegistry()

	// Old style usage
	fm.Add("pkg", "NewUser", "User")

	// Old style check - should still work
	assert.True(t, fm.Match("pkg", "NewUser", "User"))
	assert.False(t, fm.Match("pkg", "OtherFunc", "User"))

	// New functionality - get all constructors
	constructors := fm.GetFuncs("pkg", "User")
	assert.Equal(t, []string{"NewUser"}, constructors)
}
