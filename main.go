package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"goagreement/src/analyzer"
)

type TestInterface interface {
	Func(int) int
}

// MyStruct
// @implements &TestInterface
// @implements TestInterface
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

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
