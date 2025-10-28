# @testonly Annotation

The `@testonly` annotation restricts usage of types, functions, or methods to test files only. Any usage outside `*_test.go` files will be flagged as a violation.

## Motivation

Go doesn't provide built-in mechanisms to mark code as test-only. Many other languages have features like:

- `@VisibleForTesting` (Java/Android)
- `internal` test helpers
- Test-only APIs

The `@testonly` annotation fills this gap, allowing you to clearly mark test utilities and prevent their accidental use in production code.

## Syntax

```go
// @testonly
type TestHelper struct {}

// @testonly
func CreateMock() *MockService {}

// @testonly
func (m *MockService) Reset() {}
```

No parameters required - simply add `// @testonly` above the declaration.

## How It Works

GoGreement detects usage of `@testonly` items outside test files:

1. **Type usage**: Variable declarations, composite literals, type assertions
2. **Function calls**: Direct calls to `@testonly` functions
3. **Method calls**: Calls to `@testonly` methods

## Key Behaviors

1. **Test files only**: Only `*_test.go` files can use `@testonly` items
2. **No generics support**: Cannot be used with generic declarations
3. **Nested @testonly allowed**: `@testonly` code can call other `@testonly` code
4. **Per-file deduplication**: Only one error per type per file (avoids spam)
5. **Catches all usage**: Type assertions, composite literals, variable declarations
6. **Can be suppressed**: Use `@ignore` to allow usage in specific places

## Can Be Declared On

### Types

```go
// @testonly
type MockDatabase struct {
    calls int
}

// @testonly
type TestConfig struct {
    Host string
}
```

### Functions

```go
// @testonly
func CreateTestData() []User {
    return []User{{ID: 1, Name: "Test"}}
}
```

### Methods

```go
type Service struct {
    db *sql.DB
}

// @testonly
func (s *Service) ResetForTesting() {
    // Clear state for tests
}
```

## Error Codes

| Code | Description | Example |
|------|-------------|---------|
| **TONL01** | TestOnly type used in non-test context | `var m MockService` in production code |
| **TONL02** | TestOnly function called in non-test context | `CreateMock()` in production code |
| **TONL03** | TestOnly method called in non-test context | `obj.ResetForTesting()` in production code |

## Examples

### ✅ Basic Test Helper

**`helper.go`:**

```go
package myapp

// @testonly
type TestHelper struct {
    state map[string]interface{}
}

// @testonly
func NewTestHelper() *TestHelper {
    return &TestHelper{state: make(map[string]interface{})}
}

// @testonly
func (h *TestHelper) Set(key string, value interface{}) {
    h.state[key] = value
}
```

**`helper_test.go`:**

```go
package myapp

func TestSomething(t *testing.T) {
    helper := NewTestHelper()  // ✅ Allowed in test file
    helper.Set("key", "value")  // ✅ Allowed in test file
}
```

**`production.go`:**

```go
package myapp

func ProductionCode() {
    helper := NewTestHelper()  // ❌ [TONL02] function NewTestHelper is marked @testonly and can only be called in test files
}
```

### ✅ Mock Implementation

```go
// @testonly
type MockUserService struct {
    users []User
}

// @testonly
func NewMockUserService() *MockUserService {
    return &MockUserService{users: []User{}}
}

// @testonly
func (m *MockUserService) AddUser(u User) {
    m.users = append(m.users, u)
}

func (m *MockUserService) GetUser(id int) (*User, error) {
    // Real interface method - not @testonly
    for _, u := range m.users {
        if u.ID == id {
            return &u, nil
        }
    }
    return nil, errors.New("not found")
}
```

### ✅ Nested @testonly Calls

```go
// @testonly
func setupDatabase() *sql.DB {
    return nil
}

// @testonly
func createTestEnvironment() *TestEnv {
    db := setupDatabase()  // ✅ Allowed - @testonly calling @testonly
    return &TestEnv{DB: db}
}

// In test file
func TestIntegration(t *testing.T) {
    env := createTestEnvironment()  // ✅ Allowed in test
    // ...
}
```

### ❌ Type Usage in Production

```go
// @testonly
type MockCache struct {
    data map[string]string
}

func ProductionCode() {
    // ❌ [TONL01] type MockCache is marked @testonly and can only be used in test files
    var cache MockCache

    // ❌ [TONL01] type MockCache is marked @testonly and can only be used in test files
    cache = MockCache{data: make(map[string]string)}

    // ❌ [TONL01] type MockCache is marked @testonly and can only be used in test files
    if c, ok := something.(*MockCache); ok {
        _ = c
    }
}
```

### ❌ Function Call in Production

```go
// @testonly
func GenerateTestID() string {
    return "test-" + uuid.New().String()
}

func CreateUser(name string) *User {
    return &User{
        ID: GenerateTestID(),  // ❌ [TONL02] function GenerateTestID is marked @testonly and can only be called in test files
        Name: name,
    }
}
```

