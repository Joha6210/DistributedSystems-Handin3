package main

import (
	"context"
	"fmt"
	proto "handin3/grpc"
	"log"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	port           string
	messageHistory []*proto.Message
	clients        []*proto.Client
	msgChan        chan *proto.Message
	clientChans    map[string]chan *proto.Message
	uuid           string
	clk            int
	name           string
}

func main() {

	//Start up and configure logging output to file
	f, err := os.OpenFile("serverlog"+time.Now().Format("20060102150405")+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when we are done with it.
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)

	server := &ChitChatServer{}

	server.start_server()
}

func (s *ChitChatServer) start_server() {
	s.port = ":5050"
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		log.Fatalf("Did not work")
	}

	s.clientChans = make(map[string]chan *proto.Message)
	s.msgChan = make(chan *proto.Message, 100)
	s.uuid = uuid.New().String()
	s.name = "Server"

	proto.RegisterChitChatServer(grpcServer, s)
	log.Printf("gRPC server now listening on %s... \n", s.port)
	grpcServer.Serve(listener)

}

func (s *ChitChatServer) Subscribe(client *proto.Client, stream grpc.ServerStreamingServer[proto.Message]) error {
	log.Printf("Participant %s joined Chit Chat at logical time %d", client.Username, s.clk)

	oldClk := s.clk
	s.clk = s.clk + 1
	msg := proto.Message{Uuid: s.uuid, Message: fmt.Sprintf("Participant %s joined Chit Chat at logical time %d", client.Username, oldClk), Username: s.name, Clock: int32(s.clk)}
	s.PublishMessage(context.Background(), &msg)

	s.clients = append(s.clients, client)
	ch := make(chan *proto.Message, 10)
	s.clientChans[client.Uuid] = ch
	defer delete(s.clientChans, client.Uuid)

	// Send chat history
	for _, msg := range s.messageHistory {
		if err := stream.Send(msg); err != nil {
			log.Printf("Error sending history to %s: %v", client.Username, err)
			return err
		}
	}

	// Streaming loop
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				log.Printf("Stream closed for %s", client.Username)
				return nil
			}
			stream.Send(msg)
		case <-stream.Context().Done():
			log.Printf("Client disconnected: %s", client.Username)
			return nil
		}
	}
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, message *proto.Message) (*proto.Response, error) {
	s.clk = max(s.clk, int(message.Clock)) + 1
	s.messageHistory = append(s.messageHistory, message)
	for _, ch := range s.clientChans {
		select {
		case ch <- message:
		default:
		}
	}

	fmt.Printf("Message received: ")

	return &proto.Response{
		Result: true,
		Clock:  0,
	}, nil
}

// Is this needed, can the client just close the connection?
func (s *ChitChatServer) Unsubscribe(ctx context.Context, client *proto.Client) (*proto.Response, error) {
	ch, ok := s.clientChans[client.Uuid]
	if ok {
		log.Printf("Participant %s left Chit Chat at logical time %d", client.Username, s.clk)
		oldClk := s.clk
		s.clk = s.clk + 1
		msg := proto.Message{Uuid: s.uuid, Message: fmt.Sprintf("Participant %s left Chit Chat at logical time %d", client.Username, oldClk), Username: s.name, Clock: int32(s.clk)}
		s.PublishMessage(context.Background(), &msg)
		close(ch) // this will make the streaming loop in Subscribe exit
		delete(s.clientChans, client.Uuid)
	}

	// Remove from clients list
	for i, cl := range s.clients {
		if cl == client {
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}

	return &proto.Response{
		Result: true,
		Clock:  0,
	}, nil
}
