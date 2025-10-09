package storage

// FileOperation represents the different types of file operations that can be performed.
// Using human-readable string constants enables clear logging and debugging
// while maintaining type safety for operation classification.
type FileOperation string

const (
	// OperationRead represents a file read operation for metrics tracking
	OperationRead FileOperation = "read"
	// OperationWrite represents a file write operation for metrics tracking
	OperationWrite FileOperation = "write"
)

// FileMetrics holds metrics about file operations for structured logging.
// This type enables consistent performance monitoring and debugging across
// all file operations by capturing essential operation characteristics.
type FileMetrics struct {
	ContentSize int    `json:"content_size"`
	BytesRead   int    `json:"bytes_read"`
	Operation   string `json:"operation"`
}
