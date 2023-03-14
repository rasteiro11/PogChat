package user

type UserMessage interface {
	Signature() []byte
	PublicKey() []byte
	Message() []byte
}

type UserMessageOptions func(*user_message)
