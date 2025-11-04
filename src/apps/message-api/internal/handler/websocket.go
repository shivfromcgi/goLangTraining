package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"github.com/gorilla/websocket"
)

const (
	// Configuration constants
	maxConnections       = 100
	readBufferSize       = 1024
	writeBufferSize      = 1024
	writeDeadlineSeconds = 10
	closeDeadlineSeconds = 5
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow same-origin requests and localhost for development
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true // Allow requests without Origin header (e.g., non-browser clients)
			}

			// Allow localhost for development and testing
			allowedOrigins := []string{
				"http://localhost:8080",
				"http://localhost:8090",
				"http://127.0.0.1:8080",
				"http://127.0.0.1:8090",
			}

			for _, allowed := range allowedOrigins {
				if origin == allowed {
					return true
				}
			}

			// For production, you should validate against your actual domain
			return false
		},
		// Set read and write buffer sizes to configurable limits
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
	}

	// connectionCounter safely tracks active WebSocket connections
	connectionCounter = struct {
		mu    sync.Mutex
		count int
	}{}
)

// NewWebSocketHandler returns a WebSocket handler function
func NewWebSocketHandler(messageStorage *storage.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := middleware.GetTraceID(r.Context())

		// Check connection limit
		connectionCounter.mu.Lock()
		if connectionCounter.count >= maxConnections {
			connectionCounter.mu.Unlock()
			slog.WarnContext(r.Context(), "WebSocket connection limit reached", "limit", maxConnections)
			http.Error(w, "Too many connections", http.StatusTooManyRequests)
			return
		}
		connectionCounter.count++
		connectionCounter.mu.Unlock()

		// Ensure we decrement the counter when done
		defer func() {
			connectionCounter.mu.Lock()
			connectionCounter.count--
			connectionCounter.mu.Unlock()
		}()

		slog.InfoContext(r.Context(), "WebSocket connection attempt")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to upgrade to WebSocket", "error", err)
			return
		}
		defer conn.Close()

		slog.InfoContext(r.Context(), "WebSocket connection established")

		// Send last 10 messages and close
		messages, err := messageStorage.GetLastMessages(traceID, 10)
		if err != nil {
			slog.ErrorContext(r.Context(), "Failed to get messages for WebSocket", "error", err)
			return
		}

		// Send each message with proper error handling
		for _, msg := range messages {
			messageText := fmt.Sprintf("[%s] %s: %s",
				msg.Timestamp.Format("2006-01-02 15:04:05"),
				msg.User,
				msg.Message)

			// Set write deadline to prevent hanging
			if err := conn.SetWriteDeadline(time.Now().Add(writeDeadlineSeconds * time.Second)); err != nil {
				slog.ErrorContext(r.Context(), "Failed to set write deadline", "error", err)
				break
			}

			err := conn.WriteMessage(websocket.TextMessage, []byte(messageText))
			if err != nil {
				slog.ErrorContext(r.Context(), "Failed to write WebSocket message", "error", err)
				break
			}
		}

		// Send close message to gracefully close the connection
		closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Messages sent")
		if err := conn.SetWriteDeadline(time.Now().Add(closeDeadlineSeconds * time.Second)); err == nil {
			conn.WriteMessage(websocket.CloseMessage, closeMessage)
		}

		slog.InfoContext(r.Context(), "WebSocket messages sent, closing connection", "count", len(messages))
	}
}
