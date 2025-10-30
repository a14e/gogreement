package packageonly

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/codes"
	"github.com/a14e/gogreement/src/reporting"
)

// PackageOnlyViolation represents a violation of @packageonly usage
// @immutable
// implements reporting.Violation
type PackageOnlyViolation struct {
	Pos             token.Pos
	ItemName        string   // Name of the @packageonly object being used
	ItemPkgPath     string   // Package path where the item is defined
	CurrentPkgPath  string   // Current package path where the violation occurred
	AllowedPackages []string // Allowed packages for this item
	ReceiverType    string   // Receiver type for methods (empty for types/functions)
	Code            string   // Error code for this violation
}

// GetCode returns the error code for this violation
func (v PackageOnlyViolation) GetCode() string {
	return v.Code
}

// GetPos returns the position of the violation
func (v PackageOnlyViolation) GetPos() token.Pos {
	return v.Pos
}

// GetMessage returns the main error message without formatting
func (v PackageOnlyViolation) GetMessage() string {
	switch v.Code {
	case codes.PackageOnlyMethodCall:
		return fmt.Sprintf("%s.%s method is @packageonly and cannot be used from %s. Allowed packages: %s",
			v.ReceiverType, v.ItemName, v.CurrentPkgPath, fmt.Sprintf("%v", v.AllowedPackages))
	case codes.PackageOnlyTypeUsage:
		return fmt.Sprintf("%s type is @packageonly and cannot be used from %s. Allowed packages: %s",
			v.ItemName, v.CurrentPkgPath, fmt.Sprintf("%v", v.AllowedPackages))
	case codes.PackageOnlyFunctionCall:
		return fmt.Sprintf("%s function is @packageonly and cannot be used from %s. Allowed packages: %s",
			v.ItemName, v.CurrentPkgPath, fmt.Sprintf("%v", v.AllowedPackages))
	default:
		return fmt.Sprintf("%s is @packageonly and cannot be used from %s", v.ItemName, v.CurrentPkgPath)
	}
}

// ReportViolations reports packageonly violations using the new pretty formatter
// NOTE: violations should already be filtered by @ignore directives in CheckPackageOnly
func ReportViolations(pass *analysis.Pass, violations []PackageOnlyViolation) {
	reporter := reporting.NewReporter(pass, nil) // No ignore set needed, already filtered

	// Convert to generic violations and report
	for _, violation := range violations {
		reporter.ReportViolation(violation)
	}
}
