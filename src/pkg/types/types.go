package types

import (
	"time"
)

// Message represents a message in our system
type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// CreateMessageRequest represents the request body for creating a message
type CreateMessageRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// MessagesPageData represents the data passed to the messages template
type MessagesPageData struct {
	Messages    []Message `json:"messages"`
	GeneratedAt time.Time `json:"generated_at"`
	TraceID     string    `json:"trace_id"`
}

// Response represents a standard API response structure
// Generic response type to avoid interface{}
type Response[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	TraceID string `json:"trace_id"`
}

// Common response types
type MessageResponse = Response[[]Message]
type SingleMessageResponse = Response[Message]
type HealthResponse = Response[HealthStatus]
type EmptyResponse = Response[any] // For responses with no data

// HealthStatus represents the health check response structure
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
