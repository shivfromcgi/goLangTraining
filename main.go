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
)

// Embedded file systems for Assignment 4
//
//go:embed html/*
var htmlFiles embed.FS

const (
	gracefulShutdownTimeout = 30 * time.Second
	messagesFileName        = "messages.txt"
	defaultAPIVersion       = "1.0.0"
)

// Message represents a JSON message structure for the API
type Message struct {
	ID        int       `json:"id"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id"`
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

// MessagesPageData represents the data passed to the messages template (Assignment 4)
type MessagesPageData struct {
	Messages    []Message `json:"messages"`
	GeneratedAt time.Time `json:"generated_at"`
	TraceID     string    `json:"trace_id"`
}

func main() {
	// Initialize structured logging first
	setupLogging()

	slog.Info("Starting go-training-service",
		"service", "go-training-service",
		"version", defaultAPIVersion)

	// Parse CLI flags in main - this is where startup and configuration belongs
	assignmentFlag := flag.String("assignment", "", "Assignment to run: 'assignment1', 'assignment2', 'assignment3', or 'assignment4'")
	userFlag := flag.String("user", "", "User ID for message system (Assignment 1)")
	messageFlag := flag.String("message", "", "Message to append (Assignment 1)")
	clearFlag := flag.Bool("clear", false, "Clear all messages (Assignment 1)")
	filePathFlag := flag.String("file", "example.txt", "File path for storage operations (Assignment 2)")
	dataFlag := flag.String("data", "", "Data to save to file (Assignment 2)")
	portFlag := flag.Int("port", 8080, "Port for HTTP server (Assignment 3)")

	flag.Parse()

	if *assignmentFlag == "" {
		printUsage()
		os.Exit(1)
	}

	var err error
	switch *assignmentFlag {
	case "assignment1":
		err = runAssignment1(*userFlag, *messageFlag, *clearFlag)
	case "assignment2":
		err = runAssignment2(*filePathFlag, *dataFlag)
	case "assignment3":
		err = runAssignment3(*portFlag)
	case "assignment4":
		err = runAssignment4(*portFlag)
	default:
		fmt.Printf("Unknown assignment: %s\n", *assignmentFlag)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		slog.Error("Service failed to start", "error", err)
		os.Exit(1)
	}
}

// setupLogging configures the default slog logger with structured JSON output
func setupLogging() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})).With(
		"service", "go-training-service",
		"version", defaultAPIVersion,
	)

	slog.SetDefault(logger)
}

// printUsage displays comprehensive help information for the CLI application
func printUsage() {
	fmt.Println("Go Training - Unified Assignments")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  Assignment 1 - Message System:")
	fmt.Println("    go run main.go -assignment=assignment1 -user=<userID> -message=<text>")
	fmt.Println("    go run main.go -assignment=assignment1 -clear")
	fmt.Println("")
	fmt.Println("  Assignment 2 - Advanced Storage:")
	fmt.Println("    go run main.go -assignment=assignment2")
	fmt.Println("    go run main.go -assignment=assignment2 -file=<filepath> -data=<content>")
	fmt.Println("")
	fmt.Println("  Assignment 3 - HTTP JSON API:")
	fmt.Println("    go run main.go -assignment=assignment3")
	fmt.Println("    go run main.go -assignment=assignment3 -port=<port>")
	fmt.Println("")
	fmt.Println("  Assignment 4 - Web Pages:")
	fmt.Println("    go run main.go -assignment=assignment4")
	fmt.Println("    go run main.go -assignment=assignment4 -port=<port>")
}

func runAssignment1(user, message string, clear bool) error {
	fmt.Println("=== Running Assignment 1: Message System ===")

	if clear {
		err := os.Truncate(messagesFileName, 0)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("error clearing file: %w", err)
		}
		fmt.Println("All messages cleared.")
		return nil
	}

	if user == "" || message == "" {
		return fmt.Errorf("both -user and -message are required for Assignment 1")
	}

	err := appendMessage(user, message)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}

	fmt.Println("\nLast 10 Messages:")
	err = printLast10Messages()
	if err != nil {
		fmt.Printf("Error reading messages: %v\n", err)
	}

	return nil
}

