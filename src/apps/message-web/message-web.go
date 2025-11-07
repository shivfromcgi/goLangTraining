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
	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/version"
)

// Embedded file systems for web interface
//
//go:embed html/*
var htmlFiles embed.FS

const (
	gracefulShutdownTimeout = 30 * time.Second
	defaultPort             = 8090

	// Server configuration constants
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
	idleTimeout  = 60 * time.Second
)

func main() {
	setupLogging()

	slog.Info("Starting CGI Message Web Service")

	port := flag.Int("port", defaultPort, "Port for HTTP server")
	flag.Parse()

	startWebServer(*port)
}

// setupLogging configures the default slog logger following guidelines
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-web",
		"version", version.GetVersion())
	slog.SetDefault(logger)
}

// startWebServer starts the web server with ServeMux routing
func startWebServer(port int) {
	webHandler, err := handler.NewWebHandler(htmlFiles)
	if err != nil {
		slog.Error("Failed to create web handler", "error", err)
		os.Exit(1)
	}

	// Use ServeMux for consistency with API service
	mux := http.NewServeMux()

	// Web interface routes with shared middleware
	mux.Handle("/", middleware.TraceMiddleware(http.HandlerFunc(webHandler.IndexHandler)))
	mux.Handle("/messages", middleware.TraceMiddleware(http.HandlerFunc(webHandler.MessagesHandler)))
	mux.Handle("/health", middleware.TraceMiddleware(http.HandlerFunc(webHandler.HealthHandler)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
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
