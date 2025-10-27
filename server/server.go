package main

import (
	proto "ITUServer/grpc"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type ChatDatabaseServerITU_WOMEN_IN_STEM struct {
	proto.UnimplementedChatDatabaseServer
	mu      sync.RWMutex   //mutex prevents race conditions
	clients []activeClient //(pizza)slice that keeps track of the clients on the server
}

// active client that is connected to the stream
type activeClient struct {
	username string
	stream   proto.ChatDatabase_ChitChattingServer
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) ChitChatting(stream proto.ChatDatabase_ChitChattingServer) error {
	firstMessage, err := stream.Recv()
	if err != nil {
		log.Fatalf("did not recieve message: %s", err)
	}

	username := firstMessage.Username
	if username == "" {
		username = "Jesus Christensen" //default name #AMEN
	}

	//registering a client:
	s.mu.Lock()             //prevents concurrent access.
	client := activeClient{ //creating a variable of type activeClient
		username: username, //username from first message
		stream:   stream,   //bidirectional stream
	}
	s.clients = append(s.clients, client) //adds client to slice
	s.mu.Unlock()                         //unlocks

	log.Printf("user %s joined the ChitChat!<3", username)

	go s.BroadcastToAll(&proto.ChitChat{
		Message:   "User:" + username + "joined the chat!<3",
		Username:  "Server",
		Timestamp: time.Now().Format(time.RFC3339),
	})

	for {
		msg, err := stream.Recv()
		if err != nil { // when terminal is closed the user leaves the server.
			log.Printf("%v left the chat! ;(", username) //måske behøver den ikk skrive noget 
			s.RemoveClient(username)
			break
		}

		go s.BroadcastToAll(msg)
		continue
	}
	return nil
}

func main() {

	server := &ChatDatabaseServerITU_WOMEN_IN_STEM{}

	server.start_server()
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChatDatabaseServer(grpcServer, s)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}
	log.Println("Server opened")
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) BroadcastToAll(msg *proto.ChitChat) {

	for _, client := range s.clients { // for each loop
		go s.SendToOne(msg, client.stream, client.username) //starts a goroutine for each client
	}
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) SendToOne(msg *proto.ChitChat, stream proto.ChatDatabase_ChitChattingServer, username string) {
	err := stream.Send(msg) //sends message to client
	if err != nil {         //in case it fails:
		log.Printf("Failed to send message to %s: %v", username, err)
	}

}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) RemoveClient(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, client := range s.clients {
		if client.username == username {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			log.Printf("User %s removed from client list", username)

			break
		}

	}

}
