# Limitations

GoGreement is a powerful static analysis tool, but it has some limitations you should be aware of:

## 1. No Generics Support

Generics (type parameters) are not currently supported in any annotations. This means you cannot use GoGreement annotations on generic types or functions.

```go
// Not supported yet
// @immutable
type Container[T any] struct {
    value T
}
```

## 2. Import-Based Analysis Only

Due to how the `analysis` framework works, GoGreement only analyzes types and functions that are **imported** by the packages being analyzed.

This means:
- If you annotate a type in package A
- But package B never imports package A
- Package B won't see the annotations from package A

**Impact**: Cross-package enforcement only works for types that are actually imported.

## 3. Lenient Annotation Parsing

Many annotations do not fail with errors if they cannot be fully parsed. This is an intentional design decision to support:

- Comments that mention annotation keywords in the middle of text
- Additional comments after annotations
- Gradual adoption without breaking existing codebases

**Example**: These won't be recognized as valid annotations:

```go
// This is a note about @immutable types in general
// (not at the start - won't be recognized)
type MyType struct {}

// TODO: add @constructor later
// (not at the start - won't be recognized)
type Other struct {}

// @constructor
// (no function names specified - returns nil, won't be recognized)
type NeedsFunctions struct {}
```

## 4. In-Memory Fact Caching

When using the `analysis` framework directly, package facts are cached in memory. This can lead to increased memory usage for large projects with many cross-package dependencies.

**Note**: This is a property of the underlying `analysis` framework, not specific to GoGreement.

## 5. No golangci-lint Integration (Yet)

GoGreement is not yet integrated with `golangci-lint`. You need to run it as a standalone tool.

**Current status**: We are working on adding `golangci-lint` support in future releases.

## 6. Analysis Framework Limitations

Due to limitations of Go's `analysis` framework, GoGreement must be run **after** all code generation and dependency updates are complete:

```bash
# Required order:
go generate ./...         # Generate any code
go mod tidy              # Update dependencies
gogreement ./...         # Then run GoGreement
```

**Why this matters**: The `analysis` framework works on the AST (Abstract Syntax Tree) of Go code. If code generation or dependency updates happen after analysis, GoGreement may:
- Analyze outdated code structures
- Miss newly generated types and annotations
- Report false positives/negatives due to stale dependency information

## 7. Pointer vs Value Receiver Distinction Required

For `@implements` annotations, you must be explicit about pointer vs value receivers:

```go
// These are DIFFERENT:
// @implements io.Reader       // Value receiver methods
// @implements &io.Reader      // Pointer receiver methods

type MyReader struct {}
```

## Workarounds

Most limitations can be worked around:

- **No generics**: Use concrete types or wrapper types
- **Import-based analysis**: Ensure annotated types are imported where needed
- **golangci-lint**: Run GoGreement as a separate step in CI/CD