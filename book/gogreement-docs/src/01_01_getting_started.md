# Getting Started

This guide will help you install and start using GoGreement in your Go projects.

## Installation

### Recommended: Binary Installation

The easiest way to install GoGreement is using `go install`:

#### For stable version (recommended):
```bash
go install github.com/a14e/gogreement/cmd/gogreement@v0.0.1
```

#### For latest version:
```bash
go install github.com/a14e/gogreement/cmd/gogreement@latest
```

This will download and install GoGreement to your `$GOPATH/bin` (or `$HOME/go/bin` if you're using Go modules).

**Important**: Make sure your `$GOPATH/bin` (or `$HOME/go/bin`) is in your `PATH` to run `gogreement` from anywhere.

### From Source (for developers)

If you prefer to build from source or want to contribute:

```bash
git clone https://github.com/a14e/gogreement
cd gogreement
make build
```

This will create a `gogreement` (or `gogreement.exe` on Windows) binary in the project directory.

## Usage

GoGreement is built on top of Go's `golang.org/x/tools/go/analysis` framework and uses the multichecker pattern. This means you can run it like any other Go analysis tool.

### Basic Usage

Run GoGreement on your project:

```bash
gogreement ./...
```

This will analyze all packages in the current module.

### Alternative: Using go vet

You can also run GoGreement using `go vet` with the `-vettool` flag:

```bash
go vet -vettool=gogreement ./...
```

This approach has the advantage of **persistent package facts** between analyzers, which can improve cross-package analysis performance and accuracy in large multi-module projects.

### Analyzing Specific Packages

```bash
gogreement ./pkg/mypackage
gogreement ./...
```

### With Standard Multichecker Flags

Since GoGreement uses the standard `analysis` framework, it supports all standard multichecker flags:

```bash
# Run with JSON output
gogreement -json ./...

# Show analyzer documentation
gogreement -help

# Run specific analyzers only (if you extend GoGreement)
gogreement -analyzers=ImmutableChecker ./...
```

## Configuration

GoGreement can be configured using environment variables or command-line flags. **Command-line flags take priority over environment variables.**

### Configuration Options

| Option | Environment Variable | Command-Line Flag | Default | Description |
|--------|---------------------|-------------------|---------|-------------|
| **Scan Tests** | `GOGREEMENT_SCAN_TESTS` | `--config.scan-tests` | `false` | Whether to analyze test files (`*_test.go`). By default, test files are excluded. |
| **Exclude Paths** | `GOGREEMENT_EXCLUDE_PATHS` | `--config.exclude-paths` | `testdata` | Comma-separated list of path patterns to exclude. Paths are matched as substrings. |
| **Exclude Checks** | `GOGREEMENT_EXCLUDE_CHECKS` | `--config.exclude-checks` | _(empty)_ | Comma-separated list of check codes to exclude globally. Supports individual codes (`IMM01`), categories (`IMM`), or `ALL`. |

### Configuration Examples

#### Environment Variables

```bash
# Include test files in analysis
export GOGREEMENT_SCAN_TESTS=true

# Exclude multiple paths
export GOGREEMENT_EXCLUDE_PATHS=testdata,vendor,third_party

# Exclude specific checks globally
export GOGREEMENT_EXCLUDE_CHECKS=IMM01,CTOR

# Exclude entire category
export GOGREEMENT_EXCLUDE_CHECKS=TONL

# Exclude all checks (useful for debugging)
export GOGREEMENT_EXCLUDE_CHECKS=ALL

# Run analysis
gogreement ./...
```

#### Command-Line Flags

```bash
# Enable test file scanning
gogreement --config.scan-tests=true ./...

# Exclude paths
gogreement --config.exclude-paths=testdata,vendor ./...

# Exclude specific error codes
gogreement --config.exclude-checks=IMM01,CTOR02 ./...

# Exclude entire category of checks
gogreement --config.exclude-checks=IMM ./...

# Combined flags
gogreement --config.scan-tests=true --config.exclude-paths=vendor --config.exclude-checks=TONL ./...
```

### Boolean Value Formats

For boolean options like `GOGREEMENT_SCAN_TESTS`, the following values are accepted (case-insensitive):

- **True**: `true`, `1`, `yes`, `on`
- **False**: `false`, `0`, `no`, `off`, or any other value

### Multiple Values

For options that accept multiple values (`GOGREEMENT_EXCLUDE_PATHS`, `GOGREEMENT_EXCLUDE_CHECKS`):

```bash
# Comma-separated with or without spaces
export GOGREEMENT_EXCLUDE_PATHS=testdata,vendor,generated
export GOGREEMENT_EXCLUDE_PATHS="testdata, vendor, generated"

# Error codes are automatically converted to uppercase
export GOGREEMENT_EXCLUDE_CHECKS=imm01,ctor  # Becomes IMM01,CTOR
```

## Excluding Checks

### Module-Level Exclusion

Use `--config.exclude-checks` or `GOGREEMENT_EXCLUDE_CHECKS` to exclude checks across your entire project:

```bash
# Exclude all immutability checks
gogreement --config.exclude-checks=IMM ./...

# Exclude specific codes
gogreement --config.exclude-checks=IMM01,CTOR02,TONL03 ./...
```

### File and Code-Level Exclusion

Use `// @ignore` comments in your code for fine-grained control. See the [@ignore annotation](02_06_ignore.md) documentation for details.

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: GoGreement Analysis

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install GoGreement
        run: go install github.com/a14e/gogreement/cmd/gogreement@v0.0.1

      - name: Run GoGreement
        run: gogreement ./...
        env:
          GOGREEMENT_SCAN_TESTS: false
          GOGREEMENT_EXCLUDE_PATHS: testdata,vendor
```

### Makefile Integration

```makefile
.PHONY: lint
lint:
	gogreement ./...

.PHONY: lint-all
lint-all:
	GOGREEMENT_SCAN_TESTS=true gogreement ./...
```

## Next Steps

Now that you have GoGreement installed and configured:

1. **Learn about annotations**: Read the [Annotations](02_annotations.md) section
2. **Understand error codes**: Check the [Error Codes](03_codes.md) reference
3. **Review limitations**: See [Limitations](01_02_limitations.md) for known constraints