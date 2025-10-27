package main

import (
	"gogreement/src/analyzer"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	multichecker.Main(analyzer.AllAnalyzers()...)
}
