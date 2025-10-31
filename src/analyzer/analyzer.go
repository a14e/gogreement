package analyzer

import (
	"reflect"
	"sync"

	config "github.com/a14e/gogreement/src/config"

	"golang.org/x/tools/go/analysis"

	"github.com/a14e/gogreement/src/annotations"
	"github.com/a14e/gogreement/src/constructor"
	"github.com/a14e/gogreement/src/ignore"
	"github.com/a14e/gogreement/src/immutable"
	"github.com/a14e/gogreement/src/implements"
	"github.com/a14e/gogreement/src/packageonly"
	"github.com/a14e/gogreement/src/testonly"
)

// ConfigReader cache to avoid recreating config multiple times
var (
	cachedConfig = config.Empty()
	configOnce   sync.Once
)

// runConfig reads configuration from environment variables and command line flags
func runConfig(pass *analysis.Pass) (interface{}, error) {
	// Use sync.Once to ensure config is created only once
	configOnce.Do(func() {
		// Use the analyzer's flags that were parsed by multichecker
		// Note: multichecker automatically adds "config." prefix to all flag names
		// (e.g., "scan-tests" becomes "config.scan-tests" in command line)
		cachedConfig = config.ParseFlagsFromFlagSet(&pass.Analyzer.Flags)
	})

	return cachedConfig, nil
}

// ConfigReader reads configuration from environment variables and command line flags
// This analyzer must run first to provide configuration to other analyzers
// IMPORTANT: Name "config" is required for proper flag prefixing by multichecker
// Do not change the analyzer name as it affects command-line flag names
var ConfigReader = &analysis.Analyzer{
	Name:       "config",
	Doc:        "Reads configuration from environment variables and command line flags",
	Run:        runConfig,
	ResultType: reflect.TypeOf(config.Empty()),
	Flags:      *config.CreateFlagSet(),
}

// AnnotationReader reads annotations from code and exports them as facts
var AnnotationReader = &analysis.Analyzer{
	Name: "annotationreader",
	Doc:  "Reads @implements, @immutable, @constructor, @packageonly annotations from code",
	Run:  runAnnotationReader,
	Requires: []*analysis.Analyzer{
		ConfigReader,
	},
	FactTypes: []analysis.Fact{
		(*annotations.AnnotationReaderFact)(nil),
	},
	ResultType: reflect.TypeOf(annotations.PackageAnnotations{}),
}

func runAnnotationReader(pass *analysis.Pass) (interface{}, error) {
	cfg := pass.ResultOf[ConfigReader].(*config.Config)
	packageAnnotations := annotations.ReadAllAnnotations(cfg, pass)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.AnnotationReaderFact(packageAnnotations)
	pass.ExportPackageFact(&fact)

	return packageAnnotations, nil
}

// IgnoreReader reads @ignore annotations from code
var IgnoreReader = &analysis.Analyzer{
	Name: "ignorereader",
	Doc:  "Reads @ignore CODE1, CODE2 annotations from code",
	Run:  runIgnoreReader,
	Requires: []*analysis.Analyzer{
		ConfigReader,
	},
	ResultType: reflect.TypeOf(ignore.IgnoreResult{}),
}

func runIgnoreReader(pass *analysis.Pass) (interface{}, error) {
	cfg := pass.ResultOf[ConfigReader].(*config.Config)
	ignoreSet := ignore.ReadIgnoreAnnotations(cfg, pass)

	return ignore.IgnoreResult{
		IgnoreSet: ignoreSet,
	}, nil
}

// ImplementsChecker checks @implements annotations
var ImplementsChecker = &analysis.Analyzer{
	Name: "implementschecker",
	Doc:  "Checks that types implement interfaces as declared by @implements",
	Run:  runImplementsChecker,
	Requires: []*analysis.Analyzer{
		ConfigReader,
		AnnotationReader,
		IgnoreReader,
	},
	FactTypes: []analysis.Fact{
		(*annotations.ImplementsCheckerFact)(nil),
	},
}

