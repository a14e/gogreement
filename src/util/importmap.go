package util

import (
	"go/ast"
	"go/types"
	"strings"
)

// Import represents a single import from AST
// @immutable
type Import struct {
	Alias       string // explicit alias (if present) or empty
	FullPath    string // full import path like "io" or "github.com/user/pkg"
	PackageName string // actual package name from the code (e.g., "importmap" for "goagreement/src/util")
}

// ImportMap is a collection of imports with lookup methods
type ImportMap []Import

// Add adds an import spec to the map
// If pkg is provided, the actual package name will be stored
// If pkg is nil, only path information will be stored
func (m *ImportMap) Add(spec *ast.ImportSpec, pkg *types.Package) {
	if spec == nil || spec.Path == nil {
		return
	}

	fullPath := strings.Trim(spec.Path.Value, `"`)

	var alias string
	if spec.Name != nil {
		// Explicit alias: import foo "path"
		alias = spec.Name.Name
	}

	var packageName string
	if pkg != nil {
		packageName = pkg.Name()
	}

	*m = append(*m, Import{
		Alias:       alias,
		FullPath:    fullPath,
		PackageName: packageName,
	})
}

// Find searches for an import by short name with the following priority:
// 1. Explicit alias (highest priority)
// 2. Package name (actual name from package declaration)
// 3. Exact match (e.g., "io" matches "io")
// 4. Path component match (e.g., "bar" matches "foo/bar")
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

	// Priority 2: Search by actual package name
	for i := range *m {
		imp := &(*m)[i]
		if imp.PackageName != "" && imp.PackageName == shortName {
			return imp
		}
	}

	// Priority 3: Search for exact match
	// "io" should match "io", not "github.com/foo/io"
	for i := range *m {
		imp := &(*m)[i]
		if imp.FullPath == shortName {
			return imp
		}
	}

	// Priority 4: Fallback to path component match
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
