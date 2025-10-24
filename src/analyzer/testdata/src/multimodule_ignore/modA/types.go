package modA

// Service provides business logic
// @ignore LINT001
type Service struct {
	Name string
	Port int
}

// NewService creates a new service
// @ignore LINT002, LINT003
func NewService(name string, port int) *Service {
	return &Service{
		Name: name,
		Port: port,
	}
}

// Config holds configuration
type Config struct {
	Host string
	Port int
}

// GetConfig returns config
// @ignore DEPRECATED
func GetConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 8080,
	}
}
