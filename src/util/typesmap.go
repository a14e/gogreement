package util

// TypesMap provides efficient lookup for types
// Key: full package path
// Value: map of type names
type TypesMap map[string]map[string]bool

func NewTypesMap() TypesMap {
	return make(TypesMap)
}

func (m TypesMap) Add(pkgPath string, typeName string) {
	if m[pkgPath] == nil {
		m[pkgPath] = make(map[string]bool)
	}
	m[pkgPath][typeName] = true
}

func (m TypesMap) Contains(pkgPath string, typeName string) bool {
	types, ok := m[pkgPath]
	if !ok {
		return false
	}
	return types[typeName]
}

func (m TypesMap) Len() int {
	count := 0
	for _, types := range m {
		count += len(types)
	}
	return count
}
