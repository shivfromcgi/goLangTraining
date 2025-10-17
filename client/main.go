package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "cgi.com/goLangTraining/proto/message_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	defaultServerAddr = "localhost:50051"
)

func main() {
	var (
		serverAddr = flag.String("server", defaultServerAddr, "gRPC server address")
		user       = flag.String("user", "", "User for message operations")
		message    = flag.String("message", "", "Message to save")
		getLast10  = flag.Bool("get", false, "Get last 10 messages")
	)
	flag.Parse()

	conn, err := grpc.Dial(*serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMessageServiceClient(conn)
	fmt.Printf("🔌 Connected to gRPC Message Service at %s\n", *serverAddr)

	if *getLast10 {
		err := getMessages(client)
		if err != nil {
			log.Fatalf("Failed to get messages: %v", err)
		}
	} else if *user != "" && *message != "" {
		err := saveMessage(client, *user, *message)
		if err != nil {
			log.Fatalf("Failed to save message: %v", err)
		}
	} else {
		fmt.Println("\n📖 gRPC Client Usage:")
		fmt.Printf("  Save message:    go run . -user=alice -message='Hello gRPC!'\n")
		fmt.Printf("  Get messages:    go run . -get\n")
		fmt.Printf("  Custom server:   go run . -server=localhost:50051 -get\n")

		demoUser := "demo"
		demoMessage := fmt.Sprintf("gRPC Client Demo - %s", time.Now().Format("15:04:05"))

		fmt.Printf("\n1️⃣ Saving demo message...\n")
		err := saveMessage(client, demoUser, demoMessage)
		if err != nil {
			log.Fatalf("Demo failed - save message: %v", err)
		}

		fmt.Printf("\n2️⃣ Getting last 10 messages...\n")
		err = getMessages(client)
		if err != nil {
			log.Fatalf("Demo failed - get messages: %v", err)
		}
	}

	fmt.Println("\n✅ gRPC client operation completed successfully!")
}

func saveMessage(client pb.MessageServiceClient, user, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SaveMessageRequest{
		User:    user,
		Message: message,
	}

	fmt.Printf("💾 Saving message: %s -> %s\n", user, message)

	_, err := client.Save(ctx, req)
	if err != nil {
		return fmt.Errorf("save failed: %w", err)
	}

	fmt.Printf("✅ Message saved successfully!\n")
	return nil
}

func getMessages(client pb.MessageServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Printf("📨 Fetching last 10 messages...\n")

	resp, err := client.GetLast10(ctx, &emptypb.Empty{})
	if err != nil {
		return fmt.Errorf("get messages failed: %w", err)
	}

	messages := resp.GetMessages()
	if len(messages) == 0 {
		fmt.Println("📭 No messages found.")
		return nil
	}

	fmt.Printf("\n📋 Last %d Messages:\n", len(messages))
	for _, msg := range messages {
		timestamp := msg.GetTimestamp().AsTime()
		fmt.Printf("  [%d] %s (%s): %s\n",
			msg.GetId(),
			msg.GetUser(),
			timestamp.Format("2006-01-02 15:04:05"),
			msg.GetMessage())
	}

	return nil
}
