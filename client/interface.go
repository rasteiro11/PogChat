package client

type Client interface {
	LoggedIn() bool
	PublicKey() []byte
	SetLoggedIn(loggedIn bool)
	SetPublicKey(publicKey []byte)
	Close() error
	Read(buf []byte) (int, error)
	Write(buf []byte) (int, error)
	WriteToChan() chan []byte
	Receive()
	ReceiveAndDecrupt(private []byte)
}

type ClientOpts func(*client)
