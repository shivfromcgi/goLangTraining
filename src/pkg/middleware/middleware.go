package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const traceIDKey contextKey = "traceID"

// TraceMiddleware creates trace ID from headers or generates new one
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for existing trace ID in headers
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			// Generate new trace ID if none provided
			traceID = uuid.New().String()
		}

		// Add trace ID to context
		ctx := context.WithValue(r.Context(), traceIDKey, traceID)

		// Set trace ID in response headers
		w.Header().Set("X-Trace-ID", traceID)

		// Log incoming request with context
		slog.InfoContext(ctx, "Incoming HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return uuid.New().String()
}
