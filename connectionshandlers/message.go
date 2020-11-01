package connectionshandlers

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID      string `json:"id"`
	Payload string `json:"payload"`
	Time    int64  `json:"timestamp"`
}

func NewMessage(payload string) Message {
	return Message{uuid.New().String(), payload, time.Now().Unix()}
}
