package storage

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"cgi.com/goLangTraining/src/pkg/types"
)

const defaultMessagesFileName = "messages.txt"

// MessageStorage handles persistent storage of messages
type MessageStorage struct {
	filename string
}

var (
	defaultStorage *MessageStorage
	once           sync.Once
)

// GetDefaultStorage returns the default message storage instance (singleton)
func GetDefaultStorage() *MessageStorage {
	once.Do(func() {
		defaultStorage = &MessageStorage{
			filename: defaultMessagesFileName,
		}
	})
	return defaultStorage
}

// NewMessageStorage creates a new message storage instance with custom filename
// This is kept for backward compatibility and testing purposes
func NewMessageStorage(filename string) *MessageStorage {
	return &MessageStorage{
		filename: filename,
	}
}

// AddMessage appends a message to the storage file
func (ms *MessageStorage) AddMessage(user, message string) error {
	f, err := os.OpenFile(ms.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().UTC().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] %s: %s\n", timestamp, user, message)
	_, err = f.WriteString(line)
	return err
}

// ReadMessages reads all messages from storage and returns them as Message structs
func (ms *MessageStorage) ReadMessages(traceID string) ([]types.Message, error) {
	f, err := os.Open(ms.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return []types.Message{}, nil
		}
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
		message := ms.parseMessageLine(line, id, traceID)
		if message == nil {
			continue
		}

		messages = append(messages, *message)
		id++
	}

	// Explicitly check for scanner errors
	if err := scanner.Err(); err != nil {
		return []types.Message{}, err
	}

	return messages, nil
}

// ClearMessages removes all messages from storage
func (ms *MessageStorage) ClearMessages() error {
	return os.Truncate(ms.filename, 0)
}

// parseMessageLine parses a line from the message file into a Message struct
func (ms *MessageStorage) parseMessageLine(line string, id int, traceID string) *types.Message {
	// Parse format: [timestamp] user: message
	// Example: [2024-01-01 12:34:56] john: Hello world!

	re := regexp.MustCompile(`^\[(.*?)\] ([^:]+): (.+)$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 4 {
		return nil // Skip malformed lines
	}

	timestampStr := matches[1]
	user := strings.TrimSpace(matches[2])
	messageText := strings.TrimSpace(matches[3])

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		// Try parsing old format for backward compatibility
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

// GetLastMessages returns the last N messages
func (ms *MessageStorage) GetLastMessages(traceID string, limit int) ([]types.Message, error) {
	allMessages, err := ms.ReadMessages(traceID)
	if err != nil {
		return []types.Message{}, err
	}

	if len(allMessages) <= limit {
		return allMessages, nil
	}

	startIndex := len(allMessages) - limit
	return allMessages[startIndex:], nil
}
