package server

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	chatmessage "pogchat/chat_message"
	"pogchat/client"
	"pogchat/cryptography"
	"pogchat/user_message"
)

type connManager struct {
	// TODO: WE NEED A BETTER WAY TO FIND CLIENTS MAYBE HASH PUBLIC KEY ?
	logged     map[string]client.Client
	clients    map[client.Client]bool
	broadcast  chan *chatmessage.ChatMessage
	register   chan client.Client
	unregister chan client.Client
}

var ClientIsRegisteredError = errors.New("this client already exists")

var _ ConnectionManager = (*connManager)(nil)

func (man *connManager) Register(c client.Client) error {
	_, ok := man.clients[c]
	if ok {
		log.Println("[server.Register] trying to register a client that already exists")
		return ClientIsRegisteredError
	}

	man.clients[c] = true
	log.Println("[server.Register] a connection has been made!")
	return nil
}

func (man *connManager) Unregister(c client.Client) error {
	pk := base64.RawStdEncoding.EncodeToString(c.PublicKey())
	if c.LoggedIn() {
		if _, ok := man.logged[pk]; ok {
			delete(man.logged, pk)
			log.Println("[server.Unregister] a logged connection has terminated!")
		}
	} else {
		if _, ok := man.clients[c]; ok {
			delete(man.clients, c)
			log.Println("[server.Unregister] a connection has terminated!")
		}
	}
	return nil
}

func (manager *connManager) Receive(client client.Client) {
	for {
		message := make([]byte, 4096)
		length, err := client.Read(message)
		if err != nil {
			manager.unregister <- client
			err := client.Close()
			if err != nil {
				log.Printf("[server.Receive] client.Close() returned error: %+v\n", err)
				break
			}
			break
		}
		if length > 0 {
			chatMsg := &chatmessage.ChatMessage{}

			processedMsg := trimByteSeq(message, '\x00')
			err := json.Unmarshal([]byte(processedMsg), chatMsg)
			if err != nil {
				log.Printf("[server.Receive] json.Unmarshal() returned error: %+v\n", err)
				break
			}

			if client.LoggedIn() {
				manager.broadcast <- chatMsg
				continue
			}

			um, err := user_message.ParseFromJSON(chatMsg.Payload)
			if err != nil {
				log.Println("[server.Receive] could not parse json")
				continue
			}

			pk := um.FromPublicKey()

			_, err = signer.Verify(pk, um.Message(), um.Signature())
			if err != nil {
				log.Println("[server.Receive] not a valid signature")
				continue
			}

			client.SetLoggedIn(true)
			client.SetPublicKey(um.FromPublicKey())
			manager.logged[base64.RawStdEncoding.EncodeToString(pk)] = client
			delete(manager.clients, client)
		}
	}
}

func (man *connManager) Send(client client.Client) {
	defer client.Close()
	for {
		select {
		case message, ok := <-client.WriteToChan():
			if !ok {
				return
			}
			_, err := client.Write(message)
			if err != nil {
				log.Println("[server.Send] could not write to peer")
				return
			}
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

var signer cryptography.Signer = cryptography.NewSigner(
	cryptography.WithSignerHasher(crypto.SHA256),
	cryptography.WithSignerRandomizer(rand.Reader),
)

func (man *connManager) Start() {
	for {
		select {
		case connection := <-man.register:
			err := man.Register(connection)
			if err != nil {
				log.Printf("[server.Start] man.Register() returned error: %+v\n", err)
			}
		case connection := <-man.unregister:
			err := man.Unregister(connection)
			if err != nil {
				log.Printf("[server.Start] man.Register() returned error: %+v\n", err)
			}
		case chatMsg := <-man.broadcast:
			um, err := user_message.ParseFromJSON(chatMsg.Payload)
			if err != nil {
				log.Println("[server.Start] could not parse payload")
				continue
			}

			_, err = signer.Verify(um.FromPublicKey(), um.Message(), um.Signature())
			if err != nil {
				log.Println("[server.Start] not a valid signature")
				continue
			}

			peer, ok := man.logged[base64.RawStdEncoding.EncodeToString(um.ToPublicKey())]
			if !ok {
				log.Println("[server.Start] message could not be sent")
				continue
			}

			peer.WriteToChan() <- []byte(chatMsg.Payload)
		}
	}
}

type server struct {
	connManager ConnectionManager
	listener    net.Listener
	network     string
	address     string
}

func (s *server) Start() {
	log.Println("[server.NewServer] starting server")
	listener, err := net.Listen(s.network, s.address)
	if err != nil {
		fmt.Printf("[server.NewServer] net.Listen() returned error: %+v\n", err)
		return
	}
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("[server.NewServer] listener.Accept() returned error: %+v\n", err)
		}
		client := client.NewClient(client.WithConnection(connection))
		s.connManager.Register(client)
		go s.connManager.Receive(client)
		go s.connManager.Send(client)
	}
}

func WithAddress(addr string) ServerOpts {
	return func(s *server) {
		s.address = addr
	}
}

func WithNetwork(network string) ServerOpts {
	return func(s *server) {
		s.network = network
	}
}

func WithConnectionManager(manager ConnectionManager) ServerOpts {
	return func(s *server) {
		s.connManager = manager
	}
}

func NewServer(opts ...ServerOpts) Server {
	s := &server{
		address: ":42069",
		network: "tcp",
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.connManager == nil {
		s.connManager = &connManager{
			clients:    make(map[client.Client]bool),
			logged:     make(map[string]client.Client),
			broadcast:  make(chan *chatmessage.ChatMessage),
			register:   make(chan client.Client),
			unregister: make(chan client.Client),
		}
	}

	go s.connManager.Start()

	return s
}
