package main

import (
	"encoding/base64"
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
	"time"

	"github.com/marcusolsson/tui-go"
)

type UserClient struct {
	pair           key.KeyPair
	receiver       key.KeyPair
	publicKeyFile  string
	privateKeyFile string
	client         client.Client
	recChan        chan []byte
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

func (c *UserClient) GetUserName() string {
	pk := c.pair.PublicKey()
	return string(base64.RawStdEncoding.EncodeToString(c.pair.PublicKey()[len(pk)-10:]))
}

func (c *UserClient) SetReceiver(receiver key.KeyPair) {
	c.receiver = receiver
}

func (c *UserClient) SendMessage(text string) error {
	userInputMsg := user_message.NewUserMessage(
		user_message.WithFromPublicKey(c.pair.PublicKey()),
		user_message.WithToPublicKey(c.receiver.PublicKey()),
	)

	encryptedMsg, err := userInputMsg.GetEncryptedMessage([]byte(text))
	if err != nil {
		log.Printf("[server.StartClient] userInputMsg.GetEncryptedMessage() returned error: %+v\n", err)
		return err
	}

	_, err = userInputMsg.GetSignature(c.pair.PrivateKey(), encryptedMsg)
	if err != nil {
		log.Printf("[userClient.SendMessage] userInputMsg.GetSignature() returned error: %+v\n", err)
		return err
	}

	Msg, err := userInputMsg.MarshalJSON()
	if err != nil {
		log.Println("[userClient.SendMessage] could not msrhal json")
		return err
	}

	msg, err := json.Marshal(&chatmessage.ChatMessage{
		Type:    "PEER_MSG",
		Payload: string(Msg),
	})
	if err != nil {
		log.Println("[userClient.SendMessage] could not marshal json")
		return err
	}

	_, err = c.client.Write(msg)
	if err != nil {
		log.Println("[userClient.SendMessage] could not marshal json")
		return err
	}

	return nil
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
		recChan:        make(chan []byte),
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

	go c.client.ReceiveAndDecrypt(c.pair.PrivateKey(), c.recChan)

	return c, nil
}

type message struct {
	username string
	message  string
	time     string
}

var messages = []message{}

func main() {
	v := os.Getenv("SERVER")

	if v == "server" {
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
	history := tui.NewVBox()

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	input.OnSubmit(func(e *tui.Entry) {
		err := userClient.SendMessage(e.Text())
		if err != nil {
			history.Append(tui.NewHBox(
				tui.NewLabel(time.Now().String()),
				tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", userClient.GetUserName()))),
				tui.NewLabel(fmt.Sprintf("[ERROR] could not send message: %+v", err)),
				tui.NewSpacer(),
			))
		} else {
			history.Append(tui.NewHBox(
				tui.NewLabel(time.Now().String()),
				tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", userClient.GetUserName()))),
				tui.NewLabel(e.Text()),
				tui.NewSpacer(),
			))
		}
		input.SetText("")
	})

	root := tui.NewHBox(chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	go func() {
		for {
			select {
			case message, ok := <-userClient.recChan:
				if !ok {
					history.Append(tui.NewHBox(
						tui.NewLabel(time.Now().String()),
						tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", userClient.GetUserName()))),
						tui.NewLabel(fmt.Sprintf("[ERROR] could not receive message: %+v", err)),
						tui.NewSpacer(),
					))
				}
				history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().String()),
					tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", userClient.GetUserName()))),
					tui.NewLabel(string(message)),
					tui.NewSpacer(),
				))
				ui.Repaint()
			}
		}
	}()

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}

}
