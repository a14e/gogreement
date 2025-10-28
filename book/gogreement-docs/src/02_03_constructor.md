# @constructor Annotation

The `@constructor` annotation restricts object instantiation to specific functions. Objects can only be created within the designated constructor functions.

## Motivation

Go doesn't have built-in mechanisms to restrict how objects are created. This is problematic when:

- Objects require specific initialization (database connections, sockets)
- Invariants must be established at creation time
- Factory patterns are required for proper setup
- You want to ensure validation happens during construction

The `@constructor` annotation fills this gap by enforcing that objects are only created through designated functions.

## Syntax

```go
// @constructor FunctionName
// @constructor Func1, Func2, Func3
type TypeName struct {
    // fields
}
```

### Parameters

- **Function Names** (required): Comma-separated list of constructor function names
- Functions must be in the **same package** as the type
- If a specified function doesn't exist, no error is raised

## How It Works

GoGreement detects the following violations outside constructor functions:

1. **Composite literals**: `TypeName{}`
2. **new() calls**: `new(TypeName)`
3. **Var declarations**: `var x TypeName`

## Key Behaviors

1. **No generics support**: Cannot be used with generic types
2. **Same package only**: Constructor functions must be in the same package as the type
3. **Non-existent constructors OK**: No error if a named constructor doesn't exist
4. **Can be suppressed**: Use `@ignore` to allow creation in specific places
5. **Cross-package enforcement**: Works even if `@constructor` was declared in an external module

## Can Be Declared On

### Struct Types

```go
// @constructor NewDatabase
type Database struct {
    conn *sql.DB
}
```

### Interface Types

```go
// @constructor NewReader
type Reader interface {
    Read(p []byte) (n int, err error)
}
```

### Named Types

```go
// @constructor NewStatus
type Status int
```

## Error Codes

| Code | Description | Example |
|------|-------------|---------|
| **CTOR01** | Composite literal outside constructor | `db := Database{}` |
| **CTOR02** | new() call outside constructor | `db := new(Database)` |
| **CTOR03** | Var declaration creates zero-initialized instance | `var db Database` |

## Examples

### ✅ Basic Constructor Pattern

```go
// @constructor NewDatabase
type Database struct {
    conn *sql.DB
}

func NewDatabase(dsn string) (*Database, error) {
    conn, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    return &Database{conn: conn}, nil  // ✅ Allowed in constructor
}
```

### ✅ Multiple Constructors

```go
// @constructor New, NewWithDefaults, MustNew
type Config struct {
    Host string
    Port int
}

func New(host string, port int) *Config {
    return &Config{Host: host, Port: port}  // ✅ Allowed
}

func NewWithDefaults() *Config {
    return &Config{Host: "localhost", Port: 8080}  // ✅ Allowed
}

func MustNew(host string, port int) *Config {
    if port == 0 {
        panic("port required")
    }
    return &Config{Host: host, Port: port}  // ✅ Allowed
}
```

### ❌ Composite Literal Outside Constructor

```go
// @constructor NewUser
type User struct {
    ID   int
    Name string
}

func NewUser(id int, name string) *User {
    return &User{ID: id, Name: name}  // ✅ Allowed
}

func createBroken() {
    u := User{ID: 1, Name: "Alice"}  // ❌ ERROR: CTOR01
}
```

### ❌ Using new() Outside Constructor

```go
// @constructor NewBuffer
type Buffer struct {
    data []byte
}

func NewBuffer(size int) *Buffer {
    return &Buffer{data: make([]byte, size)}  // ✅ Allowed
}

func allocateBroken() {
    buf := new(Buffer)  // ❌ ERROR: CTOR02
}
```

### ❌ Var Declaration Outside Constructor

```go
// @constructor NewPoint
type Point struct {
    X, Y int
}

func NewPoint(x, y int) Point {
    return Point{X: x, Y: y}  // ✅ Allowed
}

func useBroken() {
    var p Point  // ❌ ERROR: CTOR03
    p = NewPoint(1, 2)  // Too late - already zero-initialized
}
```

### ✅ Using @ignore to Suppress

```go
// @constructor NewCache
type Cache struct {
    items map[string]string
}

func NewCache() *Cache {
    return &Cache{items: make(map[string]string)}
}

func resetCache(c *Cache) {
    // @ignore CTOR01
    *c = Cache{items: make(map[string]string)}  // ✅ Suppressed
}
```

### ✅ Cross-Package Enforcement

**Package `db`:**

