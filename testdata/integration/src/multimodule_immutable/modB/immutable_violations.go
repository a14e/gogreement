package modB // want package:"package modB"

import "multimodule_immutable/modA"

func MutateUser(u *modA.User) {
	u.Name = "modified" // want "cannot assign to field"
}

func MutateConfig(cfg *modA.Config) {
	cfg.Port = 80 // want "cannot assign to field"
}

// Test @ignore with specific code IMM01
// @ignore IMM01
func MutateUserIgnored(u *modA.User) {
	u.Name = "modified" // This should be ignored
}

// Test @ignore with category IMM
// @ignore IMM
func MutateConfigIgnored(cfg *modA.Config) {
	cfg.Port = 80 // This should be ignored via IMM category
}

// Test inline @ignore
func MutateUserInlineIgnore(u *modA.User) {
	u.Name = "modified" // @ignore IMM01
}

// Test @ignore ALL
// @ignore ALL
func MutateEverythingIgnored(u *modA.User, cfg *modA.Config) {
	u.Name = "changed"
	u.ID = 999
	cfg.Port = 80
	cfg.Host = "example.com"
}

// Test partial ignore - only IMM01 ignored, IMM03 should be detected
func PartialIgnore(u *modA.User) {
	// @ignore IMM01
	u.Name = "changed" // Ignored

	u.ID++ // want "cannot use \\+\\+"
}

// Test that non-ignored violations are still detected
func StillDetected(u *modA.User, cfg *modA.Config) {
	// @ignore CTOR
	u.Name = "changed" // want "cannot assign to field"
	cfg.Port = 80      // want "cannot assign to field"
}
