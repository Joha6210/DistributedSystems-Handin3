package main

import (
	"log"
	"net"
	proto "handin3/grpc"
	"google.golang.org/grpc" 
)

type ChitChatServer struct {
	proto.UnimplementedChitChatServer
}

func main() {
	server := &ChitChatServer{}

	server.start_server()
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work")
	}

	proto.RegisterChitChatServer(grpcServer, s)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}

}
