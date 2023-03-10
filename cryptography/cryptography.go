package cryptography

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"hash"
	"io"
)

type cryptor struct {
	r      io.Reader
	hasher hash.Hash
}

var _ Encryptor = (*cryptor)(nil)
var _ Decryptor = (*cryptor)(nil)

func (e *cryptor) Encrypt(otherPublic []byte, msg []byte) ([]byte, error) {
	publicKey, err := x509.ParsePKCS1PublicKey(otherPublic)
	if err != nil {
		return nil, err
	}
	encryptedMsg, err := rsa.EncryptOAEP(e.hasher, e.r, publicKey, msg, []byte(""))
	if err != nil {
		return nil, err
	}
	return encryptedMsg, nil
}

func (e *cryptor) Decrypt(myPrivate []byte, encryptedMsg []byte) ([]byte, error) {
	privateKey, err := x509.ParsePKCS1PrivateKey(myPrivate)
	if err != nil {
		return nil, err
	}
	decryptedMsg, err := rsa.DecryptOAEP(e.hasher, e.r, privateKey, encryptedMsg, []byte(""))
	if err != nil {
		return decryptedMsg, err
	}
	return decryptedMsg, nil
}

func WithHasher(hasher hash.Hash) CryptorOpts {
	return func(c *cryptor) {
		c.hasher = hasher
	}
}

func WithRandomizer(r io.Reader) CryptorOpts {
	return func(c *cryptor) {
		c.r = r
	}
}

func NewCryptor(opts ...CryptorOpts) Cryptor {
	c := &cryptor{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type signer struct {
	r      io.Reader
	hasher crypto.Hash
}

var _ Signer = (*signer)(nil)

func WithSignerHasher(hasher crypto.Hash) SignerOpts {
	return func(s *signer) {
		s.hasher = hasher
	}
}

func WithSignerRandomizer(r io.Reader) SignerOpts {
	return func(s *signer) {
		s.r = r
	}
}

func NewSigner(opts ...SignerOpts) Signer {
	c := &signer{}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
func (s *signer) Sign(myPrivate []byte, msg []byte) ([]byte, error) {
	privateKey, err := x509.ParsePKCS1PrivateKey(myPrivate)
	if err != nil {
		return nil, err
	}

	digest := sha256.Sum256(msg)

	signature, err := rsa.SignPKCS1v15(s.r, privateKey, s.hasher, digest[:])
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (s *signer) Verify(otherPublic []byte, msg []byte, signature []byte) (bool, error) {
	publicKey, err := x509.ParsePKCS1PublicKey(otherPublic)
	if err != nil {
		return false, err
	}

	digest := sha256.Sum256(msg)

	err = rsa.VerifyPKCS1v15(publicKey, s.hasher, digest[:], signature)
	if err != nil {
		return false, err
	}

	return true, nil
}
