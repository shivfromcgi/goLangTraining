package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"cgi.com/goLangTraining/src/apps/go-training-service/internal/types"

	"log/slog"

	"github.com/google/uuid"
)

// Handler contains HTTP handlers and middleware for the training service.
// This structure encapsulates request handling logic and promotes testability
// through dependency injection patterns.
type Handler struct{}

// New creates a new Handler instance.
// This constructor pattern allows for future dependency injection
// and maintains consistency with Go service patterns.
func New() *Handler {
	return &Handler{}
}

// TraceMiddleware adds TraceID to the context and logs requests.
// This middleware ensures all requests have distributed tracing support
// and provides consistent request logging across all endpoints.
func (h *Handler) TraceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), types.CtxKey("traceID"), traceID)

		slog.InfoContext(ctx, "Incoming HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"traceID", traceID)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Trace-ID", traceID)

		next(w, r.WithContext(ctx))
	}
}

// MessagesHandler handles both GET and POST /messages requests.
// This unified handler follows REST conventions while maintaining
// clear separation between read and write operations.
func (h *Handler) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value(types.CtxKey("traceID")).(string)

	switch r.Method {
	case http.MethodPost:
		h.createMessage(w, r, traceID)
	case http.MethodGet:
		h.getMessages(w, r, traceID)
	default:
		h.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", traceID)
	}
}

// HealthHandler handles GET /health requests.
// This handler provides structured health information for monitoring systems
// and load balancers to determine service availability.
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value(types.CtxKey("traceID")).(string)

	health := types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	h.respondWithSuccess(w, http.StatusOK, health, traceID)
}

// createMessage handles POST /messages requests.
func (h *Handler) createMessage(w http.ResponseWriter, r *http.Request, traceID string) {
	var req types.CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err, "traceID", traceID)
		h.respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", traceID)
		return
	}

	if req.User == "" || req.Message == "" {
		h.respondWithError(w, http.StatusBadRequest, "User and message are required", traceID)
		return
	}

	message := types.Message{
		ID:        int(time.Now().UnixNano() / 1000000),
		User:      req.User,
		Message:   req.Message,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}

	slog.InfoContext(r.Context(), "Message created successfully",
		"user", req.User,
		"message_id", message.ID,
		"traceID", traceID)

	h.respondWithSuccess(w, http.StatusCreated, message, traceID)
}

// getMessages handles GET /messages requests.
func (h *Handler) getMessages(w http.ResponseWriter, r *http.Request, traceID string) {
	// Placeholder implementation - would connect to repository layer
	messages := []types.Message{}

	slog.InfoContext(r.Context(), "Messages retrieved successfully",
		"message_count", len(messages),
		"traceID", traceID)

	h.respondWithSuccess(w, http.StatusOK, messages, traceID)
}

// respondWithSuccess sends a successful JSON response.
func (h *Handler) respondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}, traceID string) {
	w.WriteHeader(statusCode)
	response := types.Response{
		Success: true,
		Data:    data,
		TraceID: traceID,
	}
	json.NewEncoder(w).Encode(response)
}

// respondWithError sends an error JSON response.
func (h *Handler) respondWithError(w http.ResponseWriter, statusCode int, message string, traceID string) {
	w.WriteHeader(statusCode)
	response := types.Response{
		Success: false,
		Error:   message,
		TraceID: traceID,
	}
	json.NewEncoder(w).Encode(response)
}
