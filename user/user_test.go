package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeys(t *testing.T) {
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
			name: "with public key opt",
			opts: WithPublicKey([]byte("TIRAICHBADFTHR")),
			user: &user_message{
				PubKey: []byte("TIRAICHBADFTHR"),
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			um := NewUserMessage(tt.opts)
			assert.Equal(t, um.Message(), tt.user.Message(), "Message must be equal")
			assert.Equal(t, um.Signature(), tt.user.Signature(), "Signature must be equal")
			assert.Equal(t, um.PublicKey(), tt.user.PublicKey(), "PublicKey must be equal")
		})
	}
}
