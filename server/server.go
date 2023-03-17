package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Client struct {
	socket net.Conn
	data   chan []byte
}

func (c *Client) Receive() {
	for {
		message := make([]byte, 4096)
		length, err := c.socket.Read(message)
		if err != nil {
			c.socket.Close()
			break
		}
		if length > 0 {
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func startClient() {
	fmt.Println("[client.startClient] starting new client")
	connection, error := net.Dial("tcp", "localhost:42069")
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{socket: connection}
	go client.Receive()
	for {
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		connection.Write([]byte(strings.TrimRight(message, "\n")))
	}
}

type ConnManager struct {
	// TODO: WE NEED A BETTER WAY TO FIND CLIENTS MAYBE HASH PUBLIC KEY ?
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func (manager *ConnManager) Receive(client *Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.socket.Read(message)
		if err != nil {
			manager.unregister <- client
			client.socket.Close()
			break
		}
		if length > 0 {
			log.Println("[server.Receive] received: " + string(message))
			manager.broadcast <- message
		}
	}
}

func (man *ConnManager) Send(client *Client) {
	defer client.socket.Close()
	for {
		select {
		case message, ok := <-client.data:
			if !ok {
				return
			}
			client.socket.Write(message)
		}
	}
}

func (man *ConnManager) Start() {
	for {
		select {
		case connection := <-man.register:
			man.clients[connection] = true
			log.Println("[server.Start] a connection has been made!")
		case connection := <-man.unregister:
			if _, ok := man.clients[connection]; ok {
				close(connection.data)
				delete(man.clients, connection)
				log.Println("[server.Start] a connection has terminated!")
			}
		case message := <-man.broadcast:
			for connection := range man.clients {
				select {
				case connection.data <- message:
				default:
					close(connection.data)
					delete(man.clients, connection)
				}
			}
		}
	}
}

func StartServer() {
	log.Println("[server.NewServer] starting server")
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("[server.NewServer] net.Listen() returned error: %+v\n", err)
	}
	manager := ConnManager{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	go manager.Start()
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("[server.NewServer] listener.Accept() returned error: %+v\n", err)
		}
		client := &Client{socket: connection, data: make(chan []byte)}
		manager.register <- client
		go manager.Receive(client)
		go manager.Send(client)
	}
}
