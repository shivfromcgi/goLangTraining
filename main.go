package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cgi.com/goLangTraining/internal/handlers"
	"cgi.com/goLangTraining/internal/service"
)

// Embedded file systems for web interface
//
//go:embed html/*
var htmlFiles embed.FS

const (
	gracefulShutdownTimeout = 30 * time.Second
	defaultAPIVersion       = "1.0.0"
	defaultPort             = 8080
	messagesFileName        = "messages.txt"
)

func main() {
	// Initialize structured logging first
	setupLogging()

	slog.Info("Starting CGI Go Training Service",
		"service", "cgi-go-training",
		"version", defaultAPIVersion)

	// Parse command line flags
	var (
		port        = flag.Int("port", defaultPort, "Port for HTTP server")
		user        = flag.String("user", "", "User for CLI message operations")
		message     = flag.String("message", "", "Message for CLI operations")
		clear       = flag.Bool("clear", false, "Clear all messages")
		file        = flag.String("file", "example.txt", "File path for storage operations")
		data        = flag.String("data", "", "Data to save to file")
		cliMode     = flag.Bool("cli", false, "Run in CLI mode (no web server)")
		storageDemo = flag.Bool("storage-demo", false, "Run storage demonstration")
	)
	flag.Parse()

	// Initialize services
	storage := service.NewMessageStorage(messagesFileName)
	messageActor := service.NewMessageActor(storage)
	messageActor.Start()
	defer messageActor.Stop()

	// If CLI mode is requested, handle CLI operations and exit
	if *cliMode {
		handleCLIOperations(*user, *message, *clear, *file, *data, *storageDemo, storage)
		return
	}

	// Default behavior: start the web application
	startWebApplication(*port, messageActor)
}

// setupLogging configures the default slog logger with structured JSON output
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "cgi-go-training",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)
}

// startWebApplication starts the main web application with all features
// Fixed: Apply middleware at server level instead of individual endpoints
func startWebApplication(port int, messageActor *service.MessageActor) {
	// Create handlers
	apiHandlers := handlers.NewAPIHandlers(messageActor)
	webHandlers := handlers.NewWebHandlers(messageActor, htmlFiles)

	// Set up routes
	mux := http.NewServeMux()

	// Web interface routes
	mux.HandleFunc("/", webHandlers.IndexHandler)
	mux.HandleFunc("/web/messages", webHandlers.MessagesHandler)

	// API v1 routes
	mux.HandleFunc("/api/messages", apiHandlers.MessagesHandler)
	mux.HandleFunc("/api/health", apiHandlers.HealthHandler)

	// Legacy API routes (for backward compatibility)
	mux.HandleFunc("/messages", apiHandlers.MessagesHandler)
	mux.HandleFunc("/health", apiHandlers.HealthHandler)

	// Apply middleware to all routes at server level (addressing reviewer feedback)
	handler := handlers.TraceMiddleware(mux)

	// Create and start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("Server starting",
			"port", port,
			"address", fmt.Sprintf("http://localhost:%d", port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}

// handleCLIOperations processes command-line operations and exits
func handleCLIOperations(user, message string, clear bool, file, data string, storageDemo bool, storage *service.MessageStorage) {
	fmt.Println("=== CGI Go Training Service - CLI Mode ===")

	// Handle storage demo (Assignment 2 functionality)
	if storageDemo {
		runStorageDemo(file, data)
		return
	}

	// Clear messages if requested
	if clear {
		if err := storage.ClearMessages(); err != nil {
			fmt.Printf("❌ Error clearing messages: %v\n", err)
		} else {
			fmt.Println("✅ All messages cleared")
		}
		return
	}

	// Add new message if provided
	if user != "" && message != "" {
		if err := storage.AddMessage(user, message); err != nil {
			fmt.Printf("❌ Error adding message: %v\n", err)
		} else {
			fmt.Printf("✅ Message added: %s: %s\n", user, message)
		}
	}

	// Show last messages
	messages, err := storage.GetLastMessages("cli", 10)
	if err != nil {
		fmt.Printf("❌ Error reading messages: %v\n", err)
		return
	}

	if len(messages) == 0 {
		fmt.Println("📭 No messages found.")
		return
	}

	fmt.Printf("\n📨 Last %d Messages:\n", len(messages))
	for _, msg := range messages {
		fmt.Printf("  [%s] %s: %s\n",
			msg.Timestamp.Format("2006-01-02 15:04:05"),
			msg.User,
			msg.Message)
	}
}

// runStorageDemo demonstrates storage package functionality
func runStorageDemo(file, data string) {
	fmt.Printf("=== Storage Package Demo ===\n")
	fmt.Printf("File: %s\n", file)

	if data != "" {
		fmt.Printf("Data to save: %s\n", data)
		// Use storage package functionality - this would need to be implemented
		// based on your storage package requirements
		fmt.Println("✅ Storage demo completed (placeholder)")
	} else {
		fmt.Println("ℹ️ No data provided to save")
	}
}
