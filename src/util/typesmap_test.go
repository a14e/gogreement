package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypesMap_Add(t *testing.T) {
	tm := NewTypesMap()

	tm.Add("", "MyType")
	tm.Add("io", "Reader")
	tm.Add("io", "Writer")

	assert.True(t, tm.Contains("", "MyType"))
	assert.True(t, tm.Contains("io", "Reader"))
	assert.True(t, tm.Contains("io", "Writer"))
	assert.False(t, tm.Contains("", "Reader"))
	assert.False(t, tm.Contains("io", "MyType"))
}

func TestTypesMap_Contains(t *testing.T) {
	tm := NewTypesMap()
	tm.Add("pkg1", "Type1")
	tm.Add("pkg2", "Type2")

	tests := []struct {
		name     string
		pkgPath  string
		typeName string
		expected bool
	}{
		{"exists in pkg1", "pkg1", "Type1", true},
		{"exists in pkg2", "pkg2", "Type2", true},
		{"wrong package", "pkg1", "Type2", false},
		{"wrong type", "pkg2", "Type1", false},
		{"non-existent package", "pkg3", "Type1", false},
		{"non-existent type", "pkg1", "Type3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.Contains(tt.pkgPath, tt.typeName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTypesMap_Len(t *testing.T) {
	tm := NewTypesMap()

	assert.Equal(t, 0, tm.Len())

	tm.Add("", "Type1")
	assert.Equal(t, 1, tm.Len())

	tm.Add("", "Type2")
	assert.Equal(t, 2, tm.Len())

	tm.Add("pkg", "Type3")
	assert.Equal(t, 3, tm.Len())
}

func TestTypesMap_DuplicateAdd(t *testing.T) {
	tm := NewTypesMap()

	tm.Add("pkg", "Type1")
	tm.Add("pkg", "Type1")
	tm.Add("pkg", "Type1")

	assert.Equal(t, 1, len(tm["pkg"]))
	assert.True(t, tm.Contains("pkg", "Type1"))
}

func TestTypesMap_CurrentPackage(t *testing.T) {
	tm := NewTypesMap()

	// Add type to current package (empty string key)
	tm.Add("", "MyType")

	// Should be found with empty string
	assert.True(t, tm.Contains("", "MyType"))

	// Should NOT be found with actual package path
	assert.False(t, tm.Contains("myapp/pkg", "MyType"))
}

func TestTypesMap_MultiplePackages(t *testing.T) {
	tm := NewTypesMap()

	tm.Add("", "LocalType")
	tm.Add("io", "Reader")
	tm.Add("context", "Context")

	assert.Equal(t, 3, tm.Len())

	assert.True(t, tm.Contains("", "LocalType"))
	assert.True(t, tm.Contains("io", "Reader"))
	assert.True(t, tm.Contains("context", "Context"))
}

func TestTypesMap_EmptyMap(t *testing.T) {
	tm := NewTypesMap()

	assert.Equal(t, 0, tm.Len())
	assert.False(t, tm.Contains("", "AnyType"))
	assert.False(t, tm.Contains("pkg", "AnyType"))
}
