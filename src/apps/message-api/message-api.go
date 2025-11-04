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
	"cgi.com/goLangTraining/src/pkg/storage"
)

const (
	gracefulShutdownTimeout = 30 * time.Second
	defaultAPIVersion       = "1.0.0"
	defaultPort             = 8080
)

func main() {
	setupLogging()

	slog.Info("Starting CGI Message API Service",
		"service", "message-api",
		"version", defaultAPIVersion)

	port := flag.Int("port", defaultPort, "Port for HTTP server")
	flag.Parse()

	// Use default storage following guidelines
	messageStorage := storage.GetDefaultStorage()

	startAPIServer(*port, messageStorage)
}

// setupLogging configures the default slog logger following guidelines
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-api",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)
}

// startAPIServer starts the API server with ServeMux routing
func startAPIServer(port int, messageStorage *storage.MessageStorage) {
	apiHandler := handler.NewAPIHandler(messageStorage)

	// Use ServeMux as per reviewer feedback
	mux := http.NewServeMux()

	// API routes with proper structure
	mux.HandleFunc("/api/v1/messages", handler.TraceMiddleware(apiHandler.MessagesHandler))
	mux.HandleFunc("/api/v1/health", handler.TraceMiddleware(apiHandler.HealthHandler))

	// Legacy routes for backward compatibility
	mux.HandleFunc("/messages", handler.TraceMiddleware(apiHandler.MessagesHandler))
	mux.HandleFunc("/health", handler.TraceMiddleware(apiHandler.HealthHandler))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
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
