package client

import (
	"fmt"
	"log"
	"net"
	"pogchat/cryptography"
	"pogchat/user_message"
)

type client struct {
	loggedIn  bool
	publicKey []byte
	socket    net.Conn
	data      chan []byte
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
			fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func trimByteSeq(seq []byte, delim byte) []byte {
	finalSeq := make([]byte, 0)

	for _, b := range seq {
		if b == delim {
			break
		}
		finalSeq = append(finalSeq, b)
	}

	return finalSeq
}

func (c *client) ReceiveAndDecrypt(private []byte, rec chan []byte) {
	cryptor := cryptography.NewCryptor()
	for {
		message := make([]byte, 4096)
		length, err := c.socket.Read(message)
		if err != nil {
			c.socket.Close()
			break
		}
		if length > 0 {
			processedMsg := trimByteSeq(message, '\x00')
			um, err := user_message.ParseFromJSON(string(processedMsg))
			if err != nil {
				log.Printf("[client.ReceiveAndDecrupt] ParseFromJSON() returned error: %+v\n", err)
				continue
			}

			dec, err := cryptor.Decrypt(private, um.Message())
			if err != nil {
				log.Printf("[client.ReceiveAndDecrupt] could not decode error: %+v\n", err)
				return
			}

			rec <- dec
		}
	}
}

func (c *client) Close() error {
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
