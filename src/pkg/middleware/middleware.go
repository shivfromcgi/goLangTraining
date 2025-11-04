package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const TraceIDKey contextKey = "traceID"

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
		ctx := context.WithValue(r.Context(), TraceIDKey, traceID)

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
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return uuid.New().String()
}

// GetOrCreateTraceID extracts trace ID from context or creates a new one
// This is useful for gRPC services that might receive contexts without trace IDs
func GetOrCreateTraceID(ctx context.Context) (context.Context, string) {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return ctx, traceID
	}

	traceID := uuid.New().String()
	ctx = context.WithValue(ctx, TraceIDKey, traceID)
	return ctx, traceID
}

// RequestSizeLimitMiddleware limits the size of request bodies to prevent DoS attacks
func RequestSizeLimitMiddleware(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
