package main

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"cgi.com/goLangTraining/src/pkg/storage"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Embedded file systems for web interface
//
//go:embed html/*
var htmlFiles embed.FS

const (
	gracefulShutdownTimeout = 30 * time.Second
	messagesFileName        = "messages.txt"
	defaultAPIVersion       = "1.0.0"
	defaultPort             = 8080
)

// WebSocket upgrader for Assignment 5
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Message represents a message in our system
type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// CreateMessageRequest represents the request body for creating a message
type CreateMessageRequest struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	TraceID string      `json:"trace_id"`
}

// HealthStatus represents the health check response structure
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// MessagesPageData represents the data passed to the messages template
type MessagesPageData struct {
	Messages    []Message `json:"messages"`
	GeneratedAt time.Time `json:"generated_at"`
	TraceID     string    `json:"trace_id"`
}

func main() {
	// Initialize structured logging first
	setupLogging()

	slog.Info("Starting CGI Go Training Service",
		"service", "cgi-go-training",
		"version", defaultAPIVersion)

	// Parse command line flags
	var (
		port        = flag.Int("port", defaultPort, "Port for HTTP server")
		user        = flag.String("user", "", "User for CLI message operations")
		message     = flag.String("message", "", "Message for CLI operations")
		clear       = flag.Bool("clear", false, "Clear all messages")
		file        = flag.String("file", "example.txt", "File path for storage operations")
		data        = flag.String("data", "", "Data to save to file")
		cliMode     = flag.Bool("cli", false, "Run in CLI mode (no web server)")
		storageDemo = flag.Bool("storage-demo", false, "Run storage demonstration")
	)
	flag.Parse()

	// If CLI mode is requested, handle CLI operations and exit
	if *cliMode {
		handleCLIOperations(*user, *message, *clear, *file, *data, *storageDemo)
		return
	}

	// Default behavior: start the full web application with all features
	startWebApplication(*port)
}

// setupLogging configures the default slog logger with structured JSON output
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "cgi-go-training",
		"version", defaultAPIVersion,
	)
	slog.SetDefault(logger)
}

// handleCLIOperations processes command-line operations and exits
func handleCLIOperations(user, message string, clear bool, file, data string, storageDemo bool) {
	fmt.Println("=== CGI Go Training Service - CLI Mode ===")

	// Handle storage demo (Assignment 2 functionality)
	if storageDemo {
		runStorageDemo(file, data)
		return
	}

	// Handle message operations (Assignment 1 functionality)
	if clear {
		clearMessages()
		return
	}

	if user != "" && message != "" {
		addMessage(user, message)
		printLast10Messages()
		return
	}

	// Show usage if no valid operation specified
	fmt.Println("\nCLI Usage:")
	fmt.Println("  Add message:    go run main.go -cli -user=alice -message='Hello World'")
	fmt.Println("  Clear messages: go run main.go -cli -clear")
	fmt.Println("  Storage demo:   go run main.go -cli -storage-demo")
	fmt.Println("  Storage demo:   go run main.go -cli -storage-demo -file=test.txt -data='Custom data'")
	fmt.Println("\nWeb Server (default):")
	fmt.Println("  Start server:   go run main.go")
	fmt.Println("  Custom port:    go run main.go -port=9090")
}

