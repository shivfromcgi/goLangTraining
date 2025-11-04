package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	pb "cgi.com/goLangTraining/proto/message_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultServerAddr = "localhost:50051"
)

func main() {
	// Parse flags following guidelines - declare at top
	var (
		serverAddr = flag.String("server", defaultServerAddr, "gRPC server address")
		user       = flag.String("user", "", "User for message operations")
		message    = flag.String("message", "", "Message to save")
		getLast10  = flag.Bool("get", false, "Get last 10 messages")
	)
	flag.Parse()

	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-grpc-client",
	)
	slog.SetDefault(logger)

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to server", "error", err, "serverAddr", *serverAddr)
		os.Exit(1)
	}
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)
	slog.Info("Connected to gRPC Message Service", "serverAddr", *serverAddr)

	// Handle operations using a helper function
	if err := runClientOperations(client, *getLast10, *user, *message); err != nil {
		slog.Error("Client operation failed", "error", err)
		os.Exit(1)
	}
}

func runClientOperations(client pb.MessageServiceClient, getLast10 bool, user, message string) error {
	// Handle operations - return early following guidelines
	if getLast10 {
		return getMessages(client)
	}

	if user != "" && message != "" {
		if err := saveMessage(client, user, message); err != nil {
			return fmt.Errorf("failed to save message: %w", err)
		}

		// After saving, automatically show last 10 messages as per reviewer feedback
		slog.Info("Fetching updated messages after save")
		if err := getMessages(client); err != nil {
			return fmt.Errorf("failed to get messages after save: %w", err)
		}
		return nil
	}

	// Default demo mode
	return runDemo(client)
}

func saveMessage(client pb.MessageServiceClient, user, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SaveMessageRequest{
		User:    user,
		Message: message,
	}

	slog.Info("Saving message", "user", user, "message", message)

	_, err := client.Save(ctx, req)
	if err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	slog.Info("Message saved successfully", "user", user)
	return nil
}

func getMessages(client pb.MessageServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("Fetching last 10 messages")

	resp, err := client.GetLast10(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("get messages failed: %w", err)
	}

	messages := resp.GetMessages()
	if len(messages) == 0 {
		slog.Info("No messages found")
		return nil
	}

	slog.Info("Retrieved messages", "count", len(messages))
	for _, msg := range messages {
		timestamp := msg.GetTimestamp().AsTime()
		slog.Info("Message",
			"id", msg.GetId(),
			"user", msg.GetUser(),
			"timestamp", timestamp.Format("2006-01-02 15:04:05"),
			"message", msg.GetMessage())
	}

	return nil
}

func runDemo(client pb.MessageServiceClient) error {
	slog.Info("Running gRPC Client Demo")
	slog.Info("Usage examples",
		"save_message", "go run . -user=alice -message='Hello gRPC!'",
		"get_messages", "go run . -get",
		"custom_server", "go run . -server=localhost:50051 -get")

	demoUser := "demo"
	demoMessage := fmt.Sprintf("gRPC Client Demo - %s", time.Now().Format("15:04:05"))

	slog.Info("Step 1: Saving demo message")
	err := saveMessage(client, demoUser, demoMessage)
	if err != nil {
		return fmt.Errorf("demo failed - save message: %w", err)
	}

	slog.Info("Step 2: Getting last 10 messages")
	err = getMessages(client)
	if err != nil {
		return fmt.Errorf("demo failed - get messages: %w", err)
	}

	slog.Info("gRPC client operation completed successfully")
	return nil
}
