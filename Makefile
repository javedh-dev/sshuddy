.PHONY: build run clean test install help

# Build variables
BINARY_NAME=sshbuddy
BUILD_DIR=.
CMD_DIR=./cmd/sshbuddy
VERSION?=dev

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@go run $(CMD_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@rm -rf .go-dist/
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Install the binary to GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME)..."
	@go install $(CMD_DIR)
	@echo "Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

# Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@goreleaser build --snapshot --clean
	@echo "Builds available in dist/"

# Create a release
release:
	@echo "Creating release..."
	@goreleaser release --clean

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Run the application"
	@echo "  clean         - Remove build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  deps          - Install dependencies"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  build-all     - Build for all platforms"
	@echo "  release       - Create a release"
	@echo "  help          - Show this help message"
