package user_message

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"pogchat/cryptography"
)

type user_message struct {
	Sig     []byte               `json:"signature"`
	FromPK  []byte               `json:"from_public_key"`
	ToPK    []byte               `json:"to_public_key"`
	Msg     []byte               `json:"message"`
	cryptor cryptography.Cryptor `json:"-"`
	signer  cryptography.Signer  `json:"-"`
}

func (m *user_message) MarshalJSON() ([]byte, error) {
	return json.Marshal(*m)
}

func (m *user_message) Signature() []byte {
	return m.Sig
}

func (m *user_message) FromPublicKey() []byte {
	return m.FromPK
}

func (m *user_message) ToPublicKey() []byte {
	return m.ToPK
}

func (m *user_message) Message() []byte {
	return m.Msg
}

func (m *user_message) GetEncryptedMessage(msg []byte) ([]byte, error) {
	encryptedMsg, err := m.cryptor.Encrypt(m.ToPK, msg)
	if err != nil {
		return nil, err
	}

	m.Msg = encryptedMsg

	return m.Msg, nil
}

func (m *user_message) GetSignature(fromPrivateKey []byte, encryptedMsg []byte) ([]byte, error) {
	sig, err := m.signer.Sign(fromPrivateKey, encryptedMsg)
	if err != nil {
		return nil, err
	}

	m.Sig = sig

	return m.Sig, nil
}

func ParseFromJSON(um string) (UserMessage, error) {
	m := &user_message{}
	err := json.Unmarshal([]byte(um), m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func WithSigner(signer cryptography.Signer) UserMessageOptions {
	return func(u *user_message) {
		u.signer = signer
	}
}

func WithCryptor(cryptor cryptography.Cryptor) UserMessageOptions {
	return func(u *user_message) {
		u.cryptor = cryptor
	}
}

func WithSignature(signature []byte) UserMessageOptions {
	return func(u *user_message) {
		u.Sig = signature
	}
}

func WithToPublicKey(publicKey []byte) UserMessageOptions {
	return func(u *user_message) {
		u.ToPK = publicKey
	}
}

func WithFromPublicKey(publicKey []byte) UserMessageOptions {
	return func(u *user_message) {
		u.FromPK = publicKey
	}
}

func WithMessage(message []byte) UserMessageOptions {
	return func(u *user_message) {
		u.Msg = message
	}
}

func NewUserMessage(opts ...UserMessageOptions) UserMessage {
	um := &user_message{
		cryptor: cryptography.NewCryptor(
			cryptography.WithHasher(sha256.New()),
			cryptography.WithRandomizer(rand.Reader)),
		signer: cryptography.NewSigner(
			cryptography.WithSignerHasher(crypto.SHA256),
			cryptography.WithSignerRandomizer(rand.Reader)),
	}

	for _, opt := range opts {
		opt(um)
	}

	return um
}
