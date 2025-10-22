package testutil

import (
	"testing"

	"goagreement/src/config"
)

// WithTestConfig temporarily overrides the global config for testing
// It returns a cleanup function that should be called with defer
// Example usage:
//
//	defer testutil.WithTestConfig(t)()
//
// @testonly
func WithTestConfig(t *testing.T) func() {
	originalConfig := config.Global

	// For tests, we want to scan all files including testdata
	config.Global = config.New(false, []string{})

	// Return cleanup function
	return func() {
		config.Global = originalConfig
	}
}
