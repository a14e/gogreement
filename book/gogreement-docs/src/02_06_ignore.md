# @ignore Annotation

The `@ignore` annotation suppresses specific violations in your code. Use it when you need to bypass checks in specific scopes or when a check produces false positives.

## Motivation

Sometimes you need to violate architectural agreements for valid reasons:

- Debugging or temporary code
- Performance-critical sections
- Gradual migration to new patterns
- Working around third-party library constraints

The `@ignore` annotation provides fine-grained control over which violations to suppress and where.

## Syntax

```go
// @ignore CODE1, CODE2
// @ignore CATEGORY
// @ignore ALL
```

### Parameters

- **Error Codes** (required): Comma-separated list of codes to ignore
  - **Specific codes**: `IMM01`, `CTOR02`, `TONL03`, `PKGO01`, `IMPL01`
  - **Categories**: `IMM`, `CTOR`, `TONL`, `PKGO`, `IMPL` (ignores all codes in category)
  - **All violations**: `ALL`
- **Case-insensitive**: `imm01`, `IMM01`, `Imm01` all work (normalized to uppercase)

## Scope Types

### 1. File-Level Scope

Place `@ignore` before the `package` declaration to affect the entire file:

```go
// @ignore IMM
package mypackage

// All immutability checks are suppressed in this file
```

### 2. Block-Level Scope

Place `@ignore` before a declaration to affect that declaration:

```go
// @ignore CTOR01
type User struct {
    ID int
}

func createUser() {
    u := User{ID: 1}  // ✅ Suppressed - CTOR01 ignored for User
}
```

### 3. Inline Scope

Place `@ignore` on the same line as code to affect just that line:

```go
func modify(p *Point) {
    p.X = 10  // @ignore IMM01
}
```

## Key Behaviors

1. **Hierarchical matching**: `ALL` > Category (`IMM`) > Specific code (`IMM01`)
2. **Case-insensitive**: Codes are normalized to uppercase automatically
3. **Only affects checking**: Doesn't affect annotation scanning phase
4. **Module-level option**: Use `--exclude-checks` flag for project-wide exclusions

## Supported Annotations

| Annotation | Supported | Codes |
|------------|-----------|-------|
| **@immutable** | ✅ Yes | IMM01, IMM02, IMM03, IMM04 |
| **@constructor** | ✅ Yes | CTOR01, CTOR02, CTOR03 |
| **@testonly** | ✅ Yes | TONL01, TONL02, TONL03 |
| **@packageonly** | ✅ Yes | PKGO01, PKGO02, PKGO03 |
| **@implements** | ✅ Yes | IMPL01, IMPL02, IMPL03 |

## Examples

### ✅ File-Level Ignore

```go
// @ignore IMM, CTOR
package legacy

// All immutability and constructor checks suppressed in this file

type Config struct {
    value string
}

func mutate(c *Config) {
    c.value = "new"  // ✅ No error - IMM ignored
}

func create() {
    cfg := Config{}  // ✅ No error - CTOR ignored
}
```

### ✅ Block-Level Ignore

```go
package myapp

// @immutable
type Point struct {
    X, Y int
}

// @ignore IMM01, IMM02
func unsafeMutate(p *Point) {
    p.X = 10  // ✅ Suppressed - IMM01 ignored
    p.Y += 5  // ✅ Suppressed - IMM02 ignored
}
```

### ✅ Inline Ignore

```go
// @immutable
type Counter struct {
    value int
}

func increment(c *Counter) {
    c.value++  // @ignore IMM03
}
```

### ✅ Category-Level Ignore

```go
// @ignore IMM
func batchUpdate(points []*Point) {
    for _, p := range points {
        p.X = 0  // ✅ All IMM* codes suppressed
        p.Y = 0  // ✅ All IMM* codes suppressed
        p.X++    // ✅ All IMM* codes suppressed
    }
}
```

### ✅ Ignore All

```go
// @ignore ALL
func debugFunction() {
    // All violations suppressed here
    var db Database{}  // CTOR violations ignored
    db.conn = nil      // IMM violations ignored
}
```

### ✅ Multiple Codes

```go
func complexOperation(p *Point) {
    // @ignore IMM01, IMM02, IMM03
    p.X = 10
    p.Y += 5
    p.X++
}
```

### ✅ Case-Insensitive

