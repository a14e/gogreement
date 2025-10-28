package codes

import "iter"

// Code represents an error code with its metadata
type Code struct {
	ID          string // Unique error code (e.g., "IMM01")
	Description string // Human-readable description
}

// Error code constants for immutable violations
const (
	ImmutableFieldAssignment     = "IMM01"
	ImmutableFieldCompoundAssign = "IMM02"
	ImmutableFieldIncDec         = "IMM03"
	ImmutableIndexAssignment     = "IMM04"
	ImmutableCategoryPrefix      = "IMM"
)

// Error code constants for constructor violations
const (
	ConstructorCompositeLiteral = "CTOR01"
	ConstructorNewCall          = "CTOR02"
	ConstructorVarDeclaration   = "CTOR03"
	ConstructorCategoryPrefix   = "CTOR"
)

// Error code constants for testonly violations
const (
	TestOnlyTypeUsage      = "TONL01"
	TestOnlyFunctionCall   = "TONL02"
	TestOnlyMethodCall     = "TONL03"
	TestOnlyCategoryPrefix = "TONL"
)

// Error code constants for implements violations
// Note: @ignore directives are NOT supported for implements violations
const (
	ImplementsPackageNotFound   = "IMPL01"
	ImplementsInterfaceNotFound = "IMPL02"
	ImplementsMissingMethods    = "IMPL03"
	ImplementsCategoryPrefix    = "IMPL"
)

// CodesByCategory contains all error codes grouped by their category prefix.
// This structure is easy to read, format, and validate in tests.
// Key: category prefix (e.g., "IMM")
// Value: slice of codes belonging to that category
var CodesByCategory = map[string][]Code{
	ImmutableCategoryPrefix: {
		{ImmutableFieldAssignment, "Field of immutable type is being assigned"},
		{ImmutableFieldCompoundAssign, "Compound assignment to immutable field (e.g., +=, -=)"},
		{ImmutableFieldIncDec, "Increment/decrement of immutable field (e.g., ++, --)"},
		{ImmutableIndexAssignment, "Index assignment to immutable collection (slice/map element)"},
	},
	ConstructorCategoryPrefix: {
		{ConstructorCompositeLiteral, "Composite literal used outside allowed constructor functions"},
		{ConstructorNewCall, "new() call used outside allowed constructor functions"},
		{ConstructorVarDeclaration, "Variable declaration creates zero-initialized instance outside allowed constructor functions"},
	},
	TestOnlyCategoryPrefix: {
		{TestOnlyTypeUsage, "TestOnly type used outside test context"},
		{TestOnlyFunctionCall, "TestOnly function called outside test context"},
		{TestOnlyMethodCall, "TestOnly method called outside test context"},
	},
}

// codeToCheckList is a reverse map built from CodesByCategory.
// For each error code and category, it contains the list of codes to check for ignore directives.
// The list always starts with "ALL", followed by category prefix (if applicable), then the specific code.
//
// Example: "IMM01" -> ["ALL", "IMM", "IMM01"]
// Example: "IMM" -> ["ALL", "IMM"]
// Example: "CTOR02" -> ["ALL", "CTOR", "CTOR02"]
var codeToCheckList = func() map[string][]string {
	result := make(map[string][]string)

	// Add entries for category prefixes
	for category := range CodesByCategory {
		result[category] = []string{"ALL", category}
	}

	// Add entries for specific codes
	for category, codes := range CodesByCategory {
		for _, code := range codes {
			// Build check list: ["ALL", category, specific_code]
			result[code.ID] = []string{"ALL", category, code.ID}
		}
	}

	return result
}()

// GetCodesForCheck returns an iterator of all codes that should be checked
// for ignore directives for the given error code.
// The iterator yields codes in order: "ALL", category prefix, specific code.
//
// Example: GetCodesForCheck("IMM01") yields: "ALL", "IMM", "IMM01"
// Example: GetCodesForCheck("CTOR02") yields: "ALL", "CTOR", "CTOR02"
func GetCodesForCheck(code string) iter.Seq[string] {
	return func(yield func(string) bool) {
		checkList, exists := codeToCheckList[code]
		if !exists {
			// Unknown code, just check ALL and the code itself
			if !yield("ALL") {
				return
			}
			yield(code)
			return
		}

		// Yield all codes from the pre-built check list
		for _, c := range checkList {
			if !yield(c) {
				return
			}
		}
	}
}
