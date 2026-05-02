package inspect

import "github.com/charmbracelet/lipgloss"

const (
	layoutSidebarWidth   = 30
	layoutOuterChrome    = 4
	layoutHelpBarHeight  = 1
	layoutTabBarHeight   = 1
	layoutSidebarDivider = 1
	layoutDetailHeader   = 4 // key + type + ttl + blank line
)

const (
	colorFocus   lipgloss.Color = "62"
	colorMuted   lipgloss.Color = "240"
	colorSuccess lipgloss.Color = "40"
	colorWarning lipgloss.Color = "220"
	colorError   lipgloss.Color = "196"
)

var (
	styleError   = lipgloss.NewStyle().Foreground(colorError)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleHelpKey = lipgloss.NewStyle().Foreground(colorFocus).Bold(true)
	styleHelp    = lipgloss.NewStyle().Foreground(colorMuted)

	styleSelected = lipgloss.NewStyle().Foreground(colorFocus).Bold(true)
	styleTypeTag  = lipgloss.NewStyle().Foreground(colorSuccess)
	styleTTL      = lipgloss.NewStyle().Foreground(colorWarning)
)
