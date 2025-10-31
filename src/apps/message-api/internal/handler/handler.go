package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"cgi.com/goLangTraining/src/pkg/storage"
	"cgi.com/goLangTraining/src/pkg/types"
	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const traceIDKey contextKey = "traceID"

// APIHandler contains dependencies for HTTP handlers
type APIHandler struct {
	storage *storage.MessageStorage
}

// NewAPIHandler creates a new APIHandler instance
func NewAPIHandler(storage *storage.MessageStorage) *APIHandler {
	return &APIHandler{
		storage: storage,
	}
}

// MessagesHandler handles message-related requests
func (h *APIHandler) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createMessage(w, r)
	case http.MethodGet:
		h.getMessages(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", getTraceID(r))
	}
}

// createMessage handles POST requests to create a new message
func (h *APIHandler) createMessage(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)

	var req types.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusBadRequest, "Invalid request body", traceID)
		return
	}

	// Validate input - return early following guidelines
	if req.User == "" || req.Message == "" {
		respondWithError(w, http.StatusBadRequest, "User and message are required", traceID)
		return
	}

	// Add message directly to storage
	err := h.storage.AddMessage(req.User, req.Message)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to create message", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusInternalServerError, "Failed to create message", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Message created successfully",
		"user", req.User,
		"traceID", traceID)

	respondWithJSON(w, http.StatusCreated, types.EmptyResponse{
		Success: true,
		TraceID: traceID,
	})
}

// getMessages handles GET requests to retrieve messages
func (h *APIHandler) getMessages(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default value
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get messages directly from storage
	messages, err := h.storage.GetLastMessages(traceID, limit)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to retrieve messages", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve messages", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Messages retrieved successfully",
		"count", len(messages),
		"limit", limit,
		"traceID", traceID)

	respondWithJSON(w, http.StatusOK, types.MessageResponse{
		Success: true,
		Data:    messages,
		TraceID: traceID,
	})
}

// HealthHandler handles health check requests
func (h *APIHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)

	health := types.HealthStatus{
		Status:    "OK",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	respondWithJSON(w, http.StatusOK, types.HealthResponse{
		Success: true,
		Data:    health,
		TraceID: traceID,
	})
}

// TraceMiddleware adds trace ID to request context following guidelines
func TraceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)

		slog.InfoContext(ctx, "Incoming HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"traceID", traceID)

		w.Header().Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// getTraceID extracts trace ID from request context
func getTraceID(r *http.Request) string {
	if traceID, ok := r.Context().Value(traceIDKey).(string); ok {
		return traceID
	}
	return uuid.New().String()
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, status int, message string, traceID string) {
	respondWithJSON(w, status, types.EmptyResponse{
		Success: false,
		Error:   message,
		TraceID: traceID,
	})
}
