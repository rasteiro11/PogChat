package chatmessage

type ChatMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}
