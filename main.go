package main

import (
	"log"
	"net"
	"os"
	"pogchat/client"
	"pogchat/key"
	"pogchat/server"
	"pogchat/user_client"
)

func main() {
	v := os.Getenv("SERVER")

	if v == "server" {
		server.NewServer().Start()
	}

	connection, err := net.Dial("tcp", "localhost:42069")
	if err != nil {
		log.Fatalf("[main] net.Dial() returned error: %+v\n", err)
	}

	client := client.NewClient(client.WithConnection(connection))
	userClient, err := userclient.NewUserClient(userclient.WithClient(client))
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
	userClient.BuildUI()

	userClient.Run()
}
