package util

// FIXME a bit inconsistent api. need to refactor

// FuncMap is a two-level map for tracking functions across packages
// First level: package path ("" for current package)
// Second level: type name -> list of constructor function names
// @constructor NewFuncMap
type FuncMap map[string]map[string][]string

// NewFuncMap creates a new FuncMap
func NewFuncMap() FuncMap {
	return make(FuncMap)
}

// Add adds a function mapping to the map for a specific package
// pkgPath: package path (use "" for current package)
// funcName: name of the constructor function
// typeName: name of the type it constructs
func (fm FuncMap) Add(pkgPath string, funcName string, typeName string) {
	if fm[pkgPath] == nil {
		fm[pkgPath] = make(map[string][]string)
	}

	fm[pkgPath][typeName] = append(fm[pkgPath][typeName], funcName)
}

// Match checks if a function is a constructor for the expected type
// This is the primary method for checking if we're in a valid constructor
func (fm FuncMap) Match(pkgPath string, funcName string, expectedType string) bool {
	typeConstructors, pkgExists := fm[pkgPath]
	if !pkgExists {
		return false
	}

	constructors, typeExists := typeConstructors[expectedType]
	if !typeExists {
		return false
	}

	for _, constructor := range constructors {
		if constructor == funcName {
			return true
		}
	}

	return false
}

// GetConstructors returns list of constructor names for a type
// Returns nil if type not found or has no constructors
func (fm FuncMap) GetConstructors(pkgPath string, typeName string) []string {
	typeConstructors, pkgExists := fm[pkgPath]
	if !pkgExists {
		return nil
	}

	return typeConstructors[typeName]
}

// HasType checks if a type has constructor annotations
func (fm FuncMap) HasType(pkgPath string, typeName string) bool {
	typeConstructors, pkgExists := fm[pkgPath]
	if !pkgExists {
		return false
	}

	constructors, typeExists := typeConstructors[typeName]
	return typeExists && len(constructors) > 0
}

// Len returns the total number of functions across all packages
func (fm FuncMap) Len() int {
	total := 0
	for _, typeConstructors := range fm {
		for _, constructors := range typeConstructors {
			total += len(constructors)
		}
	}
	return total
}
