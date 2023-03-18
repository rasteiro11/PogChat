package server

import "pogchat/client"

type ConnectionManager interface {
	Receive(c client.Client)
	Send(c client.Client)
	Register(c client.Client) error
	Unregister(c client.Client) error
	Start()
}

type Server interface {
	Start()
}

type ServerOpts func(*server)
