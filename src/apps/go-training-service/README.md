# Go Training Project

A comprehensive Go training project demonstrating modern Go development practices, monorepo structure, and clean architecture patterns.

## Project Structure

This project follows a monorepo structure with clear separation between applications and shared libraries:

```
├── src/
│   ├── apps/
│   │   └── go-training-service/          # Main training service
│   │       ├── cmd/server/               # Service entrypoint
│   │       ├── internal/                 # Internal service code
│   │       │   ├── handler/              # HTTP handlers
│   │       │   ├── service/              # Business logic
│   │       │   ├── repository/           # Data access layer
│   │       │   └── types/                # Service-specific types
│   │       └── go.mod                    # Service module
│   └── pkg/
│       └── storage/                      # Shared storage library
│           ├── storage.go                # Core storage functionality
│           ├── storage_test.go           # Comprehensive tests
│           ├── types.go                  # Shared types
│           └── go.mod                    # Library module
├── scripts/
│   └── build.sh                          # Build automation script
├── build/                                # Build artifacts
├── tests/                                # Integration tests
├── go.work                               # Workspace configuration
├── Makefile                              # Build automation
└── README.md                             # This file
```

## Features

### Assignment 1 - Message System
- File-based message persistence
- User message tracking
- Clean CLI interface

### Assignment 2 - Advanced Storage
- Context-aware file operations
- Structured logging with trace IDs
- Graceful shutdown handling

### Assignment 3 - HTTP JSON API
- RESTful message API
- Middleware for request tracing
- Health check endpoints
- Graceful server shutdown

## Development Setup

### Prerequisites
- Go 1.21 or higher
- Make (for build automation)

### Quick Start

1. **Clone and setup workspace:**
   ```bash
   git clone <repository-url>
   cd go-training-project
   make deps
   ```

2. **Build the project:**
   ```bash
   make build
   ```

3. **Run tests:**
   ```bash
   make test
   ```

4. **Start the service:**
   ```bash
   # Assignment 1 - Message System
   ./build/go-training-service -assignment=assignment1 -user=alice -message="Hello World"
   
   # Assignment 2 - Storage System
   ./build/go-training-service -assignment=assignment2 -file=test.txt -data="Sample content"
   
   # Assignment 3 - HTTP API
   ./build/go-training-service -assignment=assignment3 -port=8080
   ```

## API Documentation

### Assignment 3 - HTTP Endpoints

#### Create Message
```bash
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{"user":"alice","message":"Hello API!"}'
```

#### Get All Messages
```bash
curl http://localhost:8080/messages
```

#### Health Check
```bash
curl http://localhost:8080/health
```

## Coding Standards

This project follows comprehensive Go coding standards including:

### Core Principles
- **Consistency is King**: Maintained across entire codebase
- **Build Small**: Minimal abstractions, focus on readability
- **No OO**: Idiomatic Go composition patterns
- **YAGNI**: Features implemented only when needed

### Code Structure
- Early returns to reduce nesting
- Table-driven tests for comprehensive coverage
- Structured logging with trace IDs
- Context-aware operations
- Graceful shutdown handling

### Error Handling
- No inline error throwing
- Early returns on errors
- Proper HTTP status codes
- Error context at stack boundaries

### Testing Strategy
- **Meaningful tests > code coverage**
- Table-driven test patterns
- Sandbox smoke tests
- Integration testing approach
- Minimal mocking

## Development Commands

| Command | Description |
|---------|-------------|
| `make help` | Show available commands |
| `make build` | Build all services |
| `make test` | Run all tests |
| `make clean` | Clean build artifacts |
| `make fmt` | Format all Go code |
| `make lint` | Run static analysis |
| `make run` | Run the training service |
| `make dev-run` | Run in development mode |
| `make deps` | Install dependencies |

## Architecture Decisions

### Monorepo Structure
- **Why**: Simplified dependency management and coordinated releases
- **Trade-off**: Larger repository size vs easier cross-service changes

### Go Workspaces
- **Why**: Native Go support for multi-module development
- **Benefits**: Consistent tooling across modules

### Structured Logging
- **Why**: Production-ready observability and debugging
- **Implementation**: Slog with JSON output and trace correlation

### Context-Driven Design
- **Why**: Request scoping, cancellation, and distributed tracing
- **Pattern**: Context passed through all operations

## Contributing

1. Follow the established coding standards
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Use meaningful commit messages
5. Ensure all tests pass before submitting

## License

This project is for training purposes and follows CGI coding standards and best practices.