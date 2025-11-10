package storage

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/types"
)

const defaultMessagesFileName = "messages.txt"

// messageLineRegex is compiled once for parsing message lines
// Format: [timestamp] user: message
var messageLineRegex = regexp.MustCompile(`^\[(.*?)\] ([^:]+): (.+)$`)

// Global state for singleton pattern following coding standards
var (
	defaultFilename = defaultMessagesFileName
	fileMutex       sync.RWMutex // Single mutex for all file operations
)

// sanitizeInput removes or replaces potentially dangerous characters
func sanitizeInput(input string) string {
	// Remove control characters and non-printable characters
	var result strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			result.WriteRune(r)
		} else {
			result.WriteString(" ") // Replace with space
		}
	}
	return strings.TrimSpace(result.String())
}

// AddMessageToFile appends a message to the default storage file using static singleton pattern
// This function is used ONLY by the gRPC message store service
func AddMessageToFile(ctx context.Context, user, message string) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	// Sanitize inputs to prevent log injection attacks
	sanitizedUser := sanitizeInput(user)
	sanitizedMessage := sanitizeInput(message)

	// Validate sanitized inputs are not empty
	if sanitizedUser == "" || sanitizedMessage == "" {
		err := fmt.Errorf("invalid input: user and message cannot be empty after sanitization")
		slog.ErrorContext(ctx, "Message validation failed", "error", err, "user", user, "message", message)
		return err
	}

	f, err := os.OpenFile(defaultFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to open message file", "error", err, "filename", defaultFilename)
		return err
	}
	defer f.Close()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s: %s\n", timestamp, sanitizedUser, sanitizedMessage)
	_, err = f.WriteString(line)

	if err != nil {
		slog.ErrorContext(ctx, "Failed to write message to file", "error", err, "user", sanitizedUser)
	} else {
		slog.InfoContext(ctx, "Message saved successfully", "user", sanitizedUser, "messageLength", len(sanitizedMessage))
	}

	return err
}

// ReadMessagesFromFile reads all messages from default storage and returns them as Message structs
// This function is used ONLY by the gRPC message store service
func ReadMessagesFromFile(ctx context.Context) ([]types.Message, error) {
	fileMutex.RLock()
	defer fileMutex.RUnlock()

	traceID := middleware.GetTraceID(ctx)

	f, err := os.Open(defaultFilename)
	if err != nil {
		if os.IsNotExist(err) {
			slog.InfoContext(ctx, "Message file does not exist, returning empty list", "filename", defaultFilename)
			return []types.Message{}, nil
		}
		slog.ErrorContext(ctx, "Failed to open message file for reading", "error", err, "filename", defaultFilename)
		return []types.Message{}, err
	}
	defer f.Close()

	var messages []types.Message
	scanner := bufio.NewScanner(f)
	id := 1

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse format: [timestamp] user: message
		message := parseMessageLine(line, id, traceID)
		if message == nil {
			continue
		}

		messages = append(messages, *message)
		id++
	}

	// Explicitly check for scanner errors
	if err := scanner.Err(); err != nil {
		slog.ErrorContext(ctx, "Scanner error while reading messages", "error", err)
		return []types.Message{}, err
	}

	slog.InfoContext(ctx, "Messages loaded successfully", "count", len(messages))
	return messages, nil
}

// ClearMessages removes all messages from default storage
func ClearMessages(ctx context.Context) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()

	err := os.Truncate(defaultFilename, 0)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to clear messages", "error", err, "filename", defaultFilename)
	} else {
		slog.InfoContext(ctx, "Messages cleared successfully", "filename", defaultFilename)
	}
	return err
}

// parseMessageLine parses a line from the message file into a Message struct
func parseMessageLine(line string, id int, traceID string) *types.Message {
	// Parse format: [timestamp] user: message
	// Example: [2024-01-01 12:34:56] john: Hello world!

	matches := messageLineRegex.FindStringSubmatch(line)

	if len(matches) != 4 {
		return nil // Skip malformed lines
	}

	timestampStr := matches[1]
	user := strings.TrimSpace(matches[2])
	messageText := strings.TrimSpace(matches[3])

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			// If timestamp parsing fails, use current time in UTC
			timestamp = time.Now().UTC()
		}
	}

	return &types.Message{
		ID:        id,
		User:      user,
		Message:   messageText,
		Timestamp: timestamp,
		TraceID:   traceID,
	}
}

// GetLastMessagesFromFile reads the last N messages from the file
// This function is used ONLY by the gRPC message store service
func GetLastMessagesFromFile(ctx context.Context, limit int) ([]types.Message, error) {
	messages, err := ReadMessagesFromFile(ctx)
	if err != nil {
		return nil, err
	}

	// Return last N messages
	if limit > 0 && len(messages) > limit {
		return messages[len(messages)-limit:], nil
	}

	return messages, nil
}
