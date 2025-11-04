package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cgi.com/goLangTraining/src/apps/message-api/internal/handler"
	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"github.com/gorilla/mux"
)

const (
	// gracefulShutdownTimeout defines how long the server waits for existing connections
	// to finish before forcefully shutting down. In containerized environments, this
	// allows the container orchestrator (Docker, Kubernetes) to gracefully terminate
	// the service without dropping active requests. The timeout should be less than
	// the container's terminationGracePeriodSeconds to ensure clean shutdown.
	gracefulShutdownTimeout = 30 * time.Second
	defaultAPIVersion       = "1.0.0"
	defaultPort             = 8080
)

func main() {
	// Declare flags at top following guidelines
	port := flag.Int("port", defaultPort, "Port for HTTP server")
	flag.Parse()

	// Set up slog.SetDefault following guidelines
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-api",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)

	slog.Info("Starting CGI Message API Service",
		"service", "message-api",
		"version", defaultAPIVersion)

	// Use default storage following guidelines
	messageStorage := storage.GetDefaultStorage()

	startAPIServer(*port, messageStorage)
}

// startAPIServer starts the API server with mux routing
func startAPIServer(port int, messageStorage *storage.MessageStorage) {
	// Use mux router following guidelines
	r := mux.NewRouter()

	// Create handlers with dependencies
	messagesHandler := handler.NewMessagesHandler(messageStorage)
	healthHandler := handler.NewHealthHandler()
	wsHandler := handler.NewWebSocketHandler(messageStorage)

	// API routes with proper structure
	r.Handle("/api/v1/messages", middleware.TraceMiddleware(messagesHandler))
	r.Handle("/api/v1/health", middleware.TraceMiddleware(healthHandler))
	r.Handle("/ws", middleware.TraceMiddleware(wsHandler))

	// Legacy routes for backward compatibility
	r.Handle("/messages", middleware.TraceMiddleware(messagesHandler))
	r.Handle("/health", middleware.TraceMiddleware(healthHandler))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("API server starting",
			"port", port,
			"address", fmt.Sprintf("http://localhost:%d", port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("API server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("API server exited")
}
