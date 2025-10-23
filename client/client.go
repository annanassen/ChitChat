package main

import (
	proto "ITUServer/grpc"
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	//clienten printer?

	reader := bufio.NewReader(os.Stdin)
	username := connectClient(reader)

	client := proto.NewITUDatabaseClient(conn)

	students, err := client.GetStudents(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("Not working")
	}

	for {
		fmt.Println("Enter a message")
		message, _ := reader.ReadString('\n')
		log.Println(message)

	}

}

func connectClient(reader *bufio.Reader) string {
	fmt.Println("Please enter your username")
	username, _ := reader.ReadString('\n')

	if username == "" {
		fmt.Println("Please enter a username of at least 1 character")
	} else {
		fmt.Println("Your displayname:", username)
		return username
	}
}

//dont make two goroutines (Publish and Read)
// How do we publish it via gRPC? We will try to use bidirectional streaming
