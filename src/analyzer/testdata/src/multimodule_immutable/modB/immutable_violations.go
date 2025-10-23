package modB // want package:"package modB"

import "multimodule_immutable/modA"

func MutateUser(u *modA.User) {
	u.Name = "modified" // want "cannot assign to field"
}

func MutateConfig(cfg *modA.Config) {
	cfg.Port = 80 // want "cannot assign to field"
}
