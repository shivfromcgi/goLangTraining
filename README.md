# CGI Go Training Service - Unified Application

This project demonstrates a **single cohesive Go application** where all assignments from the CGI Go Academy course are implemented as **features**, not separate modes. The architecture addresses code review feedback by eliminating unnecessary custom types, simplifying storage functions, and creating a unified application structure.

## Project Structure

```
OpenMedia_GoLang_Course/
├── main.go              # Single unified application with all features
├── go.mod               # Go module file with dependencies
├── go.work              # Workspace file for multi-module setup
├── Makefile             # Build and run commands
├── src/pkg/storage/     # Simplified storage package
│   ├── storage.go       # Direct file operations (no custom types)
│   ├── storage_test.go  # Unit tests for storage package
│   └── types.go         # Simple metrics struct only
├── html/                # Web templates for Assignment 4
│   ├── index.html       # Static home page
│   ├── messages.html    # Dynamic messages template
│   └── styles.css       # CSS styling
├── messages.txt         # Message storage file
└── README.md           # This file
```

## Key Architectural Improvements

Based on code review feedback, this version addresses:

1. **❌ Removed Custom FileOperation Type**: No more `FileOperation string` - just use plain strings
2. **❌ Simplified Storage Functions**: Merged `writeFileContent` into `SaveData` - no unnecessary abstractions  
3. **✅ Unified Application**: All assignments are features in one cohesive app, not separate modes
4. **✅ Cleaner Code Structure**: Eliminated redundant helper functions and complex patterns
4. **✅ Better Readability**: Code is more straightforward and easier to understand
5. **✅ Single Application Logic**: Features work together instead of being isolated modes

## Application Modes

### 🖥️ **Default: Unified Web Application** 
Starts a comprehensive web server with all features integrated:

```bash
# Start with default port (8080)
go run main.go

# Start with custom port
go run main.go -port=9090
```

**Features Available:**
- 📱 **Web Interface**: Home page and dynamic messages page (Assignment 4)
- 🔌 **REST API**: Messages, health check, and file operations (Assignments 1, 2, 3)
- 📊 **Structured Logging**: All operations traced with UUIDs
- 🛑 **Graceful Shutdown**: Proper cleanup on signal handling

### 🖱️ **CLI Mode**
Individual feature testing and operations:

```bash
# Message operations (Assignment 1)
go run main.go -cli -user=alice -message='Hello World'
go run main.go -cli -clear

# Storage demonstration (Assignment 2)  
go run main.go -cli -storage-demo
go run main.go -cli -storage-demo -file=test.txt -data='Custom data'

# Show CLI usage
go run main.go -cli
```

## API Endpoints

### 🔌 REST API (All assignments integrated)

```bash
# Messages API (Assignment 1 feature)
GET  /api/messages              # List all messages
POST /api/messages              # Create new message
GET  /messages                  # Legacy endpoint

# File Storage API (Assignment 2 feature)
POST /api/files                 # Save or read files

# Health Check (Assignment 3 feature)
GET  /api/health                # Service health status
GET  /health                    # Legacy endpoint
```

### 📱 Web Interface (Assignment 4)

```bash
GET  /                          # Static home page
GET  /web/messages              # Dynamic messages page
GET  /static/styles.css         # Static assets
```

## Quick Testing Examples

```bash
# Start the unified application
go run main.go -port=8090

# Test message creation (Assignment 1 via API)
curl -X POST http://localhost:8090/api/messages \
  -H 'Content-Type: application/json' \
  -d '{"user":"demo","message":"Hello unified app!"}'

# Test file storage (Assignment 2 via API)
curl -X POST http://localhost:8090/api/files \
  -H 'Content-Type: application/json' \
  -d '{"file_path":"test.txt","data":"Hello storage!","action":"save"}'

# Test health check (Assignment 3)
curl http://localhost:8090/api/health

# View web interface (Assignment 4)
open http://localhost:8090/
open http://localhost:8090/web/messages
```

## Usage

### Running Assignment 1 (Message System)
```bash
go run main.go -assignment=assignment1 -user=alice -message="Hello World"
go run main.go -assignment=assignment1 -clear
```

### Running Assignment 2 (Advanced Storage)
```bash
go run main.go -assignment=assignment2
go run main.go -assignment=assignment2 -file=custom.txt -data="My custom content"
```

### Running Assignment 3 (HTTP JSON API)
```bash
go run main.go -assignment=assignment3
go run main.go -assignment=assignment3 -port=9090
```

### Running Assignment 4 (Web Pages)
```bash
go run main.go -assignment=assignment4
go run main.go -assignment=assignment4 -port=8080

# Visit these URLs in your browser:
# http://localhost:8080/             - Static home page
# http://localhost:8080/web/messages - Dynamic messages page with last 10 messages
# http://localhost:8080/static/styles.css - CSS stylesheet
```

## Makefile Commands

```bash
make help          # Show all available commands
make build         # Build the application
make assignment1   # Run Assignment 1 with sample data
make assignment2   # Run Assignment 2 with default settings
make assignment3   # Run Assignment 3 (HTTP server)
make assignment4   # Run Assignment 4 (web pages)
make test          # Run all tests
make clean         # Clean build artifacts
```

## Assignment 4 - Web Pages Features

Assignment 4 extends the project with modern web functionality using Go's embedded file system and HTML templates:

### Core Features Implemented

1. **Static Content Serving**
   - Uses `//go:embed` to bundle HTML, CSS, and assets at build time
   - Serves static files via `http.FileServer` with embedded filesystem
   - Static home page with navigation and usage instructions

2. **Dynamic HTML Templates**
   - Dynamic messages page showing last 10 messages from `messages.txt`
   - Uses `html/template` with `template.ParseFS` for embedded templates
   - Template data includes message history, timestamps, and trace IDs

3. **Modern Web UI**
   - Responsive CSS with gradient backgrounds and animations
   - Auto-refresh functionality for real-time message updates
   - Professional styling with hover effects and smooth transitions

4. **Integrated API Access**
   - All Assignment 3 JSON API endpoints remain available
   - Web pages provide links to JSON endpoints for API access
   - Maintains backward compatibility with existing functionality

### Web Endpoints

| Endpoint | Type | Description |
|----------|------|-------------|
| `/` | Static | Home page with navigation and instructions |
| `/web/messages` | Dynamic | Template-rendered messages page |
| `/static/styles.css` | Static | CSS stylesheet (via embedded FS) |
| `/messages` | API | JSON messages endpoint (Assignment 3) |
| `/health` | API | JSON health check (Assignment 3) |

### Technical Implementation

- **Embedded Files**: Uses `embed.FS` to bundle `html/*` at compile time
- **Template Parsing**: `template.ParseFS` reads templates from embedded filesystem
- **Static Serving**: `http.StripPrefix` with `http.FileServer` for CSS/assets
- **Data Binding**: Template execution with structured `MessagesPageData`
- **Logging**: Structured logging with trace IDs for all web requests

## Architecture Principles

- **Simplicity First**: Avoid unnecessary abstractions and complex patterns
- **Single Responsibility**: Each function has a clear, focused purpose
- **No Premature Abstraction**: Keep things simple until complexity is actually needed
- **Stateless Design**: HTTP handlers and utilities are stateless functions
- **Explicit Configuration**: All startup configuration happens in main() function