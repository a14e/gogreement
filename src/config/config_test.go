package config

import (
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
		cfg := New(true, []string{"testdata"}, []string{})
		assert.True(t, cfg.ScanTests)
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
		assert.Equal(t, []string{}, cfg.ExcludeChecks)
	})

	t.Run("with ScanTests = false", func(t *testing.T) {
		cfg := New(false, []string{"testdata"}, []string{})
		assert.False(t, cfg.ScanTests)
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
		assert.Equal(t, []string{}, cfg.ExcludeChecks)
	})

	t.Run("with custom exclude paths", func(t *testing.T) {
		cfg := New(false, []string{"vendor", "node_modules"}, []string{})
		assert.False(t, cfg.ScanTests)
		assert.Equal(t, []string{"vendor", "node_modules"}, cfg.ExcludePaths)
		assert.Equal(t, []string{}, cfg.ExcludeChecks)
	})

	t.Run("with exclude checks", func(t *testing.T) {
		cfg := New(false, []string{"testdata"}, []string{"IMM01", "CTOR"})
		assert.False(t, cfg.ScanTests)
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
		assert.Equal(t, []string{"IMM01", "CTOR"}, cfg.ExcludeChecks)
	})
}

func TestWithScanTests(t *testing.T) {
	t.Run("immutability - creates new instance", func(t *testing.T) {
		original := New(false, []string{"testdata"}, []string{})
		modified := original.WithScanTests(true)

		// Original should be unchanged
		assert.False(t, original.ScanTests, "Original config should remain unchanged")

		// Modified should have new value
		assert.True(t, modified.ScanTests, "Modified config should have new value")

		// They should be different instances
		assert.NotEqual(t, original, modified, "Should create a new instance")
	})

	t.Run("change from false to true", func(t *testing.T) {
		cfg := New(false, []string{"testdata"}, []string{})
		newCfg := cfg.WithScanTests(true)

		assert.False(t, cfg.ScanTests)
		assert.True(t, newCfg.ScanTests)
	})

	t.Run("change from true to false", func(t *testing.T) {
		cfg := New(true, []string{"testdata"}, []string{})
		newCfg := cfg.WithScanTests(false)

		assert.True(t, cfg.ScanTests)
		assert.False(t, newCfg.ScanTests)
	})
}

func TestFromEnv(t *testing.T) {
	t.Run("ScanTests parsing", func(t *testing.T) {
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
				t.Setenv("GOGREEMENT_SCAN_TESTS", tt.envValue)

				cfg := fromEnvForTesting()
				assert.Equal(t, tt.expected, cfg.ScanTests, "env value %q should result in ScanTests=%v", tt.envValue, tt.expected)
			})
		}
	})

	t.Run("ExcludePaths defaults to testdata when not set", func(t *testing.T) {
		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths empty string means no exclusions", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{}, cfg.ExcludePaths, "empty string should result in no exclusions")
	})

	t.Run("ExcludePaths single path", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"vendor"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths multiple paths", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor,node_modules,tmp")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"vendor", "node_modules", "tmp"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths with spaces", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", " vendor , node_modules , tmp ")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"vendor", "node_modules", "tmp"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths with empty items", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor,,node_modules")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"vendor", "node_modules"}, cfg.ExcludePaths, "empty items should be filtered out")
	})

	t.Run("ExcludePaths with only spaces and commas", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", " , , ")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{}, cfg.ExcludePaths, "only spaces should result in empty array")
	})

	t.Run("ExcludeChecks defaults to empty when not set", func(t *testing.T) {
		cfg := fromEnvForTesting()
		assert.Equal(t, []string{}, cfg.ExcludeChecks, "default ExcludeChecks should be empty")
	})

	t.Run("ExcludeChecks empty string means no exclusions", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{}, cfg.ExcludeChecks, "empty string should result in no exclusions")
	})

	t.Run("ExcludeChecks single code", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"IMM01"}, cfg.ExcludeChecks, "code should be converted to uppercase")
	})

	t.Run("ExcludeChecks multiple codes", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01,ctor,tonl")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"IMM01", "CTOR", "TONL"}, cfg.ExcludeChecks, "codes should be converted to uppercase")
	})

	t.Run("ExcludeChecks with spaces", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", " imm01 , ctor , tonl ")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"IMM01", "CTOR", "TONL"}, cfg.ExcludeChecks, "codes should be trimmed and converted to uppercase")
	})

	t.Run("ExcludeChecks with empty items", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01,,ctor")

		cfg := fromEnvForTesting()
		assert.Equal(t, []string{"IMM01", "CTOR"}, cfg.ExcludeChecks, "empty items should be filtered out")
	})
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

func TestFromEnvCached(t *testing.T) {
	t.Run("returns cached config on subsequent calls", func(t *testing.T) {
		defer resetCache() // Reset cache before test
		t.Setenv("GOGREEMENT_SCAN_TESTS", "true")
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor,node_modules")
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01,ctor")

		// First call
		cfg1 := FromEnvCached()
		assert.True(t, cfg1.ScanTests)
		assert.Equal(t, []string{"vendor", "node_modules"}, cfg1.ExcludePaths)
		assert.Equal(t, []string{"IMM01", "CTOR"}, cfg1.ExcludeChecks)

		// Second call should return same cached instance
		cfg2 := FromEnvCached()
		assert.Same(t, cfg1, cfg2, "FromEnvCached should return the same cached instance")
		assert.Equal(t, cfg1.ScanTests, cfg2.ScanTests)
		assert.Equal(t, cfg1.ExcludePaths, cfg2.ExcludePaths)
		assert.Equal(t, cfg1.ExcludeChecks, cfg2.ExcludeChecks)
	})

	t.Run("works with default environment", func(t *testing.T) {
		defer resetCache() // Reset cache before test
		// Don't set env vars, let them be unset to test defaults

		cfg := fromEnvForTesting()
		assert.False(t, cfg.ScanTests, "default ScanTests should be false")
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths, "default ExcludePaths should be [testdata]")
	})
}

func TestConfigImmutability(t *testing.T) {
	t.Run("Config should be immutable", func(t *testing.T) {
		cfg1 := New(false, []string{"testdata"}, []string{})
		cfg2 := cfg1.WithScanTests(true)
		cfg3 := cfg2.WithScanTests(false)

		// All three should have correct values
		assert.False(t, cfg1.ScanTests)
		assert.True(t, cfg2.ScanTests)
		assert.False(t, cfg3.ScanTests)

		// Verify instances are different (addresses differ)
		assert.NotSame(t, cfg1, cfg2, "cfg1 and cfg2 should be different instances")
		assert.NotSame(t, cfg2, cfg3, "cfg2 and cfg3 should be different instances")
		assert.NotSame(t, cfg1, cfg3, "cfg1 and cfg3 should be different instances even with same values")
	})
}
