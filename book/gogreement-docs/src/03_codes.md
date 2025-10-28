# Error Codes Reference

GoGreement reports violations using structured error codes. Each code identifies a specific type of violation and can be used with the `@ignore` annotation to suppress false positives.

## Code Structure

Error codes follow the format: `[CATEGORY][NUMBER]`

- **Category**: 2-4 letter prefix identifying the annotation (e.g., `IMM`, `CTOR`, `TONL`, `IMPL`)
- **Number**: Two-digit sequential number within the category (e.g., `01`, `02`)

**Example**: `IMM01` = Immutable category, violation type 01

## All Error Codes

### IMM - Immutable Violations

Violations of `@immutable` annotations. These can be suppressed with `@ignore`.

| Code | Description | Example |
|------|-------------|---------|
| **IMM01** | Field of immutable type is being assigned | `point.X = 10` |
| **IMM02** | Compound assignment to immutable field | `point.X += 5`, `count *= 2` |
| **IMM03** | Increment/decrement of immutable field | `point.X++`, `count--` |
| **IMM04** | Index assignment to immutable collection | `obj.items[0] = value`, `obj.dict["key"] = val` |

**Suppress with**:
- `// @ignore IMM` - All immutability checks
- `// @ignore IMM01` - Specific check only

**Documentation**: [@immutable](02_02_immutable.md)

---

### CTOR - Constructor Violations

Violations of `@constructor` annotations. These can be suppressed with `@ignore`.

| Code | Description | Example |
|------|-------------|---------|
| **CTOR01** | Composite literal used outside allowed constructor functions | `db := Database{}` |
| **CTOR02** | new() call used outside allowed constructor functions | `db := new(Database)` |
| **CTOR03** | Variable declaration creates zero-initialized instance outside allowed constructor functions | `var db Database` |

**Suppress with**:
- `// @ignore CTOR` - All constructor checks
- `// @ignore CTOR01` - Specific check only

**Documentation**: [@constructor](02_03_constructor.md)

---

### TONL - TestOnly Violations

Violations of `@testonly` annotations. These can be suppressed with `@ignore`.

| Code | Description | Example |
|------|-------------|---------|
| **TONL01** | TestOnly type used outside test context | `var mock MockService` in production code |
| **TONL02** | TestOnly function called outside test context | `CreateMock()` in production code |
| **TONL03** | TestOnly method called outside test context | `service.ResetForTesting()` in production code |

**Suppress with**:
- `// @ignore TONL` - All testonly checks
- `// @ignore TONL01` - Specific check only

**Documentation**: [@testonly](02_04_testonly.md)

---

### IMPL - Implements Violations

Violations of `@implements` annotations. **Cannot be suppressed with `@ignore`** (intentional).

| Code | Description | Example |
|------|-------------|---------|
| **IMPL01** | Package not found in imports | Using `@implements pkg.Interface` without importing `pkg` |
| **IMPL02** | Interface not found in package | Interface name doesn't exist or is misspelled |
| **IMPL03** | Missing or incorrect methods | Type doesn't implement all required methods with correct signatures |

**Cannot suppress**: `@implements` violations cannot be ignored because they are under your direct control. If you don't want to implement an interface, remove the annotation.

**Documentation**: [@implements](02_01_implements.md)

---

## Using Error Codes

### With @ignore Annotation

Suppress specific violations in your code:

```go
// Suppress specific code
// @ignore IMM01
point.X = 10

// Suppress multiple codes
// @ignore IMM01, CTOR02
func unsafeOperation() {
    point.X = 10
    db := new(Database)
}

// Suppress entire category
// @ignore IMM
func batchUpdate() {
    // All IMM* codes suppressed
}

// Suppress all violations
// @ignore ALL
func debugFunction() {
    // Everything suppressed
}
```

### With Command-Line Flags

Exclude checks globally across your project:

```bash
# Exclude all immutability checks
gogreement --exclude-checks=IMM ./...

# Exclude specific codes
gogreement --exclude-checks=IMM01,CTOR02,TONL03 ./...

# Exclude multiple categories
gogreement --exclude-checks=IMM,TONL ./...
```

### With Environment Variables

Set project-wide defaults:

```bash
export GOGREEMENT_EXCLUDE_CHECKS=IMM01,CTOR
gogreement ./...
```

## Code Hierarchy

Codes follow a hierarchical structure for suppression:

```
ALL
├── IMM (Immutable)
│   ├── IMM01 (Field assignment)
│   ├── IMM02 (Compound assignment)
│   ├── IMM03 (Increment/decrement)
│   └── IMM04 (Index assignment)
├── CTOR (Constructor)
│   ├── CTOR01 (Composite literal)
│   ├── CTOR02 (new() call)
│   └── CTOR03 (Var declaration)
├── TONL (TestOnly)
│   ├── TONL01 (Type usage)
│   ├── TONL02 (Function call)
│   └── TONL03 (Method call)
└── IMPL (Implements) - Cannot suppress
    ├── IMPL01 (Package not found)
    ├── IMPL02 (Interface not found)
    └── IMPL03 (Missing methods)
```

When you suppress a code at any level, all codes below it are also suppressed:

- `@ignore ALL` → Suppresses everything
- `@ignore IMM` → Suppresses IMM01, IMM02, IMM03, IMM04
- `@ignore IMM01` → Suppresses only IMM01

## Quick Reference by Annotation

| Annotation | Codes | Suppressible |
|------------|-------|--------------|
| **@immutable** | IMM01, IMM02, IMM03, IMM04 | ✅ Yes |
| **@constructor** | CTOR01, CTOR02, CTOR03 | ✅ Yes |
| **@testonly** | TONL01, TONL02, TONL03 | ✅ Yes |
| **@implements** | IMPL01, IMPL02, IMPL03 | ❌ No |

## Error Message Format

GoGreement error messages include the error code for easy reference:

```
path/to/file.go:15:2: IMM01 - Field of immutable type is being assigned
path/to/file.go:23:5: CTOR01 - Composite literal used outside allowed constructor functions
path/to/file.go:45:10: TONL02 - TestOnly function called outside test context
```

## Best Practices

### 1. Be Specific

Use the most specific code possible when suppressing:

```go
// ✅ Good
// @ignore IMM01
point.X = 10

// ❌ Too broad
// @ignore ALL
point.X = 10
```

### 2. Document Suppressions

Always explain why you're suppressing a check:

```go
// @ignore IMM01
// REASON: Performance-critical path, avoiding allocations
// TODO: Refactor to use copy-on-write
cache.data[key] = value
```

### 3. Prefer Fixing Over Suppressing

Suppression should be the exception, not the rule:

```go
// ❌ Bad: Suppressing instead of fixing
// @ignore IMM
type Point struct { X, Y int }

// ✅ Good: Fix the architecture
// Remove @immutable if mutation is required
type Point struct { X, Y int }
```

### 4. Review Suppressions Regularly

Periodically search for `@ignore` in your codebase and review whether suppressions are still needed.

## See Also

- **[@ignore Annotation](02_05_ignore.md)**: Detailed guide on suppressing violations
- **[Getting Started - Configuration](01_01_getting_started.md#configuration)**: Module-level exclusions
- **[Annotations](02_annotations.md)**: Learn about all annotations