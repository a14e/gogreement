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

// Test @mutable field modifications - these are allowed
func UpdateUserCache(u *modA.User) {
	u.Cache = map[string]interface{}{"key": "value"} // Allowed - Cache is @mutable
}

func UpdateConfigMetadata(cfg *modA.Config) {
	cfg.Metadata = map[string]string{"env": "test"} // Allowed - Metadata is @mutable
}

// Test @mutable field with different operations
func MutableFieldOperations(u *modA.User, cfg *modA.Config) {
	u.Cache["new_key"] = "new_value" // Allowed - map access on @mutable field
	cfg.Metadata["version"] = "1.0"  // Allowed - map access on @mutable field
}

// Test mixed case - mutable allowed, non-mutable not allowed
func MixedMutableOperations(u *modA.User, cfg *modA.Config) {
	u.Cache = map[string]interface{}{"test": true} // Allowed
	u.Name = "changed"                             // want "cannot assign to field"

	cfg.Metadata = map[string]string{"debug": "true"} // Allowed
	cfg.Port = 9080                                   // want "cannot assign to field"
}
