package handler

import (
	"context"
	"log/slog"

	pb "cgi.com/goLangTraining/proto/message_service"
	"cgi.com/goLangTraining/src/pkg/middleware"
	"cgi.com/goLangTraining/src/pkg/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MessageHandler implements the MessageService gRPC service
type MessageHandler struct {
	pb.UnimplementedMessageServiceServer
}

// Handler is the singleton instance of MessageHandler
var Handler = &MessageHandler{}

// Save implements the Save RPC method
func (h *MessageHandler) Save(ctx context.Context, req *pb.SaveMessageRequest) (*emptypb.Empty, error) {
	ctx, traceID := middleware.GetOrCreateTraceID(ctx)

	slog.InfoContext(ctx, "Received Save request",
		"user", req.User,
		"message", req.Message,
		"traceID", traceID)

	// Validate input - return early following guidelines
	if req.User == "" || req.Message == "" {
		return nil, status.Error(codes.InvalidArgument, "user and message are required")
	}

	// Save message using storage service (direct file access - gRPC service is the data store)
	err := storage.AddMessageToFile(ctx, req.User, req.Message)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save message",
			"error", err,
			"user", req.User,
			"traceID", traceID)
		return nil, status.Error(codes.Internal, "failed to save message")
	}

	slog.InfoContext(ctx, "Message saved successfully",
		"user", req.User,
		"traceID", traceID)

	return &emptypb.Empty{}, nil
}

// GetLast10 implements the GetLast10 RPC method
func (h *MessageHandler) GetLast10(ctx context.Context, req *emptypb.Empty) (*pb.GetLast10Response, error) {
	ctx, traceID := middleware.GetOrCreateTraceID(ctx)

	slog.InfoContext(ctx, "Received GetLast10 request", "traceID", traceID)

	// Read messages from storage (direct file access - gRPC service is the data store)
	messages, err := storage.GetLastMessagesFromFile(ctx, 10)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read messages",
			"error", err,
			"traceID", traceID)
		return nil, status.Error(codes.Internal, "failed to retrieve messages")
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