func runAssignment2(filePath, data string) error {
	fmt.Println("=== Running Assignment 2: Advanced Storage System ===")

	traceID := uuid.New().String()
	ctx := context.WithValue(context.Background(), "traceID", traceID)

	slog.InfoContext(ctx, "Assignment 2 starting", "traceID", traceID)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var fileContent string
	if data != "" {
		fileContent = data
	} else {
		fileContent = "Hello, CGI Go Academy! - Assignment 2"
	}
	fileContent += "\nTimestamp: " + time.Now().Format(time.RFC3339)

	err := storage.SaveData(ctx, filePath, fileContent)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save data", "error", err, "filePath", filePath, "traceID", traceID)
		return fmt.Errorf("failed to save data: %w", err)
	}

	slog.InfoContext(ctx, "Data saved successfully", "filePath", filePath, "traceID", traceID)

	readData, err := storage.ReadData(ctx, filePath)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to read data", "error", err, "filePath", filePath, "traceID", traceID)
		return fmt.Errorf("failed to read data: %w", err)
	}

	preview := readData
	if len(readData) > 50 {
		preview = readData[:50] + "..."
	}
	slog.InfoContext(ctx, "Data read successfully", "filePath", filePath, "traceID", traceID, "preview", preview)

	sig := <-sigChan
	slog.InfoContext(ctx, "Received signal, shutting down", "signal", sig.String())
	slog.InfoContext(ctx, "Assignment 2 gracefully stopped")

	return nil
}

func runAssignment3(port int) error {
	fmt.Println("=== Running Assignment 3: HTTP JSON API ===")

	mux := http.NewServeMux()

	// Use stateless functions instead of handler methods
	mux.HandleFunc("/messages", traceMiddleware(messagesHandler))
	mux.HandleFunc("/health", traceMiddleware(healthHandler))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Starting HTTP server on port %d...\n", port)
		fmt.Printf("API Endpoints:\n")
		fmt.Printf("  POST http://localhost:%d/messages - Create a message\n", port)
		fmt.Printf("  GET  http://localhost:%d/messages - Get all messages\n", port)
		fmt.Printf("  GET  http://localhost:%d/health - Health check\n", port)
		fmt.Printf("\nPress Ctrl+C to stop the server...\n\n")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err, "port", port)
		}
	}()

	sig := <-sigChan
	fmt.Printf("\nReceived signal %s, shutting down server...\n", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		slog.Error("Server shutdown failed", "error", err)
		return err
	}

	fmt.Println("Assignment 3 HTTP server gracefully stopped")
	return nil
}

func runAssignment4(port int) error {
	fmt.Println("=== Running Assignment 4: Web Pages ===")

	mux := http.NewServeMux()

	// Setup static file server using embedded files
	staticFS, err := fs.Sub(htmlFiles, "html")
	if err != nil {
		return fmt.Errorf("failed to create static filesystem: %w", err)
	}

	// Static file handler for CSS, JS, and other assets
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Dynamic HTML page showing last 10 messages
	mux.HandleFunc("/web/messages", traceMiddleware(webMessagesHandler))

	// Static HTML page
	mux.HandleFunc("/", traceMiddleware(indexHandler))

	// Keep existing JSON API endpoints from Assignment 3
	mux.HandleFunc("/messages", traceMiddleware(messagesHandler))
	mux.HandleFunc("/health", traceMiddleware(healthHandler))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Starting web server on port %d...\n", port)
		fmt.Printf("Web Pages:\n")
		fmt.Printf("  GET  http://localhost:%d/ - Static home page\n", port)
		fmt.Printf("  GET  http://localhost:%d/web/messages - Dynamic messages page\n", port)
		fmt.Printf("  GET  http://localhost:%d/static/styles.css - Static CSS\n", port)
		fmt.Printf("\nAPI Endpoints:\n")
		fmt.Printf("  POST http://localhost:%d/messages - Create a message\n", port)
		fmt.Printf("  GET  http://localhost:%d/messages - Get all messages (JSON)\n", port)
		fmt.Printf("  GET  http://localhost:%d/health - Health check\n", port)
		fmt.Printf("\nPress Ctrl+C to stop the server...\n\n")

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error("Web server failed to start", "error", err, "port", port)
		}
	}()

	sig := <-sigChan
	fmt.Printf("\nReceived signal %s, shutting down web server...\n", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		slog.Error("Web server shutdown failed", "error", err)
		return err
	}

	fmt.Println("Assignment 4 web server gracefully stopped")
	return nil
}

