package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cgi.com/goLangTraining/src/pkg/storage"
)

const (
	defaultAPIVersion = "1.0.0"
	messagesFileName  = "messages.txt"
)

func main() {
	setupLogging()

	slog.Info("Starting CGI Message CLI Service",
		"service", "message-cli",
		"version", defaultAPIVersion)

	// Parse command line flags following guidelines - declare at top
	var (
		user    = flag.String("user", "", "User for CLI message operations")
		message = flag.String("message", "", "Message for CLI operations")
		clear   = flag.Bool("clear", false, "Clear all messages")
	)
	flag.Parse()

	// Initialize storage in main following guidelines
	messageStorage := storage.NewMessageStorage(messagesFileName)

	// Handle CLI operations and exit
	handleCLIOperations(*user, *message, *clear, messageStorage)
}

// setupLogging configures the default slog logger following guidelines
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-cli",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)
}

// handleCLIOperations processes command-line operations following guidelines
func handleCLIOperations(user, message string, clear bool, storage *storage.MessageStorage) {
	fmt.Println("=== CGI Message CLI Service ===")

	// Add new message first if provided
	if user != "" && message != "" {
		err := storage.AddMessage(user, message)
		if err != nil {
			fmt.Printf("❌ Error adding message: %v\n", err)
			return
		}
		fmt.Printf("✅ Message added: %s: %s\n", user, message)
	}

	// Clear messages if requested (after adding new message)
	if clear {
		err := storage.ClearMessages()
		if err != nil {
			fmt.Printf("❌ Error clearing messages: %v\n", err)
			return
		}
		fmt.Println("✅ All messages cleared")

		// If we only cleared messages, don't show the message list
		if user == "" || message == "" {
			return
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
			msg.Timestamp.UTC().Format(time.RFC3339),
			msg.User,
			msg.Message)
	}
}
