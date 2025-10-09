#!/bin/bash
set -e

# build.sh - Local build script for Go training service
# This script provides a consistent build environment for local development
# and CI/CD pipelines by standardizing build commands and error handling.

echo "=== Go Training Service Build Script ==="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    exit 1
fi

# Display Go version for build reproducibility
echo "Go version: $(go version)"

# Sync workspace modules first
echo "Syncing workspace modules..."
go work sync

# Build the main service
echo "Building go-training-service..."
cd src/apps/go-training-service
go mod tidy
go build -o ../../../build/go-training-service ./cmd/server

# Run tests with coverage
echo "Running tests..."
go test -v -race -coverprofile=../../../coverage.out ./...

# Run tests for shared packages
echo "Running shared package tests..."
cd ../../../src/pkg/storage
go test -v -race ./...

# Return to project root
cd ../../../

echo "Build completed successfully!"
echo "Binary location: build/go-training-service"
echo ""
echo "Usage examples:"
echo "  ./build/go-training-service -assignment=assignment1 -user=alice -message=\"Hello\""
echo "  ./build/go-training-service -assignment=assignment2"
echo "  ./build/go-training-service -assignment=assignment3 -port=8080"