package inspect

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return styleMuted.Render("Connecting to Redis...")
	}

	tabBar := m.renderTabBar()
	sidebar := m.renderSidebar()
	detail := m.renderDetail()

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, detail)
	help := m.renderHelp()

	appStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorMuted).
		Width(m.width - 2).
		Height(m.height - 2)

	return appStyle.Render(lipgloss.JoinVertical(lipgloss.Left, tabBar, content, help))
}

func (m Model) renderTabBar() string {
	keysTab := " 1:Keys "
	chansTab := " 2:Channels "

	if m.tab == tabKeys {
		keysTab = styleSelected.Render(keysTab)
		chansTab = styleMuted.Render(chansTab)
	} else {
		keysTab = styleMuted.Render(keysTab)
		chansTab = styleSelected.Render(chansTab)
	}

	return " " + keysTab + " " + chansTab
}

func (m Model) renderSidebar() string {
	innerHeight := m.height - layoutOuterChrome - layoutHelpBarHeight - layoutTabBarHeight

	style := lipgloss.NewStyle().
		Width(layoutSidebarWidth).
		Height(innerHeight).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderRight(true)

	if m.focus == focusSidebar || m.focus == focusFilter {
		style = style.BorderForeground(colorFocus)
	} else {
		style = style.BorderForeground(colorMuted)
	}

	if m.tab == tabKeys {
		return style.Render(m.renderKeysSidebar(innerHeight))
	}
	return style.Render(m.renderChannelsSidebar(innerHeight))
}

func (m Model) renderKeysSidebar(innerHeight int) string {
	var header string
	if m.loading {
		header = styleBold.Render("Keys") + styleMuted.Render(" scanning...")
	} else {
		header = styleBold.Render(fmt.Sprintf("Keys (%d)", len(m.filtered)))
	}

	filterLine := m.filter.View()
	listHeight := innerHeight - 3
	keyList := m.renderItemList(m.filtered, m.cursor, listHeight)

	return header + "\n" + filterLine + "\n" + keyList
}

func (m Model) renderChannelsSidebar(innerHeight int) string {
	var header string
	if m.loading {
		header = styleBold.Render("Channels") + styleMuted.Render(" refreshing...")
	} else {
		header = styleBold.Render(fmt.Sprintf("Channels (%d)", len(m.chanFiltered)))
	}

	filterLine := m.filter.View()
	listHeight := innerHeight - 3
	chanList := m.renderItemList(m.chanFiltered, m.chanCursor, listHeight)

	return header + "\n" + filterLine + "\n" + chanList
}

func (m Model) renderItemList(items []string, cursor, height int) string {
	if len(items) == 0 {
		return styleMuted.Render("(none)")
	}

	scrollOff := 0
	if cursor >= height {
		scrollOff = cursor - height + 1
	}

	var b strings.Builder
	end := scrollOff + height
	if end > len(items) {
		end = len(items)
	}

	for i := scrollOff; i < end; i++ {
		name := items[i]
		if len(name) > layoutSidebarWidth-4 {
			name = name[:layoutSidebarWidth-7] + "..."
		}

		if i == cursor {
			b.WriteString(styleSelected.Render("> "+name) + "\n")
		} else {
			b.WriteString("  " + name + "\n")
		}
	}

	return b.String()
}

func (m Model) renderDetail() string {
	innerHeight := m.height - layoutOuterChrome - layoutHelpBarHeight - layoutTabBarHeight

	style := lipgloss.NewStyle().
		Height(innerHeight).
		Padding(0, 1)

	if m.err != nil {
		return style.Render(styleError.Render("Error: " + m.err.Error()))
	}

	if m.tab == tabChannels {
		return style.Render(m.renderChannelDetail())
	}

	return style.Render(m.renderKeyDetail())
}

func (m Model) renderKeyDetail() string {
	if m.detail.Key == "" {
		return styleMuted.Render("Select a key to inspect")
	}

	keyLine := styleBold.Render("Key:  ") + m.detail.Key
	typeLine := styleBold.Render("Type: ") + styleTypeTag.Render(m.detail.Type)
	ttlLine := styleBold.Render("TTL:  ") + styleTTL.Render(formatTTL(m.detail.TTL))

	return keyLine + "\n" + typeLine + "\n" + ttlLine + "\n\n" + m.viewport.View()
}

func (m Model) renderChannelDetail() string {
	if m.subscribedTo == "" {
		return styleMuted.Render("Select a channel and press enter to subscribe")
	}

	header := styleBold.Render("Channel: ") + styleTypeTag.Render(m.subscribedTo)
	countLine := styleBold.Render("Messages: ") + fmt.Sprintf("%d", len(m.chanMessages))

	return header + "\n" + countLine + "\n\n" + m.viewport.View()
}

func formatTTL(d time.Duration) string {
	switch {
	case d == -1:
		return "no expiry"
	case d == -2:
		return "key not found"
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	default:
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
}

func (m Model) renderHelp() string {
	var keys []struct{ key, desc string }

	switch {
	case m.focus == focusFilter:
		keys = []struct{ key, desc string }{
			{"enter/esc", "done"},
			{"type", "filter"},
		}
	case m.focus == focusSidebar:
		keys = []struct{ key, desc string }{
			{"1/2", "tab"},
			{"j/k", "navigate"},
			{"/", "filter"},
			{"enter", "select"},
			{"r", "refresh"},
			{"tab", "detail"},
			{"q", "quit"},
		}
	case m.focus == focusDetail:
		keys = []struct{ key, desc string }{
			{"1/2", "tab"},
			{"j/k", "scroll"},
			{"esc/tab", "back"},
			{"r", "refresh"},
			{"q", "quit"},
		}
	}

	var items []string
	for _, k := range keys {
		items = append(items, styleHelpKey.Render(k.key)+" "+styleHelp.Render(k.desc))
	}

	return "  " + strings.Join(items, styleHelp.Render(" • "))
}
