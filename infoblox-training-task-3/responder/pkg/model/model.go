package model

import "github.com/google/uuid"

type Message struct {
	ID                      uuid.UUID
	Command, Value, Service string
}

type MessagePubSub struct {
	ID                uuid.UUID
	Response, Service string
}
