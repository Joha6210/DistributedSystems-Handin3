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

	//Start up and configure logging output to file
	f, err := os.OpenFile("logs/client"+username+"log"+time.Now().Format("20060102150405")+".log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when we are done with it.
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(serverAddr, opts)

	if err != nil {
		log.Fatalf("Something went wrong! %s", err.Error())
	}

	proto_client := proto.NewChitChatClient(conn)

	c.client = proto.Client{Uuid: uuid.New().String(), Username: username, Clock: c.clk}

	go c.handle_incoming(proto_client)

	fmt.Println("Connected successfully!")

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
			} else {
				log.Printf("Could not unsubscribe! %s \n", err)
			}

		}
		c.clk = c.clk + 1
		message := proto.Message{Uuid: c.client.Uuid, Message: text, Clock: c.clk, Username: c.client.Username, Timestamp: time.Now().Format("02-01-2006 15:04:05")}
		response, err := proto_client.PublishMessage(context.Background(), &message)

		if err != nil {
			log.Printf("Something went wrong! %s \n", err)
		}
		if !response.Result {
			log.Println("Server did not receive message!")
		}
	}

}

func (c *ChitChatClient) handle_incoming(proto_client proto.ChitChatClient) {
	stream, err := proto_client.Subscribe(context.Background(), &c.client)

	if err != nil {
		log.Fatalf("Subscribe failed: %v", err)
	} else {
		log.Println("Subscribed successfully. Listening for messages...")
	}

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

		c.clk = max(c.clk, message.Clock) + 1
		log.Printf("[%s @ %d] %s: %s \n", message.Timestamp, message.Clock, message.Username, message.Message)
		fmt.Printf("[%s @ %d] %s: %s \n", message.Timestamp, message.Clock, message.Username, message.Message)
	}
}
