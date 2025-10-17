package analyzer

import (
	"go/token"
	"regexp"
)

// ImplementsAnnotation
// parse result of "@implements MyStruct" annotation
type ImplementsAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos

	// Interface that should be implemented
	InterfaceName string // "MyInterface"
	PackageName   string // "" for the current package, "pkg" for imported
	IsPointer     bool   // true if "@implements &Interface"
}

// Compile regex once
var implementsRegex = regexp.MustCompile(
	`^\s*//\s*@implements\s+(&)?(?:(\w+)\.)?(\w+)\s*$`,
	//						   ^1   ^2		 ^3
	// 1: pointer (optional)
	// 2: package (optional)
	// 3: interface name (required)
)

// parseImplementsAnnotation parses string "@implements &pkg.Interface" or "@implements Interface"
func parseImplementsAnnotation(commentText string, typeName string, pos token.Pos) *ImplementsAnnotation {
	match := implementsRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	// match[1] = "&" or ""
	// match[2] = "pkg" or ""
	// match[3] = "Interface"

	return &ImplementsAnnotation{
		IsPointer:     match[1] == "&",
		PackageName:   match[2],
		InterfaceName: match[3],
		OnType:        typeName,
		OnTypePos:     pos,
	}
}
