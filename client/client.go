package main

import (
	"bufio"
	"context"
	"fmt"
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
}

var client proto.Client
var clk int32 = 0

func main() {
	ccclient := &ChitChatClient{}

	ccclient.start_client()
}

func (c *ChitChatClient) start_client() {

	//Default values
	username := "John Doe"
	serverAddr := "127.0.0.1:5050"

	if len(os.Args) > 1 {
		username = os.Args[1]
	}
	if len(os.Args) > 2 {
		serverAddr = os.Args[2]
	}

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(serverAddr, opts)

	if err != nil {
		log.Fatalf("Something went wrong! %s", err.Error())
	}

	proto_client := proto.NewChitChatClient(conn)

	client = proto.Client{Uuid: uuid.New().String(), Username: username, Clock: clk}

	go handle_incoming(proto_client)

	handle_message(proto_client)

	defer conn.Close()

}

func handle_message(proto_client proto.ChitChatClient) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		message := proto.Message{Uuid: client.Uuid, Message: text, Clock: clk + 1, Username: client.Username, Timestamp: time.Now().Format("02-01-2006 15:04:05")}
		response, err := proto_client.PublishMessage(context.Background(), &message)

		if err != nil {
			log.Printf("Something went wrong! %s \n", err)
		}
		if response.Result != true {
			log.Println("Server did not receive message!")
		}
	}

}

func handle_incoming(proto_client proto.ChitChatClient) {
	stream, err := proto_client.Subscribe(context.Background(), &client)

	if err != nil {
		log.Printf("Subscribe failed: %v", err)
	}

	log.Println("✅ Subscribed successfully. Listening for messages...")

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			log.Println("Server closed stream.")
			break
		}
		if err != nil {
			log.Printf("Error receiving: %v", err)
			break
		}

		fmt.Printf("[%s @ %s]: %s", message.Username, message.Timestamp, message.Message)
	}
}
