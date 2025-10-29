package main

import (
	"os"

	"github.com/a14e/gogreement/src/analyzer"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	// If no arguments provided, add --help to show multichecker help
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	multichecker.Main(analyzer.AllAnalyzers()...)
}
