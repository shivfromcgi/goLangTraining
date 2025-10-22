package service

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cgi.com/goLangTraining/internal/types"
)

// MessageActor implements the Actor pattern for message management
// This provides concurrency-safe access to messages and fan-out to listeners
type MessageActor struct {
	events    chan types.MessageEvent
	listeners map[string]*types.MessageListener
	messages  []types.Message
	storage   *MessageStorage
	mu        sync.RWMutex
	done      chan struct{}
	wg        sync.WaitGroup
}

// NewMessageActor creates a new MessageActor instance
func NewMessageActor(storage *MessageStorage) *MessageActor {
	return &MessageActor{
		events:    make(chan types.MessageEvent, 100), // Buffered channel for better performance
		listeners: make(map[string]*types.MessageListener),
		messages:  make([]types.Message, 0),
		storage:   storage,
		done:      make(chan struct{}),
	}
}

// Start begins the message actor's event processing loop
func (ma *MessageActor) Start() {
	ma.wg.Add(1)
	go ma.eventLoop()
}

// Stop gracefully shuts down the message actor
func (ma *MessageActor) Stop() {
	close(ma.done)
	ma.wg.Wait()

	// Clean up all listeners
	ma.mu.Lock()
	defer ma.mu.Unlock()
	for _, listener := range ma.listeners {
		close(listener.Channel)
		close(listener.Done)
	}
}

// eventLoop processes message events in sequence (CSP pattern)
func (ma *MessageActor) eventLoop() {
	defer ma.wg.Done()

	for {
		select {
		case event := <-ma.events:
			ma.handleEvent(event)
		case <-ma.done:
			return
		}
	}
}

// handleEvent processes individual message events
func (ma *MessageActor) handleEvent(event types.MessageEvent) {
	switch event.Type {
	case "create":
		ma.handleCreateMessage(event)
	case "get":
		ma.handleGetMessages(event)
	case "subscribe":
		ma.handleSubscribe(event)
	case "unsubscribe":
		ma.handleUnsubscribe(event)
	default:
		if event.Reply != nil {
			event.Reply <- types.MessageEventResponse{
				Error:   fmt.Errorf("unknown event type: %s", event.Type),
				Success: false,
			}
		}
	}
}

// handleCreateMessage processes message creation and fans out to listeners
func (ma *MessageActor) handleCreateMessage(event types.MessageEvent) {
	// First, save to persistent storage (file)
	err := ma.storage.AddMessage(event.Message.User, event.Message.Message)

	response := types.MessageEventResponse{Success: err == nil, Error: err}

	if err == nil {
		// Add to in-memory cache
		ma.mu.Lock()
		ma.messages = append(ma.messages, event.Message)

		// Fan out to all listeners in order (CSP pattern)
		for listenerID, listener := range ma.listeners {
			select {
			case listener.Channel <- event.Message:
				// Message sent successfully
			default:
				// Listener channel is full, skip (non-blocking)
				slog.Warn("Listener channel full, skipping message",
					"listener_id", listenerID,
					"message_id", event.Message.ID)
			}
		}
		ma.mu.Unlock()
	}

	if event.Reply != nil {
		event.Reply <- response
	}
}

// handleGetMessages retrieves all messages
func (ma *MessageActor) handleGetMessages(event types.MessageEvent) {
	// Read from persistent storage for most up-to-date data
	messages, err := ma.storage.ReadMessages(event.Message.TraceID)

	response := types.MessageEventResponse{
		Messages: messages,
		Error:    err,
		Success:  err == nil,
	}

	if event.Reply != nil {
		event.Reply <- response
	}
}

// handleSubscribe adds a new listener
func (ma *MessageActor) handleSubscribe(event types.MessageEvent) {
	listener := &types.MessageListener{
		ID:       event.Message.TraceID,        // Use trace ID as listener ID
		Channel:  make(chan types.Message, 10), // Buffered channel
		Done:     make(chan struct{}),
		Created:  time.Now(),
		LastPing: time.Now(),
	}

	ma.mu.Lock()
	ma.listeners[listener.ID] = listener
	ma.mu.Unlock()

	response := types.MessageEventResponse{Success: true}
	if event.Reply != nil {
		event.Reply <- response
	}
}

// handleUnsubscribe removes a listener
func (ma *MessageActor) handleUnsubscribe(event types.MessageEvent) {
	ma.mu.Lock()
	defer ma.mu.Unlock()

	if listener, exists := ma.listeners[event.Message.TraceID]; exists {
		close(listener.Channel)
		close(listener.Done)
		delete(ma.listeners, event.Message.TraceID)
	}

	response := types.MessageEventResponse{Success: true}
	if event.Reply != nil {
		event.Reply <- response
	}
}

// CreateMessage sends a create message event to the actor
func (ma *MessageActor) CreateMessage(user, messageText, traceID string) error {
	message := types.Message{
		ID:        int(time.Now().UnixNano() / 1000000),
		User:      user,
		Message:   messageText,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}

	reply := make(chan types.MessageEventResponse, 1)
	event := types.MessageEvent{
		Type:    "create",
		Message: message,
		Reply:   reply,
	}

	ma.events <- event

	response := <-reply
	return response.Error
}

// GetMessages retrieves all messages via the actor
func (ma *MessageActor) GetMessages(traceID string) ([]types.Message, error) {
	message := types.Message{TraceID: traceID}
	reply := make(chan types.MessageEventResponse, 1)

	event := types.MessageEvent{
		Type:    "get",
		Message: message,
		Reply:   reply,
	}

	ma.events <- event

	response := <-reply
	return response.Messages, response.Error
}

// Subscribe creates a new message listener
func (ma *MessageActor) Subscribe(traceID string) (*types.MessageListener, error) {
	message := types.Message{TraceID: traceID}
	reply := make(chan types.MessageEventResponse, 1)

	event := types.MessageEvent{
		Type:    "subscribe",
		Message: message,
		Reply:   reply,
	}

	ma.events <- event

	response := <-reply
	if response.Error != nil {
		return nil, response.Error
	}

	ma.mu.RLock()
	listener := ma.listeners[traceID]
	ma.mu.RUnlock()

	return listener, nil
}

// Unsubscribe removes a message listener
func (ma *MessageActor) Unsubscribe(traceID string) {
	message := types.Message{TraceID: traceID}

	event := types.MessageEvent{
		Type:    "unsubscribe",
		Message: message,
		Reply:   nil, // No response needed for unsubscribe
	}

	ma.events <- event
}
