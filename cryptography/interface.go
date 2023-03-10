package cryptography

type Encryptor interface {
	Encrypt(otherPublic []byte, msg []byte) ([]byte, error)
}

type Decryptor interface {
	Decrypt(myPrivate []byte, encryptedMsg []byte) ([]byte, error)
}

type CryptorOpts func(*cryptor)

type Cryptor interface {
	Encryptor
	Decryptor
}

type SignerOpts func(*signer)

type Signer interface {
	Sign(myPrivate []byte, msg []byte) ([]byte, error)
	Verify(otherPublic []byte, msg []byte, signature []byte) (bool, error)
}
