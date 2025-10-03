package storage

import (
	"context"
	"log/slog"
	"os"
)

// SaveData writes the given data to the specified file path.
// It creates the file if it doesn't exist, or overwrites it if it does.
func SaveData(ctx context.Context, filePath string, data string) error {
	traceID, _ := ctx.Value("traceID").(string)
	slog.InfoContext(ctx, "Attempting to save data",
		"filePath", filePath,
		"traceID", traceID,
		"dataLength", len(data))

	err := os.WriteFile(filePath, []byte(data), 0644)
	if err != nil {
		slog.ErrorContext(ctx, "Error writing to file",
			"error", err,
			"filePath", filePath,
			"traceID", traceID)
		return err
	}

	slog.InfoContext(ctx, "Successfully saved data to file",
		"filePath", filePath,
		"traceID", traceID)
	return nil
}

// ReadData reads and returns the content from the specified file path.
func ReadData(ctx context.Context, filePath string) (string, error) {
	traceID, _ := ctx.Value("traceID").(string)
	slog.InfoContext(ctx, "Attempting to read data",
		"filePath", filePath,
		"traceID", traceID)

	data, err := os.ReadFile(filePath)
	if err != nil {
		slog.ErrorContext(ctx, "Error reading from file",
			"error", err,
			"filePath", filePath,
			"traceID", traceID)
		return "", err
	}

	slog.InfoContext(ctx, "Successfully read data from file",
		"filePath", filePath,
		"traceID", traceID,
		"dataLength", len(data))
	return string(data), nil
}
