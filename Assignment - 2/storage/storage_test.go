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
	content := "this is a test"

	t.Run("Success", func(t *testing.T) {
		// Create a temporary directory to ensure a clean environment
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.txt")

		err := SaveData(ctx, filePath, content)
		require.NoError(t, err, "SaveData() returned an unexpected error")

		// Verify the content of the created file
		readBytes, err := os.ReadFile(filePath)
		require.NoError(t, err, "ReadFile() failed after save")
		require.Equal(t, content, string(readBytes), "file content mismatch")
	})

	t.Run("Failure on invalid path", func(t *testing.T) {
		// Attempting to write to a path that is a directory should fail
		err := SaveData(ctx, t.TempDir(), content)
		require.Error(t, err, "SaveData() was expected to return an error for an invalid path")
	})
}

func TestReadData(t *testing.T) {
	ctx := context.Background()
	content := "Hello, this is test data for reading!"

	t.Run("Success", func(t *testing.T) {
		// Create a temporary directory and write test data
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "read_test.txt")

		// First save some data
		err := SaveData(ctx, filePath, content)
		require.NoError(t, err, "SaveData() failed to setup test")

		// Now test reading it back
		readContent, err := ReadData(ctx, filePath)
		require.NoError(t, err, "ReadData() returned an unexpected error")
		require.Equal(t, content, readContent, "read content doesn't match original")
	})

	t.Run("Failure on non-existent file", func(t *testing.T) {
		nonExistentPath := "/path/that/does/not/exist/file.txt"
		_, err := ReadData(ctx, nonExistentPath)
		require.Error(t, err, "ReadData() should return error for non-existent file")
	})
}
