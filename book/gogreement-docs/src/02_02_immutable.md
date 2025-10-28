# @immutable Annotation

The `@immutable` annotation enforces immutability constraints on types. Once an object is created, its fields cannot be modified.

## Motivation

Go doesn't provide a built-in way to declare types or fields as immutable. This contrasts with many functional programming techniques that rely on immutability to provide additional stability guarantees and reduce risks in concurrent code.

**Use cases**:
- Messages sent through channels
- HTTP request/response objects
- Configuration objects
- Value objects in domain models

The `@immutable` annotation fills this gap by providing compile-time enforcement of immutability.

## Syntax

```go
// @immutable
type TypeName struct {
    // fields
}
```

No parameters required - simply add `// @immutable` above the type declaration.

## How It Works

GoGreement detects the following violations on immutable types:

1. **Field assignments**: `obj.field = value`
2. **Compound assignments**: `obj.field += value`, `obj.field -= value`, etc.
3. **Increment/decrement**: `obj.field++`, `obj.field--`
4. **Index assignments**: `obj.items[0] = value`, `obj.dict["key"] = value`

## Key Behaviors

1. **No generics support**: Cannot be used with generic types
2. **Weak immutability guarantees**:
   - Prevents field assignments and compound operations
   - Prevents assignments through methods
   - Does NOT prevent mutations through pointers or reflection
3. **Constructor exception**: Checks are ignored inside functions marked with `@constructor`
4. **Can be suppressed**: Use `@ignore` to disable checks in specific scopes
5. **Cross-package enforcement**: Works even if `@immutable` was declared in external modules

## Can Be Declared On

### Struct Types

```go
// @immutable
type Point struct {
    X, Y int
}
```

### Interface Types

```go
// @immutable
type ReadOnlyConfig interface {
    GetValue(key string) string
}
```

### Named Types

```go
// @immutable
type UserID string

// @immutable
type StatusCode int
```

## Error Codes

| Code | Description | Example |
|------|-------------|---------|
| **IMM01** | Field assignment | `point.X = 10` |
| **IMM02** | Compound assignment | `point.X += 5`, `point.Y *= 2` |
| **IMM03** | Increment/decrement | `point.X++`, `count--` |
| **IMM04** | Index assignment | `obj.items[0] = value`, `obj.dict["key"] = value` |

## Examples

### ✅ Basic Immutable Type

```go
// @immutable
// @constructor NewPoint
type Point struct {
    X, Y int
}

func NewPoint(x, y int) Point {
    return Point{X: x, Y: y}  // ✅ Allowed in constructor
}

func DoublePoint(p Point) Point {
    // ✅ Correct: Create new instance instead of mutating
    return Point{X: p.X * 2, Y: p.Y * 2}
}
```

### ✅ Immutable Configuration

```go
// @immutable
// @constructor LoadConfig
type Config struct {
    Host string
    Port int
    TLS  bool
}

func LoadConfig(path string) (*Config, error) {
    cfg := &Config{  // ✅ Allowed in constructor
        Host: "localhost",
        Port: 8080,
        TLS:  false,
    }
    return cfg, nil
}
```

### ❌ Field Assignment

```go
// @immutable
type Point struct {
    X, Y int
}

func MovePoint(p *Point) {
    p.X += 10  // ❌ ERROR: IMM02 - Compound assignment to immutable field
    p.Y = 20   // ❌ ERROR: IMM01 - Field assignment
}
```

### ❌ Increment/Decrement

```go
// @immutable
type Counter struct {
    value int
}

func Increment(c *Counter) {
    c.value++  // ❌ ERROR: IMM03 - Increment of immutable field
}

func Decrement(c *Counter) {
    c.value--  // ❌ ERROR: IMM03 - Decrement of immutable field
}
```

### ❌ Index Assignment

```go
// @immutable
type Data struct {
    items []int
    dict  map[string]int
}

func Modify(d *Data) {
    d.items[0] = 42          // ❌ ERROR: IMM04 - Index assignment to immutable collection
    d.dict["key"] = 100      // ❌ ERROR: IMM04 - Index assignment to immutable collection
}
```

