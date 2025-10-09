package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveData(t *testing.T) {
	ctx := context.Background()

	// Table-driven test cases that cover meaningful scenarios
	testCases := []struct {
		name        string
		content     string
		expectError bool
		setupFunc   func(t *testing.T) string
		description string
	}{
		{
			name:        "successful_write_small_content",
			content:     "small test content",
			expectError: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "test.txt")
			},
			description: "validates basic file write functionality with small content",
		},
		{
			name:        "successful_write_empty_content",
			content:     "",
			expectError: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "empty.txt")
			},
			description: "ensures empty content can be written without errors",
		},
		{
			name:        "failure_invalid_directory_path",
			content:     "content for invalid path",
			expectError: true,
			setupFunc: func(t *testing.T) string {
				return t.TempDir() // Directory path instead of file path
			},
			description: "verifies proper error handling when writing to directory",
		},
		{
			name:        "successful_write_large_content",
			content:     generateLargeContent(1000),
			expectError: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "large.txt")
			},
			description: "tests file write performance with larger content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.setupFunc(t)

			err := SaveData(ctx, filePath, tc.content)

			if tc.expectError {
				require.Error(t, err, "Expected SaveData to fail for case: %s", tc.description)
				return
			}

			require.NoError(t, err, "SaveData failed unexpectedly for case: %s", tc.description)

			// Verify content was written correctly
			actualContent, readErr := os.ReadFile(filePath)
			require.NoError(t, readErr, "Failed to read written file")
			require.Equal(t, tc.content, string(actualContent), "Written content mismatch")
		})
	}
}

func TestReadData(t *testing.T) {
	ctx := context.Background()

	// Table-driven test cases for read operations
	testCases := []struct {
		name        string
		setupFunc   func(t *testing.T) (string, string) // Returns (filePath, expectedContent)
		expectError bool
		description string
	}{
		{
			name: "successful_read_existing_file",
			setupFunc: func(t *testing.T) (string, string) {
				content := "test content for reading"
				filePath := filepath.Join(t.TempDir(), "read_test.txt")
				err := SaveData(ctx, filePath, content)
				require.NoError(t, err, "Setup failed")
				return filePath, content
			},
			expectError: false,
			description: "validates basic file read functionality",
		},
		{
			name: "successful_read_empty_file",
			setupFunc: func(t *testing.T) (string, string) {
				content := ""
				filePath := filepath.Join(t.TempDir(), "empty_read.txt")
				err := SaveData(ctx, filePath, content)
				require.NoError(t, err, "Setup failed")
				return filePath, content
			},
			expectError: false,
			description: "ensures empty files can be read correctly",
		},
		{
			name: "failure_nonexistent_file",
			setupFunc: func(t *testing.T) (string, string) {
				return "/nonexistent/path/file.txt", ""
			},
			expectError: true,
			description: "verifies proper error handling for missing files",
		},
		{
			name: "successful_read_large_content",
			setupFunc: func(t *testing.T) (string, string) {
				content := generateLargeContent(500)
				filePath := filepath.Join(t.TempDir(), "large_read.txt")
				err := SaveData(ctx, filePath, content)
				require.NoError(t, err, "Setup failed")
				return filePath, content
			},
			expectError: false,
			description: "tests file read performance with larger content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath, expectedContent := tc.setupFunc(t)

			actualContent, err := ReadData(ctx, filePath)

			if tc.expectError {
				require.Error(t, err, "Expected ReadData to fail for case: %s", tc.description)
				return
			}

			require.NoError(t, err, "ReadData failed unexpectedly for case: %s", tc.description)
			require.Equal(t, expectedContent, actualContent, "Read content mismatch")
		})
	}
}

// generateLargeContent creates test content of specified size for performance testing.
// This helper avoids magic numbers and provides consistent test data generation.
func generateLargeContent(sizeKB int) string {
	const baseText = "This is test content for file operations. "
	targetBytes := sizeKB * 1024

	content := ""
	for len(content) < targetBytes {
		content += baseText
	}

	// Trim to exact size to ensure consistent testing
	if len(content) > targetBytes {
		content = content[:targetBytes]
	}

	return content
}