// runImplementsChecker validates @implements annotations
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

	// Get ignore set from IgnoreReader
	ignoreSet := pass.ResultOf[IgnoreReader].(ignore.IgnoreResult).IgnoreSet

	// Load interfaces and types
	interfaceQueries := localAnnotations.ToInterfaceQuery()
	interfaces := implements.LoadInterfaces(pass, interfaceQueries)

	typeQueries := localAnnotations.ToTypeQuery()
	types := implements.LoadTypes(pass, typeQueries)

	// Validate
	missingPackages := implements.FindMissingPackages(localAnnotations.ImplementsAnnotations)
	missingInterfaces := implements.FindMissingInterfaces(localAnnotations.ImplementsAnnotations, interfaces)
	missingMethods := implements.FindMissingMethods(localAnnotations.ImplementsAnnotations, interfaces, types)

	// Report problems (filtered by ignore set)
	implements.ReportProblems(pass, missingPackages, missingInterfaces, missingMethods, ignoreSet)

	return nil, nil
}

// ImmutableChecker checks @immutable annotations
var ImmutableChecker = &analysis.Analyzer{
	Name: "immutabilitychecker",
	Doc:  "Checks that types marked as @immutable follow immutability rules",
	Run:  runImmutableChecker,
	Requires: []*analysis.Analyzer{
		ConfigReader,
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
	cfg := pass.ResultOf[ConfigReader].(*config.Config)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.ImmutableCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @immutable annotations,
	// because we need to check for violations of @immutable types from imported packages

	// Get ignore set from IgnoreReader
	ignoreSet := pass.ResultOf[IgnoreReader].(ignore.IgnoreResult).IgnoreSet

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
		ConfigReader,
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
	cfg := pass.ResultOf[ConfigReader].(*config.Config)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.ConstructorCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @constructor annotations,
	// because we need to check for violations of @constructor types from imported packages

	// Get ignore set from IgnoreReader
	ignoreSet := pass.ResultOf[IgnoreReader].(ignore.IgnoreResult).IgnoreSet

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
		ConfigReader,
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
	cfg := pass.ResultOf[ConfigReader].(*config.Config)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.TestOnlyCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @testonly annotations,
	// because we need to check for violations of @testonly items from imported packages

	// Get ignore set from IgnoreReader
	ignoreSet := pass.ResultOf[IgnoreReader].(ignore.IgnoreResult).IgnoreSet

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

// PackageOnlyChecker checks @packageonly annotations
var PackageOnlyChecker = &analysis.Analyzer{
	Name: "packageonlychecker",
	Doc:  "Checks that @packageonly items are only used in allowed packages",
	Run:  runPackageOnlyChecker,
	Requires: []*analysis.Analyzer{
		ConfigReader,
		AnnotationReader,
		IgnoreReader,
	},
	FactTypes: []analysis.Fact{
		(*annotations.PackageOnlyCheckerFact)(nil),
	},
}

func runPackageOnlyChecker(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[AnnotationReader]
	if result == nil {
		return nil, nil
	}
	localAnnotations, ok := result.(annotations.PackageAnnotations)
	if !ok {
		return nil, nil
	}
	cfg := pass.ResultOf[ConfigReader].(*config.Config)

	// Export facts before isProjectPackage check so dependencies can use them
	fact := annotations.PackageOnlyCheckerFact(localAnnotations)
	pass.ExportPackageFact(&fact)

	// Note: We still run the checker even if there are no local @packageonly annotations,
	// because we need to check for violations of @packageonly items from imported packages

	// Get ignore set from IgnoreReader
	ignoreSet := pass.ResultOf[IgnoreReader].(ignore.IgnoreResult).IgnoreSet

	// Check packageonly violations
	violations := packageonly.CheckPackageOnly(cfg, pass, &localAnnotations, ignoreSet)

	// Report violations (filtered by ignore set)
	packageonly.ReportViolations(pass, violations)

	return nil, nil
}

// AllAnalyzers returns all available analyzers
func AllAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		ConfigReader,
		AnnotationReader,
		IgnoreReader,
		ImplementsChecker,
		ImmutableChecker,
		ConstructorChecker,
		TestOnlyChecker,
		PackageOnlyChecker,
	}
}
