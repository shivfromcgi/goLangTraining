package storage

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"cgi.com/goLangTraining/src/pkg/types"
)

// MessageStorage handles persistent storage of messages
type MessageStorage struct {
	filename string
}

// NewMessageStorage creates a new message storage instance
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

	timestamp := time.Now().Format("2006-01-02 15:04:05")
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
		if line != "" {
			// Parse format: [timestamp] user: message
			message := ms.parseMessageLine(line, id, traceID)
			if message != nil {
				messages = append(messages, *message)
				id++
			}
		}
	}

	return messages, scanner.Err()
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

	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		// If timestamp parsing fails, use current time
		timestamp = time.Now()
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