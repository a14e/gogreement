package analyzer

import (
	"golang.org/x/tools/go/analysis"
)

// Analyzer is the entry point for go/analysis
var Analyzer = &analysis.Analyzer{
	Name: "goagreement",
	Doc:  "Checks code contracts via annotations",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// ========== Phase 1: Loading ==========
	annotations := ReadAllImplementsAnnotations(pass)

	interfaceQueries := toInterfaceQuery(annotations)
	interfaces := LoadInterfaces(pass, interfaceQueries)

	typeQueries := toTypeQuery(annotations)
	types := LoadTypes(pass, typeQueries)

	// ========== Phase 2: Validation ==========
	missingPackages := findMissingPackages(annotations)
	missingInterfaces := findMissingInterfaces(annotations, interfaces)
	missingMethods := findMissingMethods(annotations, interfaces, types)

	// ========== Phase 3: Reporting ==========
	reportProblems(pass, missingPackages, missingInterfaces, missingMethods)

	return nil, nil
}
