package analyzer

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/util"
)

// PackageAnnotations
// @implements &analysis.Fact
// @constructor ReadAllAnnotations
// @immutable
type PackageAnnotations struct {
	ImplementsAnnotations  []ImplementsAnnotation
	ConstructorAnnotations []ConstructorAnnotation
	ImmutableAnnotations   []ImmutableAnnotation
}

func (*PackageAnnotations) AFact() {}

// ImplementsAnnotation
// parse result of "@implements MyStruct" annotation
// @constructor parseImplementsAnnotation
// @immutable
type ImplementsAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos

	// Interface that should be implemented
	InterfaceName string // "MyInterface"
	PackageName   string // "" for the current package, "io" for imported (short name from annotation)
	IsPointer     bool   // true if "@implements &Interface"

	// Resolved package information (only available after ReadAllAnnotations)
	// NOTE: This is the only place where we have access to both AST (for comments)
	// and package imports (for resolution). Other loaders are file-agnostic.
	PackageFullPath string // Full import path: "io", "github.com/user/pkg"
	PackageNotFound bool   // true if package was referenced but not found in imports
}

// @constructor parseConstructorAnnotation
// @immutable
type ConstructorAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos

	ConstructorNames []string // ["New", "Create"]
}

// @immutable
type ImmutableAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos
}

func (p *PackageAnnotations) toInterfaceQuery() []InterfaceQuery {
	input := p.ImplementsAnnotations

	var result []InterfaceQuery

	for _, v := range input {
		if v.PackageNotFound {
			continue
		}
		x := InterfaceQuery{
			InterfaceName: v.InterfaceName,
			PackageName:   v.PackageFullPath, // Use resolved full path
		}
		result = append(result, x)
	}

	return result
}

func (p *PackageAnnotations) toTypeQuery() []TypeQuery {
	input := p.ImplementsAnnotations

	var result []TypeQuery

	var dedupMap = make(map[string]bool)

	for _, v := range input {
		if v.PackageNotFound {
			continue
		}

		if _, ok := dedupMap[v.OnType]; ok {
			continue
		}
		dedupMap[v.OnType] = true

		x := TypeQuery{
			TypeName: v.OnType,
		}
		result = append(result, x)
	}

	return result
}

// Compile regex once
var implementsRegex = regexp.MustCompile(
	`^\s*//\s*@implements\s+(&)?(?:(\w+)\.)?(\w+)\s*$`,
	//                           ^1   ^2         ^3
	// 1: pointer (optional)
	// 2: package (optional)
	// 3: interface name (required)
)

var constructorRegex = regexp.MustCompile(
	`^\s*//\s*@constructor(?:\s+(.+?))?\s*$`,
	//                              ^1
	// 1: comma-separated constructor names (optional)
)

var immutableRegex = regexp.MustCompile(
	`^\s*//\s*@immutable\s*$`,
	//                              ^1
	// 1: comma-separated constructor names (optional)
)

// parseImplementsAnnotation parses string "@implements &pkg.Interface" or "@implements Interface"
// and resolves package path immediately using importMap
func parseImplementsAnnotation(
	commentText string,
	typeName string,
	pos token.Pos,
	imports *importmap.ImportMap,
	currentPkgPath string,
) *ImplementsAnnotation {
	match := implementsRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	// match[1] = "&" or ""
	// match[2] = "pkg" or ""
	// match[3] = "Interface"

	annotation := &ImplementsAnnotation{
		IsPointer:     match[1] == "&",
		PackageName:   match[2],
		InterfaceName: match[3],
		OnType:        typeName,
		OnTypePos:     pos,
	}

	// Resolve package path immediately
	if annotation.PackageName == "" {
		// Current package
		annotation.PackageFullPath = currentPkgPath
		annotation.PackageNotFound = false
	} else {
		// Look up in imports
		imp := imports.Find(annotation.PackageName)
		if imp != nil {
			annotation.PackageFullPath = imp.FullPath
			annotation.PackageNotFound = false
		} else {
			annotation.PackageFullPath = ""
			annotation.PackageNotFound = true
		}
	}

	return annotation
}

