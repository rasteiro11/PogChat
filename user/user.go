package user

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"pogchat/cryptography"
	"pogchat/key"
	"pogchat/user_message"
)

type User interface {
	LoadKeyPair(privateFile string, publicFile string) (key.KeyPair, error)
	BuildPeerMessage(msg string) (user_message.UserMessage, error)
}

type user struct {
	cryptor       cryptography.Cryptor
	signer        cryptography.Signer
	pair          key.KeyPair
	peerPublicKey []byte
}

func (u *user) BuildPeerMessage(msg string) (user_message.UserMessage, error) {
	um := user_message.NewUserMessage(
		user_message.WithSigner(u.signer),
		user_message.WithCryptor(u.cryptor),
		user_message.WithFromPublicKey(u.pair.PublicKey()),
		user_message.WithToPublicKey(u.peerPublicKey))

	enctyptedMsg, err := um.GetEncryptedMessage([]byte(msg))
	if err != nil {
		return nil, err
	}

	_, err = um.GetSignature(u.pair.PrivateKey(), enctyptedMsg)
	if err != nil {
		return nil, err
	}

	return um, nil
}

func (u *user) LoadKeyPair(privateFile string, publicFile string) (key.KeyPair, error) {
	pair, err := key.LoadKeyPair(key.WithPublicKey(publicFile), key.WithPrivateKey(privateFile))
	if err != nil {
		return nil, err
	}

	u.pair = pair

	return pair, nil
}

func WithSigner(signer cryptography.Signer) UserOptions {
	return func(u *user) {
		u.signer = signer
	}
}

func WithCryptor(cryptor cryptography.Cryptor) UserOptions {
	return func(u *user) {
		u.cryptor = cryptor
	}
}

func WithUserKeyPair(pair key.KeyPair) UserOptions {
	return func(u *user) {
		u.pair = pair
	}

}

func WithPeerPublicKey(peerPublicKey []byte) UserOptions {
	return func(u *user) {
		u.peerPublicKey = peerPublicKey
	}

}

type UserOptions func(*user)

func NewUser(opts ...UserOptions) User {
	u := &user{
		cryptor: cryptography.NewCryptor(
			cryptography.WithHasher(sha256.New()),
			cryptography.WithRandomizer(rand.Reader)),
		signer: cryptography.NewSigner(
			cryptography.WithSignerHasher(crypto.SHA256),
			cryptography.WithSignerRandomizer(rand.Reader)),
	}

	for _, opt := range opts {
		opt(u)
	}

	return u
}
