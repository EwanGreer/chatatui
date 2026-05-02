package inspect

import (
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case keysMsg:
		m.keys = msg
		m.loading = false
		if m.tab == tabKeys {
			m.applyFilter()
			if len(m.filtered) > 0 {
				m.cursor = 0
				return m, loadKeyDetailCmd(m.rdb, m.filtered[0])
			}
		}
		return m, nil

	case keyDetailMsg:
		m.detail = KeyDetail(msg)
		m.viewport.SetContent(m.detail.Value)
		m.viewport.GotoTop()
		return m, nil

	case channelsMsg:
		m.channels = msg
		m.loading = false
		if m.tab == tabChannels {
			m.applyChanFilter()
		}
		return m, nil

	case chanDataMsg:
		formatted := formatPayload(string(msg))
		m.chanMessages = append(m.chanMessages, formatted)
		m.viewport.SetContent(strings.Join(m.chanMessages, "\n"))
		m.viewport.GotoBottom()
		if m.chanRecv != nil {
			return m, waitForChanMsg(m.chanRecv)
		}
		return m, nil

	case errMsg:
		m.err = msg
		return m, clearErrCmd()

	case clearErrMsg:
		m.err = nil
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		innerHeight := m.height - layoutOuterChrome - layoutHelpBarHeight - layoutTabBarHeight
		detailWidth := m.width - layoutOuterChrome - layoutSidebarWidth - layoutSidebarDivider
		vpHeight := innerHeight - layoutDetailHeader

		if !m.ready {
			m.viewport = viewport.New(detailWidth, vpHeight)
			m.ready = true
		} else {
			m.viewport.Width = detailWidth
			m.viewport.Height = vpHeight
		}

		m.filter.Width = layoutSidebarWidth - 4
		return m, nil

	case tea.KeyMsg:
		if m.focus == focusFilter {
			return m.updateFilter(msg)
		}

		switch msg.String() {
		case "ctrl+c":
			m.cancelSubscription()
			return m, tea.Quit
		case "q":
			m.cancelSubscription()
			return m, tea.Quit

		case "1":
			if m.focus != focusFilter {
				m.switchTab(tabKeys)
				return m, nil
			}
		case "2":
			if m.focus != focusFilter {
				m.switchTab(tabChannels)
				return m, listChannelsCmd(m.rdb)
			}

		case "tab":
			if m.focus == focusSidebar {
				m.focus = focusDetail
			} else {
				m.focus = focusSidebar
			}
			return m, nil

		case "/":
			if m.focus == focusSidebar {
				m.focus = focusFilter
				m.filter.Focus()
				return m, nil
			}

		case "r":
			m.loading = true
			if m.tab == tabKeys {
				return m, scanKeysCmd(m.rdb)
			}
			return m, listChannelsCmd(m.rdb)

		case "esc":
			if m.focus == focusDetail {
				m.focus = focusSidebar
				return m, nil
			}

		case "up", "k":
			if m.focus == focusSidebar {
				return m, m.moveCursor(-1)
			}
		case "down", "j":
			if m.focus == focusSidebar {
				return m, m.moveCursor(1)
			}

		case "enter":
			if m.focus == focusSidebar {
				return m.handleEnter()
			}
		}
	}

	if m.focus == focusDetail {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) switchTab(t tab) {
	if m.tab == t {
		return
	}
	m.cancelSubscription()
	m.tab = t
	m.focus = focusSidebar
	m.filter.SetValue("")
	m.filter.Blur()
	m.detail = KeyDetail{}
	m.viewport.SetContent("")

	if t == tabKeys {
		m.applyFilter()
		m.cursor = 0
	} else {
		m.applyChanFilter()
		m.chanCursor = 0
		m.chanMessages = nil
		m.subscribedTo = ""
	}
}

func (m *Model) moveCursor(delta int) tea.Cmd {
	if m.tab == tabKeys {
		next := m.cursor + delta
		if next < 0 || next >= len(m.filtered) {
			return nil
		}
		m.cursor = next
		return m.loadSelected()
	}
	next := m.chanCursor + delta
	if next < 0 || next >= len(m.chanFiltered) {
		return nil
	}
	m.chanCursor = next
	return nil
}

func (m *Model) handleEnter() (tea.Model, tea.Cmd) {
	if m.tab == tabKeys {
		if len(m.filtered) > 0 {
			m.focus = focusDetail
			return m, m.loadSelected()
		}
		return m, nil
	}

	if len(m.chanFiltered) == 0 {
		return m, nil
	}
	channel := m.chanFiltered[m.chanCursor]
	m.cancelSubscription()
	m.subscribedTo = channel
	m.chanMessages = nil
	m.viewport.SetContent(styleMuted.Render("Listening on " + channel + "..."))
	m.viewport.GotoTop()
	m.focus = focusDetail

	ch, cancel := subscribeChannel(m.rdb, channel)
	m.chanCancelFunc = cancel
	m.chanRecv = ch
	return m, waitForChanMsg(ch)
}

func (m *Model) cancelSubscription() {
	if m.chanCancelFunc != nil {
		m.chanCancelFunc()
		m.chanCancelFunc = nil
		m.chanRecv = nil
	}
}

func formatPayload(payload string) string {
	trimmed := strings.TrimSpace(payload)
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		var buf strings.Builder
		if err := indentJSON(&buf, trimmed); err == nil {
			return buf.String()
		}
	}
	return payload
}

func indentJSON(dst *strings.Builder, src string) error {
	var raw any
	if err := json.Unmarshal([]byte(src), &raw); err != nil {
		return err
	}
	enc := json.NewEncoder(dst)
	enc.SetIndent("", "  ")
	return enc.Encode(raw)
}

func (m *Model) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.focus = focusSidebar
		m.filter.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.filter, cmd = m.filter.Update(msg)

	if m.tab == tabKeys {
		m.applyFilter()
		if m.cursor >= len(m.filtered) {
			m.cursor = max(0, len(m.filtered)-1)
		}
		if len(m.filtered) > 0 {
			return m, tea.Batch(cmd, m.loadSelected())
		}
		m.detail = KeyDetail{}
	} else {
		m.applyChanFilter()
		if m.chanCursor >= len(m.chanFiltered) {
			m.chanCursor = max(0, len(m.chanFiltered)-1)
		}
	}

	return m, cmd
}

func (m *Model) applyFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.filtered = m.keys
		return
	}
	m.filtered = m.filtered[:0]
	for _, k := range m.keys {
		if strings.Contains(strings.ToLower(k), query) {
			m.filtered = append(m.filtered, k)
		}
	}
}

func (m *Model) applyChanFilter() {
	query := strings.ToLower(m.filter.Value())
	if query == "" {
		m.chanFiltered = m.channels
		return
	}
	m.chanFiltered = m.chanFiltered[:0]
	for _, ch := range m.channels {
		if strings.Contains(strings.ToLower(ch), query) {
			m.chanFiltered = append(m.chanFiltered, ch)
		}
	}
}

func (m *Model) loadSelected() tea.Cmd {
	if m.cursor < 0 || m.cursor >= len(m.filtered) {
		return nil
	}
	return loadKeyDetailCmd(m.rdb, m.filtered[m.cursor])
}
