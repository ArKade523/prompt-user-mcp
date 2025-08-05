.DEFAULT_GOAL := all

# Build the CLI binary
build:
	go build -o bin/prompt-mcp ./cli

# Run all tests
test:
	go test ./test/... -v

# Clean build artifacts
clean:
	rm -rf bin/

# Build and test everything
all: build test

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	go vet ./...

# Create bin directory if it doesn't exist
bin:
	mkdir -p bin

.PHONY: build test clean all deps fmt lint