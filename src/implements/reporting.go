package implements

import (
	"fmt"
	"go/token"
	"strings"

	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/reporting"
	"github.com/a14e/gogreement/src/util"

	"golang.org/x/tools/go/analysis"
)

// ========== Violation Types ==========

// @immutable
// implements reporting.Violation
type MissingPackageReport struct {
	PackageName string
	TypeName    string
	Pos         token.Pos
}

// GetCode returns the error code for this violation
func (v MissingPackageReport) GetCode() string {
	return codes.ImplementsPackageNotFound
}

// GetPos returns the position of the violation
func (v MissingPackageReport) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v MissingPackageReport) GetMessage() string {
	return fmt.Sprintf(
		"package %q referenced in @implements annotation on type \"%s\" is not imported",
		v.PackageName,
		v.TypeName,
	)
}

// @immutable
// implements reporting.Violation
type MissingInterfaceReport struct {
	InterfaceName string
	PackageName   string
	TypeName      string
	Pos           token.Pos
}

// GetCode returns the error code for this violation
func (v MissingInterfaceReport) GetCode() string {
	return codes.ImplementsInterfaceNotFound
}

// GetPos returns the position of the violation
func (v MissingInterfaceReport) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v MissingInterfaceReport) GetMessage() string {
	pkgPrefix := ""
	if v.PackageName != "" {
		pkgPrefix = v.PackageName + "."
	}
	return fmt.Sprintf(
		"interface \"%s%s\" not found for type \"%s\"",
		pkgPrefix,
		v.InterfaceName,
		v.TypeName,
	)
}

// @immutable
// implements reporting.Violation
type MissingMethodsReport struct {
	InterfaceName string
	PackageName   string
	TypeName      string
	Methods       []InterfaceMethod // Full method signatures
	Pos           token.Pos
}

// GetCode returns the error code for this violation
func (v MissingMethodsReport) GetCode() string {
	return codes.ImplementsMissingMethods
}

// GetPos returns the position of the violation
func (v MissingMethodsReport) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v MissingMethodsReport) GetMessage() string {
	pkgPrefix := ""
	if v.PackageName != "" {
		pkgPrefix = v.PackageName + "."
	}

	// Format each method signature on a new line
	var methodLines []string
	for _, method := range v.Methods {
		methodLines = append(methodLines, "  "+formatMethodSignature(method))
	}

	return fmt.Sprintf(
		"type \"%s\" does not implement interface \"%s%s\"\nmissing methods:\n%s",
		v.TypeName,
		pkgPrefix,
		v.InterfaceName,
		strings.Join(methodLines, "\n"),
	)
}

// ReportProblems reports all implements violations using the new pretty formatter.
// Supports @ignore directives for suppressing violations when needed.
func ReportProblems(
	pass *analysis.Pass,
	missingPackages []MissingPackageReport,
	missingInterfaces []MissingInterfaceReport,
	missingMethods []MissingMethodsReport,
	ignoreSet *util.IgnoreSet,
) {
	reporter := reporting.NewReporter(pass, ignoreSet)

	// Convert all violations to generic Violation interface and report
	var violations []reporting.Violation

	// Add missing packages
	for _, mp := range missingPackages {
		violations = append(violations, mp)
	}

	// Add missing interfaces
	for _, mi := range missingInterfaces {
		violations = append(violations, mi)
	}

	// Add missing methods
	for _, mm := range missingMethods {
		violations = append(violations, mm)
	}

	// Report all violations using the new pretty formatter
	reporter.ReportViolations(violations)
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
