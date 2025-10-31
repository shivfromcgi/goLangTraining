package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cgi.com/goLangTraining/src/apps/message-web/internal/handler"
	"cgi.com/goLangTraining/src/pkg/storage"
	"github.com/gorilla/mux"
)

// Embedded file systems for web interface
//
//go:embed html/*
var htmlFiles embed.FS

const (
	gracefulShutdownTimeout = 30 * time.Second
	defaultAPIVersion       = "1.0.0"
	defaultPort             = 8090
	messagesFileName        = "messages.txt"
)

func main() {
	setupLogging()

	slog.Info("Starting CGI Message Web Service",
		"service", "message-web",
		"version", defaultAPIVersion)

	port := flag.Int("port", defaultPort, "Port for HTTP server")
	flag.Parse()

	// Initialize services in main following guidelines
	messageStorage := storage.NewMessageStorage(messagesFileName)

	startWebServer(*port, messageStorage)
}

// setupLogging configures the default slog logger following guidelines
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-web",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)
}

// startWebServer starts the web server with proper mux routing
func startWebServer(port int, messageStorage *storage.MessageStorage) {
	webHandler := handler.NewWebHandler(messageStorage, htmlFiles)

	// Use mux as per guidelines, not standard http
	router := mux.NewRouter()

	// Web interface routes
	router.HandleFunc("/", handler.TraceMiddleware(webHandler.IndexHandler)).Methods("GET")
	router.HandleFunc("/messages", handler.TraceMiddleware(webHandler.MessagesHandler)).Methods("GET")
	router.HandleFunc("/health", handler.TraceMiddleware(webHandler.HealthHandler)).Methods("GET")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		slog.Info("Web server starting",
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

	slog.Info("Web server shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Web server exited")
}
