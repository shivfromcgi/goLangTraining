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
	"github.com/gorilla/mux"
)

const (
	gracefulShutdownTimeout = 30 * time.Second
	defaultAPIVersion       = "1.0.0"
	defaultPort             = 8080
	messagesFileName        = "messages.txt"
)

func main() {
	setupLogging()

	slog.Info("Starting CGI Message API Service",
		"service", "message-api",
		"version", defaultAPIVersion)

	port := flag.Int("port", defaultPort, "Port for HTTP server")
	flag.Parse()

	// Initialize storage service following guidelines - all in main
	messageStorage := storage.NewMessageStorage(messagesFileName)

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

// startAPIServer starts the API server with proper mux routing
func startAPIServer(port int, messageStorage *storage.MessageStorage) {
	apiHandler := handler.NewAPIHandler(messageStorage)

	// Use mux as per guidelines, not standard http
	router := mux.NewRouter()

	// API routes with proper structure
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/messages", handler.TraceMiddleware(apiHandler.MessagesHandler)).Methods("GET", "POST")
	api.HandleFunc("/health", handler.TraceMiddleware(apiHandler.HealthHandler)).Methods("GET")

	// Legacy routes for backward compatibility
	router.HandleFunc("/messages", handler.TraceMiddleware(apiHandler.MessagesHandler)).Methods("GET", "POST")
	router.HandleFunc("/health", handler.TraceMiddleware(apiHandler.HealthHandler)).Methods("GET")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
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
