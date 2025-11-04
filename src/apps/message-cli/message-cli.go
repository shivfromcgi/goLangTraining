package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

const (
	defaultAPIVersion = "1.0.0"
	defaultServerURL  = "ws://localhost:8080/ws"
)

func main() {
	// Declare flags at top following guidelines
	var (
		serverURL = flag.String("server", defaultServerURL, "WebSocket server URL")
	)
	flag.Parse()

	// Set up slog.SetDefault following guidelines
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-cli",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)

	slog.Info("Starting CGI Message CLI WebSocket Client",
		"service", "message-cli",
		"version", defaultAPIVersion)

	connectToWebSocket(*serverURL)
}

func connectToWebSocket(serverURL string) {
	fmt.Println("=== CGI Message CLI WebSocket Client ===")
	fmt.Printf("🔌 Connecting to %s\n", serverURL)

	u, err := url.Parse(serverURL)
	if err != nil {
		fmt.Printf("❌ Invalid server URL: %v\n", err)
		os.Exit(1)
	}

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Printf("❌ Failed to connect to WebSocket: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("✅ Connected to WebSocket server")
	fmt.Println("📨 Listening for messages... (Press Ctrl+C to exit)")

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
				fmt.Printf("� Received: %s\n", string(message))
			}
		}
	}()

	// Wait for interrupt signal or connection close
	select {
	case <-done:
		fmt.Println("\n📪 Connection closed by server")
	case <-interrupt:
		fmt.Println("\n🛑 Interrupt received, closing connection...")

		// Cleanly close the connection
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			slog.Error("Error closing WebSocket", "error", err)
			return
		}

		// Wait for the server to close the connection
		select {
		case <-done:
		}
	}

	fmt.Println("✅ CLI client exited cleanly")
}
