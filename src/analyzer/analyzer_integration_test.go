package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"goagreement/src/testutil"
)

// Cross-module integration tests for each analyzer

// TestImplementsCheckerCrossModule tests implements checking across modules
func TestImplementsCheckerCrossModule(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, ImplementsChecker, "multimodule_implements/modA", "multimodule_implements/modB")
}

// TestImmutableCheckerCrossModule tests immutability checking across modules
func TestImmutableCheckerCrossModule(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, ImmutableChecker, "multimodule_immutable/modA", "multimodule_immutable/modB")
}

// TestConstructorCheckerCrossModule tests constructor checking across modules
func TestConstructorCheckerCrossModule(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, ConstructorChecker, "multimodule_constructor/modA", "multimodule_constructor/modB")
}

// TestTestOnlyCheckerCrossModule tests testonly checking across modules
func TestTestOnlyCheckerCrossModule(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, TestOnlyChecker, "multimodule_testonly/modA", "multimodule_testonly/modB")
}
