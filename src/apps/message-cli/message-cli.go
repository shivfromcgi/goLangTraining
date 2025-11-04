package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"cgi.com/goLangTraining/src/pkg/version"
	"github.com/gorilla/websocket"
)

const (
	defaultServerURL = "ws://localhost:8080/ws"
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
		"version", version.Version,
	)
	slog.SetDefault(logger)

	slog.Info("Starting CGI Message CLI WebSocket Client",
		"service", "message-cli",
		"version", version.Version)

	if err := connectToWebSocket(*serverURL); err != nil {
		slog.Error("WebSocket client failed", "error", err)
		os.Exit(1)
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