// startWebApplication starts the main web application with all features
func startWebApplication(port int) {
	fmt.Println("=== CGI Go Training Service - Web Application ===")

	mux := http.NewServeMux()

	// Setup static file server using embedded files
	staticFS, err := fs.Sub(htmlFiles, "html")
	if err != nil {
		slog.Error("Failed to create static filesystem", "error", err)
		os.Exit(1)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Web interface routes (Assignment 4)
	mux.HandleFunc("/", traceMiddleware(indexHandler))
	mux.HandleFunc("/web/messages", traceMiddleware(webMessagesHandler))

	// REST API routes (Assignment 3)
	mux.HandleFunc("/api/messages", traceMiddleware(messagesAPIHandler))
	mux.HandleFunc("/api/health", traceMiddleware(healthHandler))

	// Legacy API routes for backward compatibility
	mux.HandleFunc("/messages", traceMiddleware(messagesAPIHandler))
	mux.HandleFunc("/health", traceMiddleware(healthHandler))

	// File storage API routes (Assignment 2)
	mux.HandleFunc("/api/files", traceMiddleware(fileStorageHandler))

	// WebSocket routes (Assignment 5)
	mux.HandleFunc("/ws", traceMiddleware(websocketHandler))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		fmt.Printf("\n🚀 Starting CGI Go Training Service on port %d\n", port)
		fmt.Printf("\n📱 Web Interface:\n")
		fmt.Printf("   http://localhost:%d/                 - Home page\n", port)
		fmt.Printf("   http://localhost:%d/web/messages     - Messages page (Assignment 4)\n", port)
		fmt.Printf("\n🔌 REST API:\n")
		fmt.Printf("   GET  http://localhost:%d/api/messages  - List messages (Assignment 1)\n", port)
		fmt.Printf("   POST http://localhost:%d/api/messages  - Create message (Assignment 1)\n", port)
		fmt.Printf("   GET  http://localhost:%d/api/health    - Health check (Assignment 3)\n", port)
		fmt.Printf("   POST http://localhost:%d/api/files     - File operations (Assignment 2)\n", port)
		fmt.Printf("\n💡 Quick Test:\n")
		fmt.Printf("   curl -X POST http://localhost:%d/api/messages -H 'Content-Type: application/json' -d '{\"user\":\"demo\",\"message\":\"Hello API!\"}'\n", port)
		fmt.Printf("\n📋 CLI Operations:\n")
		fmt.Printf("   go run main.go -cli -user=alice -message='Hello CLI'\n")
		fmt.Printf("   go run main.go -cli -storage-demo\n")
		fmt.Printf("\nPress Ctrl+C to stop the server...\n\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err, "port", port)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	fmt.Printf("\n🛑 Received signal %s, shutting down server...\n", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server shutdown failed", "error", err)
		os.Exit(1)
	}

	fmt.Println("✅ CGI Go Training Service stopped gracefully")
}

// Assignment 1: Message System Functions

func addMessage(user, message string) error {
	f, err := os.OpenFile(messagesFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	line := fmt.Sprintf("[%s] %s: %s\n", timestamp, user, message)
	_, err = f.WriteString(line)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Message added: %s: %s\n", user, message)
	return nil
}

func clearMessages() {
	err := os.Truncate(messagesFileName, 0)
	if err != nil && !os.IsNotExist(err) {
		fmt.Printf("❌ Error clearing messages: %v\n", err)
		return
	}
	fmt.Println("✅ All messages cleared")
}

func printLast10Messages() {
	f, err := os.Open(messagesFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("📭 No messages found.")
			return
		}
		fmt.Printf("❌ Error reading messages: %v\n", err)
		return
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ Error scanning messages: %v\n", err)
		return
	}

	fmt.Println("\n📨 Last 10 Messages:")
	start := 0
	if len(lines) > 10 {
		start = len(lines) - 10
	}
	for _, line := range lines[start:] {
		fmt.Println("  " + line)
	}
}

func readMessagesForAPI(traceID string) ([]Message, error) {
	f, err := os.Open(messagesFileName)
	if err != nil {
		if os.IsNotExist(err) {
			return []Message{}, nil
		}
		return []Message{}, err
	}
	defer f.Close()

	var messages []Message
	scanner := bufio.NewScanner(f)
	id := 1

	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			// Parse format: [timestamp] user: message
			message := parseMessageLine(line, id, traceID)
			if message != nil {
				messages = append(messages, *message)
				id++
			}
		}
	}

	return messages, scanner.Err()
}

// getLastMessages returns the last N messages for WebSocket (Assignment 5)
func getLastMessages(ctx context.Context, limit int) ([]Message, error) {
	traceID := ctx.Value("traceID").(string)

	// Read all messages first
	allMessages, err := readMessagesForAPI(traceID)
	if err != nil {
		return []Message{}, err
	}

	// Return last N messages
	if len(allMessages) <= limit {
		return allMessages, nil
	}

	startIndex := len(allMessages) - limit
	return allMessages[startIndex:], nil
}

