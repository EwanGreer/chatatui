package ui

import "github.com/charmbracelet/lipgloss"

// Layout constants.
const (
	layoutSidebarWidth   = 20
	layoutOuterChrome    = 4 // outer rounded border (1 each side) + internal chrome
	layoutHelpBarHeight  = 1
	layoutSidebarDivider = 1 // vertical bar between sidebar and main
	layoutHeaderHeight   = 1 // room name header above the viewport
	layoutInputHeight    = 3 // input box: 1 text row + top/bottom border
	layoutInputPadding   = 2 // horizontal padding inside input
	layoutViewportBorder = 2 // viewport border top + bottom
	layoutTypingLine     = 1 // typing indicator row beneath viewport
)

// Color palette.
const (
	colorFocus   lipgloss.Color = "62"  // active/focused element
	colorMuted   lipgloss.Color = "240" // inactive borders, muted text
	colorSuccess lipgloss.Color = "40"  // connected state
	colorWarning lipgloss.Color = "220" // reconnecting, warnings
	colorError   lipgloss.Color = "196" // errors
	colorModalBg lipgloss.Color = "235" // modal background
)

// Static styles — defined once, not recreated on every render.
var (
	styleError   = lipgloss.NewStyle().Foreground(colorError)
	styleWarning = lipgloss.NewStyle().Foreground(colorWarning)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleBold    = lipgloss.NewStyle().Bold(true)

	styleTyping   = lipgloss.NewStyle().Foreground(colorMuted).Italic(true).PaddingLeft(1)
	styleHelpKey  = lipgloss.NewStyle().Foreground(colorFocus).Bold(true)
	styleHelpDesc = lipgloss.NewStyle().Foreground(colorMuted)

	styleStateConnected    = lipgloss.NewStyle().Foreground(colorSuccess)
	styleStateConnecting   = lipgloss.NewStyle().Foreground(colorWarning)
	styleStateDisconnected = lipgloss.NewStyle().Foreground(colorError)

	styleModalTitle = lipgloss.NewStyle().Bold(true)
	styleModalHelp  = lipgloss.NewStyle().Foreground(colorMuted)
)
