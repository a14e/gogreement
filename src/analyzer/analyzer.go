package analyzer

import (
	config "github.com/a14e/gogreement/src/config"
	"reflect"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/constructor"
	"github.com/a14e/gogreement/src/ignore"
	"github.com/a14e/gogreement/src/immutable"
	"github.com/a14e/gogreement/src/implements"
	"github.com/a14e/gogreement/src/testonly"
	"github.com/a14e/gogreement/src/util"
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
	cfg := config.FromEnvCached()
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.AnnotationReaderFact(packageAnnotations)
	pass.ExportPackageFact(&fact)

	return packageAnnotations, nil
}

// IgnoreReader reads @ignore annotations from code
var IgnoreReader = &analysis.Analyzer{
	Name:       "ignorereader",
	Doc:        "Reads @ignore CODE1, CODE2 annotations from code",
	Run:        runIgnoreReader,
	ResultType: reflect.TypeOf(ignore.IgnoreResult{}),
}

func runIgnoreReader(pass *analysis.Pass) (interface{}, error) {
	cfg := config.FromEnvCached()
	ignoreSet := ignore.ReadIgnoreAnnotations(cfg, pass)

	return ignore.IgnoreResult{
		IgnoreSet: ignoreSet,
	}, nil
}

// ImplementsChecker checks @implements annotations
//
// NOTE: @ignore directives do not work with ImplementsChecker.
// Interface implementation violations cannot be suppressed.
var ImplementsChecker = &analysis.Analyzer{
	Name: "implementschecker",
	Doc:  "Checks that types implement interfaces as declared by @implements",
	Run:  runImplementsChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
		IgnoreReader,
	},
	FactTypes: []analysis.Fact{
		(*annotations.ImplementsCheckerFact)(nil),
	},
}

// runImplementsChecker validates @implements annotations
// NOTE: @ignore directives are not supported for this checker
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
	fact := annotations.ImplementsCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

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
		IgnoreReader,
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
	cfg := config.FromEnvCached()

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.ImmutableCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @immutable annotations,
	// because we need to check for violations of @immutable types from imported packages

	// Get ignore set from IgnoreReader
	var ignoreSet *util.IgnoreSet
	if ignoreResult := pass.ResultOf[IgnoreReader]; ignoreResult != nil {
		if ir, ok := ignoreResult.(ignore.IgnoreResult); ok {
			ignoreSet = ir.IgnoreSet
		}
	}

	// Check immutability violations
	violations := immutable.CheckImmutable(cfg, pass, &localAnnotations)

	// Report violations (filtered by ignore set)
	immutable.ReportViolations(pass, violations, ignoreSet)

	return nil, nil
}

// ConstructorChecker checks @constructor annotations
var ConstructorChecker = &analysis.Analyzer{
	Name: "constructorchecker",
	Doc:  "Checks that types with @constructor are only instantiated in declared constructors",
	Run:  runConstructorChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
		IgnoreReader,
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
	cfg := config.FromEnvCached()

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.ConstructorCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @constructor annotations,
	// because we need to check for violations of @constructor types from imported packages

	// Get ignore set from IgnoreReader
	var ignoreSet *util.IgnoreSet
	if ignoreResult := pass.ResultOf[IgnoreReader]; ignoreResult != nil {
		if ir, ok := ignoreResult.(ignore.IgnoreResult); ok {
			ignoreSet = ir.IgnoreSet
		}
	}

	// Check constructor violations
	violations := constructor.CheckConstructor(cfg, pass, &localAnnotations)

	// Report violations (filtered by ignore set)
	constructor.ReportViolations(pass, violations, ignoreSet)

	return nil, nil
}

// TestOnlyChecker checks @testonly annotations
var TestOnlyChecker = &analysis.Analyzer{
	Name: "testonlychecker",
	Doc:  "Checks that @testonly items are only used in test files",
	Run:  runTestOnlyChecker,
	Requires: []*analysis.Analyzer{
		AnnotationReader,
		IgnoreReader,
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
	cfg := config.FromEnvCached()

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.TestOnlyCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @testonly annotations,
	// because we need to check for violations of @testonly items from imported packages

	// Get ignore set from IgnoreReader
	var ignoreSet *util.IgnoreSet
	if ignoreResult := pass.ResultOf[IgnoreReader]; ignoreResult != nil {
		if ir, ok := ignoreResult.(ignore.IgnoreResult); ok {
			ignoreSet = ir.IgnoreSet
		}
	}

	// Check testonly violations
	// NOTE: ignoreSet is passed to CheckTestOnly for early filtering
	// This is important because reportedTypes deduplication happens during
	// violation detection. If we filter later in ReportViolations, ignored
	// violations would still mark types as "reported", preventing subsequent
	// non-ignored violations of the same type from being detected.
	violations := testonly.CheckTestOnly(cfg, pass, &localAnnotations, ignoreSet)

	// Report violations (already filtered by ignoreSet in CheckTestOnly)
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
		IgnoreReader,
	}
}
