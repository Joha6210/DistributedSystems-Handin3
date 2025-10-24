package main

import (
	"bufio"
	"context"
	"fmt"
	proto "handin3/grpc"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChitChatClient struct {
	client proto.Client
	clk    int32
}

func main() {
	ccclient := &ChitChatClient{}

	ccclient.start_client()
}

func (c *ChitChatClient) start_client() {

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	//Default values
	username := "John Doe"
	serverAddr := "127.0.0.1:5050"

	c.clk = 0

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

	c.client = proto.Client{Uuid: uuid.New().String(), Username: username, Clock: c.clk}

	go c.handle_incoming(proto_client)

	c.handle_message(proto_client, cancel)

	defer conn.Close()

}

func (c *ChitChatClient) handle_message(proto_client proto.ChitChatClient, cancel context.CancelFunc) {
	reader := bufio.NewReader(os.Stdin)

	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == `\x` {
			//Disconnect from server
			response, err := proto_client.Unsubscribe(context.Background(), &c.client)
			if response.Result {
				cancel()
				break
			}
			log.Printf("Could not unsubscribe! %s \n", err)
		}
		c.clk = c.clk + 1
		message := proto.Message{Uuid: c.client.Uuid, Message: text, Clock: c.clk, Username: c.client.Username, Timestamp: time.Now().Format("02-01-2006 15:04:05")}
		response, err := proto_client.PublishMessage(context.Background(), &message)

		if err != nil {
			log.Printf("Something went wrong! %s \n", err)
		}
		if response.Result != true {
			log.Println("Server did not receive message!")
		}
	}

}

func (c *ChitChatClient) handle_incoming(proto_client proto.ChitChatClient) {
	stream, err := proto_client.Subscribe(context.Background(), &c.client)

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

		if message.Clock > c.clk {
			c.clk = message.Clock //Update to highest clock
		}

		fmt.Printf("[%s @ %d] %s: %s \n", message.Timestamp, message.Clock, message.Username, message.Message)
	}
}
