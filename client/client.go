package main

import (
	proto "ITUServer/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type chatClient struct {
	client   proto.ChatDatabaseClient
	stream   proto.ChatDatabase_ChitChattingClient
	username string
	ctx      context.Context
	wg       sync.WaitGroup
}

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	reader := bufio.NewReader(os.Stdin)
	username := connectClient(reader)

	client := proto.NewChatDatabaseClient(conn)

	stream, err := client.ChitChatting(context.Background())
	if err != nil {
		log.Fatalf("Not working")
	}

	joinMessage := &proto.ChitChat{
		Username: username,
	}

	err = stream.Send(joinMessage)
	if err != nil {
		log.Fatalf("u aint joinin noffin")
	}

	chatClient := &chatClient{
		client:   client,
		stream:   stream,
		username: username,
		ctx:      context.Background(),
	}

	chatClient.initiateSmallTalk(reader)

}

func connectClient(reader *bufio.Reader) string {
	fmt.Println("Please enter your username")
	username, _ := reader.ReadString('\n')

	if username == "" {
		fmt.Println("Please enter a username of at least 1 character")
		return username
	} else {
		fmt.Println("Your displayname:", username)
		return username
	}

}

func (c *chatClient) initiateSmallTalk(reader *bufio.Reader) {
	fmt.Printf("Chat, or else...")
	fmt.Println("Type '/quit' to leave the chitchatting... if you dare.")
	fmt.Println("︵‿︵‿୨♡୧‿︵‿︵")

	c.wg.Add(1)
	go c.receiveMessage()

	for { //loop for messages the client sends
		message, _ := reader.ReadString('\n')

		if message == "/quit" {
			fmt.Println("you quit. but i will find you.")
			break
		}

		if len(message) > 128 {
			fmt.Println("Message is too long. Maximum is 128 characters. Yours %d", len(message))
			continue
		}

		chitChatMsg := &proto.ChitChat{
			Username:  c.username,
			Message:   message,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		err := c.stream.Send(chitChatMsg)
		if err != nil {
			log.Fatalf("Failed to send message: %v", err)
			break
		}

	}

}

func (c *chatClient) receiveMessage() {
	defer c.wg.Done()

	for {
		msg, err := c.stream.Recv()
		if err != nil {
			log.Fatalf("something went wrong:", err)
		}
		if msg.Username == "Server" {
			fmt.Println(msg.Message)
		} else {
			fmt.Printf("%s: %v", msg.Username, msg.Message)
		}
	}
}

//dont make two goroutines (Publish and Read)
// How do we publish it via gRPC? We will try to use bidirectional streaming
