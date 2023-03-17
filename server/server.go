package server

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"pogchat/cryptography"
	"pogchat/key"
	"pogchat/user_message"
	"time"
)

type Client struct {
	loggedIn  bool
	publicKey []byte
	socket    net.Conn
	data      chan []byte
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
			//fmt.Println("RECEIVED: " + string(message))
		}
	}
}

func StartClient() {

	time.Sleep(time.Second)
	fmt.Println("[client.startClient] starting new client")
	connection, error := net.Dial("tcp", "localhost:42069")
	if error != nil {
		fmt.Println(error)
	}
	client := &Client{socket: connection}
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
	msg, err := json.Marshal(&ChatMessage{
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
		msg, err := json.Marshal(&ChatMessage{
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

type ConnManager struct {
	// TODO: WE NEED A BETTER WAY TO FIND CLIENTS MAYBE HASH PUBLIC KEY ?
	logged     map[string]*Client
	logout     chan *Client
	clients    map[*Client]bool
	broadcast  chan *ChatMessage
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
			chatMsg := &ChatMessage{}

			processedMsg := trimByteSeq(message, '\x00')
			err := json.Unmarshal([]byte(processedMsg), chatMsg)
			if err != nil {
				log.Printf("[server.Receive] json.Unmarshal() returned error: %+v\n", err)
				manager.unregister <- client
				client.socket.Close()
				continue
			}

			if client.loggedIn {
				//manager.broadcast <- message
				manager.broadcast <- chatMsg
				client.data <- []byte("TALOGADOJABURRO")
				continue
			}

			fn, ok := msgProcessor["LOGIN_MSG"]
			if !ok {
				log.Println("[server.Receive] login not found")
				manager.unregister <- client
				client.socket.Close()
				continue
			}

			um, err := fn(chatMsg.Payload)
			if err != nil {
				log.Printf("[server.Receive] json.Unmarshal() returned error: %+v\n", err)
				manager.unregister <- client
				client.socket.Close()
				continue
			}

			client.loggedIn = true
			client.publicKey = um.FromPublicKey()
			manager.logged[base64.RawStdEncoding.EncodeToString(um.FromPublicKey())] = client
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
			log.Println("-----------------------------------------------------------------------------------------------")
			log.Println("[server.Send] sending message to client")
			log.Println("-----------------------------------------------------------------------------------------------")
			log.Println(string(message) + "\n")
			log.Println("-----------------------------------------------------------------------------------------------")
			_, err := client.socket.Write(message)
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

type ChatMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
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

type ChatMsgProcessor map[string]func(string) (user_message.UserMessage, error)

var msgProcessor = ChatMsgProcessor{
	"LOGIN_MSG": func(payload string) (user_message.UserMessage, error) {
		um, err := user_message.ParseFromJSON(payload)
		if err != nil {
			return nil, err
		}
		_, err = signer.Verify(um.FromPublicKey(), um.Message(), um.Signature())
		if err != nil {
			log.Println("[server.LOGIN_MSG] not a valid signature")
			return nil, err
		}

		log.Println("EVERYTHING IS OK WE LOGGED IN")

		return um, nil
	},
	"PEER_MSG": func(payload string) (user_message.UserMessage, error) {
		//um, err := user_message.ParseFromJSON(payload)
		//if err != nil {
		//	log.Println("[server.PEER_MSG] could not parse payload")
		//	return nil, err
		//}
		//_, err = signer.Verify(um.FromPublicKey(), um.Message(), um.Signature())
		//if err != nil {
		//	log.Println("[server.PEER_MSG] not a valid signature")
		//	return nil, err
		//}

		return nil, nil
	},
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
		case chatMsg := <-man.broadcast:

			fn, ok := msgProcessor[chatMsg.Type]
			if !ok {
				log.Printf("[server.Receive] msg processor for msg type %s is not implemented\n", chatMsg.Type)
				continue
			}

			_, err := fn(chatMsg.Payload)
			if err != nil {
				log.Printf("[server.Receive] msg processor returned error: %+v\n", err)
				continue
			}

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

			peer.data <- []byte(chatMsg.Payload)

			//chatMsg := &ChatMessage{}
			//processedMsg := trimByteSeq(message, '\x00')
			//fmt.Println("PROCESSED MSG: ", string(message))
			//err := json.Unmarshal([]byte(processedMsg), chatMsg)
			//if err != nil {
			//	log.Printf("[server.Start] json.Unmarshal() returned error: %+v\n", err)
			//	continue
			//}

			//log.Println("[server.Start] " + string(message))

			//r, ok := msgProcessor[chatMsg.Type]
			//if !ok {
			//	log.Printf("[server.Start] msg of type %s is not supported\n", chatMsg.Type)
			//	continue
			//}

			//_, err = r(chatMsg.Payload)
			//if err != nil {
			//	log.Printf("[server.Start] processed msg returned error: %+v\n", err)
			//	continue
			//}

			//for connection := range man.clients {
			//	select {
			//	//case connection.data <- processedMsg:
			//	default:
			//		close(connection.data)
			//		delete(man.clients, connection)
			//	}
			//}
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
		logged:     make(map[string]*Client),
		broadcast:  make(chan *ChatMessage),
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
