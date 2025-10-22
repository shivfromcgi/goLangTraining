package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"cgi.com/goLangTraining/internal/service"
	"cgi.com/goLangTraining/internal/types"
)

// APIHandlers contains dependencies for HTTP handlers
type APIHandlers struct {
	MessageActor *service.MessageActor
}

// NewAPIHandlers creates a new APIHandlers instance
func NewAPIHandlers(messageActor *service.MessageActor) *APIHandlers {
	return &APIHandlers{
		MessageActor: messageActor,
	}
}

// MessagesHandler handles both GET and POST requests for messages
func (h *APIHandlers) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID := GetTraceID(r)

	switch r.Method {
	case http.MethodPost:
		h.CreateMessage(w, r, traceID)
	case http.MethodGet:
		h.GetMessages(w, r, traceID)
	default:
		RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", traceID)
	}
}

// CreateMessage handles POST requests to create a new message
func (h *APIHandlers) CreateMessage(w http.ResponseWriter, r *http.Request, traceID string) {
	var req types.CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err, "traceID", traceID)
		RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload", traceID)
		return
	}

	if req.User == "" || req.Message == "" {
		RespondWithError(w, http.StatusBadRequest, "User and message are required", traceID)
		return
	}

	// Use message actor for concurrent-safe message creation and fan-out
	if err := h.MessageActor.CreateMessage(req.User, req.Message, traceID); err != nil {
		slog.ErrorContext(r.Context(), "Failed to save message via actor", "error", err, "traceID", traceID)
		RespondWithError(w, http.StatusInternalServerError, "Failed to save message", traceID)
		return
	}

	message := types.Message{
		ID:        int(time.Now().UnixNano() / 1000000),
		User:      req.User,
		Message:   req.Message,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}

	slog.InfoContext(r.Context(), "Message created successfully via actor",
		"user", req.User,
		"message_id", message.ID,
		"traceID", traceID)

	RespondWithSuccess(w, http.StatusCreated, message, traceID)
}

// GetMessages handles GET requests to retrieve messages
func (h *APIHandlers) GetMessages(w http.ResponseWriter, r *http.Request, traceID string) {
	messages, err := h.MessageActor.GetMessages(traceID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read messages via actor", "error", err, "traceID", traceID)
		RespondWithError(w, http.StatusInternalServerError, "Failed to read messages", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Messages retrieved successfully via actor",
		"message_count", len(messages),
		"traceID", traceID)

	RespondWithSuccess(w, http.StatusOK, messages, traceID)
}

// HealthHandler handles health check requests
func (h *APIHandlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID := GetTraceID(r)

	health := types.HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // TODO: make this configurable
	}

	RespondWithSuccess(w, http.StatusOK, health, traceID)
}