### ❌ Method Call in Production

```go
type UserRepository struct {
    db *sql.DB
}

// @testonly
func (r *UserRepository) ClearAll() error {
    _, err := r.db.Exec("DELETE FROM users")
    return err
}

func ResetProduction(repo *UserRepository) {
    repo.ClearAll()  // ❌ [TONL03] method ClearAll on UserRepository is marked @testonly and can only be called in test files
}
```

### ✅ Using @ignore to Suppress

```go
// @testonly
type DebugHelper struct {
    verbose bool
}

func debugFunction() {
    // @ignore TONL01
    helper := DebugHelper{verbose: true}  // ✅ Suppressed for debugging
    _ = helper
}
```

### ✅ Test Fixtures

```go
// @testonly
type UserFixture struct {
    Admin     User
    Regular   User
    Suspended User
}

// @testonly
func LoadUserFixtures() *UserFixture {
    return &UserFixture{
        Admin:     User{ID: 1, Name: "Admin", Role: "admin"},
        Regular:   User{ID: 2, Name: "User", Role: "user"},
        Suspended: User{ID: 3, Name: "Banned", Role: "suspended"},
    }
}

// In test file
func TestUserPermissions(t *testing.T) {
    fixtures := LoadUserFixtures()  // ✅ Allowed
    // Test with fixtures.Admin, fixtures.Regular, etc.
}
```

### ✅ Spy/Stub Pattern

```go
type Logger interface {
    Log(msg string)
}

// @testonly
type SpyLogger struct {
    messages []string
}

// @testonly
func NewSpyLogger() *SpyLogger {
    return &SpyLogger{}
}

func (s *SpyLogger) Log(msg string) {
    s.messages = append(s.messages, msg)  // Not @testonly - interface method
}

// @testonly
func (s *SpyLogger) Messages() []string {
    return s.messages
}

// In test
func TestLogging(t *testing.T) {
    spy := NewSpyLogger()  // ✅ Allowed
    service := NewService(spy)
    service.DoSomething()
    messages := spy.Messages()  // ✅ Allowed
    assert.Equal(t, 1, len(messages))
}
```

## Best Practices

### 1. Use for Test Utilities

Mark test helpers and utilities:

```go
// @testonly
func AssertNoError(t *testing.T, err error) {
    if err != nil {
        t.Fatal(err)
    }
}

// @testonly
func CreateTempDir(t *testing.T) string {
    dir, err := os.MkdirTemp("", "test")
    AssertNoError(t, err)
    t.Cleanup(func() { os.RemoveAll(dir) })
    return dir
}
```

### 2. Mark Mock Implementations

```go
type UserService interface {
    GetUser(id int) (*User, error)
    CreateUser(u *User) error
}

// @testonly
type MockUserService struct {
    users map[int]*User
}

// @testonly
func NewMockUserService() *MockUserService {
    return &MockUserService{users: make(map[int]*User)}
}
```

### 3. Test Fixtures and Factories

```go
// @testonly
func MustCreateUser(t *testing.T, name string) *User {
    user, err := CreateUser(name)
    if err != nil {
        t.Fatalf("failed to create user: %v", err)
    }
    return user
}
```

### 4. Document Test-Only Purpose

```go
// @testonly
// ResetDatabase drops all tables and recreates schema.
// WARNING: This is for testing only and will destroy all data.
func ResetDatabase(db *sql.DB) error {
    // Dangerous operation
    return nil
}
```

## Deduplication Behavior

GoGreement deduplicates type usage violations per file to avoid spam:

```go
// @testonly
type MockCache struct {}

func ProductionCode() {
    var c1 MockCache  // ❌ ERROR: TONL01
    var c2 MockCache  // ✅ No error - deduplicated
    var c3 MockCache  // ✅ No error - deduplicated

    // Only ONE error per file for MockCache type usage
}
```

**Function and method calls are NOT deduplicated** - each call reports a separate error.

## Limitations

### 1. Per-File Deduplication

Type errors are shown once per file, even if the type is used many times:

```go
// Only first usage reports error
var m1 MockService  // ❌ ERROR
var m2 MockService  // No error (deduplicated)
var m3 MockService  // No error (deduplicated)
```

### 2. Indirect Usage Not Caught

If production code uses a non-test-only function that internally uses `@testonly` code:

```go
// @testonly
func helper() int {
    return 42
}

func indirect() int {
    return helper()  // ❌ ERROR caught here
}

func production() {
    x := indirect()  // ✅ No error - indirect usage
}
```

## Related Annotations

- **[@ignore](02_05_ignore.md)**: Suppress violations when needed
- **[@implements](02_01_implements.md)**: Often used with `@testonly` for mock interfaces

## See Also

- [Error Codes Reference](03_codes.md)
- [Limitations](01_02_limitations.md)