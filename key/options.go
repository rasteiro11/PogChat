package key

func WithPrivateKey(fileName string) KeyPairOpts {
	return func(kp *keyPair) error {
		return kp.LoadPrivateKey(fileName)
	}
}

func WithPublicKey(fileName string) KeyPairOpts {
	return func(kp *keyPair) error {
		return kp.LoadPublicKey(fileName)
	}
}