func parseMessageLine(line string, id int, traceID string) *Message {
	// Simple parsing for [timestamp] user: message format
	if len(line) < 22 { // Minimum length for timestamp + user + message
		return nil
	}

	// Find end of timestamp (look for "] ")
	timestampEnd := -1
	for i := 0; i < len(line)-1; i++ {
		if line[i] == ']' && line[i+1] == ' ' {
			timestampEnd = i
			break
		}
	}

	if timestampEnd == -1 {
		return nil
	}

	remaining := line[timestampEnd+2:] // Skip "] "

	// Find ": " separator
	colonIndex := -1
	for i := 0; i < len(remaining)-1; i++ {
		if remaining[i] == ':' && remaining[i+1] == ' ' {
			colonIndex = i
			break
		}
	}

	if colonIndex == -1 {
		return nil
	}

	user := remaining[:colonIndex]
	messageText := remaining[colonIndex+2:]

	// Parse timestamp
	timestampStr := line[1:timestampEnd] // Remove [ and ]
	timestamp, err := time.Parse("2006-01-02 15:04:05", timestampStr)
	if err != nil {
		timestamp = time.Now() // Fallback
	}

	return &Message{
		ID:        id,
		User:      user,
		Message:   messageText,
		Timestamp: timestamp,
		TraceID:   traceID,
	}
}

// Assignment 2: Storage Demo Function

func runStorageDemo(filePath, data string) {
	fmt.Println("\n🗄️  Running Storage Demonstration (Assignment 2)")

	traceID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "traceID", traceID)

	// Prepare content
	content := data
	if content == "" {
		content = fmt.Sprintf("CGI Go Training - Storage Demo\nGenerated at: %s\nTrace ID: %s",
			time.Now().Format(time.RFC3339), traceID)
	}

	fmt.Printf("\n📝 Saving data to file: %s\n", filePath)

	// Save data
	err := storage.SaveData(ctx, filePath, content)
	if err != nil {
		fmt.Printf("❌ Failed to save data: %v\n", err)
		return
	}

	fmt.Println("✅ Data saved successfully")

	// Read data back
	fmt.Printf("\n📖 Reading data from file: %s\n", filePath)
	readContent, err := storage.ReadData(ctx, filePath)
	if err != nil {
		fmt.Printf("❌ Failed to read data: %v\n", err)
		return
	}

	fmt.Println("✅ Data read successfully")
	fmt.Printf("\n📄 File Content Preview:\n%s\n", readContent)

	fmt.Printf("\n🎯 Storage demonstration completed with trace ID: %s\n", traceID)
}

// HTTP Middleware and Handlers

func traceMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "traceID", traceID)

		slog.InfoContext(ctx, "Incoming HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"traceID", traceID)

		w.Header().Set("X-Trace-ID", traceID)
		next(w, r.WithContext(ctx))
	}
}

// Assignment 4: Web Interface Handlers

func indexHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	indexHTML, err := htmlFiles.ReadFile("html/index.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read index.html", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Served static index page", "traceID", traceID)
	w.Write(indexHTML)
}

func webMessagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	messages, err := readMessagesForAPI(traceID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read messages for web page", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := MessagesPageData{
		Messages:    messages,
		GeneratedAt: time.Now(),
		TraceID:     traceID,
	}

	tmpl, err := template.ParseFS(htmlFiles, "html/messages.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to parse messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to execute messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slog.InfoContext(r.Context(), "Served dynamic messages page",
		"message_count", len(messages),
		"traceID", traceID)
}

// Assignment 3: REST API Handlers

func messagesAPIHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		createMessageAPI(w, r, traceID)
	case http.MethodGet:
		getMessagesAPI(w, r, traceID)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", traceID)
	}
}

