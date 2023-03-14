package user

type user_message struct {
	Sig    []byte `json:"signature"`
	PubKey []byte `json:"public_key"`
	Msg    []byte `json:"message"`
}

func (m *user_message) Signature() []byte {
	return m.Sig
}

func (m *user_message) PublicKey() []byte {
	return m.PubKey
}

func (m *user_message) Message() []byte {
	return m.Msg
}

func WithSignature(signature []byte) UserMessageOptions {
	return func(u *user_message) {
		u.Sig = signature
	}
}

func WithPublicKey(publicKey []byte) UserMessageOptions {
	return func(u *user_message) {
		u.PubKey = publicKey
	}
}

func WithMessage(message []byte) UserMessageOptions {
	return func(u *user_message) {
		u.Msg = message
	}
}

func NewUserMessage(opts ...UserMessageOptions) UserMessage {
	um := &user_message{}

	for _, opt := range opts {
		opt(um)
	}

	return um
}
