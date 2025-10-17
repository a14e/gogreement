package analyzer

import (
	"go/types"

	"golang.org/x/tools/go/analysis"
)

type InterfaceModel struct {
	Name    string
	Package string
	Methods []InterfaceMethod
}

type InterfaceMethod struct {
	Name    string
	Inputs  []InterfaceType
	Outputs []InterfaceType
}

type InterfaceType struct {
	TypeName    string
	TypePackage string
	IsPointer   bool
	IsVariadic  bool
}

// InterfaceQuery represents what interface we're looking for
type InterfaceQuery struct {
	InterfaceName string
	PackageName   string // empty string means current package
}

// LoadInterfaces loads specified interfaces from the analysis pass
func LoadInterfaces(pass *analysis.Pass, queries []InterfaceQuery) []*InterfaceModel {
	var result []*InterfaceModel

	// Group queries by package
	pkgToInterface := make(map[string]map[string]bool)
	for _, q := range queries {
		pkg := q.PackageName
		if pkg == "" {
			pkg = pass.Pkg.Path()
		}
		if pkgToInterface[pkg] == nil {
			pkgToInterface[pkg] = make(map[string]bool)
		}
		pkgToInterface[pkg][q.InterfaceName] = true
	}

	// Collect all packages to scan (current + imports)
	packagesToScan := make([]*types.Package, 0)

	// Add current package
	if pkgToInterface[pass.Pkg.Path()] != nil {
		packagesToScan = append(packagesToScan, pass.Pkg)
	}

	// Add imported packages
	for _, imp := range pass.Pkg.Imports() {
		if pkgToInterface[imp.Path()] != nil {
			packagesToScan = append(packagesToScan, imp)
		}
	}

	// Scan all packages uniformly
	for _, pkg := range packagesToScan {
		interfaces := findInterfacesInPackage(pkg, pkgToInterface[pkg.Path()])
		result = append(result, interfaces...)
	}

	return result
}

// findInterfacesInPackage extracts interfaces from imported package using types.Package
func findInterfacesInPackage(
	pkg *types.Package,
	targetInterfaces map[string]bool,
) []*InterfaceModel {
	var result []*InterfaceModel

	scope := pkg.Scope()
	for _, name := range scope.Names() {
		if !targetInterfaces[name] {
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

		// Check if it's an interface
		iface, ok := typeName.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}

		model := &InterfaceModel{
			Name:    name,
			Package: pkg.Path(),
			Methods: extractMethodsFromInterface(iface),
		}

		result = append(result, model)
	}

	return result
}

// extractMethodsFromInterface extracts methods from types.Interface
func extractMethodsFromInterface(iface *types.Interface) []InterfaceMethod {
	var methods []InterfaceMethod

	for i := 0; i < iface.NumMethods(); i++ {
		method := iface.Method(i)
		sig := method.Type().(*types.Signature)

		methods = append(methods, InterfaceMethod{
			Name:    method.Name(),
			Inputs:  extractTypesFromTuple(sig.Params(), sig.Variadic()),
			Outputs: extractTypesFromTuple(sig.Results(), false),
		})
	}

	return methods
}

// extractTypesFromTuple converts types.Tuple to InterfaceType slice
func extractTypesFromTuple(tuple *types.Tuple, isVariadic bool) []InterfaceType {
	if tuple == nil {
		return nil
	}

	result := make([]InterfaceType, tuple.Len())

	for i := 0; i < tuple.Len(); i++ {
		param := tuple.At(i)
		result[i] = convertTypesToInterfaceType(param.Type())

		// Mark last parameter as variadic if needed
		if isVariadic && i == tuple.Len()-1 {
			result[i].IsVariadic = true
		}
	}

	return result
}

// convertTypesToInterfaceType converts types.Type to InterfaceType
func convertTypesToInterfaceType(t types.Type) InterfaceType {
	// Handle pointer
	if ptr, ok := t.(*types.Pointer); ok {
		inner := convertTypesToInterfaceType(ptr.Elem())
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

		return InterfaceType{
			TypeName:    obj.Name(),
			TypePackage: pkgPath,
			IsPointer:   false,
			IsVariadic:  false,
		}
	}

	// Handle basic types
	if basic, ok := t.(*types.Basic); ok {
		return InterfaceType{
			TypeName:    basic.Name(),
			TypePackage: "",
			IsPointer:   false,
			IsVariadic:  false,
		}
	}

	// Fallback
	return InterfaceType{
		TypeName:   t.String(),
		IsPointer:  false,
		IsVariadic: false,
	}
}
