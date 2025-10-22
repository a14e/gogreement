package main

import (
	"goagreement/src/analyzer"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(analyzer.AllAnalyzers()...)
}
