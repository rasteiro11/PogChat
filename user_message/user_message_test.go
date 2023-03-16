package user_message

import (
	"crypto"
	"crypto/rand"
	"crypto/sha256"
	"pogchat/cryptography"
	"pogchat/key"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserBuilder(t *testing.T) {
	test := []struct {
		name string
		opts UserMessageOptions
		user UserMessage
	}{
		{
			name: "with message opt",
			opts: WithMessage([]byte("TIRAICHBADFTHR")),
			user: &user_message{
				Msg: []byte("TIRAICHBADFTHR"),
			},
		},
		{
			name: "with signature opt",
			opts: WithSignature([]byte("TIRAICHBADFTHR")),
			user: &user_message{
				Sig: []byte("TIRAICHBADFTHR"),
			},
		},

		{
			name: "with from public key opt",
			opts: WithFromPublicKey([]byte("TIRAICHBADFTHR")),
			user: &user_message{
				FromPK: []byte("TIRAICHBADFTHR"),
			},
		},
		{
			name: "with to public key opt",
			opts: WithToPublicKey([]byte("TIRAICHBADFTHR")),
			user: &user_message{
				ToPK: []byte("TIRAICHBADFTHR"),
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			um := NewUserMessage(tt.opts)
			assert.Equal(t, um.Message(), tt.user.Message(), "Message must be equal")
			assert.Equal(t, um.Signature(), tt.user.Signature(), "Signature must be equal")
			assert.Equal(t, um.FromPublicKey(), tt.user.FromPublicKey(), "FromPublicKey must be equal")
			assert.Equal(t, um.ToPublicKey(), tt.user.ToPublicKey(), "ToPublicKey must be equal")
		})
	}
}

func TestUserMessage(t *testing.T) {
	c := cryptography.NewCryptor(cryptography.WithHasher(sha256.New()), cryptography.WithRandomizer(rand.Reader))
	s := cryptography.NewSigner(cryptography.WithSignerHasher(crypto.SHA256), cryptography.WithSignerRandomizer(rand.Reader))

	pairSender, _ := key.NewKeyPair(2048)
	pairReceiver, _ := key.NewKeyPair(2048)
	um := NewUserMessage(
		WithFromPublicKey(pairSender.PublicKey()),
		WithToPublicKey(pairReceiver.PublicKey()),
		WithCryptor(c),
		WithSigner(s),
	)

	test := []struct {
		name string
		msg  string
	}{
		{
			name: "with message opt",
			msg:  "hello world",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			encryptedMsg, err := um.GetEncryptedMessage([]byte(tt.msg))
			assert.Nil(t, err, "get encrypted message should be possible")

			decryptedMsg, err := c.Decrypt(pairReceiver.PrivateKey(), encryptedMsg)
			assert.Nil(t, err, "decoding must be possible")
			assert.Equal(t, tt.msg, string(decryptedMsg), "decryption must be possible")

			sig, err := um.GetSignature(pairReceiver.PrivateKey(), encryptedMsg)
			assert.Nil(t, err, "could not sign encrypted message")

			_, err = s.Verify(pairReceiver.PublicKey(), encryptedMsg, sig)
			assert.Nil(t, err, "could not validate signature")
		})
	}
}
