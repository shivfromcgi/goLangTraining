package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"cgi.com/goLangTraining/internal/service"
	"cgi.com/goLangTraining/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMessageActorConcurrentSafety tests the Actor pattern for concurrent safety
func TestMessageActorConcurrentSafety(t *testing.T) {
	// Setup: Create a clean test environment
	testFile := getTestFileName()
	setupTestEnvironment(t)
	defer func() {
		os.Remove(testFile)
		cleanupTestEnvironment(t)
	}()

	storage := service.NewMessageStorage(testFile)
	actor := service.NewMessageActor(storage)
	actor.Start()
	defer actor.Stop()

	// Test parameters
	numGoroutines := 50
	messagesPerGoroutine := 10
	totalMessages := numGoroutines * messagesPerGoroutine

	var wg sync.WaitGroup
	errors := make(chan error, totalMessages)
	messageIDs := make(chan int, totalMessages)

	t.Logf("Starting concurrent test with %d goroutines, %d messages each", numGoroutines, messagesPerGoroutine)

	// Launch multiple goroutines to create messages concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				traceID := uuid.New().String()
				user := fmt.Sprintf("user_%d", goroutineID)
				message := fmt.Sprintf("message_%d_%d", goroutineID, j)

				err := actor.CreateMessage(user, message, traceID)
				if err != nil {
					errors <- err
					return
				}

				// Generate a unique message ID for tracking
				messageID := goroutineID*messagesPerGoroutine + j
				messageIDs <- messageID
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errors)
	close(messageIDs)

	// Verify no errors occurred
	for err := range errors {
		t.Errorf("Error during concurrent message creation: %v", err)
	}

	// Verify all messages were created
	var receivedIDs []int
	for id := range messageIDs {
		receivedIDs = append(receivedIDs, id)
	}

	require.Len(t, receivedIDs, totalMessages, "Not all messages were created successfully")

	// Verify messages can be retrieved
	traceID := uuid.New().String()
	messages, err := actor.GetMessages(traceID)
	require.NoError(t, err, "Failed to retrieve messages after concurrent creation")

	// Verify we have at least the expected number of messages
	// (could be more if other tests have run)
	assert.GreaterOrEqual(t, len(messages), totalMessages,
		"Retrieved message count should be at least the number we created")

	t.Logf("Successfully created and retrieved %d messages concurrently", len(messages))
}

// TestMessageActorListenerFanOut tests the fan-out pattern with multiple listeners
func TestMessageActorListenerFanOut(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	storage := service.NewMessageStorage("test_messages.txt")
	actor := service.NewMessageActor(storage)
	actor.Start()
	defer actor.Stop()

	numListeners := 10
	numMessages := 20
	listeners := make([]*types.MessageListener, numListeners)
	receivedMessages := make([][]types.Message, numListeners)

	// Create multiple listeners
	for i := 0; i < numListeners; i++ {
		traceID := fmt.Sprintf("listener_%d_%s", i, uuid.New().String())
		listener, err := actor.Subscribe(traceID)
		require.NoError(t, err, "Failed to subscribe listener %d", i)
		listeners[i] = listener
		receivedMessages[i] = make([]types.Message, 0)
	}

	// Start goroutines to collect messages from each listener
	var wg sync.WaitGroup
	for i := 0; i < numListeners; i++ {
		wg.Add(1)
		go func(listenerIndex int) {
			defer wg.Done()

			timeout := time.After(5 * time.Second)
			messagesReceived := 0

			for messagesReceived < numMessages {
				select {
				case message, ok := <-listeners[listenerIndex].Channel:
					if !ok {
						t.Errorf("Listener %d channel closed prematurely", listenerIndex)
						return
					}
					receivedMessages[listenerIndex] = append(receivedMessages[listenerIndex], message)
					messagesReceived++

				case <-timeout:
					t.Errorf("Listener %d timed out waiting for messages. Received: %d, Expected: %d",
						listenerIndex, messagesReceived, numMessages)
					return
				}
			}
		}(i)
	}

	// Create messages after listeners are set up
	for i := 0; i < numMessages; i++ {
		traceID := uuid.New().String()
		user := fmt.Sprintf("test_user_%d", i)
		message := fmt.Sprintf("test_message_%d", i)

		err := actor.CreateMessage(user, message, traceID)
		require.NoError(t, err, "Failed to create message %d", i)

		// Small delay to ensure proper ordering
		time.Sleep(1 * time.Millisecond)
	}

	// Wait for all listeners to receive messages
	wg.Wait()

	// Verify all listeners received all messages
	for i := 0; i < numListeners; i++ {
		assert.Len(t, receivedMessages[i], numMessages,
			"Listener %d should have received %d messages", i, numMessages)

		// Verify message ordering (messages should arrive in the order they were sent)
		for j := 1; j < len(receivedMessages[i]); j++ {
			prev := receivedMessages[i][j-1]
			curr := receivedMessages[i][j]
			assert.True(t, prev.Timestamp.Before(curr.Timestamp) || prev.Timestamp.Equal(curr.Timestamp),
				"Messages not received in order for listener %d: message %d timestamp %v should be before message %d timestamp %v",
				i, j-1, prev.Timestamp, j, curr.Timestamp)
		}
	}

	// Cleanup listeners
	for i := 0; i < numListeners; i++ {
		actor.Unsubscribe(fmt.Sprintf("listener_%d_%s", i, uuid.New().String()))
	}

	t.Logf("Successfully tested fan-out pattern with %d listeners receiving %d messages each",
		numListeners, numMessages)
}

