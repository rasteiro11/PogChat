# PogChat

A peer to peer encrypted chatting system using rsa algorithm for encryption and TCP protocol for comunication between machines 
- Features
  1. Peer comunication using other user public key
  2. Privacy due to message encryption using [RSA](https://www.rfc-editor.org/rfc/rfc8017)
  3. Anonymity due to public key (it's easy to change user credentials)
  4. You can not fake user sender (every user message must be signed) 
  5. Easy to deploy server node using Docker
  6. Easy to run client
  7. Many client comunication protocol optons (TCP and WebSocket)

- How to run **server**<br>
  ``make``
- How to run **test client**<br>
  ``make me``<br>
  ``make peer``
- How to run **client**<br>
  ``RECEIVER_PUBLIC=<receiverPublicFilePath.key> SENDER_PUBLIC=<senderPublicFilePath.key> SENDER_PRIVATE=<senderPrivateFilePath.key> go run main.go``

Chat to anyone anywhere with privacy and anonymity
  
