package storage

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pb "cgi.com/goLangTraining/proto/message_service"
	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultGRPCServer = "localhost:50051"
	grpcTimeout       = 5 * time.Second
)

var (
	grpcClient     pb.MessageServiceClient
	grpcConn       *grpc.ClientConn
	grpcClientOnce sync.Once
	grpcClientErr  error
)

// InitGRPCClient initializes the gRPC client connection to the message store service
// This should be called at application startup
func InitGRPCClient(serverAddr string) error {
	var initErr error
	grpcClientOnce.Do(func() {
		if serverAddr == "" {
			serverAddr = defaultGRPCServer
		}

		slog.Info("Initializing gRPC client connection", "server", serverAddr)

		conn, err := grpc.Dial(serverAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(10*time.Second),
		)
		if err != nil {
			initErr = fmt.Errorf("failed to connect to gRPC server: %w", err)
			grpcClientErr = initErr
			return
		}

		grpcConn = conn
		grpcClient = pb.NewMessageServiceClient(conn)
		slog.Info("gRPC client connected successfully", "server", serverAddr)
	})

	if grpcClientErr != nil {
		return grpcClientErr
	}
	return initErr
}

// CloseGRPCClient closes the gRPC connection
// This should be called during application shutdown
func CloseGRPCClient() error {
	if grpcConn != nil {
		slog.Info("Closing gRPC client connection")
		return grpcConn.Close()
	}
	return nil
}

// AddMessage sends a message to the gRPC store service
func AddMessage(ctx context.Context, user, message string) error {
	if grpcClient == nil {
		return fmt.Errorf("gRPC client not initialized - call InitGRPCClient first")
	}

	traceID := middleware.GetTraceID(ctx)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	req := &pb.SaveMessageRequest{
		User:    user,
		Message: message,
	}

	slog.InfoContext(ctx, "Sending message to gRPC store",
		"user", user,
		"messageLength", len(message),
		"traceID", traceID)

	_, err := grpcClient.Save(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save message via gRPC",
			"error", err,
			"user", user,
			"traceID", traceID)
		return fmt.Errorf("gRPC save failed: %w", err)
	}

	slog.InfoContext(ctx, "Message saved successfully via gRPC",
		"user", user,
		"traceID", traceID)
	return nil
}

// GetLastMessages retrieves the last N messages from the gRPC store service
func GetLastMessages(ctx context.Context, limit int) ([]types.Message, error) {
	if grpcClient == nil {
		return nil, fmt.Errorf("gRPC client not initialized - call InitGRPCClient first")
	}

	traceID := middleware.GetTraceID(ctx)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, grpcTimeout)
	defer cancel()

	slog.InfoContext(ctx, "Requesting messages from gRPC store",
		"limit", limit,
		"traceID", traceID)

	// Note: The current gRPC service only supports GetLast10, not arbitrary limit
	// For now, we'll call GetLast10 and then take only 'limit' messages
	resp, err := grpcClient.GetLast10(ctx, &emptypb.Empty{})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve messages via gRPC",
			"error", err,
			"traceID", traceID)
		return nil, fmt.Errorf("gRPC get failed: %w", err)
	}

	// Convert protobuf messages to our types
	messages := make([]types.Message, 0, len(resp.Messages))
	for _, pbMsg := range resp.Messages {
		msg := types.Message{
			ID:        int(pbMsg.Id),
			User:      pbMsg.User,
			Message:   pbMsg.Message,
			Timestamp: pbMsg.Timestamp.AsTime(),
			TraceID:   pbMsg.TraceId,
		}
		messages = append(messages, msg)
	}

	// Limit to requested number
	if limit > 0 && limit < len(messages) {
		messages = messages[:limit]
	}

	slog.InfoContext(ctx, "Messages retrieved successfully via gRPC",
		"count", len(messages),
		"traceID", traceID)

	return messages, nil
}

// ReadMessages retrieves all messages from the gRPC store service
// For compatibility with existing code
func ReadMessages(ctx context.Context) ([]types.Message, error) {
	// Since gRPC only supports GetLast10, we'll use that
	// In a real implementation, you'd add a GetAll endpoint to the proto
	return GetLastMessages(ctx, 10)
}
