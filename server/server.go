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
}

var port string = ":5050"
var messageHistory []*proto.Message

var msgChan chan *proto.Message = make(chan *proto.Message)

func main() {
	server := &ChitChatServer{}

	server.start_server()
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChitChatServer(grpcServer, s)
	log.Printf("gRPC server now listening on %s... \n", port)
	grpcServer.Serve(listener)

}

func (s *ChitChatServer) Subscribe(client *proto.Client, stream grpc.ServerStreamingServer[proto.Message]) error {

	for _, msg := range messageHistory {
		stream.Send(msg)
	}
	go handle_subscription(stream)

	return nil
}

func handle_subscription(stream grpc.ServerStreamingServer[proto.Message]) {
	msg := <-msgChan
	stream.Send(msg)
}

func (s *ChitChatServer) PublishMessage(ctx context.Context, message *proto.Message) (*proto.Response, error) {
	messageHistory = append(messageHistory, message)
	msgChan <- message
	return &proto.Response{
		Result: true,
		Clock:  0,
	}, nil
}


//Is this needed, can the client just close the connection?
func (s *ChitChatServer) Unsubscribe(client *proto.Client) (*proto.Response, error) {
	
	return &proto.Response{
		Result: true,
		Clock:  0,
	},nil
}