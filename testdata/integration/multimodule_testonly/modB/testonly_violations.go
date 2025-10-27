package modB // want package:"package modB"

import "multimodule_testonly/modA"

func UseTestOnlyFunction() {
	_ = modA.CreateTestData() // want "marked @testonly"
}

func UseTestOnlyType() {
	var helper modA.TestHelper // want "marked @testonly"
	_ = helper
}

func UseTestOnlyMethod() {
	u := modA.NewUser(1, "test")
	_ = u.GetDebugInfo() // want "marked @testonly"
}

// Test @ignore with specific code TONL02 (function)
// @ignore TONL02
func UseTestOnlyFunctionIgnored() {
	_ = modA.CreateTestData() // This should be ignored
}

// Test @ignore with specific code TONL01 (type)
// @ignore TONL01
func UseTestOnlyTypeIgnored() {
	var helper modA.TestHelper // This should be ignored
	_ = helper
}

// Test @ignore with specific code TONL03 (method)
// @ignore TONL03
func UseTestOnlyMethodIgnored() {
	u := modA.NewUser(1, "test")
	_ = u.GetDebugInfo() // This should be ignored
}

// Test @ignore with category TONL
// @ignore TONL
func UseAllTestOnlyIgnored() {
	var helper modA.TestHelper    // Ignored via TONL
	data := modA.CreateTestData() // Ignored via TONL
	u := modA.NewUser(1, "test")
	_ = u.GetDebugInfo() // Ignored via TONL
	_ = helper
	_ = data
}

// Test inline @ignore for function call
func UseTestOnlyFunctionInlineIgnore() {
	_ = modA.CreateTestData() // @ignore TONL02
}

// Test inline @ignore for type usage
func UseTestOnlyTypeInlineIgnore() {
	var helper modA.TestHelper // @ignore TONL01
	_ = helper
}

// Test @ignore ALL
// @ignore ALL
func UseEverythingIgnored() {
	var helper modA.TestHelper
	data := modA.CreateTestData()
	u := modA.NewUser(1, "test")
	_ = u.GetDebugInfo()
	_ = helper
	_ = data
}

// Test multiple codes
// @ignore TONL01, TONL02
func UseMultipleTestOnlyIgnored() {
	var helper modA.TestHelper    // Ignored via TONL01
	data := modA.CreateTestData() // Ignored via TONL02
	_ = helper
	_ = data

	u := modA.NewUser(1, "test")
	_ = u.GetDebugInfo() // want "marked @testonly"
}

// Test partial ignore with inline - tests that ignored violations don't prevent detection of subsequent violations
func UsePartialIgnore() {
	var d1 modA.DebugConfig // @ignore TONL01
	_ = d1

	var h2 modA.DebugConfig // want "marked @testonly"
	_ = h2
}
