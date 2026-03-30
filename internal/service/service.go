package service

import (
	"time"

	"github.com/google/uuid"
)

type RoomInfo struct {
	ID   uuid.UUID
	Name string
}

type MessageInfo struct {
	ID        uuid.UUID
	Author    string
	Content   string
	CreatedAt time.Time
}

type ChatService interface {
	GetRoom(id uuid.UUID) (*RoomInfo, error)
	AddRoomMember(roomID, userID uuid.UUID) error
	GetMessageHistory(roomID uuid.UUID, limit, offset int) ([]MessageInfo, error)
	PersistMessage(content []byte, senderID, roomID uuid.UUID) (uuid.UUID, time.Time, error)
}