```go
// All of these work the same:
// @ignore IMM01
// @ignore imm01
// @ignore Imm01

func modify(p *Point) {
    p.X = 10  // @ignore imm01  (normalized to IMM01)
}
```

### ✅ Gradual Migration

When migrating to immutable types:

```go
// @immutable
type Config struct {
    timeout int
}

// Old code - suppress temporarily during migration
// @ignore IMM
func legacyUpdate(c *Config) {
    c.timeout = 5000  // ✅ Suppressed during migration
}

// New code - follows immutability
func newUpdate(c Config) Config {
    return Config{timeout: 5000}
}
```

### ✅ Performance-Critical Section

```go
// @immutable
type Stats struct {
    counts []int
}

func (s *Stats) incrementUnsafe(index int) {
    // Performance-critical: avoid allocation
    // @ignore IMM04
    s.counts[index]++
}
```

## Best Practices

### 1. Document Why

Always explain why you're ignoring a check:

```go
// @ignore IMM01
// JUSTIFICATION: Need to modify cache for performance
// TODO(user): Refactor to use copy-on-write
func updateCache(c *Cache, key string, value interface{}) {
    c.data[key] = value
}
```

### 2. Be Specific

Prefer specific codes over categories:

```go
// ✅ Good: Specific code
// @ignore IMM01
p.X = 10

// ❌ Bad: Too broad
// @ignore IMM
p.X = 10
p.Y += 5
```

### 3. Minimize Scope

Use inline scope when possible:

```go
// ✅ Good: Minimal scope
func process(p *Point) {
    validate(p)
    p.X = normalize(p.X)  // @ignore IMM01
    process(p)
}

// ❌ Bad: Too broad
// @ignore IMM01
func process(p *Point) {
    validate(p)
    p.X = normalize(p.X)
    process(p)  // IMM01 suppressed here too (unintended)
}
```

### 4. Temporary Ignores

Mark temporary ignores with TODOs:

```go
// @ignore CTOR01
// TODO(alice): Add proper constructor after refactoring
type TempConfig struct {
    value string
}
```

### 5. Review Regularly

Add comments to track ignores:

```go
// @ignore IMM01
// Added: 2024-01-15
// Reason: Performance optimization
// Review: 2024-06-01
func hotPath(data *Data) {
    data.value = compute()
}
```

## Ignore Hierarchy

GoGreement checks codes in this order:

1. **ALL** - Suppresses everything
2. **Category** (e.g., `IMM`) - Suppresses all codes in category
3. **Specific Code** (e.g., `IMM01`) - Suppresses only that code

**Example**:

```go
// @ignore ALL
// Suppresses: IMM01, IMM02, IMM03, IMM04, CTOR01, CTOR02, CTOR03, TONL01, TONL02, TONL03, PKGO01, PKGO02, PKGO03, IMPL01, IMPL02, IMPL03

// @ignore IMM
// Suppresses: IMM01, IMM02, IMM03, IMM04

// @ignore PKGO
// Suppresses: PKGO01, PKGO02, PKGO03

// @ignore IMPL
// Suppresses: IMPL01, IMPL02, IMPL03

// @ignore IMM01
// Suppresses: IMM01 only
```

## Module-Level Exclusion

For project-wide exclusions, use command-line flags instead of `@ignore`:

```bash
# Exclude all immutability checks project-wide
gogreement --exclude-checks=IMM ./...

# Exclude specific codes
gogreement --exclude-checks=IMM01,CTOR02 ./...
```

Or set environment variable:

```bash
export GOGREEMENT_EXCLUDE_CHECKS=IMM,TONL
gogreement ./...
```

See [Getting Started - Configuration](01_01_getting_started.md#configuration) for more details.

## Common Patterns

### Debugging Code

```go
// @ignore ALL
func debugDump(state *State) {
    state.value = readFromDebugger()  // Temporary debug code
}
```

### Third-Party Integration

```go
// @ignore CTOR01
// Reason: Required by external framework
type PluginConfig struct {
    Name string
}
```

### Test Utilities in Production Code

```go
// @ignore TONL02
// Reason: Needed for integration test setup in main package
func resetForIntegrationTests() {
    helper := testHelper()  // Normally not allowed
}
```

## Related

- **[Error Codes Reference](03_codes.md)**: List of all error codes
- **[Getting Started - Configuration](01_01_getting_started.md#configuration)**: Module-level exclusions

## See Also

- [Error Codes](03_codes.md)
- [Configuration](01_01_getting_started.md#configuration)