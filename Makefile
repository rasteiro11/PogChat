all:
	SERVER=server go run main.go

me:
	RECEIVER_PUBLIC=./test_credentials/receiverPublic.key SENDER_PUBLIC=./test_credentials/senderPublic.key SENDER_PRIVATE=./test_credentials/senderPrivate.key go run main.go

peer:
	RECEIVER_PUBLIC=./test_credentials/senderPublic.key SENDER_PUBLIC=./test_credentials/receiverPublic.key SENDER_PRIVATE=./test_credentials/receiverPrivate.key go run main.go

