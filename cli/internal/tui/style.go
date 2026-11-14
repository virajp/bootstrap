package tui

import "charm.land/lipgloss/v2"

// The colors of the terminal of the design "Bootstrap TUI Screens". The
// design gives one dark look of the terminal, and no light value, so the TUI
// takes no LightDark pair. The plain text and the titles keep the color of
// the terminal, so they read on a light background too.
var (
	colText   = lipgloss.Color("#d5ddd8") // The text on the cursor row.
	colDim    = lipgloss.Color("#8d9992")
	colSel    = lipgloss.Color("#7fd19b") // Selected.
	colDep    = lipgloss.Color("#74b8d6") // A dependency, and the detect hint.
	colLock   = lipgloss.Color("#e3c06b") // Required, and the scroll marks.
	colAdd    = lipgloss.Color("#9be37f")
	colRemove = lipgloss.Color("#f08a7a")
	colCursor = lipgloss.Color("#2a3a3a") // The background of the cursor row.
	colBehind = lipgloss.Color("#4d5753") // The screen behind a modal dialog.
)

// The styles of the screen.
var (
	stPlain  = lipgloss.NewStyle()
	stHead   = lipgloss.NewStyle().Bold(true)
	stDim    = lipgloss.NewStyle().Foreground(colDim)
	stSel    = lipgloss.NewStyle().Foreground(colSel)
	stDep    = lipgloss.NewStyle().Foreground(colDep)
	stLock   = lipgloss.NewStyle().Foreground(colLock)
	stAdd    = lipgloss.NewStyle().Foreground(colAdd).Bold(true)
	stRemove = lipgloss.NewStyle().Foreground(colRemove).Bold(true)
	stHint   = lipgloss.NewStyle().Foreground(colDep).Bold(true)
	stMark   = stLock
	stBehind = lipgloss.NewStyle().Foreground(colBehind)
	stButton = lipgloss.NewStyle().Background(colCursor).Foreground(colText)
)
