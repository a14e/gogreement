package importmap

import (
	"go/ast"
	"strings"
)

// Import represents a single import from AST
type Import struct {
	Alias    string // explicit alias (if present) or empty
	FullPath string // full import path like "io" or "github.com/user/pkg"
}

// ImportMap is a collection of imports with lookup methods
type ImportMap []Import

// Add adds an import spec to the map
func (m *ImportMap) Add(spec *ast.ImportSpec) {
	if spec == nil || spec.Path == nil {
		return
	}

	fullPath := strings.Trim(spec.Path.Value, `"`)

	var alias string
	if spec.Name != nil {
		// Explicit alias: import foo "path"
		alias = spec.Name.Name
	}

	*m = append(*m, Import{
		Alias:    alias,
		FullPath: fullPath,
	})
}

// Find searches for an import by short name with the following priority:
// 1. Explicit alias (highest priority)
// 2. Exact match (e.g., "io" matches "io")
// 3. Path component match (e.g., "bar" matches "foo/bar")
// Returns nil if not found
func (m *ImportMap) Find(shortName string) *Import {
	if shortName == "" {
		return nil
	}

	// Priority 1: Search by explicit alias first
	for i := range *m {
		imp := &(*m)[i]
		if imp.Alias != "" && imp.Alias == shortName {
			return imp
		}
	}

	// Priority 2: Search for exact match
	// "io" should match "io", not "github.com/foo/io"
	for i := range *m {
		imp := &(*m)[i]
		if imp.FullPath == shortName {
			return imp
		}
	}

	// Priority 3: Fallback to path component match
	// "bar" matches "foo/bar"
	for i := range *m {
		imp := &(*m)[i]

		if matchesPathComponentWithSlash(imp.FullPath, shortName) {
			return imp
		}
	}

	// Not found
	return nil
}

// matchesPathComponentWithSlash checks if fullPath ends with "/shortName"
// This is only for cases where we didn't find an exact match
// "bar" matches "foo/bar" ✓
// "io" does NOT match "fooio" ✗
func matchesPathComponentWithSlash(fullPath, shortName string) bool {
	// We need at least "/X" where X is shortName
	minLen := len(shortName) + 1 // +1 for the "/"
	if len(fullPath) < minLen {
		return false
	}

	// Check that fullPath ends with shortName
	startPos := len(fullPath) - len(shortName)
	if fullPath[startPos:] != shortName {
		return false
	}

	// Check that shortName is preceded by '/'
	return fullPath[startPos-1] == '/'
}
