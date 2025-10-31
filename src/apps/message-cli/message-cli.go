package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

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
		user        = flag.String("user", "", "User for CLI message operations")
		message     = flag.String("message", "", "Message for CLI operations")
		clear       = flag.Bool("clear", false, "Clear all messages")
		file        = flag.String("file", "example.txt", "File path for storage operations")
		data        = flag.String("data", "", "Data to save to file")
		storageDemo = flag.Bool("storage-demo", false, "Run storage demonstration")
	)
	flag.Parse()

	// Initialize storage in main following guidelines
	messageStorage := storage.NewMessageStorage(messagesFileName)

	// Handle CLI operations and exit
	handleCLIOperations(*user, *message, *clear, *file, *data, *storageDemo, messageStorage)
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
func handleCLIOperations(user, message string, clear bool, file, data string, storageDemo bool, storage *storage.MessageStorage) {
	fmt.Println("=== CGI Message CLI Service ===")

	// Handle storage demo (Assignment 2 functionality)
	if storageDemo {
		runStorageDemo(file, data)
		return
	}

	// Clear messages if requested - return early following guidelines
	if clear {
		err := storage.ClearMessages()
		if err != nil {
			fmt.Printf("❌ Error clearing messages: %v\n", err)
			return
		}
		fmt.Println("✅ All messages cleared")
		return
	}

	// Add new message if provided
	if user != "" && message != "" {
		err := storage.AddMessage(user, message)
		if err != nil {
			fmt.Printf("❌ Error adding message: %v\n", err)
			return
		}
		fmt.Printf("✅ Message added: %s: %s\n", user, message)
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
		fmt.Println("✅ Storage demo completed (placeholder)")
	} else {
		fmt.Println("ℹ️ No data provided to save")
	}
}
