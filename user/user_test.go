package user

import (
	"crypto"
	"crypto/rand"
	"pogchat/cryptography"
	"pogchat/key"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserBuilder(t *testing.T) {
	s := cryptography.NewSigner(cryptography.WithSignerHasher(crypto.SHA256), cryptography.WithSignerRandomizer(rand.Reader))

	pairSender, _ := key.NewKeyPair(2048)
	pairReceiver, _ := key.NewKeyPair(2048)

	user := NewUser(
		WithUserKeyPair(pairSender),
		WithPeerPublicKey(pairReceiver.PublicKey()))

	test := []struct {
		name         string
		verifyResult bool
		msg          string
	}{
		{
			name:         "with message opt",
			verifyResult: true,
			msg:          "hello world",
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			um, err := user.BuildPeerMessage(tt.msg)
			assert.Nil(t, err, "message could not be built")

			_, err = um.MarshalJSON()
			assert.Nil(t, err, "could not marshal user message to JSON")

			ok, err := s.Verify(pairSender.PublicKey(), um.Message(), um.Signature())
			assert.Nil(t, err, "verification must be possible")
			assert.Equal(t, ok, tt.verifyResult, "verification must be true")
		})
	}

}
