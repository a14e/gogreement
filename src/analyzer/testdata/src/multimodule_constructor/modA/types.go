package modA // want package:"package modA"

// User represents a user entity
// @immutable
// @constructor NewUser
type User struct {
	ID   int
	Name string
}

// NewUser creates a new User
func NewUser(id int, name string) *User {
	return &User{
		ID:   id,
		Name: name,
	}
}

// Config holds configuration
// @immutable
// @constructor NewConfig, NewDefaultConfig
type Config struct {
	Host string
	Port int
}

// NewConfig creates a new Config
func NewConfig(host string, port int) *Config {
	return &Config{
		Host: host,
		Port: port,
	}
}

// NewDefaultConfig creates a default Config
func NewDefaultConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 8080,
	}
}

// TestHelper is a test-only helper
// @testonly
type TestHelper struct {
	data string
}

// CreateTestData is a test-only function
// @testonly
func CreateTestData() string {
	return "test"
}
