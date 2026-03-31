package service

import (
	"errors"
	"testing"
	"time"

	"github.com/EwanGreer/chatatui/internal/repository"
	mocks "github.com/EwanGreer/chatatui/internal/service/_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestChatService_GetRoom(t *testing.T) {
	roomID := uuid.New()

	tests := []struct {
		name        string
		setup       func(*mocks.MockRoomStore)
		wantRoom    *RoomInfo
		wantErrIs   error
	}{
		{
			name: "returns RoomInfo for existing room",
			setup: func(m *mocks.MockRoomStore) {
				m.EXPECT().GetByID(roomID).Return(&repository.Room{
					BaseModel: repository.BaseModel{ID: roomID},
					Name:      "general",
				}, nil)
			},
			wantRoom: &RoomInfo{ID: roomID, Name: "general"},
		},
		{
			name: "propagates not found error",
			setup: func(m *mocks.MockRoomStore) {
				m.EXPECT().GetByID(roomID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErrIs: gorm.ErrRecordNotFound,
		},
		{
			name: "propagates unexpected store error",
			setup: func(m *mocks.MockRoomStore) {
				m.EXPECT().GetByID(roomID).Return(nil, errors.New("db unavailable"))
			},
			wantErrIs: errors.New("db unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms := mocks.NewMockRoomStore(t)
			messages := mocks.NewMockMessageStore(t)
			tt.setup(rooms)

			svc := NewChatService(rooms, messages)
			got, err := svc.GetRoom(roomID)

			if tt.wantErrIs != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.wantErrIs.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantRoom, got)
		})
	}
}

func TestChatService_AddRoomMember(t *testing.T) {
	roomID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		setup   func(*mocks.MockRoomStore)
		wantErr bool
	}{
		{
			name: "adds member successfully",
			setup: func(m *mocks.MockRoomStore) {
				m.EXPECT().AddMember(roomID, userID).Return(nil)
			},
		},
		{
			name: "propagates store error",
			setup: func(m *mocks.MockRoomStore) {
				m.EXPECT().AddMember(roomID, userID).Return(errors.New("constraint violation"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms := mocks.NewMockRoomStore(t)
			messages := mocks.NewMockMessageStore(t)
			tt.setup(rooms)

			svc := NewChatService(rooms, messages)
			err := svc.AddRoomMember(roomID, userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestChatService_GetMessageHistory(t *testing.T) {
	roomID := uuid.New()
	msgID := uuid.New()
	now := time.Now().Truncate(time.Second)

	tests := []struct {
		name     string
		setup    func(*mocks.MockMessageStore)
		want     []MessageInfo
		wantErr  bool
	}{
		{
			name: "maps repository messages to MessageInfo",
			setup: func(m *mocks.MockMessageStore) {
				m.EXPECT().GetByRoom(roomID, 50, 0).Return([]repository.Message{
					{
						BaseModel: repository.BaseModel{ID: msgID, CreatedAt: now},
						Content:   []byte("hello"),
						Sender:    repository.User{Name: "alice"},
					},
				}, nil)
			},
			want: []MessageInfo{
				{ID: msgID, Author: "alice", Content: "hello", CreatedAt: now},
			},
		},
		{
			name: "returns empty slice when no messages",
			setup: func(m *mocks.MockMessageStore) {
				m.EXPECT().GetByRoom(roomID, 50, 0).Return([]repository.Message{}, nil)
			},
			want: []MessageInfo{},
		},
		{
			name: "propagates store error",
			setup: func(m *mocks.MockMessageStore) {
				m.EXPECT().GetByRoom(roomID, 50, 0).Return(nil, errors.New("query failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms := mocks.NewMockRoomStore(t)
			messages := mocks.NewMockMessageStore(t)
			tt.setup(messages)

			svc := NewChatService(rooms, messages)
			got, err := svc.GetMessageHistory(roomID, 50, 0)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChatService_PersistMessage(t *testing.T) {
	senderID := uuid.New()
	roomID := uuid.New()
	content := []byte("hello world")

	tests := []struct {
		name    string
		setup   func(*mocks.MockMessageStore)
		wantErr bool
	}{
		{
			name: "persists message and returns ID and timestamp",
			setup: func(m *mocks.MockMessageStore) {
				m.EXPECT().Create(mockAny).RunAndReturn(func(msg *repository.Message) error {
					msg.ID = uuid.New()
					msg.CreatedAt = time.Now()
					return nil
				})
			},
		},
		{
			name: "propagates store error",
			setup: func(m *mocks.MockMessageStore) {
				m.EXPECT().Create(mockAny).Return(errors.New("insert failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rooms := mocks.NewMockRoomStore(t)
			messages := mocks.NewMockMessageStore(t)
			tt.setup(messages)

			svc := NewChatService(rooms, messages)
			id, createdAt, err := svc.PersistMessage(content, senderID, roomID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, uuid.Nil, id)
				assert.Zero(t, createdAt)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, id)
			assert.NotZero(t, createdAt)
		})
	}
}

// mockAny matches any argument — used where the exact value is set inside the function.
var mockAny = mock.MatchedBy(func(_ *repository.Message) bool { return true })