// TestMessageActorSequentialConsistency tests that messages are processed in order
func TestMessageActorSequentialConsistency(t *testing.T) {
	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	storage := service.NewMessageStorage("test_messages.txt")
	actor := service.NewMessageActor(storage)
	actor.Start()
	defer actor.Stop()

	numMessages := 100
	var wg sync.WaitGroup
	messageTimes := make([]time.Time, numMessages)

	// Create messages rapidly to test ordering
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(messageIndex int) {
			defer wg.Done()

			traceID := uuid.New().String()
			user := fmt.Sprintf("user_%d", messageIndex)
			message := fmt.Sprintf("message_%d", messageIndex)

			start := time.Now()
			err := actor.CreateMessage(user, message, traceID)
			require.NoError(t, err, "Failed to create message %d", messageIndex)
			messageTimes[messageIndex] = start
		}(i)
	}

	wg.Wait()

	// Retrieve all messages and verify they were processed in a consistent order
	traceID := uuid.New().String()
	messages, err := actor.GetMessages(traceID)
	require.NoError(t, err, "Failed to retrieve messages")

	// Verify we have at least the expected number of messages
	assert.GreaterOrEqual(t, len(messages), numMessages,
		"Should have at least %d messages", numMessages)

	// Check that messages are stored in chronological order
	// (Note: Due to concurrent creation, exact order might vary, but timestamps should be sequential)
	recentMessages := messages[len(messages)-numMessages:]
	for i := 1; i < len(recentMessages); i++ {
		prev := recentMessages[i-1]
		curr := recentMessages[i]

		// Allow for small time differences due to concurrent execution
		timeDiff := curr.Timestamp.Sub(prev.Timestamp)
		assert.True(t, timeDiff >= -100*time.Millisecond,
			"Message timestamps should be roughly in order: message %d (%v) should not be significantly before message %d (%v)",
			i, curr.Timestamp, i-1, prev.Timestamp)
	}

	t.Logf("Successfully verified sequential consistency for %d messages", numMessages)
}

