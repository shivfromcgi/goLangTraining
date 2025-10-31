#!/bin/bash

# Build script for CGI Go Training Services
# Following coding guidelines for automation scripts

set -e

echo "🔨 Building CGI Go Training Services"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build directory
BUILD_DIR="build"
mkdir -p "$BUILD_DIR"

# Function to build a service
build_service() {
    local service_dir=$1
    local service_name=$2
    local binary_name=$3
    
    echo -e "${YELLOW}Building $service_name...${NC}"
    
    cd "$service_dir" || {
        echo -e "${RED}Failed to enter directory $service_dir${NC}"
        return 1
    }
    
    # Sync workspace and build
    go work sync 2>/dev/null || true
    go mod tidy
    
    if go build -o "../../$BUILD_DIR/$binary_name" .; then
        echo -e "${GREEN}✅ Built $service_name successfully${NC}"
    else
        echo -e "${RED}❌ Failed to build $service_name${NC}"
        return 1
    fi
    
    cd - > /dev/null
}

# Function to build gRPC client
build_grpc_client() {
    echo -e "${YELLOW}Building gRPC Client...${NC}"
    
    cd "src/apps/message-grpc/cmd/client" || {
        echo -e "${RED}Failed to enter gRPC client directory${NC}"
        return 1
    }
    
    go work sync 2>/dev/null || true
    go mod tidy 2>/dev/null || go mod init cgi.com/goLangTraining/grpc-client
    
    if go build -o "../../../../$BUILD_DIR/grpc-client" .; then
        echo -e "${GREEN}✅ Built gRPC Client successfully${NC}"
    else
        echo -e "${RED}❌ Failed to build gRPC Client${NC}"
        return 1
    fi
    
    cd - > /dev/null
}

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf "$BUILD_DIR"/*

# Build all services
echo -e "${YELLOW}Starting builds...${NC}"

# Build message services
build_service "src/apps/message-cli" "Message CLI" "message-cli"
build_service "src/apps/message-api" "Message API" "message-api" 
build_service "src/apps/message-grpc" "Message gRPC Server" "message-grpc"
build_service "src/apps/message-web" "Message Web" "message-web"

# Build gRPC client
build_grpc_client

echo ""
echo -e "${GREEN}🎉 All services built successfully!${NC}"
echo "======================================"
echo -e "${YELLOW}Available binaries in $BUILD_DIR/:${NC}"
ls -la "$BUILD_DIR/"

echo ""
echo -e "${YELLOW}Usage:${NC}"
echo "  ./build/message-cli -user=alice -message='Hello CLI!'"
echo "  ./build/message-api -port=8080"
echo "  ./build/message-web -port=8090"
echo "  ./build/message-grpc"
echo "  ./build/grpc-client -user=alice -message='Hello gRPC!'"

echo ""
echo -e "${GREEN}Build completed successfully!${NC}"