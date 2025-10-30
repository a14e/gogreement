package util

import "go/types"

// FIXME do we need this?

// TypeInfo contains information about a Go type extracted from types.Type
// @immutable
// @constructor ExtractTypeInfo
type TypeInfo struct {
	// TypeName is the name of the type (e.g., "MyStruct")
	TypeName string

	// PkgPath is the full package path (e.g., "github.com/user/pkg")
	// Empty for builtin types
	PkgPath string
}

// ExtractTypeInfo extracts type name and package path from a types.Type
// Returns nil if the type is not a named type or has no package
func ExtractTypeInfo(t types.Type) *TypeInfo {
	if t == nil {
		return nil
	}

	// Remove pointer if present
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Get named type
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}

	typeName := named.Obj().Name()
	pkg := named.Obj().Pkg()
	if pkg == nil {
		return nil
	}

	return &TypeInfo{
		TypeName: typeName,
		PkgPath:  pkg.Path(),
	}
}

// ExtractTypeName extracts just the type name from a types.Type
// Returns empty string if the type is not a named type
func ExtractTypeName(t types.Type) string {
	if t == nil {
		return ""
	}

	// Remove pointer if present
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	// Get named type
	named, ok := t.(*types.Named)
	if !ok {
		return ""
	}

	return named.Obj().Name()
}
