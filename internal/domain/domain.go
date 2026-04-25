package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID     uuid.UUID
	Name   string
	APIKey string
}

type Room struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Members   []RoomMember
}

type RoomMember struct {
	UserID          uuid.UUID
	Name            string
	LastConnectedAt time.Time
}

type Message struct {
	ID        uuid.UUID
	SenderID  uuid.UUID
	Author    string
	Content   string
	RoomID    uuid.UUID
	CreatedAt time.Time
}

type WireMessageType string

const (
	WireMessageTypeChat   WireMessageType = "chat"
	WireMessageTypeSystem WireMessageType = "system"
	WireMessageTypeTyping WireMessageType = "typing"
	WireMessageTypeError  WireMessageType = "error"
)

func (t WireMessageType) String() string { return string(t) }

type WireMessage struct {
	Type      WireMessageType `json:"type"`
	ID        string          `json:"id"`
	Author    string          `json:"author"`
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
}

func (m *WireMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}
