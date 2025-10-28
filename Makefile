.PHONY: build install lint pre-build fmt test

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
	go test  ./...

pre-build: tidy fmt lint test


ci: pre-build build

run: pre-build
	go run ./cmd/gogreement ./...