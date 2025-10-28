# About GoGreement

Welcome to the GoGreement project documentation!

GoGreement is a static analysis linter that extends Go's capabilities by adding compile-time enforcements for architectural agreements. It helps teams maintain code quality through annotations like immutability, interface implementation contracts, constructor restrictions, and test-only boundaries.

## Why GoGreement?

Programming is about **agreements** between developers. The goal of this project is to help **enforce** these agreements through static analysis.

Many modern programming languages provide built-in mechanisms for:
- Ensuring immutability (`final` in Java, `readonly` in C#)
- Explicitly declaring interface implementations (`implements` keyword)
- Restricting object instantiation (private constructors, factory patterns)
- Marking code as test-only (`@VisibleForTesting` in various languages)

Go doesn't have these built-in features. GoGreement fills this gap.

## How Does It Work?

We add annotations as comments in your Go code:

```go
// @implements io.Writer
// @immutable
// @testonly
// @constructor New
```

Then run the GoGreement linter as part of your static analysis pipeline. It will report errors if the agreements are violated.

## Quick Example

### ✅ Valid Code - Agreement Enforced

```go
package mypackage

import "io"

// @implements &io.Reader
// The linter will verify this annotation
type MyReader struct {
    data []byte
    pos  int
}

// Correctly implements Read method
func (r *MyReader) Read(p []byte) (n int, err error) {
    if r.pos >= len(r.data) {
        return 0, io.EOF
    }

    n = copy(p, r.data[r.pos:])
    r.pos += n
    return n, nil
}
```

### ❌ Invalid Code - Linter Catches the Error

```go
package mypackage

import "io"

// @implements &io.Reader
type BrokenReader struct {
    data []byte
}

// ERROR: Missing Read method!
// Linter will report: IMPL03 - Missing required methods: Read
```

## More Powerful Examples

### Immutability Enforcement

```go
// @immutable
// @constructor NewPoint
type Point struct {
    X, Y int
}

func NewPoint(x, y int) Point {
    return Point{X: x, Y: y}  // ✅ Allowed in constructor
}

func MovePoint(p Point) {
    p.X += 10  // ❌ ERROR: IMM01 - Field of immutable type is being assigned
}

func ValidUsage(p Point) Point {
    // ✅ Correct: Create new instances instead of mutating
    return Point{X: p.X + 10, Y: p.Y}
}
```

### Constructor Restrictions

```go
// @constructor NewDB, MustConnect
type Database struct {
    conn *sql.DB
}

func NewDB(dsn string) (*Database, error) {
    // ✅ Allowed: Named constructor
    return &Database{conn: nil}, nil
}

func createBroken() {
    db := Database{}  // ❌ ERROR: CTOR01 - Use constructor functions
}
```

### Test-Only Boundaries

```go
// @testonly
type MockService struct {
    calls int
}

// @testonly
func CreateMock() *MockService {
    return &MockService{}
}
```

```go
// in production code (not *_test.go)
func productionFunc() {
    mock := CreateMock()  // ❌ ERROR: TONL02 - TestOnly function in non-test
}

// in test file (*_test.go)
func TestMyCode(t *testing.T) {
    mock := CreateMock()  // ✅ Allowed in tests
}
```

## What's Next?

- **[Getting Started](01_01_getting_started.md)** - Install and configure GoGreement
- **[Annotations](02_annotations.md)** - Learn about all available annotations
- **[Error Codes](03_codes.md)** - Reference for all error codes