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

// writeJSONResponse safely writes JSON response with error handling
func writeJSONResponse(w http.ResponseWriter, ctx context.Context, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.ErrorContext(ctx, "Failed to encode JSON response", "error", err)
		// At this point, headers are already sent, so we can't change the status code
		// But we can log the error for monitoring
	}
}

// MessagesHandler handles message-related requests
func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createMessage(w, r)
	case http.MethodGet:
		getMessages(w, r)
	default:
		// Inline error response
		writeJSONResponse(w, r.Context(), http.StatusMethodNotAllowed, types.EmptyResponse{
			Success: false,
			Error:   "Method not allowed",
			TraceID: middleware.GetTraceID(r.Context()),
		})
	}
}

// createMessage handles POST requests to create a new message
func createMessage(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	var req types.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err)
		// Inline error response
		writeJSONResponse(w, r.Context(), http.StatusBadRequest, types.EmptyResponse{
			Success: false,
			Error:   "Invalid request body",
			TraceID: traceID,
		})
		return
	}

	// Validate input - return early following guidelines
	if req.User == "" || req.Message == "" {
		// Inline error response
		writeJSONResponse(w, r.Context(), http.StatusBadRequest, types.EmptyResponse{
			Success: false,
			Error:   "User and message are required",
			TraceID: traceID,
		})
		return
	}

	// Validate input lengths to prevent excessively large inputs
	if len(req.User) > maxUserNameLength {
		writeJSONResponse(w, r.Context(), http.StatusBadRequest, types.EmptyResponse{
			Success: false,
			Error:   fmt.Sprintf("User name cannot exceed %d characters", maxUserNameLength),
			TraceID: traceID,
		})
		return
	}

	if len(req.Message) > maxMessageLength {
		writeJSONResponse(w, r.Context(), http.StatusBadRequest, types.EmptyResponse{
			Success: false,
			Error:   fmt.Sprintf("Message cannot exceed %d characters", maxMessageLength),
			TraceID: traceID,
		})
		return
	}

	// Save message to storage
	err := storage.AddMessage(r.Context(), req.User, req.Message)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to save message", "error", err, "user", req.User)
		// Inline error response
		writeJSONResponse(w, r.Context(), http.StatusInternalServerError, types.EmptyResponse{
			Success: false,
			Error:   "Failed to save message",
			TraceID: traceID,
		})
		return
	}
	// Broadcast the new message to all connected WebSocket clients
	messageText := fmt.Sprintf("[%s] %s: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		req.User,
		req.Message)
	hub.Broadcast([]byte(messageText))

	slog.InfoContext(r.Context(), "Message created and broadcasted successfully", "user", req.User)

	// Inline success response
	writeJSONResponse(w, r.Context(), http.StatusCreated, types.EmptyResponse{
		Success: true,
		TraceID: traceID,
	})
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
				writeJSONResponse(w, r.Context(), http.StatusBadRequest, types.EmptyResponse{
					Success: false,
					Error:   fmt.Sprintf("limit parameter cannot exceed %d", maxLimitValue),
					TraceID: traceID,
				})
				return
			}
			limit = parsedLimit
		} else {
			writeJSONResponse(w, r.Context(), http.StatusBadRequest, types.EmptyResponse{
				Success: false,
				Error:   "invalid limit parameter",
				TraceID: traceID,
			})
			return
		}
	}

	// Fetch messages from storage
	messages, err := storage.GetLastMessages(r.Context(), limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to retrieve messages", "error", err, "limit", limit)
		// Inline error response
		writeJSONResponse(w, r.Context(), http.StatusInternalServerError, types.EmptyResponse{
			Success: false,
			Error:   "Failed to retrieve messages",
			TraceID: traceID,
		})
		return
	}

	slog.InfoContext(r.Context(), "Messages retrieved successfully",
		"count", len(messages),
		"limit", limit)

	// Inline success response
	writeJSONResponse(w, r.Context(), http.StatusOK, types.MessageResponse{
		Success: true,
		Data:    messages,
		TraceID: traceID,
	})
}

// HealthHandler returns health check information
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID := middleware.GetTraceID(r.Context())

	health := types.HealthStatus{
		Status:    "OK",
		Timestamp: time.Now().UTC(),
		Version:   version.GetVersion(),
	}

	// Inline success response
	writeJSONResponse(w, r.Context(), http.StatusOK, types.HealthResponse{
		Success: true,
		Data:    health,
		TraceID: traceID,
	})
}
