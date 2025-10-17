package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	pb "cgi.com/goLangTraining/proto/message_service"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	port             = ":50051"
	messagesFileName = "messages.txt"
)

// Message represents a message in our system (matching main.go structure)
type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// messageServer implements the MessageService gRPC service
type messageServer struct {
	pb.UnimplementedMessageServiceServer
}

// Save implements the Save RPC method
func (s *messageServer) Save(ctx context.Context, req *pb.SaveMessageRequest) (*emptypb.Empty, error) {
	traceID := uuid.New().String()
	ctx = context.WithValue(ctx, "traceID", traceID)

	slog.InfoContext(ctx, "Received Save request",
		"user", req.User,
		"message", req.Message,
		"traceID", traceID)

	// Validate input
	if req.User == "" || req.Message == "" {
		return nil, fmt.Errorf("user and message are required")
	}

	// Save message using the same logic as main.go
	err := saveMessage(ctx, req.User, req.Message)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save message",
			"error", err,
			"user", req.User,
			"traceID", traceID)
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	slog.InfoContext(ctx, "Message saved successfully",
		"user", req.User,
		"traceID", traceID)

	return &emptypb.Empty{}, nil
}

// GetLast10 implements the GetLast10 RPC method
func (s *messageServer) GetLast10(ctx context.Context, req *emptypb.Empty) (*pb.GetLast10Response, error) {
	traceID := uuid.New().String()
	ctx = context.WithValue(ctx, "traceID", traceID)

	slog.InfoContext(ctx, "Received GetLast10 request", "traceID", traceID)

	// Read messages from file
	messages, err := readLast10Messages(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read messages",
			"error", err,
			"traceID", traceID)
		return nil, fmt.Errorf("failed to read messages: %w", err)
	}

	// Convert to protobuf messages
	var pbMessages []*pb.Message
	for _, msg := range messages {
		pbMsg := &pb.Message{
			Id:        int32(msg.ID),
			User:      msg.User,
			Message:   msg.Message,
			Timestamp: timestamppb.New(msg.Timestamp),
			TraceId:   msg.TraceID,
		}
		pbMessages = append(pbMessages, pbMsg)
	}

	slog.InfoContext(ctx, "Returning messages",
		"count", len(pbMessages),
		"traceID", traceID)

	return &pb.GetLast10Response{
		Messages: pbMessages,
	}, nil
}

// saveMessage saves a message to the file (similar to main.go addMessage function)
func saveMessage(ctx context.Context, user, message string) error {
	traceID, _ := ctx.Value("traceID").(string)

	f, err := os.OpenFile(messagesFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s: %s\n", timestamp, user, message)
	_, err = f.WriteString(line)
	if err != nil {
		return err
	}

	slog.InfoContext(ctx, "Message appended to file",
		"user", user,
		"message", message,
		"traceID", traceID)

	return nil
}

// readLast10Messages reads the last 10 messages from the file
func readLast10Messages(ctx context.Context) ([]Message, error) {
	traceID, _ := ctx.Value("traceID").(string)

	f, err := os.Open(messagesFileName)
	if err != nil {
		if os.IsNotExist(err) {
			slog.InfoContext(ctx, "Messages file does not exist, returning empty list", "traceID", traceID)
			return []Message{}, nil
		}
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Get last 10 lines
	start := 0
	if len(lines) > 10 {
		start = len(lines) - 10
	}

	var messages []Message
	for i, line := range lines[start:] {
		if line != "" {
			message := parseMessageLine(line, start+i+1, traceID)
			if message != nil {
				messages = append(messages, *message)
			}
		}
	}

	return messages, nil
}

// parseMessageLine parses a message line from the file (similar to main.go)
func parseMessageLine(line string, id int, traceID string) *Message {
	// Simple parsing for [timestamp] user: message format
	if len(line) < 22 { // Minimum length for timestamp + user + message
		return nil
	}

	// Find end of timestamp (look for "] ")
	timestampEnd := -1
	for i := 0; i < len(line)-1; i++ {
		if line[i] == ']' && line[i+1] == ' ' {
			timestampEnd = i
			break
		}
	}

	if timestampEnd == -1 {
		return nil
	}

	remaining := line[timestampEnd+2:] // Skip "] "

	// Find ": " separator
	colonIndex := -1
	for i := 0; i < len(remaining)-1; i++ {
		if remaining[i] == ':' && remaining[i+1] == ' ' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return nil
	}

	user := remaining[:colonIndex]
	messageText := remaining[colonIndex+2:]

	// Parse timestamp
	timestampStr := line[1:timestampEnd] // Remove [ and ]
	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		timestamp = time.Now() // Fallback
	}

	return &Message{
		ID:        id,
		User:      user,
		Message:   messageText,
		Timestamp: timestamp,
		TraceID:   traceID,
	}
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "message-store-grpc",
		"version", "1.0.0",
	)
	slog.SetDefault(logger)

	// Create TCP listener
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Create gRPC server
	s := grpc.NewServer()

	// Register message service
	pb.RegisterMessageServiceServer(s, &messageServer{})

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

	// Start server
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
