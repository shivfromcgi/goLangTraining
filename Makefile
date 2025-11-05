# CGI Go Training Project Makefile
# Restructured following coding guidelines with separate services
# This makefile provides common build, test, and maintenance tasks

.PHONY: build test clean fmt lint run help cli api grpc web build-all run-cli run-api run-grpc run-web test-all proto-gen

# Default target
help:
	@echo "CGI Go Training Project - Restructured Services"
	@echo "=============================================="
	@echo ""
	@echo "Build Commands:"
	@echo "  build-all   - Build all services using scripts/build.sh"
	@echo "  cli         - Build CLI service"
	@echo "  api         - Build API service" 
	@echo "  grpc        - Build gRPC service"
	@echo "  web         - Build web service"
	@echo ""
	@echo "Run Commands:"
	@echo "  run-cli     - Run CLI service"
	@echo "  run-api     - Run API service on port 8080"
	@echo "  run-grpc    - Run gRPC service"
	@echo "  run-web     - Run web service on port 8090"
	@echo ""
	@echo "Development:"
	@echo "  test-all    - Run tests across all services"
	@echo "  clean       - Clean build artifacts" 
	@echo "  fmt         - Format all Go code"
	@echo "  lint        - Run static analysis"
	@echo "  proto-gen   - Generate protobuf files"
	@echo ""
	@echo "  help        - Show this help message"

# Build all services using scripts
build-all:
	@echo "Building all services using scripts/build.sh..."
	@./scripts/build.sh

# Version information for builds
VERSION ?= dev
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS = -X cgi.com/goLangTraining/src/pkg/version.Version=$(VERSION) -X cgi.com/goLangTraining/src/pkg/version.BuildDate=$(BUILD_DATE) -X cgi.com/goLangTraining/src/pkg/version.GitCommit=$(GIT_COMMIT)

# Build individual services
cli:
	@echo "Building CLI service..."
	cd src/apps/message-cli && go build -ldflags "$(LDFLAGS)" -o ../../../build/message-cli .

api:
	@echo "Building API service..."
	cd src/apps/message-api && go build -ldflags "$(LDFLAGS)" -o ../../../build/message-api .

grpc:
	@echo "Building gRPC service..."
	cd src/apps/message-grpc && go build -ldflags "$(LDFLAGS)" -o ../../../build/message-grpc .
	cd src/apps/message-grpc/cmd/client && go build -ldflags "$(LDFLAGS)" -o ../../../../../build/grpc-client .

web:
	@echo "Building web service..."
	cd src/apps/message-web && go build -ldflags "$(LDFLAGS)" -o ../../../build/message-web .

# Legacy build target
build: build-all

# Run tests across all services
test-all:
	@echo "Running tests across all services..."
	go work sync
	cd src/pkg/storage && go test -v ./...
	cd src/pkg/types && go test -v ./... || echo "No tests in types package"
	@echo "Tests completed for all services"

# Legacy test target  
test: test-all

# Clean build artifacts and temporary files
clean:
	@echo "Cleaning build artifacts and temporary files..."
	rm -rf build/
	go work sync
	cd src/pkg/storage && go clean
	cd src/pkg/types && go clean  
	cd src/apps/message-cli && go clean
	cd src/apps/message-api && go clean
	cd src/apps/message-grpc && go clean
	cd src/apps/message-web && go clean
	cd proto/message_service && go clean
	find . -name "*.log" -not -path "./.git/*" -delete
	find . -name "*.tmp" -not -path "./.git/*" -delete
	find . -name "*~" -not -path "./.git/*" -delete

# Format all Go code using gofmt and goimports following guidelines
fmt:
	@echo "Formatting code..."
	go work sync
	find . -name "*.go" -not -path "./build/*" -not -path "./.git/*" | xargs gofmt -w
	@which goimports > /dev/null && find . -name "*.go" -not -path "./build/*" -not -path "./.git/*" | xargs goimports -w || echo "goimports not found, skipping"

# Run static analysis following guidelines
lint:
	@echo "Running static analysis..."
	go work sync
	cd src/pkg/storage && go vet ./...
	cd src/pkg/types && go vet ./...
	cd src/apps/message-cli && go vet .
	cd src/apps/message-api && go vet .
	cd src/apps/message-grpc && go vet .
	cd src/apps/message-web && go vet .
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not found, skipping"
	@which staticcheck > /dev/null && staticcheck ./... || echo "staticcheck not found, skipping"

# Run services following the new structure
run-cli:
	@echo "Running CLI service..."
	@echo "Usage: make run-cli USER=<user> MSG=<message>"
	@echo "Example: make run-cli USER=alice MSG='Hello World'"
	@if [ -z "$(USER)" ] || [ -z "$(MSG)" ]; then \
		echo "Running with sample data..."; \
		cd src/apps/message-cli && go run . -user=sample -message="Sample CLI message"; \
	else \
		cd src/apps/message-cli && go run . -user=$(USER) -message="$(MSG)"; \
	fi

run-api:
	@echo "Running API service..."
	@echo "Usage: make run-api PORT=<port>"
	@echo "Default port: 8080"
	@echo "API endpoints available at:"
	@echo "  http://localhost:8080/api/v1/messages - JSON API"
	@echo "  http://localhost:8080/api/v1/health - Health check"
	@if [ -z "$(PORT)" ]; then \
		cd src/apps/message-api && go run . -port=8080; \
	else \
		cd src/apps/message-api && go run . -port=$(PORT); \
	fi

run-grpc:
	@echo "Running gRPC service..."
	@echo "gRPC server will start on port :50051"
	@echo "Use 'make run-grpc-client' to test"
	cd src/apps/message-grpc && go run .

run-web:
	@echo "Running web service..."
	@echo "Usage: make run-web PORT=<port>"
	@echo "Default port: 8090"
	@echo "Web pages available at:"
	@echo "  http://localhost:8090/ - Static home page" 
	@echo "  http://localhost:8090/messages - Dynamic messages page"
	@if [ -z "$(PORT)" ]; then \
		cd src/apps/message-web && go run . -port=8090; \
	else \
		cd src/apps/message-web && go run . -port=$(PORT); \
	fi

# Legacy run target defaults to API
run: run-api

# gRPC specific targets
run-grpc-client:
	@echo "Running gRPC Client..."
	@echo "Usage: make run-grpc-client USER=<user> MSG=<message>"
	@echo "Example: make run-grpc-client USER=alice MSG='Hello gRPC!'"
	@if [ -z "$(USER)" ] || [ -z "$(MSG)" ]; then \
		echo "Running gRPC client demo..."; \
		cd src/apps/message-grpc/cmd/client && go run .; \
	else \
		cd src/apps/message-grpc/cmd/client && go run . -user=$(USER) -message="$(MSG)"; \
	fi

test-grpc:
	@echo "Testing gRPC functionality..."
	@echo "1. Save a message:"
	cd src/apps/message-grpc/cmd/client && go run . -user=test -message="Makefile test message"
	@echo "2. Get messages:"
	cd src/apps/message-grpc/cmd/client && go run . -get

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