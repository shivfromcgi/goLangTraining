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
	ctx    context.Context
	cancel context.CancelFunc
}

// Hub channels for Actor pattern - kept private with methods for access
type hubChannels struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

// WebSocket configuration constants
const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	// Singleton hub instance
	instance     *hubChannels
	instanceOnce sync.Once
	hubCtx       context.Context
	hubCancel    context.CancelFunc
)

// Start creates and runs the hub using Actor/CSP pattern with static functions
// Following coding standards: no OO patterns, static singleton approach
func Start() {
	instanceOnce.Do(func() {
		instance = &hubChannels{
			register:   make(chan *Client),
			unregister: make(chan *Client),
			broadcast:  make(chan []byte),
		}

		// Create context for hub lifecycle management
		hubCtx, hubCancel = context.WithCancel(context.Background())

		// Start the Actor goroutine
		go runHubActor()
	})
}

// Stop shuts down the hub gracefully
func Stop() {
	if hubCancel != nil {
		hubCancel()
	}
}

// runHubActor implements the Actor pattern using channels for communication
// Single goroutine ensures thread-safe access without need for mutexes
func runHubActor() {
	clients := make(map[*Client]bool)

	slog.Info("Starting message hub with CSP pattern")

	for {
		select {
		case <-hubCtx.Done():
			shutdownAllClients(clients)
			return

		case client := <-instance.register:
			registerClient(clients, client)

		case client := <-instance.unregister:
			unregisterClient(clients, client)

		case message := <-instance.broadcast:
			broadcastMessage(clients, message)
		}
	}
}

// registerClient adds a new client to the hub
func registerClient(clients map[*Client]bool, client *Client) {
	clients[client] = true
	clientCount := len(clients)

	slog.Info("Client registered",
		"clientID", client.ID,
		"totalClients", clientCount)

	// Start client goroutines
	go startWritePump(client)
	go startReadPump(client)
}

// unregisterClient removes a client from the hub
func unregisterClient(clients map[*Client]bool, client *Client) {
	if _, ok := clients[client]; ok {
		delete(clients, client)
		close(client.Send)
		clientCount := len(clients)

		slog.Info("Client unregistered",
			"clientID", client.ID,
			"totalClients", clientCount)
	}
}

// broadcastMessage sends a message to all connected clients (fan-out pattern)
func broadcastMessage(clients map[*Client]bool, message []byte) {
	clientCount := len(clients)
	slog.Info("Broadcasting message to clients",
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
			slog.Warn("Client removed due to full send buffer", "clientID", client.ID)
		}
	}
}

// shutdownAllClients closes all client connections
func shutdownAllClients(clients map[*Client]bool) {
	for client := range clients {
		close(client.Send)
		client.cancel()
		delete(clients, client)
	}
	slog.Info("All clients disconnected")
}

// Register adds a client to the hub via the register channel
func Register(client *Client) {
	if instance != nil {
		instance.register <- client
	}
}

// Unregister removes a client from the hub via the unregister channel
func Unregister(client *Client) {
	if instance != nil {
		instance.unregister <- client
	}
}

// Broadcast sends a message to all connected clients via the broadcast channel
func Broadcast(message []byte) {
	if instance != nil {
		select {
		case instance.broadcast <- message:
			// Message queued for broadcast
		default:
			slog.Warn("Broadcast channel full, message dropped")
		}
	}
}

// startWritePump pumps messages from the hub to the websocket connection
func startWritePump(client *Client) {
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
			err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				slog.Error("Failed to set write deadline", "error", err, "clientID", client.ID)
				return
			}

			if !ok {
				// The hub closed the channel
				err := client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					slog.Error("Failed to write close message", "error", err, "clientID", client.ID)
				}
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				slog.Error("Failed to get next writer", "error", err, "clientID", client.ID)
				return
			}

			_, err = w.Write(message)
			if err != nil {
				slog.Error("Failed to write message", "error", err, "clientID", client.ID)
				return
			}

			// Add queued messages to the current websocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				_, err = w.Write(<-client.Send)
				if err != nil {
					slog.Error("Failed to write queued message", "error", err, "clientID", client.ID)
					return
				}
			}

			if err := w.Close(); err != nil {
				slog.Error("Failed to close writer", "error", err, "clientID", client.ID)
				return
			}

		case <-ticker.C:
			err := client.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				slog.Error("Failed to set write deadline for ping", "error", err, "clientID", client.ID)
				return
			}

			// Split error handling onto two lines per style guide
			err = client.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				slog.Error("Failed to send ping", "error", err, "clientID", client.ID)
				return
			}
		}
	}
}

// startReadPump pumps messages from the websocket connection to the hub
func startReadPump(client *Client) {
	defer func() {
		Unregister(client)
		client.Conn.Close()
		client.cancel()
	}()

	client.Conn.SetReadLimit(maxMessageSize)

	err := client.Conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		slog.Error("Failed to set read deadline", "error", err, "clientID", client.ID)
		return
	}

	client.Conn.SetPongHandler(func(string) error {
		err := client.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return err
	})

	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			_, _, err := client.Conn.ReadMessage()
			if err != nil {
				// Log all errors EXCEPT close going away and close abnormal closure
				if !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.Error("WebSocket error", "error", err, "clientID", client.ID)
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
