package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuncMap_Add(t *testing.T) {
	fm := NewFuncMap()

	fm.Add("", "NewMyType", "MyType")
	fm.Add("io", "NewReader", "Reader")
	fm.Add("io", "NewWriter", "Writer")

	assert.True(t, fm.Match("", "NewMyType", "MyType"))
	assert.True(t, fm.Match("io", "NewReader", "Reader"))
	assert.True(t, fm.Match("io", "NewWriter", "Writer"))
}

func TestFuncMap_Match(t *testing.T) {
	fm := NewFuncMap()
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

func TestFuncMap_Len(t *testing.T) {
	fm := NewFuncMap()

	assert.Equal(t, 0, fm.Len())

	fm.Add("", "Func1", "Type1")
	assert.Equal(t, 1, fm.Len())

	fm.Add("", "Func2", "Type2")
	assert.Equal(t, 2, fm.Len())

	fm.Add("pkg", "Func3", "Type3")
	assert.Equal(t, 3, fm.Len())
}

func TestFuncMap_DuplicateAdd(t *testing.T) {
	fm := NewFuncMap()

	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "NewUser", "User")
	fm.Add("pkg", "NewUser", "Admin") // Overwrite

	assert.Equal(t, 1, len(fm["pkg"]))
	assert.True(t, fm.Match("pkg", "NewUser", "Admin")) // Last write wins
	assert.False(t, fm.Match("pkg", "NewUser", "User"))
}

func TestFuncMap_CurrentPackage(t *testing.T) {
	fm := NewFuncMap()

	// Add func to current package (empty string key)
	fm.Add("", "NewMyType", "MyType")

	// Should be found with empty string
	assert.True(t, fm.Match("", "NewMyType", "MyType"))

	// Should NOT be found with actual package path
	assert.False(t, fm.Match("myapp/pkg", "NewMyType", "MyType"))
}

func TestFuncMap_MultiplePackages(t *testing.T) {
	fm := NewFuncMap()

	fm.Add("", "NewLocal", "Local")
	fm.Add("io", "NewReader", "Reader")
	fm.Add("context", "NewContext", "Context")

	assert.Equal(t, 3, fm.Len())

	assert.True(t, fm.Match("", "NewLocal", "Local"))
	assert.True(t, fm.Match("io", "NewReader", "Reader"))
	assert.True(t, fm.Match("context", "NewContext", "Context"))
}

func TestFuncMap_EmptyMap(t *testing.T) {
	fm := NewFuncMap()

	assert.Equal(t, 0, fm.Len())
	assert.False(t, fm.Match("", "AnyFunc", "AnyType"))
	assert.False(t, fm.Match("pkg", "AnyFunc", "AnyType"))
}
