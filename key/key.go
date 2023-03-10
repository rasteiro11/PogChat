package key

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
)

var keyType = map[KeyType]string{
	RSA_PRIVATE_KEY: "RSA PRIVATE KEY",
	RSA_PUBLIC_KEY:  "RSA PUBLIC KEY",
}

func (kt KeyType) String() string {
	k, ok := keyType[kt]
	if !ok {
		panic("not implemented key")
	}
	return k
}

type keyPair struct {
	publicKey  []byte
	privateKey []byte
}

func (p *keyPair) PublicKey() []byte {
	return p.publicKey
}

func (p *keyPair) PrivateKey() []byte {
	return p.privateKey
}

func (p *keyPair) LoadPrivateKey(fileName string) error {
	blk, err := loadKey(RSA_PRIVATE_KEY, fileName)
	if err != nil {
		return err
	}

	p.privateKey = blk.Bytes

	return nil
}

func (p *keyPair) LoadPublicKey(fileName string) error {
	blk, err := loadKey(RSA_PUBLIC_KEY, fileName)
	if err != nil {
		return err
	}

	p.publicKey = blk.Bytes

	return nil
}

func loadKey(kt KeyType, fileName string) (*pem.Block, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, CreateFileError
	}

	privateKey, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, ReadFileError
	}

	blk, _ := pem.Decode(privateKey)
	if blk == nil {
		return nil, ParsePemError
	}

	if keyType[kt] != blk.Type {
		return nil, DiffKeyTypesLoadingError
	}

	return blk, nil
}

func storeKey(keyType KeyType, key []byte, fileName string) error {
	blk := pem.Block{
		Type:  keyType.String(),
		Bytes: key,
	}

	file, err := os.Create(fileName)
	if err != nil {
		return CreateFileError
	}

	err = pem.Encode(file, &blk)
	if err != nil {
		return EncodePemFileError
	}

	return nil
}

func (p *keyPair) StorePrivateKey(fileName string) error {
	return storeKey(RSA_PRIVATE_KEY, p.privateKey, fileName)
}

func (p *keyPair) StorePublicKey(fileName string) error {
	return storeKey(RSA_PUBLIC_KEY, p.publicKey, fileName)
}

func LoadKeyPair(opts ...KeyPairOpts) (KeyPair, error) {
	pair := &keyPair{}

	for _, opt := range opts {
		err := opt(pair)
		if err != nil {
			return nil, err
		}
	}

	return pair, nil
}

func NewKeyPair(bitSize int) (KeyPair, error) {
	key, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	return &keyPair{
		privateKey: x509.MarshalPKCS1PrivateKey(key),
		publicKey:  x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}, nil
}
