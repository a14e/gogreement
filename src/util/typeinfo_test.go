package util

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTypeInfo(t *testing.T) {
	t.Run("nil type", func(t *testing.T) {
		result := ExtractTypeInfo(nil)
		assert.Nil(t, result)
	})

	t.Run("basic type", func(t *testing.T) {
		// Basic types (int, string, etc.) are not Named types
		basicType := types.Typ[types.Int]
		result := ExtractTypeInfo(basicType)
		assert.Nil(t, result)
	})

	t.Run("named type", func(t *testing.T) {
		// Create a named type with package
		pkg := types.NewPackage("github.com/user/pkg", "pkg")
		typeName := types.NewTypeName(0, pkg, "MyStruct", nil)
		namedType := types.NewNamed(typeName, types.NewStruct(nil, nil), nil)

		result := ExtractTypeInfo(namedType)
		assert.NotNil(t, result)
		assert.Equal(t, "MyStruct", result.TypeName)
		assert.Equal(t, "github.com/user/pkg", result.PkgPath)
	})

	t.Run("pointer to named type", func(t *testing.T) {
		// Create a named type
		pkg := types.NewPackage("github.com/user/pkg", "pkg")
		typeName := types.NewTypeName(0, pkg, "MyStruct", nil)
		namedType := types.NewNamed(typeName, types.NewStruct(nil, nil), nil)

		// Create a pointer to it
		ptrType := types.NewPointer(namedType)

		result := ExtractTypeInfo(ptrType)
		assert.NotNil(t, result)
		assert.Equal(t, "MyStruct", result.TypeName)
		assert.Equal(t, "github.com/user/pkg", result.PkgPath)
	})

	t.Run("builtin type with no package", func(t *testing.T) {
		// Create a named type without package (builtin)
		typeName := types.NewTypeName(0, nil, "error", nil)
		namedType := types.NewNamed(typeName, types.NewStruct(nil, nil), nil)

		result := ExtractTypeInfo(namedType)
		assert.Nil(t, result) // No package, so should return nil
	})
}