func createMessageAPI(w http.ResponseWriter, r *http.Request, traceID string) {
	var req CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to decode request body", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", traceID)
		return
	}

	if req.User == "" || req.Message == "" {
		respondWithError(w, http.StatusBadRequest, "User and message are required", traceID)
		return
	}

	// Use the same message storage as CLI
	err = addMessage(req.User, req.Message)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to save message", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusInternalServerError, "Failed to save message", traceID)
		return
	}

	message := Message{
		ID:        int(time.Now().UnixNano() / 1000000),
		User:      req.User,
		Message:   req.Message,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}

	slog.InfoContext(r.Context(), "Message created successfully",
		"user", req.User,
		"message_id", message.ID,
		"traceID", traceID)

	respondWithSuccess(w, http.StatusCreated, message, traceID)
}

func getMessagesAPI(w http.ResponseWriter, r *http.Request, traceID string) {
	messages, err := readMessagesForAPI(traceID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read messages", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusInternalServerError, "Failed to read messages", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Messages retrieved successfully",
		"message_count", len(messages),
		"traceID", traceID)

	respondWithSuccess(w, http.StatusOK, messages, traceID)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)
	w.Header().Set("Content-Type", "application/json")

	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   defaultAPIVersion,
	}

	respondWithSuccess(w, http.StatusOK, health, traceID)
}

// Assignment 2: File Storage API Handler

func fileStorageHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed", traceID)
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
		Data     string `json:"data"`
		Action   string `json:"action"` // "save" or "read"
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload", traceID)
		return
	}

	if req.FilePath == "" || req.Action == "" {
		respondWithError(w, http.StatusBadRequest, "file_path and action are required", traceID)
		return
	}

	ctx := context.WithValue(r.Context(), "traceID", traceID)

	switch req.Action {
	case "save":
		if req.Data == "" {
			respondWithError(w, http.StatusBadRequest, "data is required for save action", traceID)
			return
		}

		err = storage.SaveData(ctx, req.FilePath, req.Data)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to save file", "error", err, "traceID", traceID)
			respondWithError(w, http.StatusInternalServerError, "Failed to save file", traceID)
			return
		}

		respondWithSuccess(w, http.StatusOK, map[string]string{
			"message":   "File saved successfully",
			"file_path": req.FilePath,
		}, traceID)

	case "read":
		content, err := storage.ReadData(ctx, req.FilePath)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to read file", "error", err, "traceID", traceID)
			respondWithError(w, http.StatusInternalServerError, "Failed to read file", traceID)
			return
		}

		respondWithSuccess(w, http.StatusOK, map[string]string{
			"content":   content,
			"file_path": req.FilePath,
		}, traceID)

	default:
		respondWithError(w, http.StatusBadRequest, "action must be 'save' or 'read'", traceID)
	}
}

// Utility functions for HTTP responses

func respondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}, traceID string) {
	w.WriteHeader(statusCode)
	response := Response{
		Success: true,
		Data:    data,
		TraceID: traceID,
	}
	json.NewEncoder(w).Encode(response)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string, traceID string) {
	w.WriteHeader(statusCode)
	response := Response{
		Success: false,
		Error:   message,
		TraceID: traceID,
	}
	json.NewEncoder(w).Encode(response)
}

// WebSocket handler for Assignment 5
func websocketHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.Context().Value("traceID").(string)
	slog.Info("WebSocket connection requested", "traceID", traceID)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade to WebSocket", "error", err, "traceID", traceID)
		return
	}
	defer conn.Close()

	slog.Info("WebSocket connection established", "traceID", traceID)

	// Read last 10 messages from storage
	ctx := context.WithValue(r.Context(), "traceID", traceID)
	messages, err := getLastMessages(ctx, 10)
	if err != nil {
		slog.Error("Failed to read messages", "error", err, "traceID", traceID)
		conn.WriteMessage(websocket.TextMessage, []byte("Error reading messages"))
		return
	}

	// Send each message to the client
	for _, message := range messages {
		messageJSON, err := json.Marshal(message)
		if err != nil {
			slog.Error("Failed to marshal message", "error", err, "traceID", traceID)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			slog.Error("Failed to send message over WebSocket", "error", err, "traceID", traceID)
			break
		}
	}

	slog.Info("Sent messages over WebSocket", "count", len(messages), "traceID", traceID)

	// Send completion message and close
	conn.WriteMessage(websocket.TextMessage, []byte("All messages sent. Connection will close."))
	conn.Close()
}
