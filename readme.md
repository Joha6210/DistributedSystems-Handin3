# Distributed System Design and Implementation Report

## Link to github repo

[https://github.com/Joha6210/DistributedSystems-Handin3](https://github.com/Joha6210/DistributedSystems-Handin3)

## 1. Streaming Model Selection

### 1.1 Discussion

- _Discuss, whether you are going to use server-side streaming,
    client-side streaming, or bidirectional streaming?_

- In our chit chat implementation we use server-side streaming ([1](https://grpc.io/docs/languages/go/basics/#server-side-streaming-rpc)), to send multiple messages from the server to the client. This is done to support long-lived logical flow of data from the server to the clients ([2](https://grpc.io/docs/guides/performance/))

---

## 2. System Architecture

### 2.1 Overview

- **Architecture Type:**  
  Chit chat is implemented as a server-client architecture, this allows for the server to serve multiple clients that can connect concurrently. Clients can leave and join whenever they like, and they will get the chat history send to them upon subscribing(connecting) to the server.
  
### 2.2 Components

- **Server:** The server is responsible for letting clients 'subscribe', publish messages to chit chat and unsubscribe when the client wants to leave. The server also keeps track of the message history. The server will announce when a new client either joins or leaves the server.

- **Client:** Clients can connect to the chit chat server using the 'subscribe' method, this is for the client to "register" at the server, and allows for the client to receive the message history and new messages that are being published by other clients. Registered clients can also publish new messages to the chit chat server, for other clients to receive.

## 3. RPC Methods and Message Types

### 3.1 Implemented RPC Methods

| RPC Method | Type (Unary/Server Streaming/Client Streaming/Bidirectional) | Description   |
|-------------|-------------------------------------------------------------|-------------- |
| PublishMessage (Message)| Unary             | Publishes Message object to server      |
| Subscribe (Client)      | Server Streaming  | Adds client to server                   |
| Unsubscribe (Client)    | Unary             | Removes client from server              |

### 3.2 Message Types

| Message Name | Purpose | Fields |
|---------------|----------|--------|
| message Message | Contains the message and relevant information | `string uuid = 1; string message = 2; int32 clock = 3; string username = 4; string timestamp = 5;`  |
| message Response | Contains a response from the server to a client | `bool result = 1; int32 clock = 3;`  |
| message Client |   Contains information about a client  |  `string uuid = 1; string username = 2; int32 clock = 3;`  |

## 4. Lamport Timestamp Implementation

### 4.1 Overview

_Describe how Lamport timestamps are calculated and updated across processes._

### 4.2 Algorithm Steps

1. _[Step 1: Initialize timestamp]_  
2. _[Step 2: Update on local event]_  
3. _[Step 3: Update on receiving a message]_  
4. _[Step 4: Synchronization rules]_  

### 4.3 Example

_Provide an example of timestamp evolution during message exchange._

---

## 5. Sequence Diagram

### 5.1 Interaction Flow

_Trace a sequence of RPC calls and Lamport timestamps corresponding to a specific scenario (e.g., Client X joins, publishes, leaves)._

### 5.2 Diagram

```mermaid
sequenceDiagram
    participant Client1
    participant Client2
    participant Client3
    participant Server
    Client1->>Server: Subscribe Request (LT=1)
    Server->>Client1: Return Message stream (LT=2)
    Server-->>Client1: Broadcast that client 1 has joined (LT=3)
    Client2->>Server: Subscribe Request (LT=1)
    Server->>Client2: Return Message stream (LT=2)
    Server-->>Client1: Broadcast that client 2 has joined (LT=3)
    Server-->>Client2: Broadcast that client 2 has joined (LT=3)
    Client3->>Server: Subscribe Request (LT=1)
    Server->>Client3: Return Message stream (LT=2)
    Server-->>Client1: Broadcast that client 3 has joined (LT=3)
    Server-->>Client2: Broadcast that client 3 has joined (LT=3)
    Server-->>Client3: Broadcast that client 3 has joined (LT=3)
    Client1->>Server: Publish Message (LT=3)
    Server-->>Client2: Broadcast message  (LT=3)
    Server-->>Client3: Broadcast message (LT=3)
    Server->>Client1: Confirm Publish (LT=4)
    Client1->>Server: Leave Request (LT=5)
    Server-->>Client1: Broadcast that client 1 has left (LT=3)
    Server-->>Client2: Broadcast that client 1 has left (LT=3)
    Server-->>Client3: Broadcast that client 1 has left (LT=3)
    Server->>Client1: Ack Leave (LT=6)
```

Generate proto files:
protoc -I=grpc --go_out=./grpc --go-grpc_out=./grpc grpc/chitchat.proto