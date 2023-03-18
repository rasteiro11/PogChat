package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	chatmessage "pogchat/chat_message"
	"pogchat/client"
	"pogchat/key"
	"pogchat/server"
	"pogchat/user_message"
)

type UserClient struct {
	pair           key.KeyPair
	receiver       key.KeyPair
	publicKeyFile  string
	privateKeyFile string
	client         client.Client
}

type UserClientOpts func(*UserClient)

func WithPublicKeyFile(file string) UserClientOpts {
	return func(uc *UserClient) {
		uc.publicKeyFile = file
	}
}

func WithPrivateKeyFile(file string) UserClientOpts {
	return func(uc *UserClient) {
		uc.privateKeyFile = file
	}
}

func WithClient(c client.Client) UserClientOpts {
	return func(uc *UserClient) {
		uc.client = c
	}
}

func (c *UserClient) SetReceiver(receiver key.KeyPair) {
	c.receiver = receiver
}

func (c *UserClient) HandleUserInput() {
	userInputMsg := user_message.NewUserMessage(
		user_message.WithFromPublicKey(c.pair.PublicKey()),
		user_message.WithToPublicKey(c.receiver.PublicKey()),
	)

	reader := bufio.NewReader(os.Stdin)

	//cryptor := cryptography.NewCryptor()
	for {
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')

		encryptedMsg, err := userInputMsg.GetEncryptedMessage([]byte(text))
		if err != nil {
			log.Printf("[server.StartClient] userInputMsg.GetEncryptedMessage() returned error: %+v\n", err)
			return
		}

		_, err = userInputMsg.GetSignature(c.pair.PrivateKey(), encryptedMsg)
		if err != nil {
			log.Printf("[server.StartClient] userInputMsg.GetSignature() returned error: %+v\n", err)
			return
		}

		Msg, err := userInputMsg.MarshalJSON()
		if err != nil {
			log.Println("[server.StartClient] could not msrhal json")
			return
		}

		msg, err := json.Marshal(&chatmessage.ChatMessage{
			Type:    "PEER_MSG",
			Payload: string(Msg),
		})
		if err != nil {
			log.Println("[server.StartClient] could not marshal json")
			return
		}

		c.client.Write(msg)
	}

}

func (c *UserClient) Login() error {
	um := user_message.NewUserMessage(
		user_message.WithFromPublicKey(c.pair.PublicKey()),
		user_message.WithToPublicKey(c.pair.PublicKey()),
	)

	encryptedMsg, err := um.GetEncryptedMessage([]byte("GAMER"))
	if err != nil {
		log.Println("[Login] could not encrypt msg")
		return err
	}

	_, err = um.GetSignature(c.pair.PrivateKey(), encryptedMsg)
	if err != nil {
		log.Println("[Login] could not sign message")
		return err
	}

	userMsg, err := um.MarshalJSON()
	if err != nil {
		log.Println("[server.StartClient] could not msrhal json")
		return err
	}

	msg, err := json.Marshal(&chatmessage.ChatMessage{
		Type:    "LOGIN_MSG",
		Payload: string(userMsg),
	})
	if err != nil {
		log.Println("[Login] could not marshal json")
		return err
	}

	_, err = c.client.Write(msg)
	if err != nil {
		log.Printf("[Login] c.client.Write() returned error: %+v\n", err)
		return err

	}

	log.Println("[Login] user is now logged in")
	return nil
}

func NewUserClient(opts ...UserClientOpts) (*UserClient, error) {
	c := &UserClient{
		publicKeyFile:  os.Getenv("SENDER_PUBLIC"),
		privateKeyFile: os.Getenv("SENDER_PRIVATE"),
	}

	pair, err := key.LoadKeyPair(key.WithPublicKey(c.publicKeyFile), key.WithPrivateKey(c.privateKeyFile))
	if err != nil {
		log.Println("[NewUserClient] could not load key pair")
		return nil, err
	}

	for _, opt := range opts {
		opt(c)
	}

	c.pair = pair

	go c.client.ReceiveAndDecrupt(c.pair.PrivateKey())

	return c, nil
}

func main() {
	v := os.Getenv("SERVER")

	if v == "server" {
		fmt.Println("THIS IS ENV VAR: ", v)
		server.NewServer().Start()
	}

	connection, error := net.Dial("tcp", "localhost:42069")
	if error != nil {
		fmt.Println(error)
	}

	client := client.NewClient(client.WithConnection(connection))
	userClient, err := NewUserClient(WithClient(client))
	if err != nil {
		log.Printf("[main.NewUserClient] NewUserMessage() returned error %+v\n", err)
		return
	}

	err = userClient.Login()
	if err != nil {
		log.Printf("[main.Login] userClient.Login() returned error: %+v\n", err)
		return
	}

	receiver, err := key.LoadKeyPair(key.WithPublicKey(os.Getenv("RECEIVER_PUBLIC")))
	if err != nil {
		log.Println("[main] could not load key pair")
		return
	}

	userClient.SetReceiver(receiver)

	userClient.HandleUserInput()

}
