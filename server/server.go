package server

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	chatmessage "pogchat/chat_message"
	"pogchat/client"
	"pogchat/cryptography"
	"pogchat/user_message"
)

type ConnManager struct {
	// TODO: WE NEED A BETTER WAY TO FIND CLIENTS MAYBE HASH PUBLIC KEY ?
	logged     map[string]client.Client
	logout     chan client.Client
	clients    map[client.Client]bool
	broadcast  chan *chatmessage.ChatMessage
	register   chan client.Client
	unregister chan client.Client
}

func (manager *ConnManager) Receive(client client.Client) {
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
				manager.unregister <- client
				if err != nil {
					log.Printf("[server.Receive] client.Close() returned error: %+v\n", err)
					break
				}
				break
			}

			if client.LoggedIn() {
				manager.broadcast <- chatMsg
				client.WriteToChan() <- []byte("TALOGADOJABURRO")
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
		}
	}
}

func (man *ConnManager) Send(client client.Client) {
	defer client.Close()
	for {
		select {
		case message, ok := <-client.WriteToChan():
			if !ok {
				return
			}
			log.Println("-----------------------------------------------------------------------------------------------")
			log.Println("[server.Send] sending message to client")
			log.Println("-----------------------------------------------------------------------------------------------")
			log.Println(string(message) + "\n")
			log.Println("-----------------------------------------------------------------------------------------------")
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

type ChatMsgType int

const (
	LOGIN_MSG ChatMsgType = iota
	PEER_MSG
)

var signer cryptography.Signer = cryptography.NewSigner(
	cryptography.WithSignerHasher(crypto.SHA256),
	cryptography.WithSignerRandomizer(rand.Reader),
)

func (man *ConnManager) Start() {
	for {
		select {
		case connection := <-man.register:
			man.clients[connection] = true
			log.Println("[server.Start] a connection has been made!")
		case connection := <-man.unregister:
			pk := base64.RawStdEncoding.EncodeToString(connection.PublicKey())
			if connection.LoggedIn() {
				//close(connection.data)
				delete(man.logged, pk)
				log.Printf("[server.Start] client with public key %s\n", pk)
			}
			if _, ok := man.clients[connection]; ok {
				//close(connection.data)
				delete(man.clients, connection)
				log.Println("[server.Start] a connection has terminated!")
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

func StartServer() {
	log.Println("[server.NewServer] starting server")
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Printf("[server.NewServer] net.Listen() returned error: %+v\n", err)
	}
	manager := ConnManager{
		clients:    make(map[client.Client]bool),
		logged:     make(map[string]client.Client),
		broadcast:  make(chan *chatmessage.ChatMessage),
		register:   make(chan client.Client),
		unregister: make(chan client.Client),
	}
	go manager.Start()
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("[server.NewServer] listener.Accept() returned error: %+v\n", err)
		}
		client := client.NewClient(client.WithConnection(connection))
		manager.register <- client
		go manager.Receive(client)
		go manager.Send(client)
	}
}
