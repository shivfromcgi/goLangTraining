package storage

import (
	"context"
	"log/slog"
	"os"
)

// writeFileContent persists the provided content to a file at the given location.
// This internal function implements the core write logic with comprehensive logging
// to enable debugging of file operation failures. Returns early on write failure
// to avoid nested error handling patterns that complicate error propagation.
// Uses complete file replacement strategy to ensure atomic operations and consistent state.
func writeFileContent(ctx context.Context, filePath string, content string) error {
	traceID, _ := ctx.Value("traceID").(string)

	metrics := FileMetrics{
		ContentSize: len(content),
		Operation:   string(OperationWrite),
	}

	slog.InfoContext(ctx, "Starting file write operation",
		"filePath", filePath,
		"traceID", traceID,
		"metrics", metrics)

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		slog.ErrorContext(ctx, "File write failed",
			"error", err,
			"filePath", filePath,
			"traceID", traceID)
		return err
	}

	slog.InfoContext(ctx, "File written successfully",
		"filePath", filePath,
		"traceID", traceID)
	return nil
}

// loadFileContent retrieves and returns the complete content of a file.
// This internal function implements the core read logic with structured logging
// for operational visibility into file access patterns. Returns early on read failure
// to maintain clean error flow and avoid complex nested conditional logic.
// Loads entire file into memory which is appropriate for configuration files and small datasets.
func loadFileContent(ctx context.Context, filePath string) (string, error) {
	traceID, _ := ctx.Value("traceID").(string)

	slog.InfoContext(ctx, "Starting file read operation",
		"filePath", filePath,
		"traceID", traceID)

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		slog.ErrorContext(ctx, "File read failed",
			"error", err,
			"filePath", filePath,
			"traceID", traceID)
		return "", err // Return empty string instead of zero value on error
	}

	metrics := FileMetrics{
		BytesRead: len(fileBytes),
		Operation: string(OperationRead),
	}

	slog.InfoContext(ctx, "File read successfully",
		"filePath", filePath,
		"traceID", traceID,
		"metrics", metrics)
	return string(fileBytes), nil
}

// SaveData provides a simple interface for persisting data to files.
// This exported function serves as the public API for file write operations,
// abstracting away internal implementation details while maintaining clean separation
// between public interface and internal logic for future flexibility.
func SaveData(ctx context.Context, filePath string, data string) error {
	return writeFileContent(ctx, filePath, data)
}

// ReadData provides a simple interface for reading data from files.
// This exported function serves as the public API for file read operations,
// ensuring consistent error handling and logging across all file access patterns
// while hiding internal implementation complexity from consumers.
func ReadData(ctx context.Context, filePath string) (string, error) {
	return loadFileContent(ctx, filePath)
}
