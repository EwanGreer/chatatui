package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/websocket"
)

func TestEnterKeepsMessageOverServerLimit(t *testing.T) {
	m := NewModel(Config{})
	// Update only schedules the send. A connection marker lets us check
	// whether it schedules one without opening a socket.
	m.conn = &websocket.Conn{}
	text := strings.Repeat("é", 151)
	m.input.SetValue(text)

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := next.(Model)
	if cmd != nil {
		t.Error("Enter scheduled a send for a message above the server limit")
	}
	if got := updated.input.Value(); got != text {
		t.Errorf("input = %q, want the unsent message", got)
	}
	if got := updated.charCountIndicator(); got != styleError.Render("0/300") {
		t.Errorf("counter = %q, want the zero-remaining warning", got)
	}
}

func TestEnterAcceptsMessageAtServerLimit(t *testing.T) {
	m := NewModel(Config{})
	m.conn = &websocket.Conn{}
	m.input.SetValue(strings.Repeat("é", 150))

	if got := m.charCountIndicator(); got != styleError.Render("0/300") {
		t.Errorf("counter = %q, want the zero-remaining warning", got)
	}
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("Enter did not schedule a send at the server limit")
	}
	if got := next.(Model).input.Value(); got != "" {
		t.Error("input was not cleared after scheduling a send")
	}
}