// Stateless HTTP handler functions - no receiver, no state

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

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Trace-ID", traceID)

		next(w, r.WithContext(ctx))
	}
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)

	switch r.Method {
	case http.MethodPost:
		createMessage(w, r, traceID)
	case http.MethodGet:
		getMessages(w, r, traceID)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed", traceID)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)

	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   defaultAPIVersion,
	}

	respondWithSuccess(w, http.StatusOK, health, traceID)
}

func createMessage(w http.ResponseWriter, r *http.Request, traceID string) {
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

	message := Message{
		ID:        int(time.Now().UnixNano() / 1000000),
		User:      req.User,
		Message:   req.Message,
		Timestamp: time.Now(),
		TraceID:   traceID,
	}

	// Save to file (reusing Assignment 1 logic)
	err = appendMessage(req.User, req.Message)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to save message", "error", err, "traceID", traceID)
		respondWithError(w, http.StatusInternalServerError, "Failed to save message", traceID)
		return
	}

	slog.InfoContext(r.Context(), "Message created successfully",
		"user", req.User,
		"message_id", message.ID,
		"traceID", traceID)

	respondWithSuccess(w, http.StatusCreated, message, traceID)
}

func getMessages(w http.ResponseWriter, r *http.Request, traceID string) {
	messages, err := readMessagesAsJSON(traceID)
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

func readMessagesAsJSON(traceID string) ([]Message, error) {
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
			parts := parseMessageLine(line)
			if len(parts) >= 2 {
				message := Message{
					ID:        id,
					User:      parts[0],
					Message:   parts[1],
					Timestamp: time.Now(),
					TraceID:   traceID,
				}
				messages = append(messages, message)
				id++
			}
		}
	}

	err = scanner.Err()
	if err != nil {
		return []Message{}, err
	}

	return messages, nil
}

func parseMessageLine(line string) []string {
	parts := make([]string, 2)
	if colonIndex := findFirst(line, ": "); colonIndex != -1 {
		parts[0] = line[:colonIndex]
		parts[1] = line[colonIndex+2:]
	} else {
		parts[0] = "unknown"
		parts[1] = line
	}
	return parts
}

func findFirst(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Assignment 1 helper functions
func appendMessage(user, message string) error {
	f, err := os.OpenFile(messagesFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s: %s\n", user, message)
	_, err = f.WriteString(line)
	return err
}

func printLast10Messages() error {
	f, err := os.Open(messagesFileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No messages found.")
			return nil
		}
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	start := 0
	if len(lines) > 10 {
		start = len(lines) - 10
	}
	for _, line := range lines[start:] {
		fmt.Println(line)
	}
	return nil
}

// Assignment 4 web handler functions

func indexHandler(w http.ResponseWriter, r *http.Request) {
	traceID, _ := r.Context().Value("traceID").(string)

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Read and serve the static index.html file
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

	// Set content type to HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get messages data
	messages, err := readMessagesAsJSON(traceID)
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to read messages for web page", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create template data
	data := MessagesPageData{
		Messages:    messages,
		GeneratedAt: time.Now(),
		TraceID:     traceID,
	}

	// Parse the template from embedded files
	tmpl, err := template.ParseFS(htmlFiles, "html/messages.html")
	if err != nil {
		slog.ErrorContext(r.Context(), "Failed to parse messages template", "error", err, "traceID", traceID)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Execute the template with data
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
