package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"goagreement/src/analyzer"
)

type TestInterface interface {
	Func(int) int
	Func3(int) int
	Func4(string2 string) (int, string)
}

// MyStruct
// @implements &TestInterface
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

type ImmutableB struct {
	a ImmutableA
}

func main() {
	x := ImmutableB{a: ImmutableA{a: 1}}
	println(x.a.a)
	x.a.a += 1

	multichecker.Main(
		analyzer.AnnotationReader,
		analyzer.ImplementsChecker,
		analyzer.ImmutableChecker,
	)
}
