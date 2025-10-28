# Contributing

Contributions to GoGreement are welcome! Whether it's bug fixes, new features, documentation improvements, or test coverage, all contributions help make the project better.

## Requirements

When contributing code, please ensure:

1. **Test Coverage**: Every change must be covered by tests (both unit and integration tests)
2. **Test Data Location**: Test data files are located in the `testdata` directory
3. **Code Reuse**: Reuse collection types and utilities from the `util` module instead of duplicating code
4. **Self-Documenting Code**: Use the project's own annotations on the codebase itself
5. **Documentation Updates**: Update documentation in the `book/` directory when making user-facing changes

## Project Structure

### Source Code

- `cmd/gogreement/main.go` - Entry point
- `src/analyzer/` - Analyzer registry and orchestration
- `src/annotations/` - Annotation parsing and fact types
- `src/codes/` - Error code definitions
- `src/config/` - Configuration management
- `src/constructor/` - Constructor checker
- `src/immutable/` - Immutability checker
- `src/implements/` - Interface implementation checker
- `src/testonly/` - Test-only checker
- `src/ignore/` - Ignore directive parsing
- `src/indexing/` - Cross-package fact indexing
- `src/util/` - Shared utilities
- `src/testutil/` - Testing helpers

### Tests

- `testdata/unit/` - Unit test fixtures for individual checkers
- `testdata/integration/src/` - Integration test fixtures for cross-package analysis

### Documentation

