/*
Advanced functionality - Use the "log/slog" structured logging package to log errors and when data is saved to disk

  - Use the "context" package to set a TraceID at startup to enable traceability of calls through the solution by adding it
    to all logs - Separate the core file handling logic into a different package/module to main/CLI code
  - Write unit tests to cover usefully testable code -
    Use the "os/signal" package and ensure that the application only exits when it receives the interrupt signal (ctrl+c)
*/
package main

import (
	"assignmentTwo/storage"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type ctxKey string

func main() {
	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Create a context with a TraceID
	traceID := uuid.New().String()
	ctx := context.WithValue(context.Background(), ctxKey("traceID"), traceID)

	slog.InfoContext(ctx, "Application starting", "traceID", traceID)

	// Create a channel to listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Example usage of file handler
	filePath := "example.txt"
	data := "Hello, CGI Go Academy!" // Append current time to the data
	data += "\nTimestamp: " + time.Now().Format(time.RFC3339)

	err := storage.SaveData(ctx, filePath, data)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save data", "error", err, "filePath", filePath, "traceID", traceID)
	} else {
		slog.InfoContext(ctx, "Data saved successfully", "filePath", filePath, "traceID", traceID)

		// Demonstrate reading the data back
		readData, err := storage.ReadData(ctx, filePath)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to read data", "error", err, "filePath", filePath, "traceID", traceID)
		} else {
			slog.InfoContext(ctx, "Data read successfully", "filePath", filePath, "traceID", traceID, "preview", readData[:min(len(readData), 50)])
		}
	}

	// Block until a signal is received
	sig := <-sigChan
	slog.InfoContext(ctx, "Received signal, shutting down", "signal", sig.String())

	slog.InfoContext(ctx, "Application gracefully stopped")
}
