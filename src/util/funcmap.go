package util

// FuncMap is a two-level map for tracking functions across packages
// First level: package path ("" for current package)
// Second level: function name -> associated value (e.g., type name it constructs)
type FuncMap map[string]map[string]string

// Add adds a function mapping to the map for a specific package
// Use empty string "" for current package
func (fm FuncMap) Add(pkgPath string, funcName string, value string) {
	if fm[pkgPath] == nil {
		fm[pkgPath] = make(map[string]string)
	}
	fm[pkgPath][funcName] = value
}

// Match checks if a function exists and its value matches the expected value
// Useful for checking if a function is a constructor for a specific type
func (fm FuncMap) Match(pkgPath string, funcName string, expectedValue string) bool {
	pkgFuncs, pkgExists := fm[pkgPath]
	if !pkgExists {
		return false
	}

	value, exists := pkgFuncs[funcName]
	return exists && value == expectedValue
}

// NewFuncMap creates a new FuncMap
func NewFuncMap() FuncMap {
	return make(FuncMap)
}

// Len returns the total number of functions across all packages
func (fm FuncMap) Len() int {
	total := 0
	for _, pkgFuncs := range fm {
		total += len(pkgFuncs)
	}
	return total
}