### ✅ Using @ignore to Suppress

```go
// @immutable
type Cache struct {
    data map[string]string
}

func (c *Cache) Update(key, value string) {
    // @ignore IMM04
    c.data[key] = value  // ✅ Suppressed via @ignore
}
```

### ✅ Cross-Package Immutability

**Package `models`:**

```go
package models

// @immutable
type User struct {
    ID   int
    Name string
}
```

**Package `main`:**

```go
package main

import "myapp/models"

func updateUser(u *models.User) {
    u.Name = "New Name"  // ❌ ERROR: IMM01 - User is immutable (from external package)
}
```

### ✅ Correct Pattern: Return New Instances

```go
// @immutable
// @constructor NewPerson
type Person struct {
    Name string
    Age  int
}

func NewPerson(name string, age int) Person {
    return Person{Name: name, Age: age}
}

// ✅ Correct: Return modified copy
func WithAge(p Person, newAge int) Person {
    return Person{
        Name: p.Name,
        Age:  newAge,
    }
}

// ✅ Correct: Builder pattern for construction
func (p Person) WithName(name string) Person {
    return Person{Name: name, Age: p.Age}
}
```

## Best Practices

### 1. Combine with @constructor

Always use `@constructor` with `@immutable` to control object creation:

```go
// @immutable
// @constructor NewConfig
type Config struct {
    value string
}

func NewConfig(v string) Config {
    return Config{value: v}
}
```

### 2. Return New Instances

Instead of mutating, return new instances:

```go
// ❌ Bad: Mutation
func UpdateName(u *User) {
    u.Name = "newname"  // ERROR
}

// ✅ Good: Return new instance
func WithName(u User, name string) User {
    return User{ID: u.ID, Name: name}
}
```

### 3. Use for Value Objects

Immutable types work well for value objects:

```go
// @immutable
// @constructor NewMoney
type Money struct {
    Amount   int
    Currency string
}

func (m Money) Add(other Money) (Money, error) {
    if m.Currency != other.Currency {
        return Money{}, errors.New("currency mismatch")
    }
    return Money{
        Amount:   m.Amount + other.Amount,
        Currency: m.Currency,
    }, nil
}
```

### 4. Document Immutability Intent

```go
// @immutable
// Point represents an immutable 2D coordinate.
// Use NewPoint to create instances and helper functions to derive new points.
type Point struct {
    X, Y int
}
```

## Limitations

### Weak Guarantees

`@immutable` provides **weak immutability** - it prevents direct field assignments but doesn't prevent:

- **Pointer manipulation**: Modifying through `unsafe` pointers
- **Reflection**: Mutations via `reflect` package
- **Slice/map element mutations**: Modifying elements inside slices or maps (only index assignment is caught)

```go
// @immutable
type Data struct {
    items []Item  // The slice itself can't be reassigned, but elements can be modified
}

func modify(d Data) {
    d.items = nil           // ❌ ERROR: IMM01 - Assignment
    d.items[0] = newItem    // ❌ ERROR: IMM04 - Index assignment
    d.items[0].field = 123  // ✅ No error - element field modification not caught
}
```

### Workaround

For stronger guarantees, use unexported fields with exported getter methods:

```go
// @immutable
type SafeData struct {
    items []Item  // unexported - can't be accessed directly
}

func (d SafeData) Items() []Item {
    // Return copy to prevent external modifications
    result := make([]Item, len(d.items))
    copy(result, d.items)
    return result
}
```

## Related Annotations

- **[@constructor](02_03_constructor.md)**: Control object creation
- **[@ignore](02_05_ignore.md)**: Suppress violations when needed
- **[@implements](02_01_implements.md)**: Often combined for immutable interfaces

## See Also

- [Error Codes Reference](03_codes.md)
- [Limitations](01_02_limitations.md)