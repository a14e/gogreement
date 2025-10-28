package implements

import (
	"fmt"
	"github.com/a14e/gogreement/src/codes"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ReportProblems reports all implements violations to the analysis pass.
// Note: @ignore directives are NOT supported for implements violations.
// These violations represent structural issues that must be fixed.
func ReportProblems(
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
				"[%s] package %q referenced in @implements annotation on type \"%s\" is not imported",
				codes.ImplementsPackageNotFound,
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
				"[%s] interface \"%s%s\" not found for type \"%s\"",
				codes.ImplementsInterfaceNotFound,
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
			"[%s] type \"%s\" does not implement interface \"%s%s\"\nmissing methods:\n%s",
			codes.ImplementsMissingMethods,
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
		if parts := strings.Split(t.TypePackage, "/"); len(parts) > 0 {
			shortPkg := parts[len(parts)-1]

			result.WriteString(shortPkg)
			result.WriteString(".")
		}
	}

	// Add type name
	result.WriteString(t.TypeName)

	return result.String()
}
