package handler

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"cgi.com/goLangTraining/src/pkg/types"
)

const (
	// Configuration constants for timeouts
	storageOperationTimeout = 5 * time.Second
)

// WebHandler contains dependencies for web interface handlers
type WebHandler struct {
	htmlFiles        embed.FS
	messagesTemplate *template.Template
}

// NewWebHandler creates a new WebHandler instance
func NewWebHandler(htmlFiles embed.FS) (*WebHandler, error) {
	// Parse messages template once at startup
	messagesTemplate, err := template.ParseFS(htmlFiles, "html/messages.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse messages template: %w", err)
	}

	return &WebHandler{
		htmlFiles:        htmlFiles,
		messagesTemplate: messagesTemplate,
	}, nil
}

// IndexHandler serves the static index page
func (h *WebHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	indexHTML, err := h.htmlFiles.ReadFile("html/index.html")
	if err != nil {
		// Internal error - log details but don't expose to user
		slog.ErrorContext(r.Context(), "Failed to read index.html", "error", err, "traceID", traceID)
		w.WriteHeader(http.StatusInternalServerError)
		// No body - don't expose internal state
		return
	}

	if _, err := w.Write(indexHTML); err != nil {
		slog.ErrorContext(r.Context(), "Failed to write index.html response", "error", err, "traceID", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Served static index page", "traceID", traceID)
}

// MessagesHandler serves the dynamic messages page
func (h *WebHandler) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Create timeout context for storage operation
	ctx, cancel := context.WithTimeout(r.Context(), storageOperationTimeout)
	defer cancel()

	// Use a channel to make storage operation cancellable
	type result struct {
		messages []types.Message
		err      error
	}
	resultChan := make(chan result, 1)

	go func() {
		messages, err := storage.GetLastMessages(r.Context(), 10)
		resultChan <- result{messages: messages, err: err}
	}()

	var messages []types.Message
	select {
	case res := <-resultChan:
		if res.err != nil {
			// Internal error - log details but don't expose to user
			slog.ErrorContext(r.Context(), "Failed to read messages for web page", "error", res.err, "traceID", traceID)
			w.WriteHeader(http.StatusInternalServerError)
			// No body - don't expose internal state
			return
		}
		messages = res.messages
	case <-ctx.Done():
		// Timeout error - log details
		slog.ErrorContext(r.Context(), "Storage operation timed out", "timeout", storageOperationTimeout, "traceID", traceID)
		w.WriteHeader(http.StatusRequestTimeout)
		// No body - don't expose internal state
		return
	}

	data := types.MessagesPageData{
		Messages:    messages,
		GeneratedAt: time.Now().UTC(),
		TraceID:     traceID,
	}

	if err := h.messagesTemplate.Execute(w, data); err != nil {
		// Template execution error - log details but don't expose to user
		slog.ErrorContext(r.Context(), "Failed to execute messages template", "error", err, "traceID", traceID)
		w.WriteHeader(http.StatusInternalServerError)
		// No body - don't expose internal state
		return
	}

	slog.InfoContext(r.Context(), "Served dynamic messages page",
		"message_count", len(messages),
		"traceID", traceID)
}

// HealthHandler handles health check requests
func (h *WebHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	traceID := getTraceID(r)
	w.Header().Set("X-Trace-ID", traceID)
	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		slog.ErrorContext(r.Context(), "Failed to write health check response", "error", err, "traceID", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Health check served", "traceID", traceID)
}

// getTraceID extracts trace ID from request context
func getTraceID(r *http.Request) string {
	return middleware.GetTraceID(r.Context())
}
