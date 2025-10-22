package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.NotNil(t, cfg)
	assert.False(t, cfg.ScanTests, "Default should have ScanTests = false")
}

func TestNew(t *testing.T) {
	t.Run("with ScanTests = true", func(t *testing.T) {
		cfg := New(true, nil)
		assert.True(t, cfg.ScanTests)
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
	})

	t.Run("with ScanTests = false", func(t *testing.T) {
		cfg := New(false, nil)
		assert.False(t, cfg.ScanTests)
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
	})

	t.Run("with custom exclude paths", func(t *testing.T) {
		cfg := New(false, []string{"vendor", "node_modules"})
		assert.False(t, cfg.ScanTests)
		assert.Equal(t, []string{"vendor", "node_modules"}, cfg.ExcludePaths)
	})
}

func TestWithScanTests(t *testing.T) {
	t.Run("immutability - creates new instance", func(t *testing.T) {
		original := New(false, nil)
		modified := original.WithScanTests(true)

		// Original should be unchanged
		assert.False(t, original.ScanTests, "Original config should remain unchanged")

		// Modified should have new value
		assert.True(t, modified.ScanTests, "Modified config should have new value")

		// They should be different instances
		assert.NotEqual(t, original, modified, "Should create a new instance")
	})

	t.Run("change from false to true", func(t *testing.T) {
		cfg := New(false, nil)
		newCfg := cfg.WithScanTests(true)

		assert.False(t, cfg.ScanTests)
		assert.True(t, newCfg.ScanTests)
	})

	t.Run("change from true to false", func(t *testing.T) {
		cfg := New(true, nil)
		newCfg := cfg.WithScanTests(false)

		assert.True(t, cfg.ScanTests)
		assert.False(t, newCfg.ScanTests)
	})
}

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"empty", "", false},
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"True", "True", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"on", "on", true},
		{"ON", "ON", true},
		{"false", "false", false},
		{"FALSE", "FALSE", false},
		{"0", "0", false},
		{"no", "no", false},
		{"invalid", "invalid", false},
		{"random", "xyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("GOAGREEMENT_SCAN_TESTS", tt.envValue)
			} else {
				os.Unsetenv("GOAGREEMENT_SCAN_TESTS")
			}

			// Clean up after test
			defer os.Unsetenv("GOAGREEMENT_SCAN_TESTS")

			cfg := FromEnv()
			assert.Equal(t, tt.expected, cfg.ScanTests, "env value %q should result in ScanTests=%v", tt.envValue, tt.expected)
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"TRUE", true},
		{"True", true},
		{"1", true},
		{"yes", true},
		{"YES", true},
		{"Yes", true},
		{"on", true},
		{"ON", true},
		{"On", true},
		{"false", false},
		{"FALSE", false},
		{"False", false},
		{"0", false},
		{"no", false},
		{"NO", false},
		{"off", false},
		{"", false},
		{"invalid", false},
		{"xyz", false},
		{"  true  ", true},   // with spaces
		{"  false  ", false}, // with spaces
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBool(tt.input)
			assert.Equal(t, tt.expected, result, "parseBool(%q) should return %v", tt.input, tt.expected)
		})
	}
}

func TestConfigImmutability(t *testing.T) {
	t.Run("Config should be immutable", func(t *testing.T) {
		cfg1 := New(false, nil)
		cfg2 := cfg1.WithScanTests(true)
		cfg3 := cfg2.WithScanTests(false)

		// All three should have correct values
		assert.False(t, cfg1.ScanTests)
		assert.True(t, cfg2.ScanTests)
		assert.False(t, cfg3.ScanTests)

		// Verify memory addresses are different (different instances)
		assert.NotSame(t, cfg1, cfg2, "cfg1 and cfg2 should be different instances")
		assert.NotSame(t, cfg2, cfg3, "cfg2 and cfg3 should be different instances")
		// Note: cfg1 and cfg3 have same values but should still be different instances
		assert.NotSame(t, cfg1, cfg3, "cfg1 and cfg3 should be different instances even with same values")
	})
}
