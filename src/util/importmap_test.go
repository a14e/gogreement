package importmap

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			name:         "find by suffix - exact match",
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
	t.Run("alias has highest priority", func(t *testing.T) {
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
	})

	t.Run("exact match has priority over path component", func(t *testing.T) {
		importMap := &ImportMap{}

		// Add imports in this order
		importMap.Add(&ast.ImportSpec{
			Path: &ast.BasicLit{Value: `"github.com/foo/io"`}, // matches as path component
		})
		importMap.Add(&ast.ImportSpec{
			Path: &ast.BasicLit{Value: `"myio"`}, // doesn't match
		})
		importMap.Add(&ast.ImportSpec{
			Path: &ast.BasicLit{Value: `"io"`}, // exact match - should win!
		})

		// Should find exact match "io", not "github.com/foo/io"
		result := importMap.Find("io")
		require.NotNil(t, result)
		assert.Equal(t, "io", result.FullPath)
	})

	t.Run("path component match when no exact match", func(t *testing.T) {
		importMap := &ImportMap{}

		importMap.Add(&ast.ImportSpec{
			Path: &ast.BasicLit{Value: `"github.com/foo/bar"`},
		})
		importMap.Add(&ast.ImportSpec{
			Path: &ast.BasicLit{Value: `"mybar"`}, // doesn't match
		})

		// Should find path component match
		result := importMap.Find("bar")
		require.NotNil(t, result)
		assert.Equal(t, "github.com/foo/bar", result.FullPath)
	})
}

func TestImportMapAddNil(t *testing.T) {
	importMap := &ImportMap{}

	// Should not panic or add anything
	importMap.Add(nil)
	assert.Empty(t, *importMap)

	// Add spec with nil path
	importMap.Add(&ast.ImportSpec{
		Path: nil,
	})
	assert.Empty(t, *importMap)
}

func TestMatchesPathComponentWithSlash(t *testing.T) {
	tests := []struct {
		name      string
		fullPath  string
		shortName string
		expected  bool
	}{
		{
			name:      "exact match should NOT match here",
			fullPath:  "io",
			shortName: "io",
			expected:  false, // exact matches are handled separately
		},
		{
			name:      "path with slash",
			fullPath:  "foo/bar",
			shortName: "bar",
			expected:  true,
		},
		{
			name:      "deep path",
			fullPath:  "github.com/user/project/bar",
			shortName: "bar",
			expected:  true,
		},
		{
			name:      "partial suffix should NOT match",
			fullPath:  "myio",
			shortName: "io",
			expected:  false,
		},
		{
			name:      "partial suffix with more text",
			fullPath:  "foobar",
			shortName: "bar",
			expected:  false,
		},
		{
			name:      "longer shortName",
			fullPath:  "io",
			shortName: "ioio",
			expected:  false,
		},
		{
			name:      "context stdlib",
			fullPath:  "context",
			shortName: "context",
			expected:  false, // exact match, not slash-prefixed
		},
		{
			name:      "github path",
			fullPath:  "github.com/example/baz",
			shortName: "baz",
			expected:  true,
		},
		{
			name:      "not a component boundary",
			fullPath:  "github.com/foobar",
			shortName: "bar",
			expected:  false,
		},
		{
			name:      "slash prefixed io",
			fullPath:  "github.com/foo/io",
			shortName: "io",
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesPathComponentWithSlash(tt.fullPath, tt.shortName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestImportMapFindWithSlash(t *testing.T) {
	importMap := &ImportMap{}

	importMap.Add(&ast.ImportSpec{
		Path: &ast.BasicLit{Value: `"github.com/user/project/bar"`},
	})

	// Should find by last component
	result := importMap.Find("bar")
	require.NotNil(t, result)
	assert.Equal(t, "github.com/user/project/bar", result.FullPath)
}
