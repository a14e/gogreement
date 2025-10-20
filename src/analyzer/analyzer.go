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
	annotations := ReadAllAnnotations(pass)

	interfaceQueries := annotations.toInterfaceQuery()
	interfaces := LoadInterfaces(pass, interfaceQueries)

	typeQueries := annotations.toTypeQuery()
	types := LoadTypes(pass, typeQueries)

	// ========== Phase 2: Validation ==========
	missingPackages := findMissingPackages(annotations.ImplementsAnnotations)
	missingInterfaces := findMissingInterfaces(annotations.ImplementsAnnotations, interfaces)
	missingMethods := findMissingMethods(annotations.ImplementsAnnotations, interfaces, types)

	// ========== Phase 3: Reporting ==========
	reportProblems(pass, missingPackages, missingInterfaces, missingMethods)

	return nil, nil
}
