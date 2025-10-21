package analyzer

import (
	"goagreement/src/annotations"
	"goagreement/src/constructor" // 🆕 Новый импорт
	"goagreement/src/immutable"
	"goagreement/src/implements"
	"reflect"

	"golang.org/x/tools/go/analysis"
)

// AnnotationReader reads annotations from code and exports them as facts
var AnnotationReader = &analysis.Analyzer{
	Name:       "annotationreader",
	Doc:        "Reads @implements, @immutable, @constructor annotations from code",
	Run:        runAnnotationReader,
	ResultType: reflect.TypeOf(annotations.PackageAnnotations{}),
	FactTypes: []analysis.Fact{
		new(annotations.PackageAnnotations),
	},
}

func runAnnotationReader(pass *analysis.Pass) (interface{}, error) {
	packageAnnotations := annotations.ReadAllAnnotations(pass)
	pass.ExportPackageFact(&packageAnnotations)
	return packageAnnotations, nil
}

// ImplementsChecker checks @implements annotations
var ImplementsChecker = &analysis.Analyzer{
	Name: "implementschecker",
	Doc:  "Checks that types implement interfaces as declared by @implements",
	Run:  runImplementsChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
	},
}

func runImplementsChecker(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[AnnotationReader]
	if result == nil {
		return nil, nil
	}

	localAnnotations, ok := result.(annotations.PackageAnnotations)
	if !ok {
		return nil, nil
	}

	if len(localAnnotations.ImplementsAnnotations) == 0 {
		return nil, nil
	}

	// Load interfaces and types
	interfaceQueries := localAnnotations.ToInterfaceQuery()
	interfaces := implements.LoadInterfaces(pass, interfaceQueries)

	typeQueries := localAnnotations.ToTypeQuery()
	types := implements.LoadTypes(pass, typeQueries)

	// Validate
	missingPackages := implements.FindMissingPackages(localAnnotations.ImplementsAnnotations)
	missingInterfaces := implements.FindMissingInterfaces(localAnnotations.ImplementsAnnotations, interfaces)
	missingMethods := implements.FindMissingMethods(localAnnotations.ImplementsAnnotations, interfaces, types)

	// Report problems
	implements.ReportProblems(pass, missingPackages, missingInterfaces, missingMethods)

	return nil, nil
}

// ImmutableChecker checks @immutable annotations
var ImmutableChecker = &analysis.Analyzer{
	Name: "immutabilitychecker",
	Doc:  "Checks that types marked as @immutable follow immutability rules",
	Run:  runImmutableChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
	},
}

func runImmutableChecker(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[AnnotationReader]
	if result == nil {
		return nil, nil
	}

	localAnnotations, ok := result.(annotations.PackageAnnotations)
	if !ok {
		return nil, nil
	}

	if len(localAnnotations.ImmutableAnnotations) == 0 {
		return nil, nil
	}

	// Check immutability violations
	violations := immutable.CheckImmutable(pass, localAnnotations)

	// Report violations
	immutable.ReportViolations(pass, violations)

	return nil, nil
}

// ConstructorChecker checks @constructor annotations
var ConstructorChecker = &analysis.Analyzer{
	Name: "constructorchecker",
	Doc:  "Checks that types with @constructor are only instantiated in declared constructors",
	Run:  runConstructorChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
	},
}

func runConstructorChecker(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[AnnotationReader]
	if result == nil {
		return nil, nil
	}

	localAnnotations, ok := result.(annotations.PackageAnnotations)
	if !ok {
		return nil, nil
	}

	if len(localAnnotations.ConstructorAnnotations) == 0 {
		return nil, nil
	}

	// Check constructor violations
	violations := constructor.CheckConstructor(pass, localAnnotations)

	// Report violations
	constructor.ReportViolations(pass, violations)

	return nil, nil
}

// Analyzer is the main entry point combining all checks
var Analyzer = &analysis.Analyzer{
	Name: "goagreement",
	Doc:  "Checks code contracts via annotations",
	Run:  run,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
		ImplementsChecker,
		ImmutableChecker,
		ConstructorChecker,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}
