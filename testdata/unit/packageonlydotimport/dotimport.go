package packageonlydotimport

import . "github.com/a14e/gogreement/testdata/unit/packageonlysource"

// This package dot-imports the source package and is NOT in the allowed list.
// Dot-imported @packageonly symbols appear as bare identifiers rather than
// selectors, so they must still be enforced.

func UseDotImported() {
	var x PackageOnlyType // VIOLATION (PKGO01): dot-imported @packageonly type
	_ = x

	_ = PackageOnlyFunction() // VIOLATION (PKGO02): dot-imported @packageonly function
}
