package types

import (
	"time"
)

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