// TestMessageActorStressTest performs a comprehensive stress test
func TestMessageActorStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	setupTestEnvironment(t)
	defer cleanupTestEnvironment(t)

	storage := service.NewMessageStorage("test_messages.txt")
	actor := service.NewMessageActor(storage)
	actor.Start()
	defer actor.Stop()

	// Stress test parameters
	numProducers := 20
	numConsumers := 5
	messagesPerProducer := 50
	totalMessages := numProducers * messagesPerProducer

	var wg sync.WaitGroup
	errors := make(chan error, totalMessages)

	// Consumer setup
	consumers := make([]*types.MessageListener, numConsumers)
	consumerMessages := make([][]types.Message, numConsumers)

	for i := 0; i < numConsumers; i++ {
		traceID := fmt.Sprintf("consumer_%d_%s", i, uuid.New().String())
		listener, err := actor.Subscribe(traceID)
		require.NoError(t, err, "Failed to subscribe consumer %d", i)
		consumers[i] = listener
		consumerMessages[i] = make([]types.Message, 0)

		// Start consumer goroutine
		wg.Add(1)
		go func(consumerIndex int) {
			defer wg.Done()

			timeout := time.After(30 * time.Second)

			for {
				select {
				case message, ok := <-consumers[consumerIndex].Channel:
					if !ok {
						return
					}
					consumerMessages[consumerIndex] = append(consumerMessages[consumerIndex], message)

				case <-timeout:
					t.Logf("Consumer %d timed out, received %d messages",
						consumerIndex, len(consumerMessages[consumerIndex]))
					return

				case <-consumers[consumerIndex].Done:
					return
				}
			}
		}(i)
	}

	t.Logf("Starting stress test: %d producers, %d consumers, %d total messages",
		numProducers, numConsumers, totalMessages)

	// Producer goroutines
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()

			for j := 0; j < messagesPerProducer; j++ {
				traceID := uuid.New().String()
				user := fmt.Sprintf("producer_%d", producerID)
				message := fmt.Sprintf("stress_message_%d_%d", producerID, j)

				err := actor.CreateMessage(user, message, traceID)
				if err != nil {
					errors <- err
					return
				}

				// Random small delay to simulate real-world usage
				time.Sleep(time.Duration(j%5) * time.Millisecond)
			}
		}(i)
	}

	// Wait for all producers to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		t.Log("All producers completed successfully")
	case <-time.After(45 * time.Second):
		t.Fatal("Stress test timed out")
	}

	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Errorf("Producer error: %v", err)
		errorCount++
	}

	assert.Equal(t, 0, errorCount, "No errors should occur during stress test")

	// Give consumers time to receive all messages
	time.Sleep(1 * time.Second)

	// Cleanup consumers
	for i := 0; i < numConsumers; i++ {
		actor.Unsubscribe(fmt.Sprintf("consumer_%d_%s", i, uuid.New().String()))
	}

	// Verify final state
	traceID := uuid.New().String()
	finalMessages, err := actor.GetMessages(traceID)
	require.NoError(t, err, "Failed to get final message count")

	t.Logf("Stress test completed: %d messages created, %d messages in storage",
		totalMessages, len(finalMessages))

	// Verify we have at least the expected number of messages
	assert.GreaterOrEqual(t, len(finalMessages), totalMessages,
		"Should have at least %d messages in final storage", totalMessages)
}

// Benchmark tests for performance analysis

func BenchmarkMessageActorCreate(b *testing.B) {
	benchFile := getBenchFileName()
	setupBenchmarkEnvironment(b)
	defer cleanupBenchmarkEnvironment(b, benchFile)

	storage := service.NewMessageStorage(benchFile)
	actor := service.NewMessageActor(storage)
	actor.Start()
	defer actor.Stop()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			traceID := uuid.New().String()
			user := fmt.Sprintf("bench_user_%d", i)
			message := fmt.Sprintf("bench_message_%d", i)

			err := actor.CreateMessage(user, message, traceID)
			if err != nil {
				b.Errorf("Failed to create message: %v", err)
			}
			i++
		}
	})
}

func BenchmarkMessageActorGet(b *testing.B) {
	benchFile := getBenchFileName()
	setupBenchmarkEnvironment(b)
	defer cleanupBenchmarkEnvironment(b, benchFile)

	storage := service.NewMessageStorage(benchFile)
	actor := service.NewMessageActor(storage)
	actor.Start()
	defer actor.Stop()

	// Pre-populate with some messages
	for i := 0; i < 100; i++ {
		traceID := uuid.New().String()
		user := fmt.Sprintf("setup_user_%d", i)
		message := fmt.Sprintf("setup_message_%d", i)
		actor.CreateMessage(user, message, traceID)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			traceID := uuid.New().String()
			_, err := actor.GetMessages(traceID)
			if err != nil {
				b.Errorf("Failed to get messages: %v", err)
			}
		}
	})
}

// Helper functions for test setup and cleanup

func setupTestEnvironment(t *testing.T) {
	// Create a temporary test file
	testFile := fmt.Sprintf("test_messages_%d.txt", time.Now().UnixNano())
	t.Logf("Using test file: %s", testFile)
}

func cleanupTestEnvironment(t *testing.T) {
	// Remove test file
	testFile := fmt.Sprintf("test_messages_%d.txt", time.Now().UnixNano())
	if err := os.Remove(testFile); err != nil && !os.IsNotExist(err) {
		t.Logf("Warning: Failed to cleanup test file %s: %v", testFile, err)
	}
}

func getTestFileName() string {
	return fmt.Sprintf("test_messages_%d.txt", time.Now().UnixNano())
}

func getBenchFileName() string {
	return fmt.Sprintf("bench_messages_%d.txt", time.Now().UnixNano())
}

func setupBenchmarkEnvironment(b *testing.B) {
	// Benchmark setup - no specific file operations needed here
}

func cleanupBenchmarkEnvironment(b *testing.B, filename string) {
	// Remove benchmark file
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		b.Logf("Warning: Failed to cleanup benchmark file %s: %v", filename, err)
	}
}
