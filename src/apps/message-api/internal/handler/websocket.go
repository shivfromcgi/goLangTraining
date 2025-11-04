package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// NewWebSocketHandler returns a WebSocket handler function
func NewWebSocketHandler(messageStorage *storage.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := middleware.GetTraceID(r.Context())

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

		// Send each message
		for _, msg := range messages {
			messageText := fmt.Sprintf("[%s] %s: %s",
				msg.Timestamp.Format("2006-01-02 15:04:05"),
				msg.User,
				msg.Message)

			err := conn.WriteMessage(websocket.TextMessage, []byte(messageText))
			if err != nil {
				slog.ErrorContext(r.Context(), "Failed to write WebSocket message", "error", err)
				break
			}
		}

		slog.InfoContext(r.Context(), "WebSocket messages sent, closing connection", "count", len(messages))
	}
}
