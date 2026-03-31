package service

import (
	"time"

	"github.com/egreerdp/chatatui/internal/repository"
	"github.com/google/uuid"
)

type RoomStore interface {
	GetByID(id uuid.UUID) (*repository.Room, error)
	AddMember(roomID, userID uuid.UUID) error
}

type MessageStore interface {
	Create(msg *repository.Message) error
	GetByRoom(roomID uuid.UUID, limit, offset int) ([]repository.Message, error)
}

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
