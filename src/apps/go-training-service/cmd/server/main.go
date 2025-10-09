package main

import (
	"log/slog"
	"os"

	"cgi.com/goLangTraining/src/apps/go-training-service/internal/service"
)

// main serves as the entry point for the go-training-service application.
// This lightweight main function delegates to the service package to maintain
// clean separation between application startup and business logic.
func main() {
	// Initialize structured logging for the entire application
	setupLogging()

	slog.Info("Starting go-training-service",
		"service", "go-training-service",
		"version", "1.0.0")

	// Delegate to service layer for application logic
	// This pattern keeps main.go minimal and testable
	err := service.Start()
	if err != nil {
		slog.Error("Service failed to start", "error", err)
		os.Exit(1)
	}
}

// setupLogging configures the default slog logger with structured JSON output.
// This centralized logging setup ensures consistent log format across all application components.
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "go-training-service",
		"version", "1.0.0",
	)

	slog.SetDefault(logger)
}
