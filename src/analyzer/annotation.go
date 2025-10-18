package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// ImplementsAnnotation
// parse result of "@implements MyStruct" annotation
type ImplementsAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos

	// Interface that should be implemented
	InterfaceName string // "MyInterface"
	PackageName   string // "" for the current package, "io" for imported (short name from annotation)
	IsPointer     bool   // true if "@implements &Interface"

	// Resolved package information (only available after ReadAllImplementsAnnotations)
	// NOTE: This is the only place where we have access to both AST (for comments)
	// and package imports (for resolution). Other loaders are file-agnostic.
	PackageFullPath string // Full import path: "io", "github.com/user/pkg"
	PackageNotFound bool   // true if package was referenced but not found in imports
}

func toInterfaceQuery(input []ImplementsAnnotation) []InterfaceQuery {
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

func toTypeQuery(input []ImplementsAnnotation) []TypeQuery {
	var result []TypeQuery

	var dedupMap = make(map[string]bool)

	for _, v := range input {
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
	//                   ^1   ^2     ^3
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
		// ResolvePackagePath will set packageFullPath and PackageNotFound
	}
}

// resolvePackagePath resolves a short package name to a full import path
func resolvePackagePath(annotation *ImplementsAnnotation, pkg *types.Package) {
	// Empty package name means current package
	if annotation.PackageName == "" {
		annotation.PackageFullPath = pkg.Path()
		annotation.PackageNotFound = false
		return
	}

	// Search in imports by package name (alias or actual name)
	for _, imp := range pkg.Imports() {
		// Check both package name and potential alias
		// Note: we don't have access to actual aliases from AST here,
		// so we match by package's last component name
		if imp.Name() == annotation.PackageName {
			annotation.PackageFullPath = imp.Path()
			annotation.PackageNotFound = false
			return
		}
	}

	// Package not found in imports
	annotation.PackageNotFound = true
	annotation.PackageFullPath = "" // Keep empty to signal unresolved
}

func ReadAllImplementsAnnotations(pass *analysis.Pass) []ImplementsAnnotation {
	var result []ImplementsAnnotation

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok {
				return true
			}

			if genDecl.Tok != token.TYPE {
				return true
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

				for _, comment := range doc.List {
					text := comment.Text
					if !strings.Contains(text, "@implements") {
						continue
					}

					// Parse annotation
					typeName := typeSpec.Name.Name
					pos := typeSpec.Pos()
					annotation := parseImplementsAnnotation(text, typeName, pos)
					if annotation == nil {
						continue // Failed to parse
					}

					// Resolve package path
					resolvePackagePath(annotation, pass.Pkg)

					result = append(result, *annotation)
				}
			}

			return true
		})
	}

	return result
}
