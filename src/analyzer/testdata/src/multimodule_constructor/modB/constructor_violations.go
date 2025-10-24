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
