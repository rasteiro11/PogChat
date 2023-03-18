package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	chatmessage "pogchat/chat_message"
	"pogchat/key"
	"pogchat/user_message"
	"time"
)

type client struct {
	loggedIn  bool
	publicKey []byte
	socket    net.Conn
	data      chan []byte
}

type Client interface {
	LoggedIn() bool
	PublicKey() []byte
	SetLoggedIn(loggedIn bool)
	SetPublicKey(publicKey []byte)
	Close() error
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	WriteToChan() chan []byte
}

var _ Client = (*client)(nil)

func (c *client) Receive() {
	for {
		message := make([]byte, 4096)
		length, err := c.socket.Read(message)
		if err != nil {
			c.socket.Close()
			break
		}
		if length > 0 {
			//fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func (c *client) Close() error {
	close(c.data)
	return c.socket.Close()
}

func (c *client) Read(buf []byte) (int, error) {
	return c.socket.Read(buf)
}

func (c *client) Write(b []byte) (int, error) {
	return c.socket.Write(b)
}

func (c *client) WriteToChan() chan []byte {
	return c.data
}

func (c *client) LoggedIn() bool {
	return c.loggedIn
}

func (c *client) SetLoggedIn(loggedIn bool) {
	c.loggedIn = loggedIn
}

func (c *client) PublicKey() []byte {
	return c.publicKey
}

func (c *client) SetPublicKey(publicKey []byte) {
	c.publicKey = publicKey
}

type ClientOpts func(*client)

func NewClient(opts ...ClientOpts) Client {
	c := &client{
		data: make(chan []byte),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func WithConnection(conn net.Conn) ClientOpts {
	return func(c *client) {
		c.socket = conn
	}
}

func StartClient() {

	time.Sleep(time.Second)
	fmt.Println("[client.startClient] starting new client")
	connection, error := net.Dial("tcp", "localhost:42069")
	if error != nil {
		fmt.Println(error)
	}
	client := &client{socket: connection}
	pairSender, err := key.NewKeyPair(2048)
	if err != nil {
		log.Println("[server.StartClient] could not create key pair")
		return
	}

	//pairReceiver, err := key.NewKeyPair(2048)
	//if err != nil {
	//	log.Println("[server.StartClient] could not create key pair")
	//	return
	//}

	um := user_message.NewUserMessage(
		user_message.WithFromPublicKey(pairSender.PublicKey()),
		user_message.WithToPublicKey(pairSender.PublicKey()),
	)

	encryptedMsg, err := um.GetEncryptedMessage([]byte("GAMER"))
	if err != nil {
		log.Println("[server.StartClient] could encrypt msg")
		return
	}

	_, err = um.GetSignature(pairSender.PrivateKey(), encryptedMsg)
	if err != nil {
		log.Println("[server.StartClient] could not sign message")
		return
	}

	userMsg, err := um.MarshalJSON()
	if err != nil {
		log.Println("[server.StartClient] could not msrhal json")
		return
	}

	go client.Receive()
	msg, err := json.Marshal(&chatmessage.ChatMessage{
		Type:    "LOGIN_MSG",
		Payload: string(userMsg),
	})
	if err != nil {
		log.Println("[server.StartClient] could not marshal json")
		return
	}
	connection.Write(msg)
	time.Sleep(time.Second * 5)
	for {
		msg, err := json.Marshal(&chatmessage.ChatMessage{
			Type:    "PEER_MSG",
			Payload: string(userMsg),
		})
		if err != nil {
			log.Println("[server.StartClient] could not marshal json")
			return
		}
		connection.Write(msg)
		time.Sleep(time.Second * 5)
	}
}
