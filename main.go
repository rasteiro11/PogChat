package main

import (
	"pogchat/server"
)

//func runServer() {
//	l, err := net.Listen("tcp4", ":42069")
//	defer l.Close()
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	c, err := l.Accept()
//	if err != nil {
//		fmt.Errorf("Accept failed: %+v\n", err)
//		return
//	}
//
//	go handlePeer(c)
//}

func main() {
	//pair, _ := key.NewKeyPair(2048)
	//cryptor := cryptography.NewCryptor(cryptography.WithHasher(sha1.New()), cryptography.WithRandomizer(rand.Reader))
	//go runServer()
	//time.Sleep(time.Second * 5)
	//scannerStdin := bufio.NewScanner(os.Stdin)
	//fmt.Print("Server message: ")
	//c, err := net.Dial("tcp", "127.0.0.1:42069")
	//if err != nil {
	//	fmt.Printf("ERROR: %+v\n", err)
	//	return
	//}
	//for scannerStdin.Scan() {
	//	text := scannerStdin.Text()

	//	encryptedMsg, err := cryptor.Encrypt(pair.PublicKey(), []byte(text))
	//	if err != nil {
	//		fmt.Printf("ERROR: %+v\n", err)
	//	}

	//	// send to server
	//	fmt.Println("SENT THIS TO SERVER: ", base64.RawURLEncoding.EncodeToString(encryptedMsg))
	//	_, err = fmt.Fprintf(c, base64.RawURLEncoding.EncodeToString(encryptedMsg)+"\n")
	//	if err != nil {
	//		fmt.Printf("ERROR SEND TO SERVER: %+v\n", err)
	//	}

	//	break
	//}

	//// listen for reply
	//serverResponse, _ := bufio.NewReader(c).ReadString('\n')

	//dec, err := base64.RawURLEncoding.DecodeString(string(serverResponse[:len(serverResponse)-1]))
	//if err != nil {
	//	fmt.Printf("ERROR DECODING BASE64: %+v\n", err)
	//}

	//decryptedMsg, err := cryptor.Decrypt(pair.PrivateKey(), dec)
	//if err != nil {
	//	fmt.Printf("ERROR: %+v\n", err)
	//}

	//fmt.Println("FINAL RESPONSE: ", string(decryptedMsg))

	server.StartServer()
}

//func handlePeer(c net.Conn) {
//	for {
//		data, err := bufio.NewReader(c).ReadSlice('\n')
//		if err != nil {
//			fmt.Printf("ERROR: %+v\n", err)
//			break
//		}
//
//		nBytes, err := c.Write([]byte(data))
//		if err != nil {
//			fmt.Printf("ERROR: %+v\n", err)
//			break
//		}
//
//		if nBytes == 0 {
//			fmt.Printf("Closing connection with: %s\n", c.RemoteAddr().String())
//			return
//		}
//	}
//	c.Close()
//}
