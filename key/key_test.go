package key

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGenerateKeys(t *testing.T) {
	pair, err := NewKeyPair(2048)
	assert.Nil(t, err, "could not generate pair keys")

	test := []struct {
		name string
		f    func(KeyPair, *testing.T)
	}{
		{
			name: "store public key store",
			f: func(pair KeyPair, t *testing.T) {
				err = pair.StorePublicKey("public.key")
				assert.Nil(t, err, "could not store public key")
			},
		},
		{
			name: "store private key store",
			f: func(pair KeyPair, t *testing.T) {
				err = pair.StorePrivateKey("private.key")
				assert.Nil(t, err, "could not store private key")
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			tt.f(pair, t)
		})
	}
}

func TestLoadKeys(t *testing.T) {
	test := []struct {
		name string
		f    func(*testing.T)
	}{
		{
			name: "load public key store",
			f: func(t *testing.T) {
				pair, err := LoadKeyPair(WithPublicKey("public.key"))
				assert.Nil(t, err, "did not load public key")
				assert.NotNil(t, pair.PublicKey(), "public key did not load")
				assert.Nil(t, pair.PrivateKey(), "private key must be nil")
			},
		},
		{
			name: "load private key store",
			f: func(t *testing.T) {
				pair, err := LoadKeyPair(WithPrivateKey("private.key"))
				assert.Nil(t, err, "did not load private key")
				assert.NotNil(t, pair.PrivateKey(), "private key did not load")
				assert.Nil(t, pair.PublicKey(), "public key must be nil")
			},
		},
		{
			name: "load public and private keys",
			f: func(t *testing.T) {
				pair, err := LoadKeyPair(WithPrivateKey("private.key"), WithPublicKey("public.key"))
				assert.Nil(t, err, "did not load private key")
				assert.NotNil(t, pair.PrivateKey(), "private key did not load")
				assert.NotNil(t, pair.PublicKey(), "public key did not load")
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			tt.f(t)
		})
	}
}

func TestClean(t *testing.T) {
	os.Remove("public.key")
	os.Remove("private.key")
}
