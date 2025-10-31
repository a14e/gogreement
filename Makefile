.PHONY: build install lint pre-build fmt test

# TEST_ARGS allows passing additional arguments to go test
# Examples:
#   make test                           # Run tests with default settings
#   make test TEST_ARGS="-v"           # Run tests with verbose output
#   make test TEST_ARGS="-coverprofile=coverage.txt"  # Generate coverage profile
#   make test TEST_ARGS="-race -cover" # Run tests with race detection and coverage

build: pre-build
	go build ./cmd/gogreement

install: tidy
	go install go.uber.org/nilaway/cmd/nilaway@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

tidy:
	go mod tidy

lint:
	golangci-lint run
	nilaway ./...

fmt:
	go fmt ./...

test:
	go test $(if $(TEST_ARGS),$(TEST_ARGS),) ./...

pre-build: tidy fmt lint test


ci: pre-build build

run: pre-build
	go run ./cmd/gogreement ./...