package user

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"pogchat/cryptography"
	"pogchat/key"
	"testing"
)

func TestUserBuilder(t *testing.T) {
	s := cryptography.NewSigner(cryptography.WithSignerHasher(crypto.SHA256), cryptography.WithSignerRandomizer(rand.Reader))

	pairSender, _ := key.NewKeyPair(2048)
	pairReceiver, _ := key.NewKeyPair(2048)

	user := NewUser(
		WithUserKeyPair(pairSender),
		WithPeerPublicKey(pairReceiver.PublicKey()))

	msg := "MSG GAMER"

	um, _ := user.BuildPeerMessage(msg)

	toBeSent, _ := um.MarshalJSON()

	fmt.Println("MESSAGE SERIALIZED: ", string(toBeSent))

	fmt.Println("ENCRYPTED MSG: ", base64.RawStdEncoding.EncodeToString(um.Message()))
	fmt.Println("SIGNATURE: ", base64.RawStdEncoding.EncodeToString(um.Signature()))

	ok, _ := s.Verify(pairSender.PublicKey(), um.Message(), um.Signature())
	if ok {
		fmt.Println("GAMERS GAMING")
	}

	t.Fatalf("WE ARE GETTING TO THE END")

}
