package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/EwanGreer/chatatui/internal/domain"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/websocket"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.fetchMe(), m.fetchRooms(), m.tickCmd())
}

func (m Model) fetchMe() tea.Cmd {
	return func() tea.Msg {
		var me struct {
			Name string `json:"name"`
		}
		if err := m.api.do("GET", "/me", nil, &me, http.StatusOK); err != nil {
			return errMsg(err)
		}
		return meMsg(me.Name)
	}
}

func (m Model) fetchRooms() tea.Cmd {
	return func() tea.Msg {
		var rooms []Room
		if err := m.api.do("GET", "/rooms", nil, &rooms, http.StatusOK); err != nil {
			return errMsg(err)
		}
		return roomsMsg(rooms)
	}
}

func (m Model) createRoom(name string) tea.Cmd {
	return func() tea.Msg {
		var room Room
		if err := m.api.do("POST", "/rooms", map[string]string{"name": name}, &room, http.StatusCreated); err != nil {
			return errMsg(err)
		}
		return roomCreatedMsg(room)
	}
}

func clearFlashCmd() tea.Cmd {
	return tea.Tick(4*time.Second, func(time.Time) tea.Msg {
		return clearFlashMsg{}
	})
}

func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) connectToRoom(roomID string) tea.Cmd {
	return func() tea.Msg {
		if m.conn != nil {
			_ = m.conn.Close(websocket.StatusNormalClosure, "switching rooms")
		}

		url := m.config.wsURL("/ws/" + roomID)

		ctx := context.Background()
		conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{
			HTTPHeader: http.Header{
				"Authorization": []string{m.config.APIKey},
			},
		})
		if err != nil {
			return errMsg(err)
		}

		return connectedMsg{roomID: roomID, conn: conn}
	}
}

func (m *Model) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		if m.conn == nil {
			return nil
		}

		_, data, err := m.conn.Read(context.Background())
		if err != nil {
			// Ignore normal close errors (happens when switching rooms)
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return nil
			}
			return errMsg(err)
		}

		var wire wireMessage
		if err := json.Unmarshal(data, &wire); err == nil {
			if wire.Type == domain.WireMessageTypeTyping.String() {
				return typingMsg(wire.Author)
			}
			return incomingMsg{formatted: formatWireMessage(data), author: wire.Author}
		}

		return incomingMsg{formatted: string(data)}
	}
}

func sendMessageCmd(conn *websocket.Conn, text string) tea.Cmd {
	return func() tea.Msg {
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(text)); err != nil {
			return errMsg(err)
		}
		return nil
	}
}

func sendTypingCmd(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		msg := &domain.WireMessage{Type: domain.WireMessageTypeTyping}
		data, err := msg.Marshal()
		if err != nil {
			return errMsg(err)
		}
		if err := conn.Write(context.Background(), websocket.MessageText, data); err != nil {
			return errMsg(err)
		}
		return nil
	}
}
