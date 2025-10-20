package analyzer

import (
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ========== Error Types ==========

type MissingPackageReport struct {
	PackageName string
	TypeName    string
	Pos         token.Pos
}

type MissingInterfaceReport struct {
	InterfaceName string
	PackageName   string
	TypeName      string
	Pos           token.Pos
}

type MissingMethodsReport struct {
	InterfaceName string
	PackageName   string
	TypeName      string
	Methods       []InterfaceMethod // Full method signatures
	Pos           token.Pos
}

// ========== Phase 2: Validation Functions ==========

// findMissingPackages identifies annotations with unresolved package references
func findMissingPackages(annotations []ImplementsAnnotation) []MissingPackageReport {
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

// findMissingInterfaces identifies annotations where the interface was not found
func findMissingInterfaces(
	annotations []ImplementsAnnotation,
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

// findMissingMethods identifies types that don't implement required interfaces
func findMissingMethods(
	annotations []ImplementsAnnotation,
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
			continue // Already reported in findMissingInterfaces
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

// ========== Phase 3: Reporting Functions ==========

func reportProblems(
	pass *analysis.Pass,
	missingPackages []MissingPackageReport,
	missingInterfaces []MissingInterfaceReport,
	missingMethods []MissingMethodsReport,
) {
	// Report missing packages
	for _, mp := range missingPackages {
		pass.Report(analysis.Diagnostic{
			Pos: mp.Pos,
			Message: fmt.Sprintf(
				"package %q referenced in @implements annotation on type \"%s\" is not imported",
				mp.PackageName,
				mp.TypeName,
			),
		})
	}

	// Report missing interfaces
	for _, mi := range missingInterfaces {
		pkgPrefix := ""
		if mi.PackageName != "" {
			pkgPrefix = mi.PackageName + "."
		}
		pass.Report(analysis.Diagnostic{
			Pos: mi.Pos,
			Message: fmt.Sprintf(
				"interface \"%s%s\" not found for type \"%s\"",
				pkgPrefix,
				mi.InterfaceName,
				mi.TypeName,
			),
		})
	}

	// Report missing methods
	for _, mm := range missingMethods {
		pkgPrefix := ""
		if mm.PackageName != "" {
			pkgPrefix = mm.PackageName + "."
		}

		// Format each method signature on a new line
		var methodLines []string
		for _, method := range mm.Methods {
			methodLines = append(methodLines, "  "+formatMethodSignature(method))
		}

		message := fmt.Sprintf(
			"type \"%s\" does not implement interface \"%s%s\"\nmissing methods:\n%s",
			mm.TypeName,
			pkgPrefix,
			mm.InterfaceName,
			strings.Join(methodLines, "\n"),
		)

		pass.Report(analysis.Diagnostic{
			Pos:     mm.Pos,
			Message: message,
		})
	}
}

// formatMethodSignature formats a method signature for display
// Example: Read(p []byte) (n int, err error)
// formatMethodSignature formats a method signature for display
// Example: Read(p []byte) (n int, err error)
func formatMethodSignature(method InterfaceMethod) string {

	var result strings.Builder

	// Format inputs
	inputs := formatTypeList(method.Inputs)

	// Format outputs
	outputs := formatTypeList(method.Outputs)

	// Build signature
	result.WriteString(method.Name)
	result.WriteString("(")
	result.WriteString(inputs)
	result.WriteString(")")

	if outputs != "" {
		// Wrap in parens only if multiple outputs
		if len(method.Outputs) > 1 {

			result.WriteString(" (")
			result.WriteString(outputs)
			result.WriteString(")")
		} else {
			// Single output without parens
			result.WriteString(" ")
			result.WriteString(outputs)
		}
	}

	return result.String()
}

// formatTypeList formats a list of types for display
// Example: p []byte, n int, err error
func formatTypeList(types []InterfaceType) string {
	if len(types) == 0 {
		return ""
	}

	var parts []string
	for _, t := range types {
		parts = append(parts, formatType(t))
	}

	return strings.Join(parts, ", ")
}

// formatType formats a single type for display
// Examples: int, *string, []byte, io.Reader, ...string
func formatType(t InterfaceType) string {
	var result strings.Builder

	// Add variadic prefix
	if t.IsVariadic {
		result.WriteString("...")
	}

	// Add pointer prefix
	if t.IsPointer {
		result.WriteString("*")
	}

	// Add package prefix
	if t.TypePackage != "" {
		// Extract short package name from full path
		parts := strings.Split(t.TypePackage, "/")
		shortPkg := parts[len(parts)-1]

		result.WriteString(shortPkg)
		result.WriteString(".")
	}

	// Add type name
	result.WriteString(t.TypeName)

	return result.String()
}
