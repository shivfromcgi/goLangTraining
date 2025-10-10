# Go Training Project Makefile
# This makefile provides common build, test, and maintenance tasks for the simplified project

.PHONY: build test clean fmt lint run help assignment1 assignment2 assignment3

# Default target
help:
	@echo "Go Training Project - Available Commands:"
	@echo "  build       - Build the main application"
	@echo "  test        - Run all tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  fmt         - Format all Go code"
	@echo "  lint        - Run static analysis"
	@echo "  run         - Run assignment 3 (HTTP server)"
	@echo "  assignment1 - Run assignment 1 (message system)"
	@echo "  assignment2 - Run assignment 2 (storage system)"
	@echo "  assignment3 - Run assignment 3 (HTTP server)"
	@echo "  help        - Show this help message"

# Build the main application
build:
	@echo "Building main application..."
	go build -o build/go-training-service .
	@echo "Build complete. Binary available at: ./build/go-training-service"

# Run all tests across the workspace
test:
	@echo "Running tests across workspace..."
	go work sync
	cd src/pkg/storage && go test -v ./...
	go test -v .

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf build/
	go clean
	cd src/pkg/storage && go clean

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
	go vet .
	cd src/pkg/storage && go vet ./...
	# Add golangci-lint if available
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not found, skipping"

# Run the HTTP server (Assignment 3)
run: assignment3

# Assignment targets
assignment1:
	@echo "Running Assignment 1 - Message System..."
	@echo "Usage: make assignment1 USER=<user> MSG=<message>"
	@echo "Example: make assignment1 USER=alice MSG='Hello World'"
	@if [ -z "$(USER)" ] || [ -z "$(MSG)" ]; then \
		echo "Running with sample data..."; \
		go run main.go -assignment=assignment1 -user=sample -message="Sample message from Makefile"; \
	else \
		go run main.go -assignment=assignment1 -user=$(USER) -message="$(MSG)"; \
	fi

assignment2:
	@echo "Running Assignment 2 - Storage System..."
	@echo "Usage: make assignment2 FILE=<file> DATA=<data>"
	@echo "Example: make assignment2 FILE=test.txt DATA='Hello Storage'"
	@if [ -z "$(FILE)" ]; then \
		echo "Running with default file..."; \
		go run main.go -assignment=assignment2; \
	else \
		go run main.go -assignment=assignment2 -file=$(FILE) -data="$(DATA)"; \
	fi

assignment3:
	@echo "Running Assignment 3 - HTTP Server..."
	@echo "Usage: make assignment3 PORT=<port>"
	@echo "Default port: 8080"
	@if [ -z "$(PORT)" ]; then \
		go run main.go -assignment=assignment3 -port=8080; \
	else \
		go run main.go -assignment=assignment3 -port=$(PORT); \
	fi

# Development targets
dev-test:
	@echo "Running tests in watch mode (requires entr)..."
	find . -name "*.go" | entr -r make test

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go work sync
	go mod tidy
	cd src/pkg/storage && go mod tidy