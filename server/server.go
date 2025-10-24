package main

import (
	"context"
	proto "handin3/grpc"
	"log"
	"net"

	"google.golang.org/grpc"
)

type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	port           string
	messageHistory []*proto.Message
	clients        []*proto.Client
	msgChan        chan *proto.Message
	clientChans    map[string]chan *proto.Message
}

func main() {
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

	proto.RegisterChitChatServer(grpcServer, s)
	log.Printf("gRPC server now listening on %s... \n", s.port)
	grpcServer.Serve(listener)

}

func (s *ChitChatServer) Subscribe(client *proto.Client, stream grpc.ServerStreamingServer[proto.Message]) error {
	log.Printf("Client subscribed: %s", client.Username)

	s.clients = append(s.clients, client)
	ch := make(chan *proto.Message, 10)
	s.clientChans[client.Uuid] = ch
	defer delete(s.clientChans, client.Uuid)

	// Send message history
	for _, msg := range s.messageHistory {
		if err := stream.Send(msg); err != nil {
			log.Printf("Error sending history to %s: %v", client.Username, err)
			return err
		}
	}

	// Streaming loop
	for {
		select {
		case msg := <-ch:
			if err := stream.Send(msg); err != nil {
				log.Printf("Error sending message to %s: %v", client.Username, err)
				return err
			}
		case <-stream.Context().Done():
			log.Printf("Client disconnected: %s", client.Username)
			return nil
		}
	}
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, message *proto.Message) (*proto.Response, error) {
	s.messageHistory = append(s.messageHistory, message)
	for _, ch := range s.clientChans {
		// Non-blocking send
		select {
		case ch <- message:
		default:
		}
	}

	return &proto.Response{
		Result: true,
		Clock:  0,
	}, nil
}

// Is this needed, can the client just close the connection?
func (s *ChitChatServer) Unsubscribe(ctx context.Context, client *proto.Client) (*proto.Response, error) {

	for i, cl := range s.clients {
		if cl == client {
			// Remove the element at index i
			s.clients = append(s.clients[:i], s.clients[i+1:]...)
			break
		}
	}

	return &proto.Response{
		Result: true,
		Clock:  0,
	}, nil
}
