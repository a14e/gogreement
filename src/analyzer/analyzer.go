package analyzer

import (
	"reflect"

	"golang.org/x/tools/go/analysis"

	"goagreement/src/annotations"
	"goagreement/src/constructor"
	"goagreement/src/immutable"
	"goagreement/src/implements"
	"goagreement/src/testonly"
)

// AnnotationReader reads annotations from code and exports them as facts
var AnnotationReader = &analysis.Analyzer{
	Name: "annotationreader",
	Doc:  "Reads @implements, @immutable, @constructor annotations from code",
	Run:  runAnnotationReader,
	FactTypes: []analysis.Fact{
		(*annotations.AnnotationReaderFact)(nil),
	},
	ResultType: reflect.TypeOf(annotations.PackageAnnotations{}),
}

func runAnnotationReader(pass *analysis.Pass) (interface{}, error) {
	packageAnnotations := annotations.ReadAllAnnotations(pass)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := (*annotations.AnnotationReaderFact)(&packageAnnotations)
	pass.ExportPackageFact(fact)

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
	FactTypes: []analysis.Fact{
		(*annotations.ImplementsCheckerFact)(nil),
	},
}

// Does not use facts between packages, so result works very well
func runImplementsChecker(pass *analysis.Pass) (interface{}, error) {

	result := pass.ResultOf[AnnotationReader]
	if result == nil {
		return nil, nil
	}
	localAnnotations, ok := result.(annotations.PackageAnnotations)
	if !ok {
		return nil, nil
	}

	// Export facts before isProjectPackage check so dependencies can use them
	fact := (*annotations.ImplementsCheckerFact)(&localAnnotations)
	pass.ExportPackageFact(fact)

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
	FactTypes: []analysis.Fact{
		(*annotations.ImmutableCheckerFact)(nil),
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

	// Export facts before isProjectPackage check so dependencies can use them
	fact := (*annotations.ImmutableCheckerFact)(&localAnnotations)
	pass.ExportPackageFact(fact)

	// Note: We still run the checker even if there are no local @immutable annotations,
	// because we need to check for violations of @immutable types from imported packages

	// Check immutability violations
	violations := immutable.CheckImmutable(pass, &localAnnotations)

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
	FactTypes: []analysis.Fact{
		(*annotations.ConstructorCheckerFact)(nil),
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

	// Export facts before isProjectPackage check so dependencies can use them
	fact := (*annotations.ConstructorCheckerFact)(&localAnnotations)
	pass.ExportPackageFact(fact)

	// Note: We still run the checker even if there are no local @constructor annotations,
	// because we need to check for violations of @constructor types from imported packages

	// Check constructor violations
	violations := constructor.CheckConstructor(pass, &localAnnotations)

	// Report violations
	constructor.ReportViolations(pass, violations)

	return nil, nil
}

// TestOnlyChecker checks @testonly annotations
var TestOnlyChecker = &analysis.Analyzer{
	Name: "testonlychecker",
	Doc:  "Checks that @testonly items are only used in test files",
	Run:  runTestOnlyChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
	},
	FactTypes: []analysis.Fact{
		(*annotations.TestOnlyCheckerFact)(nil),
	},
}

func runTestOnlyChecker(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[AnnotationReader]
	if result == nil {
		return nil, nil
	}
	localAnnotations, ok := result.(annotations.PackageAnnotations)
	if !ok {
		return nil, nil
	}

	// Export facts before isProjectPackage check so dependencies can use them
	fact := (*annotations.TestOnlyCheckerFact)(&localAnnotations)
	pass.ExportPackageFact(fact)

	// Note: We still run the checker even if there are no local @testonly annotations,
	// because we need to check for violations of @testonly items from imported packages

	// Check testonly violations
	violations := testonly.CheckTestOnly(pass, &localAnnotations)

	// Report violations
	testonly.ReportViolations(pass, violations)

	return nil, nil
}

// AllAnalyzers returns all available analyzers
func AllAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		AnnotationReader,
		ImplementsChecker,
		ImmutableChecker,
		ConstructorChecker,
		TestOnlyChecker,
	}
}
