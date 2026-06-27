package analyzer

import (
	"testing"

	"github.com/a14e/gogreement/src/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/analysis"
)

// TestRunConfigParsesPerPass ensures configuration is parsed on every pass and
// not frozen in a process-global cache, so an env change is reflected.
func TestRunConfigParsesPerPass(t *testing.T) {
	pass := &analysis.Pass{Analyzer: ConfigReader}

	t.Setenv("GOGREEMENT_ENV_ONLY", "1")

	t.Setenv("GOGREEMENT_SCAN_TESTS", "true")
	r1, err := runConfig(pass)
	require.NoError(t, err)
	cfg1, ok := r1.(*config.Config)
	require.True(t, ok)

	t.Setenv("GOGREEMENT_SCAN_TESTS", "false")
	r2, err := runConfig(pass)
	require.NoError(t, err)
	cfg2, ok := r2.(*config.Config)
	require.True(t, ok)

	assert.True(t, cfg1.ScanTests, "first parse should reflect GOGREEMENT_SCAN_TESTS=true")
	assert.False(t, cfg2.ScanTests, "config must be re-parsed per pass; a stale global cache would keep it true")
}
