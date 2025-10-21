package implements

import (
	"go/types"
	"goagreement/src/annotations"

	"golang.org/x/tools/go/analysis"
)

// TypeModel represents a parsed named type with its methods
// @immutable
type TypeModel struct {
	Name           string
	Package        string
	UnderlyingType string // "struct", "int", "string", etc. // FIXME Do We need this?
	Methods        []TypeMethod
}

// TypeMethod represents a method of a type
// @immutable
type TypeMethod struct {
	Name              string
	Inputs            []MethodType
	Outputs           []MethodType
	ReceiverIsPointer bool // true if receiver is *T, false if T
}

// MethodType represents a type in method signature
// @immutable
// @constructor convertTypesToMethodType, extractMethodTypesFromTuple
type MethodType struct {
	TypeName    string
	TypePackage string
	IsPointer   bool
	IsVariadic  bool
}

// LoadTypes loads specified named types from the current package
func LoadTypes(pass *analysis.Pass, queries []annotations.TypeQuery) []*TypeModel {
	var result []*TypeModel

	// Create a set of type names we're looking for
	targetTypes := make(map[string]bool)
	for _, q := range queries {
		targetTypes[q.TypeName] = true
	}

	// Scan only current package
	result = findTypesInPackage(pass.Pkg, targetTypes)

	return result
}

// findTypesInPackage extracts named types and their methods from package
func findTypesInPackage(
	pkg *types.Package,
	targetTypes map[string]bool,
) []*TypeModel {
	var result []*TypeModel

	scope := pkg.Scope()
	for _, name := range scope.Names() {
		if !targetTypes[name] {
			continue
		}

		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		// Check if it's a type name
		typeName, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}

		// We want named types (struct, int, string, etc.)
		namedType, ok := typeName.Type().(*types.Named)
		if !ok {
			continue
		}

		// Determine underlying type for debugging/reporting
		underlyingType := getUnderlyingTypeName(namedType.Underlying())

		// Extract methods for this type
		methods := extractMethodsFromNamedType(namedType)

		model := &TypeModel{
			Name:           name,
			Package:        pkg.Path(),
			UnderlyingType: underlyingType,
			Methods:        methods,
		}

		result = append(result, model)
	}

	return result
}

// getUnderlyingTypeName returns a string representation of the underlying type
func getUnderlyingTypeName(t types.Type) string {
	switch ut := t.(type) {
	case *types.Struct:
		return "struct"
	case *types.Basic:
		return ut.Name()
	case *types.Slice:
		return "slice"
	case *types.Array:
		return "array"
	case *types.Map:
		return "map"
	case *types.Chan:
		return "chan"
	case *types.Signature:
		return "func"
	case *types.Interface:
		return "interface"
	case *types.Pointer:
		return "pointer"
	default:
		return t.String()
	}
}

// extractMethodsFromNamedType extracts all methods (value + pointer receivers)
func extractMethodsFromNamedType(named *types.Named) []TypeMethod {
	var methods []TypeMethod

	// Get method set for *T (includes both T and *T receivers)
	ptrType := types.NewPointer(named)
	methodSet := types.NewMethodSet(ptrType)

	for i := 0; i < methodSet.Len(); i++ {
		selection := methodSet.At(i)
		method := selection.Obj().(*types.Func)
		sig := method.Type().(*types.Signature)

		// Determine if receiver is pointer
		recvIsPointer := isPointerReceiver(sig.Recv().Type())

		methods = append(methods, TypeMethod{
			Name:              method.Name(),
			Inputs:            extractMethodTypesFromTuple(sig.Params(), sig.Variadic()),
			Outputs:           extractMethodTypesFromTuple(sig.Results(), false),
			ReceiverIsPointer: recvIsPointer,
		})
	}

	return methods
}

// isPointerReceiver checks if receiver type is a pointer
func isPointerReceiver(t types.Type) bool {
	_, ok := t.(*types.Pointer)
	return ok
}

// extractMethodTypesFromTuple converts types.Tuple to MethodType slice
func extractMethodTypesFromTuple(tuple *types.Tuple, isVariadic bool) []MethodType {
	if tuple == nil {
		return nil
	}

	result := make([]MethodType, tuple.Len())

	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)
		result[i] = convertTypesToMethodType(param.Type())

		// Mark last parameter as variadic if needed
		if isVariadic && i == tuple.Len()-1 {
			result[i].IsVariadic = true

			// Unwrap slice type for variadic: []string -> string
			if slice, ok := param.Type().(*types.Slice); ok {
				result[i] = convertTypesToMethodType(slice.Elem())
				result[i].IsVariadic = true
			}
		}
	}

	return result
}

// convertTypesToMethodType converts types.Type to MethodType
func convertTypesToMethodType(t types.Type) MethodType {
	// Handle pointer
	if ptr, ok := t.(*types.Pointer); ok {
		inner := convertTypesToMethodType(ptr.Elem())
		inner.IsPointer = true
		return inner
	}

	// Handle named types
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		pkg := obj.Pkg()
		pkgPath := ""
		if pkg != nil {
			pkgPath = pkg.Path()
		}

		return MethodType{
			TypeName:    obj.Name(),
			TypePackage: pkgPath,
			IsPointer:   false,
			IsVariadic:  false,
		}
	}

	// Handle basic types
	if basic, ok := t.(*types.Basic); ok {
		return MethodType{
			TypeName:    basic.Name(),
			TypePackage: "",
			IsPointer:   false,
			IsVariadic:  false,
		}
	}

	// Fallback
	return MethodType{
		TypeName:   t.String(),
		IsPointer:  false,
		IsVariadic: false,
	}
}
