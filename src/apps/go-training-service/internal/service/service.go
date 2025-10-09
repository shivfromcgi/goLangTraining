package service

import (
	"bufio"
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
	"cgi.com/goLangTraining/src/pkg/storage"

	"github.com/google/uuid"
)

const (
	gracefulShutdownTimeout = 30 * time.Second
	messagesFileName        = "messages.txt"
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

	// Handle the clear flag first
	if clear {
		err := os.Truncate(messagesFileName, 0)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error clearing file: %w", err)
		}
		fmt.Println("All messages cleared.")
		return nil
	}

	// Validate required flags for message operations
	if user == "" || message == "" {
		return fmt.Errorf("both -user and -message are required for Assignment 1")
	}

	// Append the message to disk
	err := appendMessage(user, message)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	// Retrieve and print the last 10 messages
	fmt.Println("\nLast 10 Messages:")
	err = printLast10Messages()
	if err != nil {
		fmt.Printf("Error reading messages: %v\n", err)
	}

	return nil
}

func runAssignment2(filePath, data string) error {
	fmt.Println("=== Running Assignment 2: Advanced Storage System ===")

	// Create a context with a TraceID for distributed tracing
	traceID := uuid.New().String()
	ctx := context.WithValue(context.Background(), types.CtxKey("traceID"), traceID)

	slog.InfoContext(ctx, "Assignment 2 starting", "traceID", traceID)

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Determine data content
	var fileContent string
	if data != "" {
		fileContent = data
	} else {
		fileContent = "Hello, CGI Go Academy! - Assignment 2"
	}
	fileContent += "\nTimestamp: " + time.Now().Format(time.RFC3339)

	// Save data using the storage package
	err := storage.SaveData(ctx, filePath, fileContent)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save data", "error", err, "filePath", filePath, "traceID", traceID)
		return fmt.Errorf("failed to save data: %w", err)
	}

	slog.InfoContext(ctx, "Data saved successfully", "filePath", filePath, "traceID", traceID)

	// Demonstrate reading the data back
	readData, err := storage.ReadData(ctx, filePath)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read data", "error", err, "filePath", filePath, "traceID", traceID)
		return fmt.Errorf("failed to read data: %w", err)
	}

	preview := readData
	if len(readData) > 50 {
		preview = readData[:50] + "..."
	}
	slog.InfoContext(ctx, "Data read successfully", "filePath", filePath, "traceID", traceID, "preview", preview)

	// Block until a signal is received
	sig := <-sigChan
	slog.InfoContext(ctx, "Received signal, shutting down", "signal", sig.String())
	slog.InfoContext(ctx, "Assignment 2 gracefully stopped")

	return nil
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

// Assignment 1 helper functions
func appendMessage(user, message string) error {
	f, err := os.OpenFile(messagesFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Format: userID: message
	line := fmt.Sprintf("%s: %s\n", user, message)
	_, err = f.WriteString(line)
	return err
}

func printLast10Messages() error {
	f, err := os.Open(messagesFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No messages found.")
			return nil
		}
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	// Print last 10
	start := 0
	if len(lines) > 10 {
		start = len(lines) - 10
	}
	for _, line := range lines[start:] {
		fmt.Println(line)
	}
	return nil
}
