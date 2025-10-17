CGI Go Training Service – Unified Application

A single cohesive Go application integrating all CGI Go Academy assignments into one streamlined project.
This version focuses on simplicity, modularity, and production-grade structure while maintaining easy extensibility.

Project Overview

This repository consolidates multiple Go assignments into a single, maintainable codebase with REST, CLI, and gRPC layers.

Key Highlights

Unified architecture across all assignments

Simplified file storage (no custom types)

Integrated REST + gRPC + Web + CLI modes

Structured logging and graceful shutdown

Tested and build-ready via Makefile

Project Structure
OpenMedia_GoLang_Course/
├── main.go              # Unified entry point
├── go.mod               # Dependencies
├── go.work              # Workspace config
├── Makefile             # Build & run shortcuts
├── src/pkg/storage/     # Simplified file storage
│   ├── storage.go
│   ├── storage_test.go
│   └── types.go
├── html/                # Web templates (Assignment 4)
│   ├── index.html
│   ├── messages.html
│   └── styles.css
├── messages.txt         # Message storage
└── README.md

Setup & Run
Prerequisites

Go 1.22+

Make

Build & Run
make build
make run

Run Web Server
go run main.go -port=8080

Run CLI Mode
go run main.go -cli -user=alice -message='Hello World'

gRPC Implementation

Includes complete gRPC setup with Protocol Buffers.

Quick Start

make build-grpc
make run-grpc-server
make run-grpc-client

Component	Description
proto/	Proto definitions
server/	gRPC server implementation
client/	gRPC client module
Port	:50051

REST API Endpoints
Endpoint	Method	Description
/api/messages	GET / POST	Retrieve or create messages
/api/files	POST	Save file data
/api/health	GET	Health check
/	GET	Web home
/web/messages	GET	Dynamic message view
/static/styles.css	GET	Static CSS file
🧪 Testing

Run all tests:

make test


Example API test:

curl -X POST http://localhost:8080/api/messages \
  -H 'Content-Type: application/json' \
  -d '{"user":"demo","message":"Hello unified app!"}'

Design Principles

Simplicity First: Focus on readable, maintainable code

Single Responsibility: Each function has one purpose

No Premature Abstraction: Use only necessary complexity

Stateless Design: All handlers are independent

Graceful Lifecycle: Clean startup and shutdown

Makefile Commands
Command	Description
make help	List commands
make build	Build binary
make run	Run server
make test	Run tests
make assignment1..4	Run specific modules
make clean	Clean build artifacts


Technologies Used

Language: Go

Frameworks: net/http, gRPC

Storage: Local file system

Templates: html/template, embed.FS

Testing: go test

Build System: Make

License

This project is developed as part of the CGI Go Academy Training Program.
Use, modify, and extend for learning purposes.