package main

import (
	"flag"

	"goagreement/src/analyzer"
	"goagreement/src/config"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	// Parse custom flags
	scanTests := flag.Bool("scan-tests", false, "enable scanning of test files (*_test.go)")
	flag.Parse()

	// Update global configuration immutably
	// Start with environment config, then override with flag if specified
	if *scanTests {
		config.Global = config.Global.WithScanTests(true)
	}

	// Run all analyzers separately - multichecker will handle dependencies
	multichecker.Main(
		analyzer.AnnotationReader,
		analyzer.ImplementsChecker,
		analyzer.ImmutableChecker,
		analyzer.ConstructorChecker,
	)
}
