package testonlyviolations

import "testing"

// This file is a test file, so it CAN use @testonly items

func TestUseTestOnly(t *testing.T) {
	h := TestHelper{Data: "test"} // OK: in test file
	_ = h

	data := CreateMockData() // OK: in test file
	_ = data

	w := &Worker{}
	w.Reset() // OK: in test file
}
