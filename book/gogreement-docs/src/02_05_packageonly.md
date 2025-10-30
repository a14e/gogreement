# @packageonly Annotation

The `@packageonly` annotation restricts usage of types, functions, or methods to specific packages.

## Motivation

Go doesn't provide built-in mechanisms to restrict which packages can use specific code. This is needed for maintaining clean architecture:

- Controllers should only be used in routing layer
- Database models should only be accessed by repository layer
- Internal utilities should stay within specific modules

The `@packageonly` annotation fills this gap.

## Syntax

```go
// @packageonly
type Helper struct {}

// @packageonly pkg1, pkg2
func SpecialFunction() {}

// @packageonly myapp/internal/auth
func (s *Service) AdminMethod() {}
```

### Parameters

- **Package list** (optional): Comma-separated list of allowed packages
- Can specify **package names** (e.g., `testing`) or **full paths** (e.g., `myapp/internal/auth`)
- If no packages specified, only the **current package** is allowed
- **Current package is always included automatically**

## How It Works

GoGreement detects usage of `@packageonly` items outside allowed packages:

1. **Type usage**: Variable declarations, composite literals, type assertions
2. **Function calls**: Direct calls to `@packageonly` functions
3. **Method calls**: Calls to `@packageonly` methods

## Key Behaviors

1. **Current package always allowed**: The declaring package is automatically included
2. **No generics support**: Cannot be used with generic declarations
3. **Per-file deduplication**: Only one error per type per file (avoids spam)
4. **Catches all usage**: Type assertions, composite literals, variable declarations
5. **Can be suppressed**: Use `@ignore` to allow usage in specific places
6. **Cross-package enforcement**: Works even if `@packageonly` was declared in an external module

## Can Be Declared On

### Types

```go
// @packageonly routes
type Controller struct {
    service *Service
}
```

### Functions

```go
// @packageonly myapp/internal/admin
func ExecuteAdminCommand(cmd string) error {
    return nil
}
```

### Methods

```go
type Repository struct {
    db *sql.DB
}

// @packageonly testing
func (r *Repository) ClearAll() error {
    return nil
}
```

## Error Codes

| Code | Description | Example |
|------|-------------|---------|
| **PKGO01** | PackageOnly type used outside allowed packages | `var h Helper` in unauthorized package |
| **PKGO02** | PackageOnly function called outside allowed packages | `ExecuteAdminCommand()` in unauthorized package |
| **PKGO03** | PackageOnly method called outside allowed packages | `repo.ClearAll()` in unauthorized package |

## Examples

### ✅ Basic Usage

**`helpers/helper.go`:**

```go
package helpers

// @packageonly services, handlers
type InternalHelper struct {
    state map[string]interface{}
}

// @packageonly services
func ProcessInternal(data string) error {
    return nil
}
```

**`services/service.go`:**

```go
package services

import "myapp/helpers"

func BusinessLogic() {
    h := helpers.InternalHelper{}  // ✅ Allowed - services in list
    helpers.ProcessInternal("data")  // ✅ Allowed
}
```

**`main.go`:**

```go
package main

import "myapp/helpers"

func main() {
    // ❌ [PKGO01] type InternalHelper is marked @packageonly and can only be used in packages: [services, handlers, helpers]
    h := helpers.InternalHelper{}

    // ❌ [PKGO02] function ProcessInternal is marked @packageonly and can only be called in packages: [services, helpers]
    helpers.ProcessInternal("data")
}
```

### ✅ Current Package Always Allowed

```go
package auth

// @packageonly
type SessionManager struct {
    sessions map[string]*Session
}

func NewSessionManager() *SessionManager {
    return &SessionManager{  // ✅ Allowed - same package
        sessions: make(map[string]*Session),
    }
}
```

### ✅ Method-Level Restrictions

```go
package repository

type UserRepository struct {
    db *sql.DB
}

// Public method - no restrictions
func (r *UserRepository) GetUser(id int) (*User, error) {
    return nil, nil
}

// @packageonly testing
func (r *UserRepository) InsertTestData(users []User) error {
    return nil
}
```

**Usage:**

```go
package handlers

import "myapp/repository"

func HandleRequest(repo *repository.UserRepository) {
    user, _ := repo.GetUser(1)  // ✅ Allowed - public method
    // ❌ [PKGO03] method InsertTestData on UserRepository is marked @packageonly and can only be called in packages: [testing, repository]
    repo.InsertTestData(nil)
}
```

### ✅ Using @ignore to Suppress

```go
// @packageonly services
type InternalCache struct {
    data map[string]string
}

func debugFunction() {
    // @ignore PKGO01
    cache := InternalCache{}  // ✅ Suppressed
}
```

### ✅ Cross-Package Enforcement

**Package `external/lib`:**

```go
package lib

// @packageonly myapp/internal/core
type InternalAPI struct {
    token string
}
```

**Package `myapp/main`:**

```go
package main

import "external/lib"

func main() {
    // ❌ [PKGO01] type InternalAPI is marked @packageonly and can only be used in packages: [myapp/internal/core, lib]
    api := lib.InternalAPI{}
}
```

**Package `myapp/internal/core`:**

```go
package core

import "external/lib"

func CoreLogic() {
    api := lib.InternalAPI{}  // ✅ Allowed
}
```

## Best Practices

### Architectural Boundaries

```go
// @packageonly routes
type HTTPController struct {
    service *BusinessService
}

// @packageonly repository
type DatabaseConnection struct {
    db *sql.DB
}
```

### Administrative Functions

```go
// @packageonly admin, internal/admin
func ResetAllData() error {
    return nil
}
```

## Deduplication Behavior

Type usage violations are deduplicated per file:

```go
// @packageonly admin
type AdminHelper struct {}

func UnauthorizedCode() {
    var h1 AdminHelper  // ❌ ERROR: PKGO01
    var h2 AdminHelper  // ✅ No error - deduplicated
    var h3 AdminHelper  // ✅ No error - deduplicated
}
```

**Function and method calls are NOT deduplicated** - each call reports a separate error.

## Limitations

### Package Name Ambiguity

If multiple packages have the same name, use full paths:

```go
// ❌ Ambiguous
// @packageonly util

// ✅ Clear
// @packageonly myapp/internal/util
```

## Related Annotations

- **[@ignore](02_06_ignore.md)**: Suppress violations when needed
- **[@testonly](02_04_testonly.md)**: Simpler version that restricts to test files only

## See Also

- [Error Codes Reference](03_codes.md)
- [Limitations](01_02_limitations.md)
