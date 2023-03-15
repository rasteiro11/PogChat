package user_message

import "encoding/json"

type UserMessage interface {
	Signature() []byte
	FromPublicKey() []byte
	ToPublicKey() []byte
	Message() []byte
	GetEncryptedMessage(msg []byte) ([]byte, error)
	GetSignature(fromPrivateKey []byte, encryptedMsg []byte) ([]byte, error)
	json.Marshaler
}

type UserMessageOptions func(*user_message)
