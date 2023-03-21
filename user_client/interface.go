package userclient

import "pogchat/key"

type UserClient interface {
	GetUsername() string
	GetPeername() string
	SetReceiver(r key.KeyPair)
	SendMessage(text string) error
	Login() error
	BuildUI() error
	Run() error
}

var _ UserClient = (*userClient)(nil)

type UserClientOpts func(*userClient)
