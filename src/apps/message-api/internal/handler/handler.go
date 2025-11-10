package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"cgi.com/goLangTraining/src/pkg/hub"
	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"cgi.com/goLangTraining/src/pkg/types"
	"cgi.com/goLangTraining/src/pkg/version"
)

const (
	// Configuration constants for input validation
	maxUserNameLength = 100
	maxMessageLength  = 1000
	maxLimitValue     = 1000
	defaultLimit      = 10
)

// writeInternalError writes a 500 error with no body (no internal state exposed)
// and adds trace ID to response headers for debugging
func writeInternalError(w http.ResponseWriter, ctx context.Context) {
	traceID := middleware.GetTraceID(ctx)
	w.Header().Set("X-Trace-ID", traceID)
	w.WriteHeader(http.StatusInternalServerError)
	// No body - don't expose internal state
}

// MessagesHandler handles message-related requests
func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	switch r.Method {
	case http.MethodPost:
		createMessage(w, r)
	case http.MethodGet:
		getMessages(w, r)
	default:
		// 405 Method Not Allowed - simple error message is acceptable
		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}

// createMessage handles POST requests to create a new message
func createMessage(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	var req types.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err, "traceID", traceID)
		// 400 Bad Request - simple message acceptable for client errors
		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	// Validate input - return early following guidelines
	if req.User == "" || req.Message == "" {
		slog.InfoContext(r.Context(), "Validation failed: empty user or message", "traceID", traceID)
		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user and message required"})
		return
	}

	// Validate input lengths to prevent excessively large inputs
	if len(req.User) > maxUserNameLength {
		slog.InfoContext(r.Context(), "Validation failed: user name too long", "length", len(req.User), "traceID", traceID)
		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "user name too long"})
		return
	}

	if len(req.Message) > maxMessageLength {
		slog.InfoContext(r.Context(), "Validation failed: message too long", "length", len(req.Message), "traceID", traceID)
		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "message too long"})
		return
	}

	// Save message to storage
	err := storage.AddMessage(r.Context(), req.User, req.Message)
	if err != nil {
		// Internal error - log details but return 500 with no body
		slog.ErrorContext(r.Context(), "Failed to save message", "error", err, "user", req.User, "traceID", traceID)
		writeInternalError(w, r.Context())
		return
	}

	// Broadcast the new message to all connected WebSocket clients
	messageText := fmt.Sprintf("[%s] %s: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		req.User,
		req.Message)
	hub.Broadcast([]byte(messageText))

	slog.InfoContext(r.Context(), "Message created and broadcasted successfully", "user", req.User, "traceID", traceID)

	// Success response
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// getMessages handles GET requests to retrieve messages
func getMessages(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	// Parse limit parameter with validation
	limitStr := r.URL.Query().Get("limit")
	limit := defaultLimit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			// Enforce maximum limit to prevent resource exhaustion
			if parsedLimit > maxLimitValue {
				slog.InfoContext(r.Context(), "Validation failed: limit too large", "limit", parsedLimit, "traceID", traceID)
				w.Header().Set("X-Trace-ID", traceID)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "limit too large"})
				return
			}
			limit = parsedLimit
		} else {
			slog.InfoContext(r.Context(), "Validation failed: invalid limit", "limitStr", limitStr, "traceID", traceID)
			w.Header().Set("X-Trace-ID", traceID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid limit"})
			return
		}
	}

	// Fetch messages from storage
	messages, err := storage.GetLastMessages(r.Context(), limit)
	if err != nil {
		// Internal error - log details but return 500 with no body
		slog.ErrorContext(r.Context(), "Failed to retrieve messages", "error", err, "limit", limit, "traceID", traceID)
		writeInternalError(w, r.Context())
		return
	}

	slog.InfoContext(r.Context(), "Messages retrieved successfully",
		"count", len(messages),
		"limit", limit,
		"traceID", traceID)

	// Success response with data
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    messages,
	})
}

// HealthHandler returns health check information
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	health := map[string]interface{}{
		"status":    "OK",
		"timestamp": time.Now().UTC(),
		"version":   version.GetVersion(),
	}

	// Success response
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    health,
	})
}
