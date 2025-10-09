package service

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cgi.com/goLangTraining/src/apps/go-training-service/internal/handler"
	"cgi.com/goLangTraining/src/apps/go-training-service/internal/types"
)

const (
	gracefulShutdownTimeout = 30 * time.Second
)

// Start initializes and starts the go-training-service application.
// This function centralizes application startup logic including CLI parsing,
// service routing, and graceful shutdown handling.
func Start() error {
	// Parse CLI flags for service configuration
	assignmentFlag := flag.String("assignment", "", "Assignment to run: 'assignment1', 'assignment2', or 'assignment3'")
	userFlag := flag.String("user", "", "User ID for message system (Assignment 1)")
	messageFlag := flag.String("message", "", "Message to append (Assignment 1)")
	clearFlag := flag.Bool("clear", false, "Clear all messages (Assignment 1)")
	filePathFlag := flag.String("file", "example.txt", "File path for storage operations (Assignment 2)")
	dataFlag := flag.String("data", "", "Data to save to file (Assignment 2)")
	portFlag := flag.Int("port", 8080, "Port for HTTP server (Assignment 3)")

	flag.Parse()

	if *assignmentFlag == "" {
		printUsage()
		return fmt.Errorf("no assignment specified")
	}

	assignment := types.Assignment(*assignmentFlag)
	switch assignment {
	case types.AssignmentOne:
		return runAssignment1(*userFlag, *messageFlag, *clearFlag)
	case types.AssignmentTwo:
		return runAssignment2(*filePathFlag, *dataFlag)
	case types.AssignmentThree:
		return runAssignment3(*portFlag)
	default:
		printUsage()
		return fmt.Errorf("unknown assignment: %s", assignment)
	}
}

// printUsage displays comprehensive help information for the CLI application.
func printUsage() {
	fmt.Println("Go Training - Unified Assignments")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  Assignment 1 - Message System:")
	fmt.Println("    go run main.go -assignment=assignment1 -user=<userID> -message=<text>")
	fmt.Println("    go run main.go -assignment=assignment1 -clear")
	fmt.Println("")
	fmt.Println("  Assignment 2 - Advanced Storage:")
	fmt.Println("    go run main.go -assignment=assignment2")
	fmt.Println("    go run main.go -assignment=assignment2 -file=<filepath> -data=<content>")
	fmt.Println("")
	fmt.Println("  Assignment 3 - HTTP JSON API:")
	fmt.Println("    go run main.go -assignment=assignment3")
	fmt.Println("    go run main.go -assignment=assignment3 -port=<port>")
}

func runAssignment1(user, message string, clear bool) error {
	fmt.Println("=== Running Assignment 1: Message System ===")
	// Implementation will be moved here from main.go
	return fmt.Errorf("assignment 1 not yet implemented in new structure")
}

func runAssignment2(filePath, data string) error {
	fmt.Println("=== Running Assignment 2: Advanced Storage System ===")
	// Implementation will be moved here from main.go
	return fmt.Errorf("assignment 2 not yet implemented in new structure")
}

func runAssignment3(port int) error {
	fmt.Println("=== Running Assignment 3: HTTP JSON API ===")

	// Create HTTP server with handler
	mux := http.NewServeMux()
	h := handler.New()

	mux.HandleFunc("/messages", h.TraceMiddleware(h.MessagesHandler))
	mux.HandleFunc("/health", h.TraceMiddleware(h.HealthHandler))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	// Create signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		fmt.Printf("Starting HTTP server on port %d...\n", port)
		fmt.Printf("API Endpoints:\n")
		fmt.Printf("  POST http://localhost:%d/messages - Create a message\n", port)
		fmt.Printf("  GET  http://localhost:%d/messages - Get all messages\n", port)
		fmt.Printf("  GET  http://localhost:%d/health - Health check\n", port)
		fmt.Printf("\nPress Ctrl+C to stop the server...\n\n")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err, "port", port)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	fmt.Printf("\nReceived signal %s, shutting down server...\n", sig.String())

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		slog.Error("Server shutdown failed", "error", err)
		return err
	}

	fmt.Println("Assignment 3 HTTP server gracefully stopped")
	return nil
}
