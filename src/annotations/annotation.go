package annotations

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"

	"github.com/cloudflare/ahocorasick"
	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/config"
	"github.com/a14e/gogreement/src/util"
)

// PackageAnnotations
// @implements &analysis.Fact
// @immutable
type PackageAnnotations struct {
	ImplementsAnnotations  []ImplementsAnnotation
	ConstructorAnnotations []ConstructorAnnotation
	ImmutableAnnotations   []ImmutableAnnotation
	TestonlyAnnotations    []TestOnlyAnnotation
	MutableAnnotations     []MutableAnnotation
}

func (*PackageAnnotations) AFact() {}

// AnnotationWrapper is an interface for all fact types that wrap PackageAnnotations
// This allows generic access to annotations while maintaining unique fact types per analyzer
type AnnotationWrapper interface {
	analysis.Fact
	GetAnnotations() *PackageAnnotations
	Empty() AnnotationWrapper
}

// Unique fact types for each analyzer to enable cross-package fact sharing
// Each analyzer needs its own fact type because go/analysis doesn't allow
// multiple analyzers to export facts with the same type

// AnnotationReaderFact is used by AnnotationReader analyzer
// @implements &analysis.Fact
// @implements &AnnotationWrapper
type AnnotationReaderFact PackageAnnotations

func (*AnnotationReaderFact) AFact() {}

func (f *AnnotationReaderFact) GetAnnotations() *PackageAnnotations {
	return (*PackageAnnotations)(f)
}

func (*AnnotationReaderFact) Empty() AnnotationWrapper {
	return &AnnotationReaderFact{}
}

// ImplementsCheckerFact is used by ImplementsChecker analyzer
// @implements &analysis.Fact
// @implements &AnnotationWrapper
type ImplementsCheckerFact PackageAnnotations

func (*ImplementsCheckerFact) AFact() {}

func (f *ImplementsCheckerFact) GetAnnotations() *PackageAnnotations {
	return (*PackageAnnotations)(f)
}

func (*ImplementsCheckerFact) Empty() AnnotationWrapper {
	return &ImplementsCheckerFact{}
}

// ImmutableCheckerFact is used by ImmutableChecker analyzer
// @implements &analysis.Fact
// @implements &AnnotationWrapper
type ImmutableCheckerFact PackageAnnotations

func (*ImmutableCheckerFact) AFact() {}

func (f *ImmutableCheckerFact) GetAnnotations() *PackageAnnotations {
	return (*PackageAnnotations)(f)
}

func (*ImmutableCheckerFact) Empty() AnnotationWrapper {
	return &ImmutableCheckerFact{}
}

// ConstructorCheckerFact is used by ConstructorChecker analyzer
// @implements &analysis.Fact
// @implements &AnnotationWrapper
type ConstructorCheckerFact PackageAnnotations

func (*ConstructorCheckerFact) AFact() {}

func (f *ConstructorCheckerFact) GetAnnotations() *PackageAnnotations {
	return (*PackageAnnotations)(f)
}

func (*ConstructorCheckerFact) Empty() AnnotationWrapper {
	return &ConstructorCheckerFact{}
}

// TestOnlyCheckerFact is used by TestOnlyChecker analyzer
// @implements &analysis.Fact
// @implements &AnnotationWrapper
type TestOnlyCheckerFact PackageAnnotations

func (*TestOnlyCheckerFact) AFact() {}

func (f *TestOnlyCheckerFact) GetAnnotations() *PackageAnnotations {
	return (*PackageAnnotations)(f)
}

func (*TestOnlyCheckerFact) Empty() AnnotationWrapper {
	return &TestOnlyCheckerFact{}
}

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

// ConstructorAnnotation
// @constructor parseConstructorAnnotation
// @immutable
type ConstructorAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos

	ConstructorNames []string // ["New", "Create"]
}

// ImmutableAnnotation
// @immutable
// @constructor parseImmutableAnnotation
type ImmutableAnnotation struct {
	// Type on which annotation is placed
	OnType    string // "MyStruct"
	OnTypePos token.Pos
}

// TestOnlyKind represents the kind of declaration @testonly is placed on
type TestOnlyKind int

