package service

import (
	"time"

	"github.com/EwanGreer/chatatui/internal/domain"
	"github.com/google/uuid"
)

type ChatService struct {
	rooms    RoomStore
	messages MessageStore
}

func NewChatService(rooms RoomStore, messages MessageStore) *ChatService {
	return &ChatService{rooms: rooms, messages: messages}
}

func (s *ChatService) GetRoom(id uuid.UUID) (*domain.Room, error) {
	return s.rooms.GetByID(id)
}

func (s *ChatService) AddRoomMember(roomID, userID uuid.UUID) error {
	return s.rooms.AddMember(roomID, userID)
}

func (s *ChatService) GetMessageHistory(roomID uuid.UUID, limit, offset int) ([]domain.Message, error) {
	return s.messages.GetByRoom(roomID, limit, offset)
}

func (s *ChatService) PersistMessage(content []byte, senderID, roomID uuid.UUID) (uuid.UUID, time.Time, error) {
	msg := &domain.Message{
		Content:  string(content),
		SenderID: senderID,
		RoomID:   roomID,
	}
	if err := s.messages.Create(msg); err != nil {
		return uuid.Nil, time.Time{}, err
	}
	return msg.ID, msg.CreatedAt, nil
}
