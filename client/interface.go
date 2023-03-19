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
	ReceiveAndDecrypt(private []byte, rec chan []byte)
}

type ClientOpts func(*client)
