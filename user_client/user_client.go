package userclient

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	chatmessage "pogchat/chat_message"
	"pogchat/client"
	"pogchat/key"
	"pogchat/user_message"
	"time"

	"github.com/marcusolsson/tui-go"
)

type userClient struct {
	pair           key.KeyPair
	receiver       key.KeyPair
	publicKeyFile  string
	privateKeyFile string
	client         client.Client
	recChan        chan []byte
	ui             tui.UI
	history        *tui.Box
}

func WithPublicKeyFile(file string) UserClientOpts {
	return func(uc *userClient) {
		uc.publicKeyFile = file
	}
}

func WithPrivateKeyFile(file string) UserClientOpts {
	return func(uc *userClient) {
		uc.privateKeyFile = file
	}
}

func WithClient(c client.Client) UserClientOpts {
	return func(uc *userClient) {
		uc.client = c
	}
}

func (c *userClient) GetUsername() string {
	pk := c.pair.PublicKey()
	return string(base64.RawStdEncoding.EncodeToString(pk[len(pk)-10:]))
}

func (c *userClient) GetPeername() string {
	pk := c.receiver.PublicKey()
	return string(base64.RawStdEncoding.EncodeToString(pk[len(pk)-10:]))
}

func (c *userClient) SetReceiver(receiver key.KeyPair) {
	c.receiver = receiver
}

func (c *userClient) SendMessage(text string) error {
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

func (c *userClient) Login() error {
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

func (c *userClient) BuildUI() error {
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
		err := c.SendMessage(e.Text())
		if err != nil {
			history.Append(tui.NewHBox(
				tui.NewLabel(time.Now().String()),
				tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", c.GetUsername()))),
				tui.NewLabel(fmt.Sprintf("[ERROR] could not send message: %+v", err)),
				tui.NewSpacer(),
			))
		} else {
			history.Append(tui.NewHBox(
				tui.NewLabel(time.Now().String()),
				tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", c.GetUsername()))),
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

	c.ui = ui
	c.history = history

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	return nil
}

func (u *userClient) Run() error {
	go func() {
		for {
			select {
			case message, ok := <-u.recChan:
				if !ok {
					u.history.Append(tui.NewHBox(
						tui.NewLabel(time.Now().String()),
						tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", u.GetUsername()))),
						tui.NewLabel(fmt.Sprintf("[ERROR] could not receive message: %+v", errors.New("something went wrong decrypting message"))),
						tui.NewSpacer(),
					))
				}
				u.history.Append(tui.NewHBox(
					tui.NewLabel(time.Now().String()),
					tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", u.GetUsername()))),
					tui.NewLabel(string(message)),
					tui.NewSpacer(),
				))
				u.ui.Repaint()
			}
		}
	}()
	if err := u.ui.Run(); err != nil {
		return err
	}
	return nil
}

func NewUserClient(opts ...UserClientOpts) (*userClient, error) {
	c := &userClient{
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
