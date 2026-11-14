package tui

import "charm.land/bubbles/v2/key"

// keyMap holds the keys of the main screen. The key line shows the short
// view of the help component, and ? shows the full view in a modal dialog.
type keyMap struct {
	Move, Pane, Scroll, Space, AcceptAll, Tab, Search, Help, Quit key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Move:   key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "move")),
		Pane:   key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←→", "pane")),
		Scroll: key.NewBinding(key.WithKeys("shift+up", "shift+down"), key.WithHelp("shift ↑↓", "scroll the details")),
		Space:  key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "select")),
		// AcceptAll is a key of the Detected tab only (keysFor).
		AcceptAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "accept all"), key.WithDisabled()),
		Tab:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
		Search:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// ShortHelp gives the keys of the key line.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Move, k.Pane, k.Space, k.AcceptAll, k.Tab, k.Search, k.Help, k.Quit}
}

// FullHelp gives each key of the screen, one line for each key.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Move, k.Pane, k.Scroll, k.Space, k.AcceptAll, k.Tab, k.Search, k.Help, k.Quit}}
}

// keysFor gives the keys of the tab <tab>: on the Detected tab, space
// accepts a proposal and a accepts all of them.
func (k keyMap) keysFor(tab int) keyMap {
	if tab == detectedTab {
		k.Space.SetHelp("space", "accept")
		k.AcceptAll.SetEnabled(true)
	}
	return k
}
