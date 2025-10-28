package config

import (
	"flag"
	"go/ast"
	"iter"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// Config holds the configuration for gogreement analyzers
// @immutable
// @constructor New
type Config struct {
	// ScanTests determines whether test files should be analyzed
	// By default, test files (*_test.go) are excluded from analysis
	// Environment variable: GOGREEMENT_SCAN_TESTS=true|false
	// Command line flag: --scan-tests=true|false
	ScanTests bool

	// ExcludePaths is a list of path patterns to exclude from analysis
	// Paths are matched as substrings (e.g. "testdata" will exclude any path containing "testdata")
	// Environment variable: GOGREEMENT_EXCLUDE_PATHS=path1,path2,path3
	// Command line flag: --exclude-paths=path1,path2,path3
	// Default: ["testdata"]
	ExcludePaths []string

	// ExcludeChecks is a list of check codes to exclude from analysis
	// Can be individual codes (IMM01) or categories (IMM) or ALL
	// Environment variable: GOGREEMENT_EXCLUDE_CHECKS=IMM01,CTOR,TONL
	// Command line flag: --exclude-checks=IMM01,CTOR,TONL
	// Default: [] (no exclusions)
	ExcludeChecks []string
}

// Default returns the default configuration
func Default() *Config {
	return New(false, []string{"testdata"}, []string{})
}

func Empty() *Config {
	return New(false, []string{}, []string{})
}

// New creates a new Config with specified settings
func New(scanTests bool, excludePaths []string, excludeChecks []string) *Config {
	return &Config{
		ScanTests:     scanTests,
		ExcludePaths:  excludePaths,
		ExcludeChecks: excludeChecks,
	}
}

// FromEnv creates a new Config from environment variables and command line flags.
// Command line flags take priority over environment variables.
func FromEnv() *Config {
	return fromEnvWithFlags(true)
}

// fromEnvWithFlags creates a new Config from environment variables and optionally command line flags.
// If parseFlags is false, only environment variables are used (for testing).
func fromEnvWithFlags(parseFlags bool) *Config {
	// Get environment values first
	scanTests := false
	excludePaths := []string{"testdata"} // Default
	excludeChecks := []string{}          // Default - no exclusions

	if envVal := os.Getenv("GOGREEMENT_SCAN_TESTS"); envVal != "" {
		scanTests = parseBool(envVal)
	}

	excludePaths = parseEnvValue("GOGREEMENT_EXCLUDE_PATHS", false, excludePaths)
	excludeChecks = parseEnvValue("GOGREEMENT_EXCLUDE_CHECKS", true, excludeChecks)

	if parseFlags {
		// Setup flags with env values as defaults
		scanTestsFlag := flag.Bool("scan-tests", scanTests, "Enable analysis of test files")
		excludePathsFlag := flag.String("exclude-paths", strings.Join(excludePaths, ","), "Comma-separated list of paths to exclude from analysis")
		excludeChecksFlag := flag.String("exclude-checks", strings.Join(excludeChecks, ","), "Comma-separated list of check codes to exclude from analysis")

		flag.Parse()

		// Parse flag values (flags take priority)
		finalExcludePaths := parseFlagValue(*excludePathsFlag, false)
		finalExcludeChecks := parseFlagValue(*excludeChecksFlag, true)

		return New(*scanTestsFlag, finalExcludePaths, finalExcludeChecks)
	}

	// Return config from environment variables only
	return New(scanTests, excludePaths, excludeChecks)
}

// cache for FromEnvCached to avoid reallocations
var (
	envCache     *Config
	envCacheOnce sync.Once
)

// parseStringList parses a comma-separated string into a slice of strings
// Converts to uppercase if specified
func parseStringList(input string, toUpper bool) []string {
	if input == "" {
		return []string{}
	}

	parts := strings.Split(input, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			if toUpper {
				trimmed = strings.ToUpper(trimmed)
			}
			result = append(result, trimmed)
		}
	}
	return result
}

// parseEnvValue gets and parses an environment variable
func parseEnvValue(key string, toUpper bool, defaultValue []string) []string {
	if envVal, set := os.LookupEnv(key); set {
		return parseStringList(envVal, toUpper)
	}
	return defaultValue
}

// parseFlagValue parses a command line flag value
func parseFlagValue(value string, toUpper bool) []string {
	return parseStringList(value, toUpper)
}

// FromEnvCached creates a new Config from environment variables with caching.
// The first call will parse environment variables and cache the result.
// Subsequent calls will return the cached config without allocation.
func FromEnvCached() *Config {
	envCacheOnce.Do(func() {
		envCache = fromEnvWithFlags(true)
	})
	return envCache
}

// resetCache resets the cached config for testing purposes
// @testonly
func resetCache() {
	envCache = Empty()
	envCacheOnce = sync.Once{}
}

// fromEnvForTesting creates a new Config from environment variables only (no flags)
// @testonly
func fromEnvForTesting() *Config {
	return fromEnvWithFlags(false)
}

// WithScanTests returns a new Config with ScanTests set to the specified value
func (c *Config) WithScanTests(scanTests bool) *Config {
	return New(scanTests, c.ExcludePaths, c.ExcludeChecks)
}

// WithExcludePaths returns a new Config with ExcludePaths set to the specified value
func (c *Config) WithExcludePaths(excludePaths []string) *Config {
	return New(c.ScanTests, excludePaths, c.ExcludeChecks)
}

// WithExcludeChecks returns a new Config with ExcludeChecks set to the specified value
func (c *Config) WithExcludeChecks(excludeChecks []string) *Config {
	return New(c.ScanTests, c.ExcludePaths, excludeChecks)
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
func (c *Config) FilterFiles(pass *analysis.Pass) iter.Seq[*ast.File] {

	return func(yield func(*ast.File) bool) {
		for _, file := range pass.Files {
			if !c.ShouldSkipFile(pass, file) {
				if !yield(file) {
					return
				}
			}
		}

	}
}
