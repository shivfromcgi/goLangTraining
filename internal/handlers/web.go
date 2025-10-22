package handlers

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"cgi.com/goLangTraining/internal/service"
	"cgi.com/goLangTraining/internal/types"
)

// WebHandlers contains dependencies for web interface handlers
type WebHandlers struct {
	MessageActor *service.MessageActor
	HTMLFiles    embed.FS
}

// NewWebHandlers creates a new WebHandlers instance
func NewWebHandlers(messageActor *service.MessageActor, htmlFiles embed.FS) *WebHandlers {
	return &WebHandlers{
		MessageActor: messageActor,
		HTMLFiles:    htmlFiles,
	}
}

// IndexHandler serves the static index page
func (h *WebHandlers) IndexHandler(w http.ResponseWriter, r *http.Request) {
	traceID := GetTraceID(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	indexHTML, err := h.HTMLFiles.ReadFile("html/index.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read index.html", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Served static index page", "traceID", traceID)
	w.Write(indexHTML)
}

// MessagesHandler serves the dynamic messages page
func (h *WebHandlers) MessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID := GetTraceID(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	messages, err := h.MessageActor.GetMessages(traceID)
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

	tmpl, err := template.ParseFS(h.HTMLFiles, "html/messages.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to parse messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err = tmpl.Execute(w, data); err != nil {
		slog.ErrorContext(r.Context(), "Failed to execute messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Served dynamic messages page",
		"message_count", len(messages),
		"traceID", traceID)
}
