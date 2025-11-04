CGI Go Training Service – Restructured Architecture

A modular Go application following coding guidelines with separate services for CLI, API, gRPC, and Web interfaces.
This restructured version follows proper Go project organization, separation of concerns, and production-ready patterns.

## Project Overview

This repository has been restructured following coding guidelines to provide:
- **Separate Services**: CLI, API, gRPC, and Web services as independent applications
- **Shared Libraries**: Common types and storage functionality in `src/pkg/`
- **Proper Structure**: Following recommended Go project layout patterns
- **Build Automation**: Scripts and Makefile for streamlined development

## Key Features

✅ **Modular Architecture**: Each service is independent and focused  
✅ **Shared Packages**: Common functionality in `src/pkg/`  
✅ **Coding Guidelines**: Follows established Go best practices  
✅ **Build Scripts**: Automated build and deployment scripts  
✅ **Multiple Interfaces**: CLI, REST API, gRPC, and Web UI  
✅ **Production Ready**: Structured logging, health checks, graceful shutdown  

## Project Structure

```
OpenMedia_GoLang_Course/
├── scripts/                    # Build and automation scripts
│   ├── build.sh               # Main build script
│   └── Dockerfile             # Container build definition
├── src/
│   ├── apps/                  # Individual services
│   │   ├── message-cli/       # CLI service
│   │   ├── message-api/       # REST API service
│   │   ├── message-grpc/      # gRPC service with client
│   │   └── message-web/       # Web interface service
│   └── pkg/                   # Shared packages
│       ├── storage/           # Message storage functionality
│       └── types/             # Shared type definitions
├── proto/                     # Protocol buffer definitions
├── build/                     # Compiled binaries (generated)
├── go.work                    # Go workspace configuration
├── Makefile                   # Build and run tasks
└── README.md                  # This file
```

## Prerequisites

- **Go 1.22+**
- **Make** (for build automation)
- **Protocol Buffers** (for gRPC development)

## Quick Start

### Build All Services
```bash
# Use automated build script
make build-all
# or
./scripts/build.sh
```

### Run Individual Services

#### CLI Service
```bash
# Run CLI service directly
make run-cli USER=alice MSG="Hello CLI!"

# Or build and run binary
make cli
./build/message-cli -user=alice -message="Hello CLI!"
```

#### API Service
```bash
# Run API service (default port 8080)
make run-api

# Or with custom port
make run-api PORT=9000

# Or build and run binary
make api
./build/message-api -port=8080
```

#### gRPC Service
```bash
# Start gRPC server (port :50051)
make run-grpc

# Test with gRPC client
make run-grpc-client USER=alice MSG="Hello gRPC!"

# Or build and run binaries
make grpc
./build/message-grpc &
./build/grpc-client -user=alice -message="Hello gRPC!"
```

#### Web Service
```bash
# Run web service (default port 8090)
make run-web

# Or with custom port  
make run-web PORT=8000

# Or build and run binary
make web
./build/message-web -port=8090
```

## Service Details

### CLI Service (`src/apps/message-cli/`)
- **Purpose**: Command-line message operations
- **Features**: Add messages, clear messages, list messages
- **Usage**: `./build/message-cli -user=alice -message="Hello!"`

### API Service (`src/apps/message-api/`)
- **Purpose**: REST API for message management
- **Port**: 8080 (default)
- **Endpoints**:
  - `GET /api/v1/messages` - Retrieve messages
  - `POST /api/v1/messages` - Create message
  - `GET /api/v1/health` - Health check
- **Example**:
  ```bash
  curl -X POST http://localhost:8080/api/v1/messages \
    -H 'Content-Type: application/json' \
    -d '{"user":"alice","message":"Hello API!"}'
  ```

### gRPC Service (`src/apps/message-grpc/`)
- **Purpose**: gRPC server and client for message operations
- **Port**: :50051
- **Services**: `Save`, `GetLast10`
- **Client**: Located in `cmd/client/`

### Web Service (`src/apps/message-web/`)
- **Purpose**: Web interface for viewing messages
- **Port**: 8090 (default)  
- **Pages**:
  - `/` - Static home page
  - `/messages` - Dynamic message listing
  - `/health` - Health check

## Testing

```bash
# Run all tests
make test-all

# Test gRPC functionality
make test-grpc
```

## Development

### Code Formatting
```bash
# Format all Go code
make fmt

# Run static analysis  
make lint
```

### Building Docker Images
```bash
# Build container using provided Dockerfile
docker build -f scripts/Dockerfile -t go-training-services .

# Run containerized API service
docker run -p 8080:8080 go-training-services
```

## Architecture Highlights

### Design Principles
- **Separation of Concerns**: Each service has a single responsibility
- **Shared Libraries**: Common functionality in `src/pkg/`  
- **Following Guidelines**: Adheres to established Go coding standards
- **Modular Structure**: Easy to extend and maintain
- **Production Ready**: Structured logging, health checks, graceful shutdown

### Service Independence
Each service in `src/apps/` is:
- **Self-contained**: Has its own `go.mod` and dependencies
- **Independently Deployable**: Can be built and run separately
- **Focused**: Single responsibility (CLI, API, gRPC, or Web)
- **Testable**: Individual testing and validation

### Shared Packages
- **`src/pkg/types/`**: Common data structures and types
- **`src/pkg/storage/`**: Message storage functionality
- **Protocol Buffers**: Shared gRPC definitions in `proto/`

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build-all` | Build all services |
| `make run-cli` | Run CLI service |
| `make run-api` | Run API service |
| `make run-grpc` | Run gRPC service |
| `make run-web` | Run web service |
| `make test-all` | Run all tests |
| `make clean` | Clean build artifacts |
| `make fmt` | Format code |
| `make lint` | Run static analysis |

## Technologies Used

- **Language**: Go 1.22+
- **Frameworks**: http.ServeMux, gRPC, html/template
- **Storage**: File-based message storage
- **Build**: Make, shell scripts, Docker
- **Testing**: go test, testify
- **Protocols**: HTTP REST, gRPC, Protocol Buffers

## Project Status

This restructured version addresses the review feedback:
- ✅ **Proper folder structure** following coding guidelines
- ✅ **Separated CLI, API, gRPC, and Web** into individual services
- ✅ **Removed empty files** (`main`, `goLangTraining`)
- ✅ **Clear gRPC client/server** organization in `message-grpc/`
- ✅ **Follows coding standards** for Go project layout

## License

This project is developed as part of the CGI Go Academy Training Program.
Use, modify, and extend for learning purposes.