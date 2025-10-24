package implements

import (
	"go/token"
	"goagreement/src/annotations"
)

// ========== Error Types ==========

// @immutable
type MissingPackageReport struct {
	PackageName string
	TypeName    string
	Pos         token.Pos
}

// @immutable
type MissingInterfaceReport struct {
	InterfaceName string
	PackageName   string
	TypeName      string
	Pos           token.Pos
}

// @immutable
type MissingMethodsReport struct {
	InterfaceName string
	PackageName   string
	TypeName      string
	Methods       []InterfaceMethod // Full method signatures
	Pos           token.Pos
}

// FindMissingPackages identifies annotations with unresolved package references
func FindMissingPackages(annotations []annotations.ImplementsAnnotation) []MissingPackageReport {
	var result []MissingPackageReport

	for _, ann := range annotations {
		if ann.PackageNotFound {
			result = append(result, MissingPackageReport{
				PackageName: ann.PackageName,
				TypeName:    ann.OnType,
				Pos:         ann.OnTypePos,
			})
		}
	}

	return result
}

// FindMissingInterfaces identifies annotations where the interface was not found
func FindMissingInterfaces(
	annotations []annotations.ImplementsAnnotation,
	interfaces []*InterfaceModel,
) []MissingInterfaceReport {
	var result []MissingInterfaceReport

	// Create index of found interfaces: "package.Interface" -> true
	foundInterfaces := make(map[string]bool)
	for _, iface := range interfaces {
		key := iface.Package + "." + iface.Name
		foundInterfaces[key] = true
	}

	for _, ann := range annotations {
		// Skip if package was not found (already reported in Phase 1)
		if ann.PackageNotFound {
			continue
		}

		// Check if interface exists
		key := ann.PackageFullPath + "." + ann.InterfaceName
		if !foundInterfaces[key] {
			result = append(result, MissingInterfaceReport{
				InterfaceName: ann.InterfaceName,
				PackageName:   ann.PackageName, // Use short name for display
				TypeName:      ann.OnType,
				Pos:           ann.OnTypePos,
			})
		}
	}

	return result
}

// FindMissingMethods identifies types that don't implement required interfaces
func FindMissingMethods(
	annotations []annotations.ImplementsAnnotation,
	interfaces []*InterfaceModel,
	types []*TypeModel,
) []MissingMethodsReport {
	var result []MissingMethodsReport

	// Create index of interfaces by full key
	interfaceIndex := make(map[string]*InterfaceModel)
	for _, iface := range interfaces {
		key := iface.Package + "." + iface.Name
		interfaceIndex[key] = iface
	}

	// Create index of types by name
	typeIndex := make(map[string]*TypeModel)
	for _, t := range types {
		typeIndex[t.Name] = t
	}

	for _, ann := range annotations {
		// Skip if package or interface not found (already reported)
		if ann.PackageNotFound {
			continue
		}

		ifaceKey := ann.PackageFullPath + "." + ann.InterfaceName
		iface, ifaceExists := interfaceIndex[ifaceKey]
		if !ifaceExists {
			continue // Already reported in FindMissingInterfaces
		}

		typeModel, typeExists := typeIndex[ann.OnType]
		if !typeExists {
			// Type not found - should not happen but skip
			continue
		}

		// Check if type implements interface
		missing := checkImplementation(typeModel, iface, ann.IsPointer)
		if len(missing) > 0 {
			result = append(result, MissingMethodsReport{
				InterfaceName: ann.InterfaceName,
				PackageName:   ann.PackageName,
				TypeName:      ann.OnType,
				Methods:       missing,
				Pos:           ann.OnTypePos,
			})
		}
	}

	return result
}

// checkImplementation checks if type implements interface
// Returns list of missing methods with full signatures
func checkImplementation(
	typeModel *TypeModel,
	iface *InterfaceModel,
	requirePointer bool,
) []InterfaceMethod {
	var missing []InterfaceMethod

	// Create index of type's methods
	typeMethods := make(map[string]TypeMethod)
	for _, method := range typeModel.Methods {
		// Filter methods based on pointer requirement
		if requirePointer {
			// For &Interface, we need pointer receiver methods
			// (but value receiver methods are also OK per Go spec:
			// method set of *T includes methods with receiver T or *T)
			typeMethods[method.Name] = method
		} else {
			// For Interface (no &), we need value receiver methods only
			if !method.ReceiverIsPointer {
				typeMethods[method.Name] = method
			}
		}
	}

	// Check each interface method
	for _, ifaceMethod := range iface.Methods {
		typeMethod, exists := typeMethods[ifaceMethod.Name]
		if !exists {
			missing = append(missing, ifaceMethod)
			continue
		}

		// Check signature match
		if !signaturesMatch(typeMethod, ifaceMethod) {
			missing = append(missing, ifaceMethod)
		}
	}

	return missing
}

// signaturesMatch checks if type method matches interface method signature
func signaturesMatch(typeMethod TypeMethod, ifaceMethod InterfaceMethod) bool {
	// Check input count
	if len(typeMethod.Inputs) != len(ifaceMethod.Inputs) {
		return false
	}

	// Check output count
	if len(typeMethod.Outputs) != len(ifaceMethod.Outputs) {
		return false
	}

	// Check each input type
	for i := range typeMethod.Inputs {
		if !typesMatch(&typeMethod.Inputs[i], &ifaceMethod.Inputs[i]) {
			return false
		}
	}

	// Check each output type
	for i := range typeMethod.Outputs {
		if !typesMatch(&typeMethod.Outputs[i], &ifaceMethod.Outputs[i]) {
			return false
		}
	}

	return true
}

// typesMatch checks if two types are the same
func typesMatch(t1 *MethodType, t2 *InterfaceType) bool {
	return t1.TypeName == t2.TypeName &&
		t1.TypePackage == t2.TypePackage &&
		t1.IsPointer == t2.IsPointer &&
		t1.IsVariadic == t2.IsVariadic
}
