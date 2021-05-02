package events

import (
	"time"

	"github.com/google/uuid"
)

// MessageType
type MessageType string

const (
	Query   MessageType = "query"
	Command MessageType = "command"
	Default MessageType = "default"
)

type Message struct {
	ID         string      `json:"id"`
	Payload    string      `json:"payload"`
	MessagType MessageType `json:"message_type"`
	Timestamp  int64       `json:"timestamp"`
}

func NewMessage(payload string, messagType MessageType) Message {
	return Message{
		uuid.New().String(),
		payload,
		messagType,
		time.Now().Unix(),
	}
}