const (
	TestOnlyOnType   TestOnlyKind = iota // @testonly on type (struct, interface, etc)
	TestOnlyOnFunc                       // @testonly on function
	TestOnlyOnMethod                     // @testonly on method
)

// TestOnlyAnnotation
// @immutable
type TestOnlyAnnotation struct {
	// Kind of declaration: type, func, or method
	Kind TestOnlyKind

	// Name of the object: type name, function name, or method name
	// Examples: "MyStruct", "MyFunction", "MyMethod"
	ObjectName string
	Pos        token.Pos

	// Receiver type (only for methods, empty otherwise)
	// Example: "MyStruct" for method receivers
	ReceiverType string
}

// MutableAnnotation
// @immutable
// @constructor parseMutableAnnotation
type MutableAnnotation struct {
	// Type on which the field is defined
	OnType string // "MyStruct"

	// Field name that is marked as mutable
	FieldName string // "MutableField"

	// Position of the field declaration
	Pos token.Pos
}

// TypeQuery represents what type we're looking for
// @immutable
type TypeQuery struct {
	TypeName string
	// No PackageName - we only search in the current package
}

// InterfaceQuery represents what interface we're looking for
// @immutable
type InterfaceQuery struct {
	InterfaceName string
	PackageName   string // empty string means current package
}

func (p *PackageAnnotations) ToInterfaceQuery() []InterfaceQuery {
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

func (p *PackageAnnotations) ToTypeQuery() []TypeQuery {
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
	`^\s*//\s*@implements\s+(&)?(?:(\w+)\.)?(\w+)(?:\s+.*)?$`,
	//                           ^1   ^2         ^3
	// 1: pointer (optional)
	// 2: package (optional)
	// 3: interface name (required)
)

var constructorRegex = regexp.MustCompile(
	`^\s*//\s*@constructor(?:\s+([a-zA-Z_][a-zA-Z0-9_]*(?:\s*,\s*[a-zA-Z_][a-zA-Z0-9_]*)*(?:\s*,)?))?(?:\s+.*)?$`,
	//                              ^1
	// 1: comma-separated constructor names (only valid Go identifiers, optional trailing comma)
)

var immutableRegex = regexp.MustCompile(
	`^\s*//\s*@immutable(?:\s+.*)?$`,
	//                              ^1
	// 1: comma-separated constructor names (optional)
)

var testonlyRegex = regexp.MustCompile(
	`^\s*//\s*@testonly(?:\s+.*)?$`,
	//                              ^1
	// 1: comma-separated constructor names (optional)
)

var mutableRegex = regexp.MustCompile(
	`^\s*//\s*@mutable(?:\s+.*)?$`,
)

// parseImplementsAnnotation parses string "@implements &pkg.Interface" or "@implements Interface"
// and resolves package path immediately using importMap
func parseImplementsAnnotation(
	commentText string,
	typeName string,
	pos token.Pos,
	imports *util.ImportMap,
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

	// match[1] = "New,Create" or "" (regex now captures only valid identifiers)
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

func parseTestOnlyAnnotation(commentText string, objectName string, pos token.Pos, kind TestOnlyKind, receiverType string) *TestOnlyAnnotation {
	match := testonlyRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	return &TestOnlyAnnotation{
		Kind:         kind,
		ObjectName:   objectName,
		Pos:          pos,
		ReceiverType: receiverType,
	}
}

func parseMutableAnnotation(commentText string, typeName string, fieldName string, pos token.Pos) *MutableAnnotation {
	match := mutableRegex.FindStringSubmatch(commentText)
	if match == nil {
		return nil
	}

	return &MutableAnnotation{
		OnType:    typeName,
		FieldName: fieldName,
		Pos:       pos,
	}
}

// getFuncKindAndReceiver determines if a function declaration is a method or function
// Returns: (kind, receiverType)
// - For methods: (TestOnlyOnMethod, "MyStruct")
// - For functions: (TestOnlyOnFunc, "")
func getFuncKindAndReceiver(funcDecl *ast.FuncDecl) (TestOnlyKind, string) {
	if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
		// It's a method
		receiverType := ExtractReceiverType(funcDecl.Recv.List[0].Type)
		return TestOnlyOnMethod, receiverType
	}
	// It's a function
	return TestOnlyOnFunc, ""
}

// ExtractReceiverType extracts the receiver type name from a receiver type expression
// Examples: *MyStruct -> MyStruct, MyStruct -> MyStruct
func ExtractReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		// Pointer receiver: *MyStruct
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		// Value receiver: MyStruct
		return t.Name
	}
	return ""
}

