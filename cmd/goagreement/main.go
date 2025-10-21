package main

import (
	"goagreement/src/analyzer"
	"goagreement/src/annotations"

	"honnef.co/go/tools/analysis/lint"
	"honnef.co/go/tools/lintcmd"
)

type TestInterface interface {
	Func(int) int
	Func3(int) int
	Func4(string2 string) (int, string)
}

// MyStruct
// @immutable
// @usein main.go, main
// @constructor New
// @shouldcall Cleanup
// @shouldcalloneof Func, Func2
type MyStruct struct {
	// @notnil
	x interface{}
}

func (*MyStruct) Func(x int) int {
	return x
}

func (*MyStruct) Func2(x int) int {
	return x
}

func (*MyStruct) Cleanup(x int) int {
	return x
}

func New() *MyStruct {
	return nil
}

// @immutable
type ImmutableA struct {
	a int
}

// @immutable
type ImmutableB struct {
	a ImmutableA
}

// @immutable
type ImmutableС struct {
	a annotations.TypeQuery
}

func main() {

	// otherwise it doesn't work use facts =(

	analyzers := []*lint.Analyzer{
		{
			Doc:      &lint.RawDocumentation{},
			Analyzer: analyzer.AnnotationReader,
		},
		{
			Doc:      &lint.RawDocumentation{},
			Analyzer: analyzer.ImplementsChecker,
		},
		{
			Doc:      &lint.RawDocumentation{},
			Analyzer: analyzer.ImmutableChecker,
		},
		{
			Doc:      &lint.RawDocumentation{},
			Analyzer: analyzer.ConstructorChecker,
		},
	}

	cmd := lintcmd.NewCommand("goagreement")
	cmd.AddAnalyzers(analyzers...)

	cmd.Run()
}
