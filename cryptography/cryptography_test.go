package cryptography

import (
	"crypto"
	"crypto/rand"
	"crypto/sha1"
	"pogchat/key"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptMessage(t *testing.T) {
	pair, err := key.NewKeyPair(2048)
	assert.Nil(t, err, "could not generate pair keys")
	encryptedMsg := []byte("")

	test := []struct {
		name string
		f    func(*testing.T)
	}{
		{
			name: "encrypt message",
			f: func(t *testing.T) {
				cryptor := NewCryptor(WithHasher(sha1.New()), WithRandomizer(rand.Reader))
				encryptedMsg, err = cryptor.Encrypt(pair.PublicKey(), []byte("TIRAICHBADFTHR"))
				assert.Nil(t, err, "could not encrypt message")
			},
		},
		{
			name: "decrypt message",
			f: func(t *testing.T) {
				cryptor := NewCryptor(WithHasher(sha1.New()), WithRandomizer(rand.Reader))
				decryptedMsg, err := cryptor.Decrypt(pair.PrivateKey(), encryptedMsg)
				assert.Nil(t, err, "could not decrypt message")
				assert.Equal(t, decryptedMsg, []byte("TIRAICHBADFTHR"), "both messages must be equal")
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			tt.f(t)
		})
	}
	assert.NotNil(t, encryptedMsg, "encryption failed")

}

func TestSignMessage(t *testing.T) {
	pair, err := key.NewKeyPair(2048)
	assert.Nil(t, err, "could not generate pair keys")
	encryptedMsg := []byte("")
	test := []struct {
		name string
		f    func(*testing.T)
	}{
		{
			name: "sign message",
			f: func(t *testing.T) {
				cryptor := NewCryptor(WithHasher(sha1.New()), WithRandomizer(rand.Reader))
				encryptedMsg, err = cryptor.Encrypt(pair.PublicKey(), []byte("TIRAICHBADFTHR"))
				assert.Nil(t, err, "could not encrypt message")
				signer := NewSigner(WithSignerHasher(crypto.SHA256), WithSignerRandomizer(rand.Reader))
				signature, err := signer.Sign(pair.PrivateKey(), []byte("TIRAICHBADFTHR"))
				assert.Nil(t, err, "could not sign message")
				valid, err := signer.Verify(pair.PublicKey(), []byte("TIRAICHBADFTHR"), signature)
				assert.Nil(t, err, "error during verification")
				assert.Equal(t, valid, true, "verifification failed")
			},
		},
		{
			name: "sign validation error",
			f: func(t *testing.T) {
				cryptor := NewCryptor(WithHasher(sha1.New()), WithRandomizer(rand.Reader))
				encryptedMsg, err = cryptor.Encrypt(pair.PublicKey(), []byte("TIRAICHBADFTHR"))
				assert.Nil(t, err, "could not encrypt message")
				signer := NewSigner(WithSignerHasher(crypto.SHA256), WithSignerRandomizer(rand.Reader))
				signature, err := signer.Sign(pair.PrivateKey(), []byte("TIRAICHBADFTHR"))
				assert.Nil(t, err, "could not sign message")
				valid, err := signer.Verify(pair.PublicKey(), []byte("TIRAICHBADFTHR DEVE FALHAR"), signature)
				assert.NotNil(t, err, "error during verification")
				assert.Equal(t, valid, false, "verifification must fail")
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			tt.f(t)
		})
	}

	assert.NotNil(t, encryptedMsg, "encryption failed")
}
