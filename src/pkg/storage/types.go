package storage

// FileMetrics holds metrics about file operations for structured logging.
// This type enables consistent performance monitoring and debugging across
// all file operations by capturing essential operation characteristics.
type FileMetrics struct {
	ContentSize int    `json:"content_size"`
	BytesRead   int    `json:"bytes_read"`
	Operation   string `json:"operation"`
}
