package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "cgi.com/goLangTraining/proto/message_service"
	"cgi.com/goLangTraining/src/apps/message-grpc/internal/handler"
	"google.golang.org/grpc"
)

const (
	port              = ":50051"
	defaultAPIVersion = "1.0.0"
)

func main() {
	setupLogging()

	slog.Info("Starting CGI Message gRPC Service",
		"service", "message-grpc",
		"version", defaultAPIVersion)

	// Create TCP listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		slog.Error("Failed to listen", "error", err, "port", port)
		os.Exit(1)
	}

	// Create gRPC server
	s := grpc.NewServer()

	// Register message service
	messageHandler := handler.NewMessageHandler()
	pb.RegisterMessageServiceServer(s, messageHandler)

	slog.Info("Starting gRPC Message Store Server",
		"port", port,
		"service", "MessageService")

	fmt.Printf("🚀 gRPC Message Store Server started on port %s\n", port)
	fmt.Printf("📋 Available services:\n")
	fmt.Printf("   - Save(SaveMessageRequest) -> Empty\n")
	fmt.Printf("   - GetLast10(Empty) -> GetLast10Response\n")
	fmt.Printf("\n💡 Test with grpcurl:\n")
	fmt.Printf("   grpcurl -plaintext -d '{\"user\":\"alice\",\"message\":\"Hello gRPC!\"}' localhost:50051 message_service.MessageService/Save\n")
	fmt.Printf("   grpcurl -plaintext localhost:50051 message_service.MessageService/GetLast10\n")

	// Start server in goroutine for graceful shutdown
	go func() {
		if err := s.Serve(lis); err != nil {
			slog.Error("Failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("gRPC server shutting down...")

	// Graceful shutdown of gRPC server
	s.GracefulStop()

	slog.Info("gRPC server exited")
}

// setupLogging configures the default slog logger following guidelines
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-grpc",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)
}
