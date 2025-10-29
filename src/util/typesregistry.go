package util

// TypeAssociationRegistry is a two-level map for tracking associations with types across packages
// Can be used for functions, fields, methods, etc.
// First level: package path ("" for current package)
// Second level: type name -> list of associated names (constructors, methods, fields, etc.)
// @constructor NewTypeAssociationRegistry
type TypeAssociationRegistry map[string]map[string][]string

// NewTypeAssociationRegistry creates a new TypeAssociationRegistry
func NewTypeAssociationRegistry() TypeAssociationRegistry {
	return make(TypeAssociationRegistry)
}

// Add adds an association to the map for a specific package
// pkgPath: package path (use "" for current package)
// associatedName: name of the associated item (function, field, method, etc.)
// typeName: name of the type this item relates to
func (tar TypeAssociationRegistry) Add(pkgPath string, associatedName string, typeName string) {
	if tar[pkgPath] == nil {
		tar[pkgPath] = make(map[string][]string)
	}

	tar[pkgPath][typeName] = append(tar[pkgPath][typeName], associatedName)
}

// Match checks if an item is associated with the expected type
// This is the primary method for checking if we're in a valid context (constructor, method, field, etc.)
func (tar TypeAssociationRegistry) Match(pkgPath string, associatedName string, expectedType string) bool {
	typeItems, pkgExists := tar[pkgPath]
	if !pkgExists {
		return false
	}

	items, typeExists := typeItems[expectedType]
	if !typeExists {
		return false
	}

	for _, item := range items {
		if item == associatedName {
			return true
		}
	}

	return false
}

// GetAssociated returns list of associated names for a type
// Returns nil if type not found or has no associated items
func (tar TypeAssociationRegistry) GetAssociated(pkgPath string, typeName string) []string {
	typeItems, pkgExists := tar[pkgPath]
	if !pkgExists {
		return nil
	}

	return typeItems[typeName]
}

// HasType checks if a type has any associated items
func (tar TypeAssociationRegistry) HasType(pkgPath string, typeName string) bool {
	typeItems, pkgExists := tar[pkgPath]
	if !pkgExists {
		return false
	}

	items, typeExists := typeItems[typeName]
	return typeExists && len(items) > 0
}

// Len returns the total number of associations across all packages
func (tar TypeAssociationRegistry) Len() int {
	total := 0
	for _, typeItems := range tar {
		for _, items := range typeItems {
			total += len(items)
		}
	}
	return total
}

// Empty returns true if the registry contains no associations
func (tar TypeAssociationRegistry) Empty() bool {
	return len(tar) == 0
}