```go
package db

// @constructor Open
type Connection struct {
    dsn string
}

func Open(dsn string) (*Connection, error) {
    return &Connection{dsn: dsn}, nil
}
```

**Package `main`:**

```go
package main

import "myapp/db"

func main() {
    // ❌ ERROR: CTOR01 - Connection requires constructor
    conn := db.Connection{dsn: "localhost"}

    // ✅ Correct: Use constructor
    conn, err := db.Open("localhost")
}
```

### ✅ With Validation

```go
// @constructor NewEmail
type Email string

func NewEmail(s string) (Email, error) {
    if !strings.Contains(s, "@") {
        return "", errors.New("invalid email")
    }
    return Email(s), nil  // ✅ Allowed
}

func validate(input string) {
    // ❌ ERROR: CTOR01
    email := Email(input)  // Bypass validation!

    // ✅ Correct: Use constructor
    email, err := NewEmail(input)
}
```

### ✅ Builder Pattern

```go
// @constructor NewRequestBuilder
type RequestBuilder struct {
    method string
    url    string
    headers map[string]string
}

func NewRequestBuilder() *RequestBuilder {
    return &RequestBuilder{  // ✅ Allowed
        headers: make(map[string]string),
    }
}

func (rb *RequestBuilder) Method(m string) *RequestBuilder {
    rb.method = m
    return rb
}

func (rb *RequestBuilder) URL(u string) *RequestBuilder {
    rb.url = u
    return rb
}

func (rb *RequestBuilder) Build() *http.Request {
    // Build actual request
    return nil
}
```

## Best Practices

### 1. Validate in Constructors

Use constructors to enforce invariants:

```go
// @constructor NewPositiveInt
type PositiveInt int

func NewPositiveInt(n int) (PositiveInt, error) {
    if n <= 0 {
        return 0, errors.New("must be positive")
    }
    return PositiveInt(n), nil
}
```

### 2. Initialize Resources

Use constructors for resource acquisition:

```go
// @constructor OpenFile
type File struct {
    handle *os.File
}

func OpenFile(path string) (*File, error) {
    handle, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    return &File{handle: handle}, nil
}
```

### 3. Provide Multiple Constructors

Offer convenience constructors:

```go
// @constructor New, NewDefault, NewFromConfig
type Server struct {
    host string
    port int
}

func New(host string, port int) *Server {
    return &Server{host: host, port: port}
}

func NewDefault() *Server {
    return New("localhost", 8080)
}

func NewFromConfig(cfg *Config) *Server {
    return New(cfg.Host, cfg.Port)
}
```

### 4. Document Constructor Requirements

```go
// @constructor NewPool
// Pool manages a pool of database connections.
// Must be created with NewPool to ensure proper initialization.
type Pool struct {
    conns []*sql.DB
}

func NewPool(size int, dsn string) (*Pool, error) {
    // Initialize pool
    return &Pool{}, nil
}
```

### 5. Combine with @immutable

```go
// @immutable
// @constructor NewConfig
type Config struct {
    timeout time.Duration
}

func NewConfig(timeout time.Duration) *Config {
    return &Config{timeout: timeout}
}
```

## Common Patterns

### Singleton Pattern

```go
// @constructor GetInstance
type Singleton struct {
    data string
}

var instance *Singleton
var once sync.Once

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{data: "initialized"}
    })
    return instance
}
```

### Factory Pattern

```go
// @constructor NewLogger
type Logger interface {
    Log(msg string)
}

type logger struct {
    level string
}

func NewLogger(level string) Logger {
    return &logger{level: level}
}
```

## Limitations

### 1. Non-Existent Constructors

If you misspell a constructor name, no error is raised:

```go
// @constructor NewUzer  // Typo! Should be NewUser
type User struct {
    name string
}

// NewUser is defined, but annotation says NewUzer
func NewUser(name string) *User {
    return &User{name: name}
}

func main() {
    u := User{}  // ❌ Will report CTOR01, even though constructor exists
}
```

**Solution**: Be careful with constructor names.

### 2. Same Package Only

Constructors must be in the same package:

```go
package models

// @constructor factory.CreateUser
// ❌ Won't work - factory is different package
type User struct {}
```

## Related Annotations

- **[@immutable](02_02_immutable.md)**: Often combined to ensure objects can't be mutated after construction
- **[@ignore](02_05_ignore.md)**: Suppress violations when needed

## See Also

- [Error Codes Reference](03_codes.md)
- [Limitations](01_02_limitations.md)