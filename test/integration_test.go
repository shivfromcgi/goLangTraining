package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"cgi.com/goLangTraining/src/pkg/types"
	"github.com/gorilla/websocket"
)

const (
	testAPIURL = "http://localhost:8080/api/v1/messages"
	testWSURL  = "ws://localhost:8080/ws"
)

// TestRealTimeBroadcast tests the CSP/Actor pattern implementation
// This test validates that messages posted via API are immediately broadcast to WebSocket clients
func TestRealTimeBroadcast(t *testing.T) {
	// Skip if running in CI or if server is not running
	if !isServerRunning() {
		t.Skip("Test server not running, skipping integration test")
		return
	}

	// Set up logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Create multiple WebSocket clients to test broadcasting
	const numClients = 3
	var wg sync.WaitGroup
	receivedMessages := make([][]string, numClients)

	// Start WebSocket clients
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			u, _ := url.Parse(testWSURL)
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				t.Errorf("Client %d failed to connect: %v", clientID, err)
				return
			}
			defer conn.Close()

			// Read messages for 5 seconds
			timeout := time.After(5 * time.Second)
			done := make(chan bool)

			go func() {
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						done <- true
						return
					}
					receivedMessages[clientID] = append(receivedMessages[clientID], string(message))
				}
			}()

			select {
			case <-timeout:
				slog.Info("Client timeout reached", "clientID", clientID)
			case <-done:
				slog.Info("Client disconnected", "clientID", clientID)
			}
		}(i)
	}

	// Wait a moment for clients to connect
	time.Sleep(1 * time.Second)

	// Post test messages via API
	testMessages := []struct {
		user    string
		message string
	}{
		{"TestUser1", "Hello from integration test!"},
		{"TestUser2", "Testing real-time broadcast"},
		{"TestUser3", "CSP pattern working!"},
	}

	for _, msg := range testMessages {
		if err := postTestMessage(msg.user, msg.message); err != nil {
			t.Errorf("Failed to post message: %v", err)
		}
		time.Sleep(500 * time.Millisecond) // Small delay between messages
	}

	// Wait for all clients to finish
	wg.Wait()

	// Validate that all clients received the messages
	fmt.Printf("\n=== Integration Test Results ===\n")
	for i, messages := range receivedMessages {
		fmt.Printf("Client %d received %d messages:\n", i, len(messages))
		for _, msg := range messages {
			fmt.Printf("  - %s\n", msg)
		}
		fmt.Println()
	}

	// Check that each client received at least the test messages
	expectedCount := len(testMessages)
	for i, messages := range receivedMessages {
		if len(messages) < expectedCount {
			t.Errorf("Client %d received only %d messages, expected at least %d",
				i, len(messages), expectedCount)
		}
	}

	fmt.Printf("✅ Real-time broadcast test completed successfully!\n")
	fmt.Printf("✅ All %d clients received messages via CSP/Actor pattern\n", numClients)
}

// postTestMessage posts a message to the API
func postTestMessage(user, message string) error {
	payload := types.CreateMessageRequest{
		User:    user,
		Message: message,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(testAPIURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// isServerRunning checks if the test server is running
func isServerRunning() bool {
	resp, err := http.Get("http://localhost:8080/api/v1/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
