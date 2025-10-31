# Annotations

All enforcement mechanisms in GoGreement are based on **annotations**. Annotations are special comments that start with `@` and follow a specific format:

```go
// @annotation param1 param2
```

## How Annotations Work

1. **Write annotations** as comments above types, functions, or methods
2. **Run GoGreement** analyzer on your code
3. **Get violations** reported if agreements are broken

## Available Annotations

GoGreement supports five core annotations:

| Annotation | Purpose | Applied To |
|------------|---------|-----------|
| **[@implements](02_01_implements.md)** | Enforce interface implementation contracts | Types |
| **[@immutable](02_02_immutable.md)** | Prevent field mutations after creation | Types |
| **[@constructor](02_03_constructor.md)** | Restrict object creation to specific functions | Types |
| **[@testonly](02_04_testonly.md)** | Limit usage to test files only | Types, Functions, Methods |
| **[@ignore](02_06_ignore.md)** | Suppress specific violations | Files, Blocks, Lines |

## Annotation Syntax Rules

### 1. Must Start Comment Line

Annotations must be at the **beginning** of a comment line (after `//` and whitespace):

```go
// @immutable        ✅ Valid
//   @immutable      ✅ Valid (whitespace OK)
// TODO: @immutable  ❌ Invalid (not at start)
```

### 2. Case-Sensitive Keywords

Annotation keywords are case-sensitive and must be lowercase:

```go
// @immutable   ✅ Valid
// @Immutable   ❌ Invalid
// @IMMUTABLE   ❌ Invalid
```

### 3. Additional Comments Allowed

You can add comments after annotation parameters:

```go
// @constructor New, Create  // These are the factory functions
// @implements &io.Reader    // Pointer receiver required
```

### 4. Multiple Annotations

You can use multiple annotations on the same declaration:

```go
// @immutable
// @constructor NewPoint
type Point struct {
    X, Y int
}
```

## Annotation Scope

Annotations are only recognized on **top-level declarations**:

```go
// ✅ Top-level declaration - annotation works
// @immutable
type Config struct {
    Host string
}

func Example() {
    // ❌ Not a top-level declaration - annotation ignored
    // @immutable
    type LocalConfig struct {
        Host string
    }
}
```

Annotations on nested types, local functions, or any declarations inside functions are ignored by GoGreement.

## Annotation Processing

GoGreement uses a two-phase approach:

1. **Reading Phase** (`AnnotationReader` analyzer)
   - Scans all files for annotations on top-level declarations
   - Parses and validates syntax
   - Exports annotations as package facts

2. **Checking Phase** (Individual checkers)
   - Import annotations from dependencies
   - Build cross-package indices
   - Detect violations
   - Report errors with specific codes

This design enables **cross-package enforcement** - annotations in one package affect analysis in packages that import it.

## Next Steps

Learn about each annotation in detail:

- **[@implements](02_01_implements.md)** - Ensure types implement interfaces
- **[@immutable](02_02_immutable.md)** - Enforce immutability
- **[@constructor](02_03_constructor.md)** - Control object creation
- **[@testonly](02_04_testonly.md)** - Restrict to tests
- **[@ignore](02_06_ignore.md)** - Suppress violations