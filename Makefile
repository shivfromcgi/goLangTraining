# Go Training Project Makefile
# This makefile provides common build, test, and maintenance tasks for the monorepo

.PHONY: build test clean fmt lint run help

# Default target
help:
	@echo "Go Training Project - Available Commands:"
	@echo "  build     - Build all services"
	@echo "  test      - Run all tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  fmt       - Format all Go code"
	@echo "  lint      - Run static analysis"
	@echo "  run       - Run the training service"
	@echo "  help      - Show this help message"

# Build all services in the monorepo
build:
	@echo "Building all services..."
	cd src/apps/go-training-service && go build -o ../../../build/go-training-service ./cmd/server
	@echo "Build complete. Binary available at: ./build/go-training-service"

# Run all tests across the workspace
test:
	@echo "Running tests across workspace..."
	go work sync
	cd src/pkg/storage && go test -v ./...
	cd src/apps/go-training-service && go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/
	cd src/pkg/storage && go clean
	cd src/apps/go-training-service && go clean

# Format all Go code using gofmt and goimports
fmt:
	@echo "Formatting code..."
	go work sync
	find . -name "*.go" -not -path "./build/*" | xargs gofmt -w
	find . -name "*.go" -not -path "./build/*" | xargs goimports -w

# Run static analysis
lint:
	@echo "Running static analysis..."
	go work sync
	cd src/pkg/storage && go vet ./...
	cd src/apps/go-training-service && go vet ./...
	# Add golangci-lint if available
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not found, skipping"

# Run the training service
run: build
	@echo "Starting Go Training Service..."
	./build/go-training-service -assignment=assignment3 -port=8080

# Development targets
dev-run:
	@echo "Running service in development mode..."
	cd src/apps/go-training-service && go run ./cmd/server -assignment=assignment3 -port=8080

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go work sync
	cd src/pkg/storage && go mod tidy
	cd src/apps/go-training-service && go mod tidy