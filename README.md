# GoGreement

[![CI](https://github.com/a14e/gogreement/workflows/CI/badge.svg)](https://github.com/a14e/gogreement/actions)
[![codecov](https://codecov.io/github/a14e/gogreement/graph/badge.svg?token=CZWXY3URMF)](https://codecov.io/github/a14e/gogreement)
[![Documentation](https://img.shields.io/badge/mdBook-docs-blue.svg)](https://a14e.github.io/gogreement)
[![Go Reference](https://pkg.go.dev/badge/github.com/a14e/gogreement.svg)](https://pkg.go.dev/github.com/a14e/gogreement)
[![Go Report Card](https://goreportcard.com/badge/github.com/a14e/gogreement)](https://goreportcard.com/report/github.com/a14e/gogreement)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A static analyzer for Go that enforces developer agreements through code annotations.

## What is it?

GoGreement lets you document and enforce contracts directly in your code using annotations. Mark types as immutable, enforce constructor usage, restrict code to tests, or verify interface implementations—the linter will catch violations at build time.

```go
// Mark a type as immutable - prevent field modifications
// @immutable
type Config struct {
    Host string
    Port int
}

func main() {
    cfg := Config{Host: "localhost", Port: 8080}
    cfg.Port = 9000  // [IMM01] immutability violation in type "Config": cannot assign to field "Port" of immutable type
}
```

## Installation

### For stable version (recommended):
```bash
go install github.com/a14e/gogreement/cmd/gogreement@v0.0.1
```

### For latest version:
```bash
go install github.com/a14e/gogreement/cmd/gogreement@latest
```

Run on your project:

```bash
gogreement ./...
```

## Quick Examples

### Prevent mutations with `@immutable`

```go
// @immutable
type Point struct {
    X, Y int
}

p := Point{X: 1, Y: 2}
p.X = 10  // [IMM01] immutability violation in type "Point": cannot assign to field "X" of immutable type
```

### Enforce constructor usage with `@constructor`

```go
// @constructor NewUser
type User struct {
    id   string
    name string
}

func NewUser(name string) *User {
    return &User{id: generateID(), name: name}
}

// This won't compile:
u := User{name: "John"}  // [CTOR01] type instantiation must be in constructor (allowed: [NewUser])
```

### Restrict code to tests with `@testonly`

```go
// @testonly
func CreateTestDatabase() *DB {
    // Only callable from test files
}

// In production code:
db := CreateTestDatabase()  // [TONL02] function CreateTestDatabase is marked @testonly and can only be called in test files
```

### Verify interface implementations with `@implements`

```go
// @implements io.Reader
type BufferedReader struct {
    buf []byte
}

func (r *BufferedReader) Read(p []byte) (n int, err error) {
    // Implementation
}

// If you remove Read() or change its signature, you get an error:
// [IMPL03] type "BufferedReader" does not implement interface "io.Reader"
// missing methods:
//   Read([]byte) (int, error)
```

## Configuration

Control behavior with environment variables or command flags:

```bash
# Skip test files
gogreement --scan-tests=false ./...

# Exclude specific paths
export GOGREEMENT_EXCLUDE_PATHS="vendor,generated"
gogreement ./...

# Disable specific checks
gogreement --exclude-checks=IMM,CTOR ./...
```

## Why use it?

- **Documentation that's enforced**: Annotations serve as both documentation and contracts
- **Catch bugs early**: Violations are caught at build time, not in production
- **Team coordination**: Formalize agreements about how code should be used
- **No runtime overhead**: Pure static analysis

## Documentation

Full documentation is available at [a14e.github.io/gogreement](https://a14e.github.io/gogreement/)

## Contributing

Contributions are welcome. See the [contributing guide](https://a14e.github.io/gogreement/04_contributing.html) for details on:

- Setting up the development environment
- Running tests
- Adding new checkers
- Code style guidelines

## Disclaimer

This tool is provided as-is. While it aims to catch common mistakes and enforce agreements, it's a static analyzer and may have limitations. Always review its output and use it as part of your development workflow, not as a replacement for testing and code review.

## License

MIT License - see LICENSE-MIT file for details