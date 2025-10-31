package config

import (
	"flag"
	"go/ast"
	"iter"
	"os"
	"strconv"
	"strings"

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

// CreateFlagSet creates and returns a flagset with gogreement-specific flags.
// This allows the flags to be registered in the analyzer and appear in help.
// IMPORTANT: Flag names are automatically prefixed with "config" by multichecker framework
// when used in the config analyzer. Do not modify flag names as it affects CLI.
func CreateFlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("gogreement", flag.ExitOnError)

	// Get defaults from environment to avoid duplicating logic
	defaultConfig := FromEnv()

	// Setup flags with env values as defaults from FromEnv()
	fs.Bool("scan-tests", defaultConfig.ScanTests, "Enable analysis of test files")
	fs.String("exclude-paths", strings.Join(defaultConfig.ExcludePaths, ","), "Comma-separated list of paths to exclude from analysis")
	fs.String("exclude-checks", strings.Join(defaultConfig.ExcludeChecks, ","), "Comma-separated list of check codes to exclude from analysis")

	return fs
}

// ParseFlagsFromFlagSet parses configuration from a FlagSet and environment variables.
// Command line flags from the FlagSet take priority over environment variables.
// If GOGREEMENT_ENV_ONLY is set, will use only environment variables (for testing).
func ParseFlagsFromFlagSet(fs *flag.FlagSet) *Config {
	if fs == nil {
		return Empty()
	}

	// Check for env-only mode (for testing)
	if os.Getenv("GOGREEMENT_ENV_ONLY") != "" {
		// In test mode, ignore flag values and use only environment variables
		return FromEnv()
	}

	// Get flag values
	scanTests := fs.Lookup("scan-tests").Value.(flag.Getter).Get().(bool)
	excludePathsStr := fs.Lookup("exclude-paths").Value.String()
	excludeChecksStr := fs.Lookup("exclude-checks").Value.String()

	// Parse flag values
	finalExcludePaths := parseStringList(excludePathsStr, false)
	finalExcludeChecks := parseStringList(excludeChecksStr, true)

	return New(scanTests, finalExcludePaths, finalExcludeChecks)
}

// FromEnv creates a new Config from environment variables.
func FromEnv() *Config {
	// Get environment values
	scanTests := false
	excludePaths := []string{"testdata"} // Default
	excludeChecks := []string{}          // Default - no exclusions

	if envVal := os.Getenv("GOGREEMENT_SCAN_TESTS"); envVal != "" {
		scanTests = parseBool(envVal)
	}

	excludePaths = parseEnvValue("GOGREEMENT_EXCLUDE_PATHS", false, excludePaths)
	excludeChecks = parseEnvValue("GOGREEMENT_EXCLUDE_CHECKS", true, excludeChecks)

	return New(scanTests, excludePaths, excludeChecks)
}

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
