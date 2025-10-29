# @implements Annotation

The `@implements` annotation enforces that a type implements a specific interface. When you add this annotation, GoGreement verifies at analysis time that all required methods are present with correct signatures.

## Motivation

Historically, Go had no direct way to explicitly declare that a struct implements an interface. You could write:

```go
var _ MyInterface = (*MyStruct)(nil)
```

But this is not ideal - it's cryptic, error-prone, and doesn't clearly express intent.

The `@implements` annotation fills this gap by providing a **clear, explicit declaration** that is verified by the linter.

## Syntax

```go
// @implements InterfaceName
// @implements PackageName.InterfaceName
// @implements &InterfaceName
// @implements &PackageName.InterfaceName
```

### Parameters

- **Interface Name** (required): Name of the interface to implement
- **Package Prefix** (optional): Package name for external interfaces
- **Pointer Marker `&`** (optional): Indicates pointer receiver methods

## How It Works

1. **Annotation is parsed** when GoGreement scans the file
2. **Interface definition is loaded** from the specified package
3. **Method signatures are compared** between the type and interface
4. **Violations are reported** if methods are missing or signatures don't match

### Key Behaviors

1. **No generics support**: Cannot be used with generic types or interfaces
2. **Imports required**: External interfaces must be imported (even with `import _ "package"` if not used)
3. **Pointer vs value**: `@implements Interface` and `@implements &Interface` are different contracts
4. **Signature matching**: Validation is based on method signature comparison
5. **No multi-interface syntax**: Use separate lines for multiple interfaces
6. **Strict parsing**: Extra characters before the annotation will cause it to be ignored
7. **Receiver compatibility**: Pointer receiver methods can satisfy value receiver requirements (following Go's standard method set rules), but value receiver methods cannot satisfy pointer receiver requirements

## Can Be Declared On

### Struct Types

```go
// @implements &io.Reader
type MyReader struct {
    data []byte
    pos  int
}

func (r *MyReader) Read(p []byte) (n int, err error) {
    return 0, nil
}
```

### Named Types

```go
// @implements fmt.Stringer
type Status int

func (s Status) String() string {
    return "status"
}
```

## Error Codes

| Code | Description | Example |
|------|-------------|---------|
| **IMPL01** | Package not found in imports | Using `@implements pkg.Interface` without importing `pkg` |
| **IMPL02** | Interface not found in package | Interface name doesn't exist or is misspelled |
| **IMPL03** | Missing or incorrect methods | Type doesn't implement all required methods with correct signatures |

## Examples

### ✅ Basic Interface Implementation

```go
package main

import "io"

// @implements &io.Reader
type ByteReader struct {
    data []byte
    pos  int
}

func (r *ByteReader) Read(p []byte) (n int, err error) {
    if r.pos >= len(r.data) {
        return 0, io.EOF
    }
    n = copy(p, r.data[r.pos:])
    r.pos += n
    return n, nil
}
```

### ✅ Multiple Interfaces

```go
// @implements &io.Reader
// @implements &io.Closer
type ReaderCloser struct {
    file *os.File
}

func (rc *ReaderCloser) Read(p []byte) (n int, err error) {
    return rc.file.Read(p)
}

func (rc *ReaderCloser) Close() error {
    return rc.file.Close()
}
```

### ✅ Current Package Interface

```go
type Validator interface {
    Validate() error
}

// @implements Validator
type User struct {
    Name string
}

func (u User) Validate() error {
    if u.Name == "" {
        return errors.New("name required")
    }
    return nil
}
```

### ✅ Value vs Pointer Receivers

```go
// @implements fmt.Stringer
type Status int

func (s Status) String() string {
    return fmt.Sprintf("Status(%d)", s)
}

// @implements &io.Writer
type Buffer struct {
    data []byte
}

func (b *Buffer) Write(p []byte) (n int, err error) {
    b.data = append(b.data, p...)
    return len(p), nil
}
```

### ❌ Missing Method

```go
// @implements &io.ReadWriter
type BrokenRW struct {}

func (rw *BrokenRW) Read(p []byte) (n int, err error) {
    return 0, nil
}

// [IMPL03] type "BrokenRW" does not implement interface "&io.ReadWriter"
// missing methods:
//   Write([]byte) (int, error)
```

### ❌ Wrong Signature

```go
// @implements &io.Reader
type BadReader struct {}

// [IMPL03] type "BadReader" does not implement interface "&io.Reader"
// missing methods:
//   Read([]byte) (int, error)
func (r *BadReader) Read(p []byte) int {
    return 0
}
```

### ❌ Package Not Imported

```go
// @implements http.Handler
// [IMPL01] package "http" referenced in @implements annotation on type "MyHandler" is not imported
type MyHandler struct {}
```

**Fix**: Add import

```go
import "net/http"

// @implements http.Handler
type MyHandler struct {}

func (h MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### ❌ Pointer vs Value Mismatch

```go
// @implements io.Writer
type Writer struct {}

func (w *Writer) Write(p []byte) (n int, err error) {
    return len(p), nil
}

// [IMPL03] type "Writer" does not implement interface "io.Writer"
// missing methods:
//   Write([]byte) (int, error)
```

**Fix**: Use `&` in annotation

```go
// @implements &io.Writer
type Writer struct {}

func (w *Writer) Write(p []byte) (n int, err error) {
    return len(p), nil
}
```

## Best Practices

### 1. Always Import Interfaces

Even if the interface isn't used directly, add an import:

```go
import _ "io"

// @implements &io.Reader
type MyReader struct {}
```

### 2. Use Pointer Marker Correctly

Match the receiver type in your implementation:

- Value receivers → `@implements Interface`
- Pointer receivers → `@implements &Interface`

### 3. One Interface Per Line

```go
// ✅ Good
// @implements &io.Reader
// @implements &io.Closer
type RC struct {}

// ❌ Bad - Not supported
// @implements &io.Reader, &io.Closer
type RC struct {}
```

### 4. Document Why

Add comments explaining the purpose:

```go
// @implements &http.Handler handles API requests for user management
type UserHandler struct {
    db *sql.DB
}
```

## Related Annotations

- **[@immutable](02_02_immutable.md)**: Often combined with `@implements` for immutable data structures
- **[@testonly](02_04_testonly.md)**: Use for mock implementations in tests

## See Also

- [Error Codes Reference](03_codes.md)