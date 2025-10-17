# Go Training Project Makefile
# This makefile provides common build, test, and maintenance tasks for the simplified project

.PHONY: build test clean fmt lint run help assignment1 assignment2 assignment3 assignment4 build-grpc run-grpc-server run-grpc-client test-grpc proto-gen

# Default target
help:
	@echo "Go Training Project - Available Commands:"
	@echo ""
	@echo "Main Application:"
	@echo "  build       - Build the main application"
	@echo "  test        - Run all tests"
	@echo "  clean       - Clean build artifacts"
	@echo "  fmt         - Format all Go code"
	@echo "  lint        - Run static analysis"
	@echo "  run         - Run assignment 4 (web pages)"
	@echo ""
	@echo "Assignment Targets:"
	@echo "  assignment1 - Run assignment 1 (message system)"
	@echo "  assignment2 - Run assignment 2 (storage system)"
	@echo "  assignment3 - Run assignment 3 (HTTP server)"
	@echo "  assignment4 - Run assignment 4 (web pages)"
	@echo ""
	@echo "gRPC Implementation:"
	@echo "  build-grpc       - Build gRPC server and client"
	@echo "  run-grpc-server  - Start gRPC message store server"
	@echo "  run-grpc-client  - Run gRPC client demo"
	@echo "  test-grpc        - Test gRPC functionality"
	@echo "  proto-gen        - Generate protobuf files"
	@echo ""
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
	@echo "Cleaning build artifacts and temporary files..."
	rm -rf build/
	go clean
	cd src/pkg/storage && go clean
	cd proto/message_service && go clean
	cd store && go clean  
	cd client && go clean
	find . -name "*.log" -not -path "./.git/*" -delete
	find . -name "*.tmp" -not -path "./.git/*" -delete
	find . -name "*~" -not -path "./.git/*" -delete

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

# Run the web server (Assignment 4)
run: assignment4

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

assignment4:
	@echo "Running Assignment 4 - Web Pages..."
	@echo "Usage: make assignment4 PORT=<port>"
	@echo "Default port: 8080"
	@echo "Web Pages available at:"
	@echo "  http://localhost:8080/ - Static home page"
	@echo "  http://localhost:8080/web/messages - Dynamic messages page"
	@if [ -z "$(PORT)" ]; then \
		go run main.go -assignment=assignment4 -port=8080; \
	else \
		go run main.go -assignment=assignment4 -port=$(PORT); \
	fi

# gRPC targets
build-grpc:
	@echo "Building gRPC components..."
	cd store && go build -o ../build/grpc-store-server .
	cd client && go build -o ../build/grpc-client .
	@echo "gRPC build complete:"
	@echo "  Server: ./build/grpc-store-server"
	@echo "  Client: ./build/grpc-client"

run-grpc-server:
	@echo "Starting gRPC Message Store Server..."
	cd store && go run .

run-grpc-client:
	@echo "Running gRPC Client demo..."
	cd client && go run .

test-grpc:
	@echo "Testing gRPC functionality..."
	@echo "1. Save a message:"
	cd client && go run . -user=test -message="Makefile test message"
	@echo "2. Get messages:"
	cd client && go run . -get

proto-gen:
	@echo "Generating protobuf files..."
	protoc --go_out=proto/message_service --go-grpc_out=proto/message_service proto/message_service.proto
	@echo "Protobuf files generated in proto/message_service/"

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
	cd proto/message_service && go mod tidy
	cd store && go mod tidy
	cd client && go mod tidy