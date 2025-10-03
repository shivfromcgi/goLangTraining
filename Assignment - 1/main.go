package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
)

const fileName = "messages.txt"

func main() {
	// Define CLI flags
	userFlag := flag.String("user", "", "User ID (required)")
	msgFlag := flag.String("message", "", "Message to append (required)")
	clearFlag := flag.Bool("clear", false, "Clear all messages")
	flag.Parse()

	// --- Handle the clear flag FIRST ---
	if *clearFlag {
		err := os.Truncate(fileName, 0)
		if err != nil && !os.IsNotExist(err) {
			// Ignore if the file doesn’t exist yet
			fmt.Printf("Error clearing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All messages cleared.")
		return
	}

	// Validate flags
	if *userFlag == "" || *msgFlag == "" {
		fmt.Println("Usage: go run main.go -user=<userID> -message=<text>  OR  -clear")
		os.Exit(1)
	}

	// Append the message to disk
	if err := appendMessage(*userFlag, *msgFlag); err != nil {
		fmt.Printf("Error writing message: %v\n", err)
		os.Exit(1)
	}

	// Retrieve and print the last 10 messages
	fmt.Println("\nLast 10 Messages:")
	if err := printLast10(); err != nil {
		fmt.Printf("Error reading messages: %v\n", err)
	}
}

func appendMessage(user, message string) error {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Format: userID: message
	line := fmt.Sprintf("%s: %s\n", user, message)
	_, err = f.WriteString(line)
	return err
}

func printLast10() error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Print last 10
	start := 0
	if len(lines) > 10 {
		start = len(lines) - 10
	}
	for _, line := range lines[start:] {
		fmt.Println(line)
	}
	return nil
}
