package model

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type ResponderMode struct {
	gorm.Model
	Mode bool `gorm:"default:true"`
}

type Message struct {
	ID                      uuid.UUID
	Command, Value, Service string
}

type MessagePubSub struct {
	ID                uuid.UUID
	Response, Service string
}
