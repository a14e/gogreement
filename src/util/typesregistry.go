package util

// TypeFuncRegistry is a two-level map for tracking functions associated with types across packages
// First level: package path ("" for current package)
// Second level: type name -> list of associated function names (constructors, methods, etc.)
// @constructor NewTypeFuncRegistry
type TypeFuncRegistry map[string]map[string][]string

// NewTypeFuncRegistry creates a new TypeFuncRegistry
func NewTypeFuncRegistry() TypeFuncRegistry {
	return make(TypeFuncRegistry)
}

// Add adds a function mapping to the map for a specific package
// pkgPath: package path (use "" for current package)
// funcName: name of the associated function
// typeName: name of the type this function relates to
func (tfr TypeFuncRegistry) Add(pkgPath string, funcName string, typeName string) {
	if tfr[pkgPath] == nil {
		tfr[pkgPath] = make(map[string][]string)
	}

	tfr[pkgPath][typeName] = append(tfr[pkgPath][typeName], funcName)
}

// Match checks if a function is associated with the expected type
// This is the primary method for checking if we're in a valid context (constructor, method, etc.)
func (tfr TypeFuncRegistry) Match(pkgPath string, funcName string, expectedType string) bool {
	typeFuncs, pkgExists := tfr[pkgPath]
	if !pkgExists {
		return false
	}

	funcs, typeExists := typeFuncs[expectedType]
	if !typeExists {
		return false
	}

	for _, fn := range funcs {
		if fn == funcName {
			return true
		}
	}

	return false
}

// GetFuncs returns list of associated function names for a type
// Returns nil if type not found or has no associated functions
func (tfr TypeFuncRegistry) GetFuncs(pkgPath string, typeName string) []string {
	typeFuncs, pkgExists := tfr[pkgPath]
	if !pkgExists {
		return nil
	}

	return typeFuncs[typeName]
}

// HasType checks if a type has any associated functions
func (tfr TypeFuncRegistry) HasType(pkgPath string, typeName string) bool {
	typeFuncs, pkgExists := tfr[pkgPath]
	if !pkgExists {
		return false
	}

	funcs, typeExists := typeFuncs[typeName]
	return typeExists && len(funcs) > 0
}

// Len returns the total number of functions across all packages
func (tfr TypeFuncRegistry) Len() int {
	total := 0
	for _, typeFuncs := range tfr {
		for _, funcs := range typeFuncs {
			total += len(funcs)
		}
	}
	return total
}

// Empty returns true if the registry contains no functions
func (tfr TypeFuncRegistry) Empty() bool {
	return len(tfr) == 0
}
