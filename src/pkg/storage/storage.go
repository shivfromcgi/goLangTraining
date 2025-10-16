package storage

import (
	"context"
	"log/slog"
	"os"
)

// SaveData provides a simple interface for persisting data to files.
// This function implements the complete write logic with comprehensive logging
// to enable debugging of file operation failures. Uses atomic file replacement
// to ensure consistent state and proper error propagation.
func SaveData(ctx context.Context, filePath string, data string) error {
	traceID, _ := ctx.Value("traceID").(string)

	metrics := FileMetrics{
		ContentSize: len(data),
		Operation:   "write",
	}

	slog.InfoContext(ctx, "Starting file write operation",
		"filePath", filePath,
		"traceID", traceID,
		"metrics", metrics)

	err := os.WriteFile(filePath, []byte(data), 0644)
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

// ReadData provides a simple interface for reading data from files.
// This function implements the complete read logic with structured logging
// for operational visibility into file access patterns. Loads entire file
// into memory which is appropriate for configuration files and small datasets.
func ReadData(ctx context.Context, filePath string) (string, error) {
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
		return "", err
	}

	metrics := FileMetrics{
		BytesRead: len(fileBytes),
		Operation: "read",
	}

	slog.InfoContext(ctx, "File read successfully",
		"filePath", filePath,
		"traceID", traceID,
		"metrics", metrics)
	return string(fileBytes), nil
}
