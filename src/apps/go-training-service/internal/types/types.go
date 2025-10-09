package types

import "time"

// CtxKey represents context key type to avoid string collisions.
// Using a custom type prevents accidental key conflicts when storing
// values in context, ensuring reliable context value retrieval.
type CtxKey string

// Assignment represents different assignment types for CLI selection.
// Using typed constants instead of raw strings provides compile-time
// validation and prevents typos in assignment routing logic.
type Assignment string

const (
	// AssignmentOne represents the message system assignment
	AssignmentOne Assignment = "assignment1"
	// AssignmentTwo represents the advanced storage assignment
	AssignmentTwo Assignment = "assignment2"
	// AssignmentThree represents the HTTP JSON API assignment
	AssignmentThree Assignment = "assignment3"
)

// Message represents a JSON message structure for the API.
// This structure provides consistent message format across all API endpoints
// with embedded tracing support for distributed debugging capabilities.
type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id"`
}

// CreateMessageRequest represents the request body for creating a message.
// Separating request and response structures allows independent evolution
// of API contracts without affecting internal message representation.
type CreateMessageRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// Response represents a standard API response structure.
// This unified response format ensures consistent error handling patterns
// and simplifies client-side response processing across all endpoints.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	TraceID string      `json:"trace_id"`
}

// HealthStatus represents the health check response structure.
// Structured health responses enable automated monitoring systems
// to parse service status and version information programmatically.
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
