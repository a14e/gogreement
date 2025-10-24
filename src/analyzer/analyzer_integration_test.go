package analyzer

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis/analysistest"

	"goagreement/src/ignore"
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

// TestIgnoreReaderCrossModule tests ignore reader across modules
func TestIgnoreReaderCrossModule(t *testing.T) {
	defer testutil.WithTestConfig(t)()

	testdata := analysistest.TestData()
	results := analysistest.Run(t, testdata, IgnoreReader, "multimodule_ignore/modA", "multimodule_ignore/modB")

	// Verify we got results for both modules
	require.Len(t, results, 2, "expected results for two packages")

	// Check modA results
	modAResult := results[0].Result
	require.NotNil(t, modAResult, "expected non-nil result for modA")

	modAIgnoreResult, ok := modAResult.(ignore.IgnoreResult)
	require.True(t, ok, "expected IgnoreResult type for modA")

	// modA has 3 @ignore annotations: LINT001, LINT002+LINT003, DEPRECATED
	require.Equal(t, 3, modAIgnoreResult.IgnoreSet.Len(), "expected 3 ignore annotations in modA")

	// Check modB results
	modBResult := results[1].Result
	require.NotNil(t, modBResult, "expected non-nil result for modB")

	modBIgnoreResult, ok := modBResult.(ignore.IgnoreResult)
	require.True(t, ok, "expected IgnoreResult type for modB")

	// modB has 2 @ignore annotations: LINT004, LINT005+LINT006+LINT007
	require.Equal(t, 2, modBIgnoreResult.IgnoreSet.Len(), "expected 2 ignore annotations in modB")
}
