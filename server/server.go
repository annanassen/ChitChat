package main

import (
	proto "ITUServer/grpc"
	"log"
	"net"
	"strconv"
	"sync"

	"google.golang.org/grpc"
)

type ChatDatabaseServerITU_WOMEN_IN_STEM struct {
	proto.UnimplementedChatDatabaseServer
	mu      sync.RWMutex   //mutex prevents race conditions
	clients []activeClient //(pizza)slice that keeps track of the clients on the server
	clock   int64          //logical clock (lamport)
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

	log.Printf("[Server] %s has been registered in the client list!", username)

	time := s.IncrementClock() //increments the clock when a client joins the server
	go s.BroadcastToAll(&proto.ChitChat{
		Message:          "User " + username + " joined the chat (at logical time: " + strconv.FormatInt(time, 10) + ")",
		Username:         "Server",
		LamportTimestamp: time,
	})

	for {
		msg, err := stream.Recv()

		if err != nil { // when terminal is closed the user leaves the server.
			s.RemoveClient(username)
			leaveTime := s.IncrementClock() //increments clock when a client leaves the server
			go s.BroadcastToAll(&proto.ChitChat{
				Message:          "User " + username + " left the chat (at logical time: " + strconv.FormatInt(leaveTime, 10) + ")",
				Username:         "Server",
				LamportTimestamp: leaveTime,
			})
			break
		}

		log.Printf("[Server] Received message from %s", msg.Username) // If it wasn't because a client left

		serverTime := s.UpdateClock(msg.LamportTimestamp) //increments clock when it recieves a message
		msg.LamportTimestamp = serverTime                 //
		go s.BroadcastToAll(msg)
		continue
	}
	return nil
}

func main() {

	server := &ChatDatabaseServerITU_WOMEN_IN_STEM{}

	log.Println("[Server] Starting server...")
	server.start_server()
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}
	proto.RegisterChatDatabaseServer(grpcServer, s)

	log.Println("[Server] Startup successful.")

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Something went wrong...")
	}
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) BroadcastToAll(msg *proto.ChitChat) {

	for _, client := range s.clients { // for each loop
		go s.SendToOne(msg, client.stream, client.username) //starts a goroutine for each client
	}
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) SendToOne(msg *proto.ChitChat, stream proto.ChatDatabase_ChitChattingServer, username string) {
	err := stream.Send(msg) //sends message to client
	if err != nil {         //in case it fails:
		log.Printf("[Server] Failed to send message to %s: %v", username, err)
	}
	log.Printf("[Server] Successfully sent %s's message to %s", msg.Username, username)

}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) RemoveClient(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, client := range s.clients {
		if client.username == username {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			log.Printf("[Server] Removed user %s from the client list", username)
			break
		}

	}

}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) IncrementClock() int64 {
	s.clock++
	return s.clock
}

func (s *ChatDatabaseServerITU_WOMEN_IN_STEM) UpdateClock(receivedTime int64) int64 {
	if receivedTime > s.clock {
		s.clock = receivedTime
	}
	s.clock++
	return s.clock
}
