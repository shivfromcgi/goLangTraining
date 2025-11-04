package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"cgi.com/goLangTraining/src/pkg/types"
)

// NewMessagesHandler returns a handler function for message-related requests
func NewMessagesHandler(messageStorage *storage.MessageStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createMessage(w, r, messageStorage)
		case http.MethodGet:
			getMessages(w, r, messageStorage)
		default:
			// Inline error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(types.EmptyResponse{
				Success: false,
				Error:   "Method not allowed",
				TraceID: middleware.GetTraceID(r.Context()),
			})
		}
	}
}

// createMessage handles POST requests to create a new message
func createMessage(w http.ResponseWriter, r *http.Request, messageStorage *storage.MessageStorage) {
	traceID := middleware.GetTraceID(r.Context())

	var req types.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err)
		// Inline error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.EmptyResponse{
			Success: false,
			Error:   "Invalid request body",
			TraceID: traceID,
		})
		return
	}

	// Validate input - return early following guidelines
	if req.User == "" || req.Message == "" {
		// Inline error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(types.EmptyResponse{
			Success: false,
			Error:   "User and message are required",
			TraceID: traceID,
		})
		return
	}

	// Add message directly to storage
	err := messageStorage.AddMessage(req.User, req.Message)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to create message", "error", err)
		// Inline error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.EmptyResponse{
			Success: false,
			Error:   "Failed to create message",
			TraceID: traceID,
		})
		return
	}

	slog.InfoContext(r.Context(), "Message created successfully", "user", req.User)

	// Inline success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(types.EmptyResponse{
		Success: true,
		TraceID: traceID,
	})
}

// getMessages handles GET requests to retrieve messages
func getMessages(w http.ResponseWriter, r *http.Request, messageStorage *storage.MessageStorage) {
	traceID := middleware.GetTraceID(r.Context())

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get messages directly from storage
	messages, err := messageStorage.GetLastMessages(traceID, limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to retrieve messages", "error", err)
		// Inline error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(types.EmptyResponse{
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(types.MessageResponse{
		Success: true,
		Data:    messages,
		TraceID: traceID,
	})
}

// NewHealthHandler returns a handler function for health check requests
func NewHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := middleware.GetTraceID(r.Context())

		health := types.HealthStatus{
			Status:    "OK",
			Timestamp: time.Now(),
			Version:   "1.0.0",
		}

		// Inline success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(types.HealthResponse{
			Success: true,
			Data:    health,
			TraceID: traceID,
		})
	}
}
