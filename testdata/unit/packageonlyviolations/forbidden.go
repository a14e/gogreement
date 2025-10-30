package packageonlyviolations

import "github.com/a14e/gogreement/testdata/unit/packageonlysource"

// This file uses @packageonly items from packageonlysource in forbiddenpkg (should cause violations)

func usePackageOnlyTypeViolation() {
	var x packageonlysource.PackageOnlyType // This should cause violation - current package is forbiddenpkg, not allowedpkg
	x.Method()
}

func callPackageOnlyFunctionViolation() {
	result := packageonlysource.PackageOnlyFunction() // This should cause violation
	_ = result
}

func callPackageOnlyMethodViolation() {
	var s packageonlysource.PackageOnlyStruct
	s.PackageOnlyMethod() // This should cause violation
}

// Examples with @ignore directives
func usePackageOnlyTypeIgnored() {
	// @ignore PKGO01
	var x packageonlysource.PackageOnlyType // This should NOT cause violation (ignored)
	x.Method()
}

func callPackageOnlyFunctionIgnored() {
	// @ignore PKGO02
	result := packageonlysource.PackageOnlyFunction() // This should NOT cause violation (ignored)
	_ = result
}

func callPackageOnlyMethodIgnored() {
	// @ignore PKGO03
	var s packageonlysource.PackageOnlyStruct
	s.PackageOnlyMethod() // This should NOT cause violation (ignored)
}

// Example with category ignore
func usePackageOnlyTypeCategoryIgnored() {
	// @ignore PKGO
	var x packageonlysource.PackageOnlyType // This should NOT cause violation (category ignored)
	x.Method()
}

func useRegularItems() {
	var regularType packageonlysource.RegularType
	regularType.Method() // This should be fine

	result := packageonlysource.RegularFunction() // This should be fine
	_ = result

	var regularStruct packageonlysource.RegularStruct
	regularStruct.RegularMethod() // This should be fine
}
