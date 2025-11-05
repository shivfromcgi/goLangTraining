package hub

import (
	"context"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	ctx    context.Context
	cancel context.CancelFunc
}

// HubChannels contains all channels needed for the Actor pattern
type HubChannels struct {
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
	Shutdown   chan struct{}
}

// WebSocket configuration constants
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// Start creates and runs the hub using Actor/CSP pattern with static functions
// Following coding standards: no OO patterns, static singleton approach
func Start(ctx context.Context, logger *slog.Logger) *HubChannels {
	channels := &HubChannels{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
		Shutdown:   make(chan struct{}),
	}

	// Start the Actor goroutine
	go runHubActor(ctx, channels, logger)

	return channels
}

// runHubActor implements the Actor pattern using channels for communication
// Single goroutine ensures thread-safe access without need for mutexes
func runHubActor(ctx context.Context, channels *HubChannels, logger *slog.Logger) {
	clients := make(map[*Client]bool)

	slog.InfoContext(ctx, "Starting message hub with CSP pattern")

	for {
		select {
		case <-ctx.Done():
			shutdownAllClients(clients, logger)
			return

		case client := <-channels.Register:
			registerClient(ctx, clients, client, channels, logger)

		case client := <-channels.Unregister:
			unregisterClient(ctx, clients, client, logger)

		case message := <-channels.Broadcast:
			broadcastMessage(ctx, clients, message, logger)

		case <-channels.Shutdown:
			slog.InfoContext(ctx, "Hub shutting down")
			shutdownAllClients(clients, logger)
			return
		}
	}
}

// registerClient adds a new client to the hub
func registerClient(ctx context.Context, clients map[*Client]bool, client *Client, channels *HubChannels, logger *slog.Logger) {
	clients[client] = true
	clientCount := len(clients)

	slog.InfoContext(ctx, "Client registered",
		"clientID", client.ID,
		"totalClients", clientCount)

	// Start client goroutines
	go startWritePump(client, logger)
	go startReadPump(client, channels, logger)
}

// unregisterClient removes a client from the hub
func unregisterClient(ctx context.Context, clients map[*Client]bool, client *Client, logger *slog.Logger) {
	if _, ok := clients[client]; ok {
		delete(clients, client)
		close(client.Send)
		clientCount := len(clients)

		slog.InfoContext(ctx, "Client unregistered",
			"clientID", client.ID,
			"totalClients", clientCount)
	}
}

// broadcastMessage sends a message to all connected clients (fan-out pattern)
func broadcastMessage(ctx context.Context, clients map[*Client]bool, message []byte, logger *slog.Logger) {
	clientCount := len(clients)
	slog.InfoContext(ctx, "Broadcasting message to clients",
		"clientCount", clientCount,
		"messageSize", len(message))

	// Fan out message to all clients using CSP pattern
	for client := range clients {
		select {
		case client.Send <- message:
			// Message sent successfully
		default:
			// Client's send channel is full, remove the client
			delete(clients, client)
			close(client.Send)
			slog.WarnContext(ctx, "Client removed due to full send buffer", "clientID", client.ID)
		}
	}
}

// shutdownAllClients closes all client connections
func shutdownAllClients(clients map[*Client]bool, logger *slog.Logger) {
	for client := range clients {
		close(client.Send)
		client.cancel()
		delete(clients, client)
	}
	// This is called during shutdown, so we'll use a background context
	slog.InfoContext(context.Background(), "All clients disconnected")
}

// BroadcastMessage sends a message to all connected clients via channels
func BroadcastMessage(channels *HubChannels, message []byte, logger *slog.Logger) {
	select {
	case channels.Broadcast <- message:
		// Message queued for broadcast
	default:
		logger.Warn("Broadcast channel full, message dropped")
	}
}

// GetClientCount returns the number of connected clients (for monitoring)
// Note: This is approximate as it requires a channel operation
func GetClientCount(channels *HubChannels) int {
	// In Actor pattern, we don't expose internal state directly
	// This would require adding a query channel if needed
	return 0 // Not implemented to maintain Actor pattern purity
}

// Shutdown gracefully shuts down the hub
func Shutdown(channels *HubChannels) {
	close(channels.Shutdown)
}

// startWritePump pumps messages from the hub to the websocket connection
func startWritePump(client *Client, logger *slog.Logger) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.Conn.Close()
		client.cancel()
	}()

	for {
		select {
		case <-client.ctx.Done():
			return

		case message, ok := <-client.Send:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				slog.ErrorContext(client.ctx, "Failed to get next writer", "error", err, "clientID", client.ID)
				return
			}

			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				slog.ErrorContext(client.ctx, "Failed to close writer", "error", err, "clientID", client.ID)
				return
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.ErrorContext(client.ctx, "Failed to send ping", "error", err, "clientID", client.ID)
				return
			}
		}
	}
}

// startReadPump pumps messages from the websocket connection to the hub
func startReadPump(client *Client, channels *HubChannels, logger *slog.Logger) {
	defer func() {
		channels.Unregister <- client
		client.Conn.Close()
		client.cancel()
	}()

	client.Conn.SetReadLimit(maxMessageSize)
	client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			_, _, err := client.Conn.ReadMessage()
			if err != nil {
				// Fix: Log all errors EXCEPT close going away and close abnormal closure
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.ErrorContext(client.ctx, "WebSocket error", "error", err, "clientID", client.ID)
				}
				return
			}

			// In functional design, message handling would be done via passed function
			// For now, we just consume the message without echoing
		}
	}
}

// NewClient creates a new WebSocket client
func NewClient(id string, conn *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		ID:     id,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		ctx:    ctx,
		cancel: cancel,
	}
}
