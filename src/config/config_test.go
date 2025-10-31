package config

import (
	"bytes"
	"encoding/gob"
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

				cfg := FromEnv()
				assert.Equal(t, tt.expected, cfg.ScanTests, "env value %q should result in ScanTests=%v", tt.envValue, tt.expected)
			})
		}
	})

	t.Run("ExcludePaths defaults to testdata when not set", func(t *testing.T) {
		cfg := FromEnv()
		assert.Equal(t, []string{"testdata"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths empty string means no exclusions", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "")

		cfg := FromEnv()
		assert.Equal(t, []string{}, cfg.ExcludePaths, "empty string should result in no exclusions")
	})

	t.Run("ExcludePaths single path", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor")

		cfg := FromEnv()
		assert.Equal(t, []string{"vendor"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths multiple paths", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor,node_modules,tmp")

		cfg := FromEnv()
		assert.Equal(t, []string{"vendor", "node_modules", "tmp"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths with spaces", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", " vendor , node_modules , tmp ")

		cfg := FromEnv()
		assert.Equal(t, []string{"vendor", "node_modules", "tmp"}, cfg.ExcludePaths)
	})

	t.Run("ExcludePaths with empty items", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", "vendor,,node_modules")

		cfg := FromEnv()
		assert.Equal(t, []string{"vendor", "node_modules"}, cfg.ExcludePaths, "empty items should be filtered out")
	})

	t.Run("ExcludePaths with only spaces and commas", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_PATHS", " , , ")

		cfg := FromEnv()
		assert.Equal(t, []string{}, cfg.ExcludePaths, "only spaces should result in empty array")
	})

	t.Run("ExcludeChecks defaults to empty when not set", func(t *testing.T) {
		cfg := FromEnv()
		assert.Equal(t, []string{}, cfg.ExcludeChecks, "default ExcludeChecks should be empty")
	})

	t.Run("ExcludeChecks empty string means no exclusions", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "")

		cfg := FromEnv()
		assert.Equal(t, []string{}, cfg.ExcludeChecks, "empty string should result in no exclusions")
	})

	t.Run("ExcludeChecks single code", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01")

		cfg := FromEnv()
		assert.Equal(t, []string{"IMM01"}, cfg.ExcludeChecks, "code should be converted to uppercase")
	})

	t.Run("ExcludeChecks multiple codes", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01,ctor,tonl")

		cfg := FromEnv()
		assert.Equal(t, []string{"IMM01", "CTOR", "TONL"}, cfg.ExcludeChecks, "codes should be converted to uppercase")
	})

	t.Run("ExcludeChecks with spaces", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", " imm01 , ctor , tonl ")

		cfg := FromEnv()
		assert.Equal(t, []string{"IMM01", "CTOR", "TONL"}, cfg.ExcludeChecks, "codes should be trimmed and converted to uppercase")
	})

	t.Run("ExcludeChecks with empty items", func(t *testing.T) {
		t.Setenv("GOGREEMENT_EXCLUDE_CHECKS", "imm01,,ctor")

		cfg := FromEnv()
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

func TestFromEnvDefaults(t *testing.T) {
	t.Run("works with default environment", func(t *testing.T) {
		// Don't set env vars, let them be unset to test defaults

		cfg := FromEnv()
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

func TestConfigGobSerialization(t *testing.T) {
	t.Run("config can be serialized and deserialized with gob", func(t *testing.T) {
		// Create a test config with various values
		original := New(true, []string{"vendor", "node_modules", "testdata"}, []string{"IMM01", "CTOR", "TONL"})

		// Serialize to gob
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		err := encoder.Encode(original)
		assert.NoError(t, err, "Config should be gob-encodable")

		// Deserialize from gob
		decoder := gob.NewDecoder(&buf)
		var deserialized Config

		err = decoder.Decode(&deserialized)
		assert.NoError(t, err, "Config should be gob-decodable")

		// Verify all fields match
		assert.Equal(t, original.ScanTests, deserialized.ScanTests, "ScanTests should match after gob serialization")
		assert.Equal(t, original.ExcludePaths, deserialized.ExcludePaths, "ExcludePaths should match after gob serialization")
		assert.Equal(t, original.ExcludeChecks, deserialized.ExcludeChecks, "ExcludeChecks should match after gob serialization")
	})

	t.Run("empty config can be serialized and deserialized", func(t *testing.T) {
		original := Empty()

		// Serialize to gob
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		err := encoder.Encode(original)
		assert.NoError(t, err, "Empty config should be gob-encodable")

		// Deserialize from gob
		decoder := gob.NewDecoder(&buf)
		var deserialized Config

		err = decoder.Decode(&deserialized)
		assert.NoError(t, err, "Empty config should be gob-decodable")

		// Verify all fields match (note: gob converts empty slices to nil)
		assert.Equal(t, original.ScanTests, deserialized.ScanTests)
		assert.Equal(t, []string(nil), deserialized.ExcludePaths)  // gob converts empty slice to nil
		assert.Equal(t, []string(nil), deserialized.ExcludeChecks) // gob converts empty slice to nil
	})

	t.Run("default config can be serialized and deserialized", func(t *testing.T) {
		original := Default()

		// Serialize to gob
		var buf bytes.Buffer
		encoder := gob.NewEncoder(&buf)

		err := encoder.Encode(original)
		assert.NoError(t, err, "Default config should be gob-encodable")

		// Deserialize from gob
		decoder := gob.NewDecoder(&buf)
		var deserialized Config

		err = decoder.Decode(&deserialized)
		assert.NoError(t, err, "Default config should be gob-decodable")

		// Verify all fields match (Default has ["testdata"] in ExcludePaths, so it's preserved)
		assert.Equal(t, original.ScanTests, deserialized.ScanTests)
		assert.Equal(t, original.ExcludePaths, deserialized.ExcludePaths)
		assert.Equal(t, []string(nil), deserialized.ExcludeChecks) // gob converts empty slice to nil
	})
}
