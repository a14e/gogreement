package modB // want package:"package modB"

import "multimodule_constructor/modA"

func CreateUserDirectly() *modA.User {
	return &modA.User{ // want "type instantiation must be in constructor"
		ID:   1,
		Name: "test",
	}
}

func CreateConfigDirectly() *modA.Config {
	return &modA.Config{ // want "type instantiation must be in constructor"
		Host: "localhost",
		Port: 8080,
	}
}

// Test @ignore with specific code CTOR01
// @ignore CTOR01
func CreateUserIgnored() *modA.User {
	return &modA.User{ // This should be ignored
		ID:   1,
		Name: "test",
	}
}

// Test @ignore with category CTOR
// @ignore CTOR
func CreateConfigIgnored() *modA.Config {
	return &modA.Config{ // This should be ignored via CTOR category
		Host: "localhost",
		Port: 8080,
	}
}

// Test inline @ignore
func CreateUserInlineIgnore() *modA.User {
	return &modA.User{ // @ignore CTOR01
		ID:   1,
		Name: "test",
	}
}

// Test @ignore with new() call (CTOR02)
// @ignore CTOR02
func CreateUserWithNew() *modA.User {
	return new(modA.User) // This should be ignored
}

// Test @ignore ALL
// @ignore ALL
func CreateEverythingIgnored() (*modA.User, *modA.Config) {
	u := &modA.User{ID: 1, Name: "test"}
	c := &modA.Config{Host: "localhost", Port: 8080}
	return u, c
}

// Test that CTOR ignore doesn't affect other violations
func CreateUserWithNewDetected() *modA.User {
	// @ignore IMM
	return new(modA.User) // want "type instantiation with new"
}

// Test multiple codes
// @ignore CTOR01, CTOR02
func CreateMultipleIgnored() (*modA.User, *modA.Config) {
	u := &modA.User{ID: 1, Name: "test"} // Ignored via CTOR01
	c := new(modA.Config)                // Ignored via CTOR02
	return u, c
}

// Test var declarations
func CreateVarDeclarations() {
	var user modA.User     // want "zero-initialized variable declaration must be in constructor"
	var config modA.Config // want "zero-initialized variable declaration must be in constructor"

	// Use variables to avoid "declared and not used" errors
	_ = user
	_ = config

	// Pointer var should be allowed
	var userPtr *modA.User     // OK: nil pointer
	var configPtr *modA.Config // OK: nil pointer
	_ = userPtr
	_ = configPtr
}

// Test @ignore with var declarations (CTOR03)
// @ignore CTOR03
func CreateVarIgnored() {
	var user modA.User     // This should be ignored
	var config modA.Config // This should be ignored
	_ = user
	_ = config
}