// parseConstructorAnnotation parses string "@constructor New" or "@constructor New, Create"
func parseConstructorAnnotation(commentText string, typeName string, pos token.Pos) *ConstructorAnnotation {
	match := constructorRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	// match[1] = "New, Create" or ""
	namesStr := strings.TrimSpace(match[1])

	// If no names provided, return nil (user must specify constructor names explicitly)
	if namesStr == "" {
		return nil
	}

	// Split by comma and trim each name
	var names []string
	parts := strings.Split(namesStr, ",")
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name != "" {
			names = append(names, name)
		}
	}

	// If after trimming we have no names, return nil
	if len(names) == 0 {
		return nil
	}

	return &ConstructorAnnotation{
		OnType:           typeName,
		OnTypePos:        pos,
		ConstructorNames: names,
	}
}

func parseImmutableAnnotation(commentText string, typeName string, pos token.Pos) *ImmutableAnnotation {
	match := immutableRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	return &ImmutableAnnotation{
		OnType:    typeName,
		OnTypePos: pos,
	}
}

// Import represents a single import from AST
type Import struct {
	Alias    string          // explicit alias (if present) or empty
	FullPath string          // full import path like "io" or "github.com/user/pkg"
	Spec     *ast.ImportSpec // original AST node
}

// ImportMap is a collection of imports with lookup methods
type ImportMap []Import

// Add adds an import spec to the map
func (m *ImportMap) Add(spec *ast.ImportSpec) {
	if spec == nil || spec.Path == nil {
		return
	}

	fullPath := strings.Trim(spec.Path.Value, `"`)

	var alias string
	if spec.Name != nil {
		// Explicit alias: import foo "path"
		alias = spec.Name.Name
	}

	*m = append(*m, Import{
		Alias:    alias,
		FullPath: fullPath,
		Spec:     spec,
	})
}

// Find searches for an import by short name
// Returns nil if not found
func (m *ImportMap) Find(shortName string) *Import {
	if shortName == "" {
		return nil
	}

	// Step 1: Search by explicit alias first
	for i := range *m {
		imp := &(*m)[i]
		if imp.Alias != "" && imp.Alias == shortName {
			return imp
		}
	}

	// Step 2: Fallback to suffix match (last path component)
	for i := range *m {
		imp := &(*m)[i]
		// Extract last component from path
		parts := strings.Split(imp.FullPath, "/")
		lastComponent := parts[len(parts)-1]

		if lastComponent == shortName {
			return imp
		}
	}

	// Not found
	return nil
}

func ReadAllAnnotations(pass *analysis.Pass) PackageAnnotations {
	var implements []ImplementsAnnotation
	var constructors []ConstructorAnnotation
	var immutables []ImmutableAnnotation

	currentPkgPath := pass.Pkg.Path()

	for _, file := range pass.Files {
		// Build import map for this file
		imports := &importmap.ImportMap{}
		for _, imp := range file.Imports {
			imports.Add(imp)
		}

		for _, n := range file.Decls {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok {
				continue
			}

			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				// docs can be in type or genDecl
				doc := genDecl.Doc
				if typeSpec.Doc != nil {
					doc = typeSpec.Doc
				}

				if doc == nil {
					continue
				}

				typeName := typeSpec.Name.Name
				pos := typeSpec.Pos()

				for _, comment := range doc.List {
					text := comment.Text

					// Micro-optimization: skip comments without annotations
					if !strings.Contains(text, "@") {
						continue
					}

					// Parse @implements
					if strings.Contains(text, "@implements") {
						annotation := parseImplementsAnnotation(text, typeName, pos, imports, currentPkgPath)
						if annotation != nil {
							implements = append(implements, *annotation)
						}
					}

					// Parse @constructor
					if strings.Contains(text, "@constructor") {
						annotation := parseConstructorAnnotation(text, typeName, pos)
						if annotation != nil {
							constructors = append(constructors, *annotation)
						}
					}

					// Parse @immutable
					if strings.Contains(text, "@immutable") {
						annotation := parseImmutableAnnotation(text, typeName, pos)
						if annotation != nil {
							immutables = append(immutables, *annotation)
						}
					}
				}
			}
		}

	}

	return PackageAnnotations{
		ImplementsAnnotations:  implements,
		ConstructorAnnotations: constructors,
		ImmutableAnnotations:   immutables,
	}
}
