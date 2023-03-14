package user

type UserMessage interface {
	Signature() []byte
	FromPublicKey() []byte
	ToPublicKey() []byte
	Message() []byte
	GetEncryptedMessage([]byte) ([]byte, error)
	GetSignature([]byte, []byte) ([]byte, error)
}

type UserMessageOptions func(*user_message)
