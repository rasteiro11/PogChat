package key

import "errors"

var (
	CreateFileError          = errors.New("os.Create() returned error")
	EncodePemFileError       = errors.New("pem.Encode() returned error")
	ReadFileError            = errors.New("read file returned error")
	ParsePemError            = errors.New("could not parse pem key from file")
	DiffKeyTypesLoadingError = errors.New("wrong key types during load")
)

type KeyType int

const (
	RSA_PRIVATE_KEY KeyType = iota
	RSA_PUBLIC_KEY
	NOT_IMPLEMENTED_KEY
)

type KeyPairOpts func(*keyPair) error

type KeyPairLoader interface {
	LoadPrivateKey(fileName string) error
	LoadPublicKey(fileName string) error
}

type KeyPairSerializer interface {
	StorePrivateKey(fileName string) error
	StorePublicKey(fileName string) error
}

type KeyPair interface {
	KeyPairLoader
	KeyPairSerializer
	PrivateKey() []byte
	PublicKey() []byte
}

type KeyPairFactory interface {
	NewKeyPair(bitSize int) (KeyPair, error)
}
