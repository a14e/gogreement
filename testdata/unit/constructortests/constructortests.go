package constructortests

// User has a constructor annotation
// @constructor NewUser, NewUserWithVar
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

// Var declarations that should violate constructor rules
func VarDeclarationViolations() {
	var user User     // ❌ VIOLATION: zero-initialized User outside constructor
	var config Config // ❌ VIOLATION: zero-initialized Config outside constructor

	// Use variables to avoid "declared and not used" errors
	_ = user
	_ = config

	// Pointer var declarations should be allowed (they only create nil pointers)
	var userPtr *User // ✅ OK: nil pointer
	_ = userPtr

	// Types without constructor annotation should be allowed
	var service Service // ✅ OK: no @constructor annotation
	_ = service
}

// Var declarations in constructors should be allowed
func NewUserWithVar(name string, age int) *User {
	var user User // ✅ OK: in constructor
	user.Name = name
	user.Age = age
	return &user
}

// Email is a named type whose construction must go through NewEmail.
// Building it via a type conversion outside the constructor bypasses validation.
// @constructor NewEmail
type Email string

func NewEmail(s string) Email {
	return Email(s) // ✅ OK: conversion in constructor
}

func MakeEmailWrong(s string) Email {
	return Email(s) // ❌ VIOLATION: type conversion outside constructor
}

// Widget is constructed only via NewWidget.
// @constructor NewWidget
type Widget struct {
	Name string
}

func NewWidget(name string) *Widget {
	return &Widget{Name: name} // ✅ OK: in the declared constructor
}

// Factory has a method whose name collides with Widget's constructor name.
// A method is never the declared (free-function) constructor, so the literal
// inside it must still be flagged.
type Factory struct{}

func (f *Factory) NewWidget() *Widget {
	return &Widget{Name: "wrong"} // ❌ VIOLATION: a method is not the declared constructor
}

// MakeUserPtrPtr uses new(*User), which allocates a **User and never
// instantiates a User, so it must not be flagged.
func MakeUserPtrPtr() **User {
	return new(*User) // ✅ OK: new(*User) does not construct a User
}

// Gadget is constructed only via NewGadget. The package-level instantiation
// right after the constructor verifies the enclosing-function context does not
// leak across declarations.
// @constructor NewGadget
type Gadget struct {
	Name string
}

func NewGadget() *Gadget {
	return &Gadget{} // ✅ OK: in the declared constructor
}

var packageGadget = Gadget{Name: "pkg"} // ❌ VIOLATION: package-level instantiation (no constructor leak)
