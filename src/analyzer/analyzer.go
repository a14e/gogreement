package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer is the entry point for go/analysis
var Analyzer = &analysis.Analyzer{
	Name: "goagreement",
	Doc:  "Checks code contracts via annotations",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	annotations := readAllImplementsAnnotations(pass)

	for _, ann := range annotations {
		msg := fmt.Sprintf(
			"Type '%s' should implement %s%s%s",
			ann.OnType,
			formatPointer(ann.IsPointer),
			formatPackage(ann.PackageName),
			ann.InterfaceName,
		)

		pass.Report(analysis.Diagnostic{
			Pos:     ann.OnTypePos,
			Message: msg,
		})
	}

	return nil, nil
}

func formatPointer(isPointer bool) string {
	if isPointer {
		return "&"
	}
	return ""
}

func formatPackage(pkg string) string {
	if pkg == "" {
		return ""
	}
	return pkg + "."
}

func readAllImplementsAnnotations(pass *analysis.Pass) []ImplementsAnnotation {
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

					result = append(result, *annotation)
				}
			}

			return true
		})
	}

	return result
}
