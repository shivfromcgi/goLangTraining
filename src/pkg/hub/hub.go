package hub

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
	ctx    context.Context
	cancel context.CancelFunc
}

// Hub implements the Actor/CSP pattern for message broadcasting
// It maintains active WebSocket connections and broadcasts messages to all clients
type Hub struct {
	// Channel for registering new clients
	Register chan *Client

	// Channel for unregistering clients
	Unregister chan *Client

	// Channel for broadcasting messages to all clients
	Broadcast chan []byte

	// Map of active clients
	clients map[*Client]bool

	// Channel for graceful shutdown
	shutdown chan struct{}

	// Logger for hub operations
	logger *slog.Logger

	// Mutex for protecting clients map during read operations
	clientsMu sync.RWMutex
}

// WebSocket configuration constants
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

// NewHub creates a new Hub instance implementing the Actor pattern
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
		clients:    make(map[*Client]bool),
		shutdown:   make(chan struct{}),
		logger:     logger,
	}
}

// Run starts the Hub's main event loop (Actor pattern implementation)
// This method implements the CSP pattern using channels for communication
func (h *Hub) Run() {
	h.logger.Info("Starting message hub with CSP pattern")

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case message := <-h.Broadcast:
			h.broadcastMessage(message)

		case <-h.shutdown:
			h.logger.Info("Hub shutting down")
			h.shutdownAllClients()
			return
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.clientsMu.Lock()
	h.clients[client] = true
	clientCount := len(h.clients)
	h.clientsMu.Unlock()

	h.logger.Info("Client registered",
		"clientID", client.ID,
		"totalClients", clientCount)

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.clientsMu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.Send)
		clientCount := len(h.clients)
		h.clientsMu.Unlock()

		h.logger.Info("Client unregistered",
			"clientID", client.ID,
			"totalClients", clientCount)
	} else {
		h.clientsMu.Unlock()
	}
}

// broadcastMessage sends a message to all connected clients (fan-out pattern)
func (h *Hub) broadcastMessage(message []byte) {
	h.clientsMu.RLock()
	clientCount := len(h.clients)
	h.logger.Info("Broadcasting message to clients",
		"clientCount", clientCount,
		"messageSize", len(message))

	// Fan out message to all clients using CSP pattern
	for client := range h.clients {
		select {
		case client.Send <- message:
			// Message sent successfully
		default:
			// Client's send channel is full, remove the client
			delete(h.clients, client)
			close(client.Send)
			h.logger.Warn("Client removed due to full send buffer", "clientID", client.ID)
		}
	}
	h.clientsMu.RUnlock()
}

// shutdownAllClients closes all client connections
func (h *Hub) shutdownAllClients() {
	h.clientsMu.Lock()
	for client := range h.clients {
		close(client.Send)
		client.cancel()
		delete(h.clients, client)
	}
	h.clientsMu.Unlock()
	h.logger.Info("All clients disconnected")
}

// Shutdown gracefully shuts down the hub
func (h *Hub) Shutdown() {
	close(h.shutdown)
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(message []byte) {
	select {
	case h.Broadcast <- message:
		// Message queued for broadcast
	default:
		h.logger.Warn("Broadcast channel full, message dropped")
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	return len(h.clients)
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
		c.cancel()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return

		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.Hub.logger.Error("Failed to get next writer", "error", err, "clientID", c.ID)
				return
			}

			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				c.Hub.logger.Error("Failed to close writer", "error", err, "clientID", c.ID)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.Hub.logger.Error("Failed to send ping", "error", err, "clientID", c.ID)
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
		c.cancel()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.Hub.logger.Error("WebSocket error", "error", err, "clientID", c.ID)
				}
				return
			}

			// Echo received messages back to all clients (optional behavior)
			c.Hub.BroadcastMessage(message)
		}
	}
}

// NewClient creates a new WebSocket client
func NewClient(id string, conn *websocket.Conn, hub *Hub) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		ID:     id,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    hub,
		ctx:    ctx,
		cancel: cancel,
	}
}
