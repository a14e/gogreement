package modB // want package:"package modB"

import "multimodule_testonly/modA"

func UseTestOnlyFunction() {
	_ = modA.CreateTestData() // want "marked @testonly"
}

func UseTestOnlyType() {
	var helper modA.TestHelper // want "marked @testonly"
	_ = helper
}
