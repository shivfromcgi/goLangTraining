package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"cgi.com/goLangTraining/internal/types"
	"github.com/google/uuid"
)

// TraceMiddleware adds trace ID to request context
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "traceID", traceID)

		slog.InfoContext(ctx, "Incoming HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"traceID", traceID)

		w.Header().Set("X-Trace-ID", traceID)
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RespondWithSuccess sends a successful response with typed data
func RespondWithSuccess[T any](w http.ResponseWriter, statusCode int, data T, traceID string) {
	w.WriteHeader(statusCode)
	response := types.Response[T]{
		Success: true,
		Data:    data,
		TraceID: traceID,
	}
	json.NewEncoder(w).Encode(response)
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, statusCode int, message string, traceID string) {
	w.WriteHeader(statusCode)
	response := types.EmptyResponse{
		Success: false,
		Error:   message,
		TraceID: traceID,
	}
	json.NewEncoder(w).Encode(response)
}

// GetTraceID extracts trace ID from request context
func GetTraceID(r *http.Request) string {
	if traceID, ok := r.Context().Value("traceID").(string); ok {
		return traceID
	}
	return uuid.New().String()
}