var matcher = ahocorasick.NewStringMatcher([]string{
	"@implements",
	"@constructor",
	"@immutable",
	"@testonly",
	"@mutable",
	"@usein",
})

func ReadAllAnnotations(
	cfg *config.Config,
	pass *analysis.Pass,
) PackageAnnotations {
	var implements []ImplementsAnnotation
	var constructors []ConstructorAnnotation
	var immutables []ImmutableAnnotation
	var testonly []TestOnlyAnnotation
	var mutables []MutableAnnotation

	currentPkgPath := pass.Pkg.Path()

	// Filter files based on configuration (skip test files by default)
	filesToScan := cfg.FilterFiles(pass)

	for file := range filesToScan {
		// Build import map for this file
		imports := &util.ImportMap{}
		for _, imp := range file.Imports {
			imports.Add(imp, pass.Pkg)
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
					if !matcher.Contains([]byte(text)) {
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

							// Read field annotations for this immutable type
							fieldMutables := readFieldAnnotationsForType(typeSpec, typeName)
							mutables = append(mutables, fieldMutables...)
						}
					}

					// Parse @testonly
					if strings.Contains(text, "@testonly") {
						annotation := parseTestOnlyAnnotation(text, typeName, pos, TestOnlyOnType, "")
						if annotation != nil {
							testonly = append(testonly, *annotation)
						}
					}
				}
			}
		}

		// Process function and method declarations for @testonly
		for _, n := range file.Decls {
			funcDecl, ok := n.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Doc == nil {
				continue
			}

			funcName := funcDecl.Name.Name
			pos := funcDecl.Pos()

			// Determine if it's a method or function
			kind, receiverType := getFuncKindAndReceiver(funcDecl)

			for _, comment := range funcDecl.Doc.List {
				text := comment.Text

				// Micro-optimization: skip comments without annotations
				if !matcher.Contains([]byte(text)) {
					continue
				}

				// Parse @testonly
				if strings.Contains(text, "@testonly") {
					annotation := parseTestOnlyAnnotation(text, funcName, pos, kind, receiverType)
					if annotation != nil {
						testonly = append(testonly, *annotation)
					}
				}
			}
		}

	}

	return PackageAnnotations{
		ImplementsAnnotations:  implements,
		ConstructorAnnotations: constructors,
		ImmutableAnnotations:   immutables,
		TestonlyAnnotations:    testonly,
		MutableAnnotations:     mutables,
	}
}

// readFieldAnnotationsForType scans struct fields for annotations (currently only @mutable)
// Called only for types that have @immutable annotation
func readFieldAnnotationsForType(typeSpec *ast.TypeSpec, typeName string) []MutableAnnotation {
	var mutables []MutableAnnotation

	// Only process struct types
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		return mutables
	}

	// Iterate through struct fields
	for _, field := range structType.Fields.List {
		// Skip fields without names (embedded fields)
		if len(field.Names) == 0 {
			continue
		}

		// Check for field documentation comments
		if field.Doc == nil {
			continue
		}

		// Process each field name (multiple fields can be declared together)
		for _, fieldName := range field.Names {
			pos := fieldName.Pos()

			// Check each comment for @mutable annotation
			for _, comment := range field.Doc.List {
				text := comment.Text

				// Micro-optimization: skip comments without annotations
				if !matcher.Contains([]byte(text)) {
					continue
				}

				// Parse @mutable
				if strings.Contains(text, "@mutable") {
					annotation := parseMutableAnnotation(text, typeName, fieldName.Name, pos)
					if annotation != nil {
						mutables = append(mutables, *annotation)
					}
				}
			}
		}
	}

	return mutables
}
