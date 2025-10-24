package main

import (
	"bufio"
	"context"
	proto "handin3/grpc"
	"io"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChitChatClient struct {
	proto.UnimplementedChitChatServer
}

var client proto.Client
var clk int32 = 0

func main() {
	client := &ChitChatClient{}

	client.start_client()
}

func (c *ChitChatClient) start_client() {
	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient("127.0.0.1:5050", opts)

	if err != nil {
		log.Fatalf("Something went wrong! %s", err.Error())
	}

	proto_client := proto.NewChitChatClient(conn)

	client = proto.Client{Uuid: uuid.New().String(), Username: "John Doe", Clock: clk}

	go handle_incoming(proto_client)

	handle_message(proto_client)

	defer conn.Close()

}

func handle_message(proto_client proto.ChitChatClient) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		message := proto.Message{Uuid: client.Uuid, Message: text, Clock: clk + 1, Username: client.Username, Timestamp: time.Now().Format("24-10-2025 11:37:05")}
		response, err := proto_client.PublishMessage(context.Background(), &message)

		if err != nil {
			log.Printf("Something went wrong! %s \n", err)
		}

		log.Printf("%t", response.Result)
		log.Printf("%s", message.Timestamp)
	}

}

func handle_incoming(proto_client proto.ChitChatClient) {
	stream, err := proto_client.Subscribe(context.Background(), &client)

	if err != nil {
		log.Printf("Subscribe failed: %v", err)
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Stream closed!")
			break
		}
		if err != nil {
			log.Fatalf("%v.Subscribe(_) = _, %v", proto_client, err)
		}
		log.Println(message.Message)
	}
}
