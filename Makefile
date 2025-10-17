.PHONY: build install swagger lint pre-build  fmt

build: pre-build
	go build

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

pre-build: tidy fmt lint