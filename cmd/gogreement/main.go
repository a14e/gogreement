package main

import (
	"github.com/a14e/gogreement/src/analyzer"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	multichecker.Main(analyzer.AllAnalyzers()...)
}
