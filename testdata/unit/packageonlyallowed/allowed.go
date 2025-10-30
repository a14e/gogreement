package packageonlyallowed

import "github.com/a14e/gogreement/testdata/unit/packageonlysource"

// Allowed package using @packageonly items correctly

func usePackageOnlyType() {
	var x packageonlysource.PackageOnlyType // This should be allowed
	x.Method()
}

func callPackageOnlyFunction() {
	result := packageonlysource.PackageOnlyFunction() // This should be allowed
	_ = result
}

func callPackageOnlyMethod() {
	var s packageonlysource.PackageOnlyStruct
	s.PackageOnlyMethod() // This should be allowed
}

func useRegularItems() {
	var regularType packageonlysource.RegularType
	regularType.Method()

	result := packageonlysource.RegularFunction()
	_ = result

	var regularStruct packageonlysource.RegularStruct
	regularStruct.RegularMethod()
}
