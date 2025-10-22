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
