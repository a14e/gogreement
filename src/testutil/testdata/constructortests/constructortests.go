package constructortests

// User has a constructor annotation
// @constructor NewUser
type User struct {
	Name string
	Age  int
}

func NewUser(name string, age int) *User {
	return &User{Name: name, Age: age} // ✅ OK: in constructor
}

func CreateUserWrong(name string) *User {
	return &User{Name: name} // ❌ VIOLATION: not in declared constructor
}

func MakeUserWithNew() *User {
	return new(User) // ❌ VIOLATION: new() outside constructor
}

// Config has multiple constructors
// @constructor NewConfig, NewDefaultConfig
type Config struct {
	Host string
	Port int
}

func NewConfig(host string, port int) *Config {
	return &Config{Host: host, Port: port} // ✅ OK: in constructor
}

func NewDefaultConfig() *Config {
	cfg := &Config{} // ✅ OK: in constructor
	cfg.Host = "localhost"
	cfg.Port = 8080
	return cfg
}

func MakeConfigWrong() *Config {
	return &Config{Host: "test"} // ❌ VIOLATION
}

// Database with pointer literal
// @constructor NewDatabase
type Database struct {
	conn string
}

func NewDatabase(conn string) *Database {
	db := Database{conn: conn} // ✅ OK: in constructor
	return &db
}

func OpenDatabase() *Database {
	db := Database{conn: "test"} // ❌ VIOLATION
	return &db
}

// Service without constructor annotation (should not report violations)
type Service struct {
	name string
}

func CreateService() *Service {
	return &Service{name: "test"} // ✅ OK: no @constructor annotation
}

func MakeService() *Service {
	return new(Service) // ✅ OK: no @constructor annotation
}

// Point with value receiver constructor
// @constructor NewPoint
type Point struct {
	X, Y int
}

func NewPoint(x, y int) Point {
	return Point{X: x, Y: y} // ✅ OK: in constructor
}

func MakePoint() Point {
	return Point{X: 1, Y: 2} // ❌ VIOLATION
}

// Helper function that creates nested types
func HelperFunction() {
	_ = &User{Name: "nested"} // ❌ VIOLATION
	_ = new(Config)           // ❌ VIOLATION
}
