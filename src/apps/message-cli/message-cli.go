package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"cgi.com/goLangTraining/src/pkg/types"
	"cgi.com/goLangTraining/src/pkg/version"
	"github.com/gorilla/websocket"
)

const (
	defaultServerURL = "ws://localhost:8080/ws"
	defaultAPIURL    = "http://localhost:8080/api/v1/messages"
)

func main() {
	// Declare flags at top following guidelines
	var (
		serverURL = flag.String("server", defaultServerURL, "WebSocket server URL")
		apiURL    = flag.String("api", defaultAPIURL, "HTTP API URL for posting messages")
		user      = flag.String("user", "", "User name for posting message (required with -message)")
		message   = flag.String("message", "", "Message to post (requires -user)")
		listen    = flag.Bool("listen", false, "Listen for real-time messages via WebSocket")
	)
	flag.Parse()

	// Set up slog.SetDefault following guidelines
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-cli",
		"version", version.GetVersion(),
	)
	slog.SetDefault(logger)

	slog.Info("Starting CGI Message CLI")

	// Determine mode of operation
	if *user != "" && *message != "" {
		// Post message mode
		if err := postMessage(*apiURL, *user, *message); err != nil {
			slog.Error("Failed to post message", "error", err)
			os.Exit(1)
		}

		// If listen flag is also set, continue to WebSocket listening
		if !*listen {
			return
		}
	} else if *user != "" || *message != "" {
		// Only one of user/message provided - error
		fmt.Fprintf(os.Stderr, "Error: Both -user and -message flags are required to post a message\n")
		fmt.Fprintf(os.Stderr, "Usage examples:\n")
		fmt.Fprintf(os.Stderr, "  %s -listen                              # Listen for messages\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -user john -message \"Hello World\"   # Post a message\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -user john -message \"Hello\" -listen # Post and listen\n", os.Args[0])
		os.Exit(1)
	}

	// Default or explicit listen mode
	if *listen || (*user == "" && *message == "") {
		if err := connectToWebSocket(*serverURL); err != nil {
			slog.Error("WebSocket client failed", "error", err)
			os.Exit(1)
		}
	}
}

func connectToWebSocket(serverURL string) error {
	slog.Info("Starting WebSocket client connection", "serverURL", serverURL)

	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL %s: %w", serverURL, err)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket %s: %w", serverURL, err)
	}
	defer conn.Close()

	slog.Info("Connected to WebSocket server successfully", "serverURL", serverURL)
	slog.Info("Listening for messages - Press Ctrl+C to exit")

	// Set up signal handling for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Channel to signal when we're done reading
	done := make(chan struct{})

	// Start goroutine to read messages
	go func() {
		defer close(done)
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.Error("WebSocket error", "error", err)
				}
				return
			}

			if messageType == websocket.TextMessage {
				slog.Info("Message received", "message", string(message))
			}
		}
	}()

	// Wait for interrupt signal or connection close
	select {
	case <-done:
		slog.Info("Connection closed by server")
	case <-interrupt:
		slog.Info("Interrupt received, closing connection")

		// Cleanly close the connection
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			slog.Error("Error closing WebSocket", "error", err)
			return fmt.Errorf("error closing WebSocket: %w", err)
		}

		// Wait for the server to close the connection
		<-done
	}

	slog.Info("CLI client exited cleanly")
	return nil
}

// postMessage sends a message to the HTTP API
func postMessage(apiURL, user, message string) error {
	slog.Info("Posting message to API", "apiURL", apiURL, "user", user)

	// Validate input
	if strings.TrimSpace(user) == "" || strings.TrimSpace(message) == "" {
		return fmt.Errorf("user and message cannot be empty")
	}

	// Create request payload
	payload := types.CreateMessageRequest{
		User:    user,
		Message: message,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	slog.Info("Message posted successfully", "user", user, "status", resp.StatusCode)
	return nil
}
