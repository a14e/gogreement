package modB // want package:"package modB"

import (
	"fmt"

	modA "multimodule_packageonly/modA"
)

func UsePackageOnlyItems() {
	// These usages should be allowed - modB is in the allowed list
	it := modA.NewInternalType("test")
	result := it.Process()
	fmt.Println(result)

	// Use package-only variable
	fmt.Println(modA.InternalGlobal)

	// Use package-only type directly
	var internalType modA.InternalType
	internalType.Value = "direct"
	fmt.Println(internalType.Process())

	// Use public items - always allowed
	pt := modA.NewPublicType("public")
	fmt.Printf("Public type: %+v\n", pt)
}

func UsePublicItems() {
	// These usages should always be allowed
	pt := modA.NewPublicType("test")
	fmt.Printf("Public type data: %s\n", pt.Data)
}

// This function should cause violations - modB is using items restricted to modC
func UseRestrictedItems() {
	rt := modA.NewRestrictedType("secret")     // want "\\[PKGO02\\].*NewRestrictedType function is @packageonly.*"
	fmt.Printf("Secret: %s\n", rt.GetSecret()) // want "\\[PKGO03\\].*RestrictedType.GetSecret method is @packageonly.*"

	var restricted modA.RestrictedType // want "\\[PKGO01\\].*RestrictedType type is @packageonly.*"
	restricted.Secret = "direct secret"
	fmt.Printf("Direct secret: %s\n", restricted.GetSecret()) // want "\\[PKGO03\\].*RestrictedType.GetSecret method is @packageonly.*"
}

// @ignore PKGO01, PKGO02, PKGO03
func UseRestrictedItemsWithIgnore() {
	// These violations should be suppressed by @ignore
	rt := modA.NewRestrictedType("ignored secret")
	fmt.Printf("Ignored secret: %s\n", rt.GetSecret())

	var restricted modA.RestrictedType
	restricted.Secret = "ignored direct"
	fmt.Printf("Ignored direct: %s\n", restricted.GetSecret())
}
