package modB

import "multimodule_ignore/modA"

// UseService uses the service from modA
// @ignore LINT004
func UseService() {
	svc := modA.NewService("test", 3000)
	_ = svc
}

// ProcessData processes data
func ProcessData() {
	// @ignore LINT005, LINT006, LINT007
	cfg := modA.GetConfig()
	_ = cfg
}
