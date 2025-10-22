package types

import (
	"net/http"
	"time"
)

// MessageEvent represents an event in the message system (CSP pattern)
type MessageEvent struct {
	Type    string                    `json:"type"`    // "create", "get", "subscribe", "unsubscribe"
	Message Message                   `json:"message"` // The message data
	Reply   chan MessageEventResponse `json:"-"`       // Response channel
}

// MessageEventResponse represents a response to a message event
type MessageEventResponse struct {
	Messages []Message `json:"messages,omitempty"`
	Error    error     `json:"error,omitempty"`
	Success  bool      `json:"success"`
}

// MessageListener represents a listener/subscriber in the message system
type MessageListener struct {
	ID       string        `json:"id"`
	Channel  chan Message  `json:"-"`                   // Channel to send messages to this listener
	Done     chan struct{} `json:"-"`                   // Channel to signal when listener is done
	Request  *http.Request `json:"-"`                   // Associated HTTP request for context
	Created  time.Time     `json:"created"`             // When this listener was created
	LastPing time.Time     `json:"last_ping,omitempty"` // Last activity timestamp
}
