package handler

import (
	"context"
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"cgi.com/goLangTraining/src/pkg/storage"
	"cgi.com/goLangTraining/src/pkg/types"
	"github.com/google/uuid"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const traceIDKey contextKey = "traceID"

// WebHandler contains dependencies for web interface handlers
type WebHandler struct {
	storage   *storage.MessageStorage
	htmlFiles embed.FS
}

// NewWebHandler creates a new WebHandler instance
func NewWebHandler(storage *storage.MessageStorage, htmlFiles embed.FS) *WebHandler {
	return &WebHandler{
		storage:   storage,
		htmlFiles: htmlFiles,
	}
}

// IndexHandler serves the static index page
func (h *WebHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	indexHTML, err := h.htmlFiles.ReadFile("html/index.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read index.html", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Served static index page", "traceID", traceID)
	w.Write(indexHTML)
}

// MessagesHandler serves the dynamic messages page
func (h *WebHandler) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	messages, err := h.storage.GetLastMessages(traceID, 10)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read messages for web page", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := types.MessagesPageData{
		Messages:    messages,
		GeneratedAt: time.Now(),
		TraceID:     traceID,
	}

	tmpl, err := template.ParseFS(h.htmlFiles, "html/messages.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to parse messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to execute messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Served dynamic messages page",
		"message_count", len(messages),
		"traceID", traceID)
}

// HealthHandler handles health check requests
func (h *WebHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)
	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	slog.InfoContext(r.Context(), "Health check served", "traceID", traceID)
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
