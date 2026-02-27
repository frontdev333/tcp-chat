package internal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ChatMessage struct {
	ClientID    string    `json:"client_id"`
	ClientType  string    `json:"client_type"`
	Content     string    `json:"content"`
	MessageID   string    `json:"message_id"`
	MessageType string    `json:"message_type"`
	Timestamp   time.Time `json:"timestamp"`
}

func FormatMessage(msg ChatMessage) string {
	timestamp := msg.Timestamp.Format("15:04:05")

	var client string
	if msg.ClientType == "user" {
		client = msg.ClientID + ":"
	} else {
		client = "***"
	}

	return fmt.Sprintf("[%s] %s %s", timestamp, client, msg.Content)
}

func ParseIncomingMessage(raw string, senderID string) ChatMessage {

	clientId := fmt.Sprintf("<User_%s>", senderID)
	msgId := uuid.New().String()

	return ChatMessage{
		ClientID:    clientId,
		ClientType:  "user",
		Content:     raw,
		MessageID:   msgId,
		MessageType: "user",
		Timestamp:   time.Now(),
	}
}