- `book/gogreement-docs/` - User documentation built with [mdBook](https://github.com/rust-lang/mdBook)

## Development Workflow

### 1. Setup

```bash
# Clone the repository
git clone https://github.com/a14e/gogreement
cd gogreement

# Install development tools
make install
```

### 2. Make Changes

```bash
# Run tests frequently
make test

# Run linters
make lint

# Format code
make fmt
```

### 3. Testing

```bash
# Run all tests
make test

# Run specific test
go test ./src/immutable/...

# Run integration tests
go test ./src/analyzer/...
```

**Note**: Integration tests use a multi-module structure in `testdata/integration/src`. Each test scenario (e.g., `multimodule_implements/`) is a separate Go module in that directory. The `analysistest.Run()` function automatically adds `/src` to the testdata path and handles module loading.

### 4. Pre-Commit Checks

Before committing, ensure all checks pass:

```bash
# Run full pre-build check
make pre-build
```

This runs:
- `go mod tidy`
- `make fmt`
- `make lint` (golangci-lint and nilaway)
- `make test`

### 5. Build

```bash
make build
```

## Adding New Features

### Adding a New Checker

To add a new annotation checker:

1. **Create checker package** in `src/newchecker/`
   ```
   src/newchecker/
   ├── checker.go
   ├── checker_test.go
   └── reporting.go
   ```

2. **Add annotation type** to `PackageAnnotations` struct in `src/annotations/annotation.go` (if your checker needs a new annotation type):

   ```go
   type PackageAnnotations struct {
       ImplementsAnnotations  []ImplementsAnnotation
       ConstructorAnnotations []ConstructorAnnotation
       ImmutableAnnotations   []ImmutableAnnotation
       TestonlyAnnotations    []TestOnlyAnnotation
       NewCheckerAnnotations  []NewCheckerAnnotation  // Add your annotation slice here
   }
   ```

3. **Define fact type** in `src/annotations/annotation.go`

   **Important**: Each checker must have its own fact type. This is a workaround for a limitation in the Go analysis framework—facts are not shared between different analyzers. By creating separate types (even though they wrap the same `PackageAnnotations`), we allow each checker to export and import facts independently.

   ```go
   type NewCheckerFact PackageAnnotations
   func (*NewCheckerFact) AFact() {}
   func (f *NewCheckerFact) GetAnnotations() *PackageAnnotations { return (*PackageAnnotations)(f) }
   func (*NewCheckerFact) Empty() AnnotationWrapper { return &NewCheckerFact{} }
   ```

4. **Add error codes** in `src/codes/codes.go`
   ```go
   const (
       NewCheckerViolation01 = "NEWC01"
       NewCheckerViolation02 = "NEWC02"
   )

   var CodesByCategory = map[string][]Code{
       // ...
       "NEWC": {
           {NewCheckerViolation01, "Description of NEWC01"},
           {NewCheckerViolation02, "Description of NEWC02"},
       },
   }
   ```

5. **Register analyzer** in `src/analyzer/analyzer.go`
   ```go
   func AllAnalyzers() []*analysis.Analyzer {
       return []*analysis.Analyzer{
           AnnotationReader,
           IgnoreReader,
           NewChecker,  // Add here
           // ...
       }
   }
   ```

6. **Create unit tests** in `testdata/unit/newcheckertests/`

7. **Create integration tests** in `testdata/integration/src/multimodule_newchecker/`

8. **Update documentation** in `book/gogreement-docs/src/02_0X_newchecker.md`

### Adding a New Error Code

Follow these steps when adding a new violation type:

1. **Update** `src/codes/codes.go`:
   ```go
   const (
       ExistingCode01 = "EXIST01"
       ExistingCode02 = "EXIST02"
       NewCode03      = "EXIST03"  // Add new code
   )

   var CodesByCategory = map[string][]Code{
       "EXIST": {
           {ExistingCode01, "Description"},
           {ExistingCode02, "Description"},
           {NewCode03, "Description of new code"},  // Add entry
       },
   }
   ```

2. **Update checker** to use the new code:
   ```go
   import "github.com/a14e/github.com/a14e/gogreement/src/codes"

   violation := Violation{
       Code: codes.NewCode03,
       // ...
   }
   ```

3. **Tests automatically verify**:
   - Code uniqueness
   - Category prefix correctness
   - Reverse mapping

### Best Practices for Error Codes

- **Use different codes for different violation types**: Each distinct violation should have its own code
- **Enable selective suppression**: Users can then use `@ignore EXIST03` to suppress only that specific violation
- **Examples**:
  - `IMM01` for field assignments, `IMM02` for compound assignments
  - `TONL01` for type usage, `TONL02` for function calls

## Testing Guidelines

### Unit Tests

- Located in same package as the code being tested
- Use `testutil.CreateTestPass()` for creating test passes
- Use `testutil.GetUnitTestdataPath()` for test fixtures

```go
func TestMyChecker(t *testing.T) {
    pass := testutil.CreateTestPass(t, "packagename")
    violations := checker.CheckSomething(cfg, pass, annotations)
    require.Equal(t, expectedViolations, violations)
}
```

### Integration Tests

- Located in `src/analyzer/analyzer_integration_test.go`
- Use `testutil.GetIntegrationTestdataPath()` for fixtures
- Test cross-package analysis and package facts

```go
func TestCrossPackage(t *testing.T) {
    testdata := testutil.GetRootTestdataPath() + "/integration"
    analysistest.Run(t, testdata, Checker, "multimodule_x/modA", "multimodule_x/modB")
}
```

## Code Style

- **Comments**: All code comments must be in English
- **Formatting**: Use `make fmt` to format code
- **Linting**: Code must pass `golangci-lint` and `nilaway`
- **No Redundant Comments**: Comments should explain *why*, not *what*. If the code is self-explanatory, comments are unnecessary.

## Third-Party Dependencies

When adding external libraries:

1. Ensure the library has an **MIT-compatible license**
2. Update `THIRD_PARTY_LICENSES` file
3. **Only include direct dependencies** (not transitive or test dependencies)

```bash
# Add dependency
go get github.com/example/library

# Update third-party licenses
# Add entry to THIRD_PARTY_LICENSES with:
# - Library name
# - Version
# - License type
# - License text
```

## Documentation

Documentation is built using [mdBook](https://github.com/rust-lang/mdBook).

### Building Docs Locally

```bash
cd book/gogreement-docs

# Install mdBook
cargo install mdbook

# Serve docs locally
mdbook serve

# Build docs
mdbook build
```

### Documentation Structure

- `book/gogreement-docs/src/SUMMARY.md` - Table of contents
- `book/gogreement-docs/src/01_*.md` - Getting started guides
- `book/gogreement-docs/src/02_*.md` - Annotation references
- `book/gogreement-docs/src/03_codes.md` - Error codes reference
- `book/gogreement-docs/src/04_contributing.md` - This file

## Pull Request Guidelines

Before submitting a pull request:

1. **Run all checks**:
   ```bash
   make pre-build
   ```

2. **Write descriptive commit messages**:
   ```
   Add support for generic types in @immutable

   - Implement generic type parsing
   - Add tests for generic immutable types
   - Update documentation
   ```

3. **Update documentation** if your change affects user-facing behavior

4. **Add tests** for all new functionality

5. **Keep changes focused**: One PR should address one issue or feature

## Getting Help

- **Issues**: Report bugs or request features at [github.com/a14e/gogreement/issues](https://github.com/a14e/gogreement/issues)
- **Discussions**: Ask questions or discuss ideas in GitHub Discussions

## License

By contributing to GoGreement, you agree that your contributions will be licensed under the same license as the project.