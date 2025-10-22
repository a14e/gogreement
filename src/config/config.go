package config

import (
	"go/ast"
	"os"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Config holds the configuration for goagreement analyzers
// @immutable
// @constructor New, FromEnv, Default
type Config struct {
	// ScanTests determines whether test files should be analyzed
	// By default, test files (*_test.go) are excluded from analysis
	// Environment variable: GOAGREEMENT_SCAN_TESTS=true|false
	ScanTests bool

	// ExcludePaths is a list of path patterns to exclude from analysis
	// Paths are matched as substrings (e.g. "testdata" will exclude any path containing "testdata")
	// Environment variable: GOAGREEMENT_EXCLUDE_PATHS=path1,path2,path3
	// Default: ["testdata"]
	ExcludePaths []string
}

// Default returns the default configuration
func Default() *Config {
	return New(false, []string{"testdata"})
}

// New creates a new Config with specified settings
func New(scanTests bool, excludePaths []string) *Config {
	if excludePaths == nil {
		excludePaths = []string{"testdata"}
	}
	return &Config{
		ScanTests:    scanTests,
		ExcludePaths: excludePaths,
	}
}

// FromEnv creates a new Config from environment variables
func FromEnv() *Config {
	scanTests := false
	excludePaths := []string{"testdata"} // Default

	if envVal := os.Getenv("GOAGREEMENT_SCAN_TESTS"); envVal != "" {
		scanTests = parseBool(envVal)
	}

	if envVal := os.Getenv("GOAGREEMENT_EXCLUDE_PATHS"); envVal != "" {
		// Split by comma and trim each path
		parts := strings.Split(envVal, ",")
		excludePaths = make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				excludePaths = append(excludePaths, trimmed)
			}
		}
	}

	return New(scanTests, excludePaths)
}

// WithScanTests returns a new Config with ScanTests set to the specified value
func (c *Config) WithScanTests(scanTests bool) *Config {
	return New(scanTests, c.ExcludePaths)
}

// WithExcludePaths returns a new Config with ExcludePaths set to the specified value
func (c *Config) WithExcludePaths(excludePaths []string) *Config {
	return New(c.ScanTests, excludePaths)
}

// parseBool parses a string to boolean
// Accepts: "true", "1", "yes", "on" (case-insensitive) as true
// Everything else is false
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))

	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}

	// Also accept "yes" and "on"
	return s == "yes" || s == "on"
}

// ShouldSkipFile returns true if the file should be skipped based on configuration
func (c *Config) ShouldSkipFile(pass *analysis.Pass, file *ast.File) bool {
	position := pass.Fset.Position(file.Pos())
	filename := position.Filename

	// Check exclude paths first (always exclude testdata by default)
	for _, excludePath := range c.ExcludePaths {
		if strings.Contains(filename, excludePath) {
			return true // Skip files in excluded paths
		}
	}

	// Skip test files when ScanTests is false
	if !c.ScanTests && strings.HasSuffix(filename, "_test.go") {
		return true
	}

	return false
}

// FilterFiles returns only the files that should be analyzed based on configuration
func (c *Config) FilterFiles(pass *analysis.Pass) []*ast.File {
	filtered := make([]*ast.File, 0, len(pass.Files))
	for _, file := range pass.Files {
		if !c.ShouldSkipFile(pass, file) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// Global configuration instance - initialized from environment by default
var Global = FromEnv()
