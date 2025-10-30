package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttachmentsMap_BasicOperations(t *testing.T) {
	am := &AttachmentsMap{}

	// Test nil safety
	assert.False(t, am.HasPkgAttachment("pkg1", "tag1"))
	assert.False(t, am.HasPkgFunctionAttachment("pkg1", "func1", "tag1"))
	assert.False(t, am.HasPkgTypeAttachment("pkg1", "Type1", "tag1"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg1", "Type1", "field1", "tag1"))
	assert.False(t, am.HasPkgTypeMethodAttachment("pkg1", "Type1", "method1", "tag1"))

	// Add package attachment
	am.AddPkgAttachment("pkg1", "package-tag")
	assert.True(t, am.HasPkgAttachment("pkg1", "package-tag"))
	assert.False(t, am.HasPkgAttachment("pkg1", "nonexistent"))
	assert.False(t, am.HasPkgAttachment("pkg2", "package-tag"))

	// Add function attachment
	am.AddPkgFunctionAttachment("pkg1", "Function1", "function-tag")
	assert.True(t, am.HasPkgFunctionAttachment("pkg1", "Function1", "function-tag"))
	assert.False(t, am.HasPkgFunctionAttachment("pkg1", "Function1", "nonexistent"))
	assert.False(t, am.HasPkgFunctionAttachment("pkg1", "Function2", "function-tag"))

	// Add type attachment
	am.AddPkgTypeAttachment("pkg1", "MyType", "type-tag")
	assert.True(t, am.HasPkgTypeAttachment("pkg1", "MyType", "type-tag"))
	assert.False(t, am.HasPkgTypeAttachment("pkg1", "MyType", "nonexistent"))
	assert.False(t, am.HasPkgTypeAttachment("pkg1", "OtherType", "type-tag"))

	// Add type field attachment
	am.AddPkgTypeFieldAttachment("pkg1", "MyType", "Field1", "field-tag")
	assert.True(t, am.HasPkgTypeFieldAttachment("pkg1", "MyType", "Field1", "field-tag"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg1", "MyType", "Field1", "nonexistent"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg1", "MyType", "Field2", "field-tag"))

	// Add type method attachment
	am.AddPkgTypeMethodAttachment("pkg1", "MyType", "Method1", "method-tag")
	assert.True(t, am.HasPkgTypeMethodAttachment("pkg1", "MyType", "Method1", "method-tag"))
	assert.False(t, am.HasPkgTypeMethodAttachment("pkg1", "MyType", "Method1", "nonexistent"))
	assert.False(t, am.HasPkgTypeMethodAttachment("pkg1", "MyType", "Method2", "method-tag"))
}

func TestAttachmentsMap_MultipleAttachments(t *testing.T) {
	am := &AttachmentsMap{}

	// Add multiple attachments to same item
	am.AddPkgAttachment("pkg1", "tag1")
	am.AddPkgAttachment("pkg1", "tag2")
	am.AddPkgAttachment("pkg1", "tag3")

	assert.True(t, am.HasPkgAttachment("pkg1", "tag1"))
	assert.True(t, am.HasPkgAttachment("pkg1", "tag2"))
	assert.True(t, am.HasPkgAttachment("pkg1", "tag3"))

	// Add multiple attachments to different packages
	am.AddPkgAttachment("pkg2", "tag1")
	am.AddPkgAttachment("pkg2", "tag2")

	assert.True(t, am.HasPkgAttachment("pkg2", "tag1"))
	assert.True(t, am.HasPkgAttachment("pkg2", "tag2"))
	assert.False(t, am.HasPkgAttachment("pkg2", "tag3")) // doesn't exist in pkg2
}

func TestAttachmentsMap_MultiplePackagesAndTypes(t *testing.T) {
	am := &AttachmentsMap{}

	// Package 1
	am.AddPkgTypeAttachment("pkg1", "TypeA", "type-tag")
	am.AddPkgTypeFieldAttachment("pkg1", "TypeA", "field1", "field-tag")
	am.AddPkgTypeMethodAttachment("pkg1", "TypeA", "method1", "method-tag")

	// Package 2 - same type name, different package
	am.AddPkgTypeAttachment("pkg2", "TypeA", "different-type-tag")
	am.AddPkgTypeFieldAttachment("pkg2", "TypeA", "field1", "different-field-tag")

	// Verify package isolation
	assert.True(t, am.HasPkgTypeAttachment("pkg1", "TypeA", "type-tag"))
	assert.False(t, am.HasPkgTypeAttachment("pkg1", "TypeA", "different-type-tag"))

	assert.True(t, am.HasPkgTypeAttachment("pkg2", "TypeA", "different-type-tag"))
	assert.False(t, am.HasPkgTypeAttachment("pkg2", "TypeA", "type-tag"))

	assert.True(t, am.HasPkgTypeFieldAttachment("pkg1", "TypeA", "field1", "field-tag"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg1", "TypeA", "field1", "different-field-tag"))

	assert.True(t, am.HasPkgTypeFieldAttachment("pkg2", "TypeA", "field1", "different-field-tag"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg2", "TypeA", "field1", "field-tag"))
}

func TestAttachmentsMap_NilMap(t *testing.T) {
	var am *AttachmentsMap

	// All operations should return false safely
	assert.False(t, am.HasPkgAttachment("pkg1", "tag1"))
	assert.False(t, am.HasPkgFunctionAttachment("pkg1", "func1", "tag1"))
	assert.False(t, am.HasPkgTypeAttachment("pkg1", "Type1", "tag1"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg1", "Type1", "field1", "tag1"))
	assert.False(t, am.HasPkgTypeMethodAttachment("pkg1", "Type1", "method1", "tag1"))
}

func TestPackageAttachments_BasicOperations(t *testing.T) {
	pa := &PackageAttachments{}

	// Test nil safety
	assert.False(t, pa.HasAttachment("tag1"))
	assert.False(t, pa.HasFunctionAttachment("func1", "tag1"))
	assert.False(t, pa.HasTypeAttachment("Type1", "tag1"))
	assert.False(t, pa.HasTypeFieldAttachment("Type1", "field1", "tag1"))
	assert.False(t, pa.HasTypeMethodAttachment("Type1", "method1", "tag1"))

	// Add attachments
	pa.AddAttachment("local-tag")
	pa.AddFunctionAttachment("Function1", "function-tag")
	pa.AddTypeAttachment("MyType", "type-tag")
	pa.AddTypeFieldAttachment("MyType", "Field1", "field-tag")
	pa.AddTypeMethodAttachment("MyType", "Method1", "method-tag")

	// Verify attachments exist
	assert.True(t, pa.HasAttachment("local-tag"))
	assert.True(t, pa.HasFunctionAttachment("Function1", "function-tag"))
	assert.True(t, pa.HasTypeAttachment("MyType", "type-tag"))
	assert.True(t, pa.HasTypeFieldAttachment("MyType", "Field1", "field-tag"))
	assert.True(t, pa.HasTypeMethodAttachment("MyType", "Method1", "method-tag"))
}

func TestTypeAttachments_BasicOperations(t *testing.T) {
	ta := &TypeAttachments{}

	// Test nil safety
	assert.False(t, ta.HasAttachment("tag1"))
	assert.False(t, ta.HasFieldAttachment("field1", "tag1"))
	assert.False(t, ta.HasMethodAttachment("method1", "tag1"))

	// Add attachments
	ta.AddAttachment("type-tag")
	ta.AddFieldAttachment("Field1", "field-tag")
	ta.AddMethodAttachment("Method1", "method-tag")

	// Verify attachments exist
	assert.True(t, ta.HasAttachment("type-tag"))
	assert.True(t, ta.HasFieldAttachment("Field1", "field-tag"))
	assert.True(t, ta.HasMethodAttachment("Method1", "method-tag"))

	// Verify non-existent attachments
	assert.False(t, ta.HasAttachment("nonexistent"))
	assert.False(t, ta.HasFieldAttachment("Field1", "nonexistent"))
	assert.False(t, ta.HasFieldAttachment("NonExistentField", "field-tag"))
	assert.False(t, ta.HasMethodAttachment("Method1", "nonexistent"))
	assert.False(t, ta.HasMethodAttachment("NonExistentMethod", "method-tag"))
}

func TestAttachmentsMap_ComplexScenario(t *testing.T) {
	am := &AttachmentsMap{}

	// Simulate a complex package structure
	pkgPath := "github.com/example/mypackage"

	// Package-level attachment
	am.AddPkgAttachment(pkgPath, "package-only")

	// Function attachment
	am.AddPkgFunctionAttachment(pkgPath, "CreateInstance", "package-only,testing")

	// Type with multiple attachments
	am.AddPkgTypeAttachment(pkgPath, "Service", "immutable")
	am.AddPkgTypeFieldAttachment(pkgPath, "Service", "config", "mutable")
	am.AddPkgTypeMethodAttachment(pkgPath, "Service", "Start", "package-only")
	am.AddPkgTypeMethodAttachment(pkgPath, "Service", "Stop", "package-only")

	// Another type
	am.AddPkgTypeAttachment(pkgPath, "Helper", "package-only")
	am.AddPkgTypeMethodAttachment(pkgPath, "Helper", "DoWork", "package-only,testing")

	// Verify all attachments
	assert.True(t, am.HasPkgAttachment(pkgPath, "package-only"))
	assert.True(t, am.HasPkgFunctionAttachment(pkgPath, "CreateInstance", "package-only,testing"))
	assert.True(t, am.HasPkgTypeAttachment(pkgPath, "Service", "immutable"))
	assert.True(t, am.HasPkgTypeFieldAttachment(pkgPath, "Service", "config", "mutable"))
	assert.True(t, am.HasPkgTypeMethodAttachment(pkgPath, "Service", "Start", "package-only"))
	assert.True(t, am.HasPkgTypeMethodAttachment(pkgPath, "Service", "Stop", "package-only"))
	assert.True(t, am.HasPkgTypeAttachment(pkgPath, "Helper", "package-only"))
	assert.True(t, am.HasPkgTypeMethodAttachment(pkgPath, "Helper", "DoWork", "package-only,testing"))

	// Verify cross-contamination doesn't happen
	assert.False(t, am.HasPkgTypeAttachment(pkgPath, "Helper", "immutable"))
	assert.False(t, am.HasPkgTypeFieldAttachment(pkgPath, "Service", "config", "package-only"))
	assert.False(t, am.HasPkgTypeMethodAttachment(pkgPath, "Helper", "Start", "package-only"))
}

func TestAttachmentsMap_MultipleTagsPerItem(t *testing.T) {
	am := &AttachmentsMap{}

	// Add multiple tags to same function
	am.AddPkgFunctionAttachment("pkg1", "Func1", "tag1")
	am.AddPkgFunctionAttachment("pkg1", "Func1", "tag2")
	am.AddPkgFunctionAttachment("pkg1", "Func1", "tag3")

	// Add multiple tags to same type field
	am.AddPkgTypeFieldAttachment("pkg1", "Type1", "Field1", "field-tag1")
	am.AddPkgTypeFieldAttachment("pkg1", "Type1", "Field1", "field-tag2")

	// Verify all tags exist
	assert.True(t, am.HasPkgFunctionAttachment("pkg1", "Func1", "tag1"))
	assert.True(t, am.HasPkgFunctionAttachment("pkg1", "Func1", "tag2"))
	assert.True(t, am.HasPkgFunctionAttachment("pkg1", "Func1", "tag3"))
	assert.True(t, am.HasPkgTypeFieldAttachment("pkg1", "Type1", "Field1", "field-tag1"))
	assert.True(t, am.HasPkgTypeFieldAttachment("pkg1", "Type1", "Field1", "field-tag2"))

	// Verify non-existent tags
	assert.False(t, am.HasPkgFunctionAttachment("pkg1", "Func1", "nonexistent"))
	assert.False(t, am.HasPkgTypeFieldAttachment("pkg1", "Type1", "Field1", "nonexistent"))
}

func TestPackageAttachments_Nil(t *testing.T) {
	var pa *PackageAttachments

	// All operations should return false safely
	assert.False(t, pa.HasAttachment("tag1"))
	assert.False(t, pa.HasFunctionAttachment("func1", "tag1"))
	assert.False(t, pa.HasTypeAttachment("Type1", "tag1"))
	assert.False(t, pa.HasTypeFieldAttachment("Type1", "field1", "tag1"))
	assert.False(t, pa.HasTypeMethodAttachment("Type1", "method1", "tag1"))
}

func TestTypeAttachments_Nil(t *testing.T) {
	var ta *TypeAttachments

	// All operations should return false safely
	assert.False(t, ta.HasAttachment("tag1"))
	assert.False(t, ta.HasFieldAttachment("field1", "tag1"))
	assert.False(t, ta.HasMethodAttachment("method1", "tag1"))
}

func TestAttachmentsMap_ZeroValue(t *testing.T) {
	var am AttachmentsMap

	// Should work with zero value
	am.AddPkgAttachment("pkg1", "tag1")
	require.True(t, am.HasPkgAttachment("pkg1", "tag1"))
}
