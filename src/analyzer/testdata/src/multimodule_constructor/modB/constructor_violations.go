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
