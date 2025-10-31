package handler

import (
	"context"
	"fmt"
	"log/slog"

	pb "cgi.com/goLangTraining/proto/message_service"
	"cgi.com/goLangTraining/src/pkg/storage"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const traceIDKey contextKey = "traceID"

// MessageHandler implements the MessageService gRPC service
type MessageHandler struct {
	pb.UnimplementedMessageServiceServer
	storage *storage.MessageStorage
}

// NewMessageHandler creates a new MessageHandler instance
func NewMessageHandler(storage *storage.MessageStorage) *MessageHandler {
	return &MessageHandler{
		storage: storage,
	}
}

// Save implements the Save RPC method
func (h *MessageHandler) Save(ctx context.Context, req *pb.SaveMessageRequest) (*emptypb.Empty, error) {
	traceID := uuid.New().String()
	ctx = context.WithValue(ctx, traceIDKey, traceID)

	slog.InfoContext(ctx, "Received Save request",
		"user", req.User,
		"message", req.Message,
		"traceID", traceID)

	// Validate input - return early following guidelines
	if req.User == "" || req.Message == "" {
		return nil, fmt.Errorf("user and message are required")
	}

	// Save message using storage service
	err := h.storage.AddMessage(req.User, req.Message)
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
func (h *MessageHandler) GetLast10(ctx context.Context, req *emptypb.Empty) (*pb.GetLast10Response, error) {
	traceID := uuid.New().String()
	ctx = context.WithValue(ctx, traceIDKey, traceID)

	slog.InfoContext(ctx, "Received GetLast10 request", "traceID", traceID)

	// Read messages from storage
	messages, err := h.storage.GetLastMessages(traceID, 10)
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
