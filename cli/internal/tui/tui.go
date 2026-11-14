// Package tui is the full-screen TUI of `bootstrap tui` and `bootstrap mac
// tui`, which `init` and `mac init` open in a terminal. It is for a person;
// the command line is for an agent. internal/cli parses the arguments, does
// the preflight and takes the clone, then calls Run.
//
// The TUI builds the change list of internal/core one key at a time, and
// calls core for each read and each write. It never imports internal/cli,
// and no string literal of it names a tool: each text of a tool comes from
// the data.
//
// Contract: docs/design/cli.md.
package tui

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/virajp/bootstrap/cli/internal/core"
	"github.com/virajp/bootstrap/cli/internal/fail"
)

// Options are the inputs of Run.
type Options struct {
	// Side is the side, its values file and its clone, after the preflight.
	Side core.Side
	// Init is init in a terminal: a values file that is not there, or that
	// has an old form, is the first change of the list.
	Init bool
	// NoBootstrap skips each mise bootstrap step of the Mac sync at Apply.
	NoBootstrap bool
	// Home is the home folder, shown as ~ in the header.
	Home   string
	Stdin  io.Reader
	Stdout io.Writer
}

// Run runs the TUI to its end and gives its exit code. A quit by the user
// gives code 0. A problem before the TUI opens (the size of the terminal,
// the values file) gives its code and the error, which the caller prints.
// Run never exits the process.
func Run(o Options) (fail.Code, error) {
	m := newModel(o)
	p := tea.NewProgram(m, tea.WithInput(o.Stdin), tea.WithOutput(o.Stdout))
	end, err := p.Run()
	if errors.Is(err, tea.ErrInterrupted) {
		return fail.OK, nil
	}
	if err != nil {
		return fail.Other, &core.Problem{Code: fail.Other, Msg: "the TUI stopped: " + err.Error()}
	}
	if fm, ok := end.(*Model); ok && fm.err != nil {
		return fail.CodeOf(fm.err), fm.err
	}
	return fail.OK, nil
}

// The phases of the TUI.
type phase int

const (
	starting phase = iota // No size of the terminal yet.
	loading               // The size is good; the values file loads.
	ready                 // The screen shows.
)

// Model is the Bubble Tea model of the TUI.
type Model struct {
	opt   Options
	keys  keyMap
	help  help.Model
	phase phase
	w, h  int
	err   error // The problem that stops the TUI before it opens.

	// The change list of core: init, the items and the answers.
	list core.List
	// fixed are the changes that the user cannot remove: the new values
	// file, or the migration of its old form.
	fixed []string
	// state is the values file and its selection, when it has the new form.
	state *core.State

	tabs  []*tab
	cur   int    // The tab that shows.
	modal *modal // The modal dialog over the screen, or nil.
	tick  int    // The ticks since the cursor moved: the scroll of its text.
}

// tickInterval is the time of one step of the scroll of a long text.
const tickInterval = 160 * time.Millisecond

// The names of the tabs, in their order.
var tabNames = []string{"Tools", "Packs", "Detected"}

func newModel(o Options) *Model {
	h := help.New()
	h.ShortSeparator = "  "
	h.FullSeparator = "  "
	h.Styles.ShortKey, h.Styles.ShortDesc, h.Styles.ShortSeparator = stDim, stDim, stDim
	h.Styles.FullKey, h.Styles.FullDesc, h.Styles.FullSeparator = stHint, stPlain, stDim
	h.Styles.Ellipsis = stDim
	h.SetWidth(screenW - 2)
	m := &Model{opt: o, keys: newKeyMap(), help: h, list: core.List{Init: o.Init}}
	for _, n := range tabNames {
		t := &tab{name: n}
		t.left.title = "Categories"
		if n == "Packs" {
			t.left.title = "Packs"
		}
		m.tabs = append(m.tabs, t)
	}
	return m
}

// changes gives the count of the changes of the header: the fixed changes
// and the items of the change list.
func (m *Model) changes() int { return len(m.fixed) + len(m.list.Items) }

// Messages of the model.
type (
	loadedMsg struct {
		fixed string
		state *core.State
		err   error
	}
	tickMsg struct{}
)

// Init is a part of tea.Model. The TUI waits for the size of the terminal.
func (m *Model) Init() tea.Cmd { return nil }

// load reads the values file. init takes a missing values file, or an old
// form, as the first change of the list; tui needs a values file of the new
// form.
func (m *Model) load() tea.Msg {
	sd := m.opt.Side
	if m.opt.Init {
		text, err := core.InitChange(sd)
		if err != nil || text != "" {
			return loadedMsg{fixed: text, err: err}
		}
	}
	st, err := core.Open(sd)
	return loadedMsg{state: st, err: err}
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

// small tells if the terminal is smaller than the screen.
func (m *Model) small() bool { return m.w < screenW || m.h < screenH }

// sizeText gives the first line of the size message of <cmd>.
func (m *Model) sizeText(cmd string) string {
	return fmt.Sprintf("%s needs a terminal of %d × %d or larger. This terminal is %d × %d.", cmd, screenW, screenH, m.w, m.h)
}

// command gives the command of the TUI: init, or tui.
func (m *Model) command() string {
	if m.opt.Init {
		return core.CmdName(m.opt.Side.Mac, "init")
	}
	return core.CmdName(m.opt.Side.Mac, "tui")
}

// Update is a part of tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		if m.phase != starting {
			return m, nil
		}
		if m.small() {
			// The TUI does not start in a small terminal (ruling 17).
			cmd := m.command()
			fix := "make the terminal larger, then run " + cmd + " again"
			if m.opt.Init {
				fix += ", or run " + cmd + " --no-tui"
			}
			m.err = &core.Problem{Code: fail.Usage, Msg: m.sizeText(cmd), Fix: fix}
			return m, tea.Quit
		}
		m.phase = loading
		return m, m.load
	case loadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		if msg.fixed != "" {
			m.fixed = append(m.fixed, msg.fixed)
		}
		m.state = msg.state
		m.phase = ready
		return m, m.tickCmd()
	case tickMsg:
		m.tick++
		return m, m.tickCmd()
	case tea.KeyPressMsg:
		if m.phase != ready {
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		if m.modal != nil {
			return m, m.modal.key(m, msg)
		}
		if m.small() {
			if key.Matches(msg, m.keys.Quit) {
				return m, m.quit()
			}
			return m, nil
		}
		return m, m.mainKey(msg)
	}
	return m, nil
}

// mainKey handles a key of the main screen.
func (m *Model) mainKey(msg tea.KeyPressMsg) tea.Cmd {
	t := m.tabs[m.cur]
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m.quit()
	case key.Matches(msg, m.keys.Help):
		m.modal = m.helpModal()
	case key.Matches(msg, m.keys.Scroll):
		if msg.String() == "shift+up" {
			t.scrollBottom(-1)
		} else {
			t.scrollBottom(1)
		}
	case key.Matches(msg, m.keys.Move):
		if msg.String() == "up" {
			t.active().move(-1)
		} else {
			t.active().move(1)
		}
		m.tick = 0
	case key.Matches(msg, m.keys.Pane):
		t.onRight = msg.String() == "right"
		m.tick = 0
	case key.Matches(msg, m.keys.Tab):
		m.cur = (m.cur + 1) % len(m.tabs)
		m.tick = 0
	}
	return nil
}

// quit ends the TUI with an empty change list, else asks "Discard N
// changes? [y/N]" (rulings 23 and 24).
func (m *Model) quit() tea.Cmd {
	n := m.changes()
	if n == 0 {
		return tea.Quit
	}
	m.modal = &modal{
		title: "Quit",
		width: 50,
		body:  []line{{text("Discard " + plural(n, "change") + "? [y/N]")}},
		keys:  "y discard and quit  any other key goes back",
		onKey: func(m *Model, k tea.KeyPressMsg) tea.Cmd {
			m.modal = nil
			if k.String() == "y" {
				return tea.Quit
			}
			return nil
		},
	}
	return nil
}

// helpModal gives the help dialog: each key of the screen, one line for
// each key, with the full view of the help component (D4).
func (m *Model) helpModal() *modal {
	var body []line
	for _, l := range strings.Split(m.help.FullHelpView(m.keys.FullHelp()), "\n") {
		body = append(body, line{{text: l, raw: true}})
	}
	return &modal{
		title: "Keys",
		width: 50,
		body:  body,
		keys:  "↑↓ scroll  esc close",
		onKey: func(m *Model, k tea.KeyPressMsg) tea.Cmd {
			switch k.String() {
			case "esc":
				m.modal = nil
			case "up":
				m.modal.scroll(-1)
			case "down":
				m.modal.scroll(1)
			}
			return nil
		},
	}
}

// View is a part of tea.Model.
func (m *Model) View() tea.View {
	if m.phase != ready {
		return tea.NewView("")
	}
	var lines []line
	switch {
	case m.small():
		lines = m.smallLines()
	default:
		lines = m.screen()
		if m.modal != nil {
			lines = m.modal.over(lines)
		}
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		out[i] = l.render()
	}
	v := tea.NewView(strings.Join(out, "\n"))
	v.AltScreen = true
	return v
}

// smallLines gives the message of a terminal that became too small during
// a run: the change list stays (ruling 17).
func (m *Model) smallLines() []line {
	kept := "Make the terminal larger to continue."
	if n := m.changes(); n > 0 {
		kept += " Your " + plural(n, "change") + " are kept."
		if n == 1 {
			kept = "Make the terminal larger to continue. Your 1 change is kept."
		}
	}
	last := line{styled(" q quit", stDim)}
	if m.modal != nil && len(m.modal.body) > 0 {
		last = append(line{text(" ")}, m.modal.body[len(m.modal.body)-1]...)
	}
	return []line{
		{},
		{styled(" "+m.sizeText(core.CmdName(m.opt.Side.Mac, "tui")), stRemove)},
		{styled(" "+kept, stDim)},
		{},
		last,
	}
}

// screen gives the 25 lines of the main screen.
func (m *Model) screen() []line {
	t := m.tabs[m.cur]
	out := m.header()
	out = append(out, line{styled(strings.Repeat("─", leftW)+"┬"+strings.Repeat("─", rightW), stDim)})
	titleStyle := func(on bool) lipgloss.Style {
		if on {
			return stHead
		}
		return stDim
	}
	out = append(out, line{
		styled(" "+fit(t.left.title, leftW-1), titleStyle(!t.onRight)),
		styled("│", stDim),
		styled(" "+fit(t.right.title, rightW-1), titleStyle(t.onRight)),
	})
	left := t.left.lines(leftW, !t.onRight, m.tick)
	right := t.right.lines(rightW, t.onRight, m.tick)
	for i := range paneH {
		l := append(line{}, left[i]...)
		l = append(l, styled("│", stDim))
		out = append(out, append(l, right[i]...))
	}
	out = append(out, line{styled(strings.Repeat("─", leftW)+"┴"+strings.Repeat("─", rightW), stDim)})
	out = append(out, t.bottomLines()...)
	out = append(out, line{styled(strings.Repeat("─", screenW), stDim)})
	out = append(out, line{text(" "), {text: m.help.ShortHelpView(m.keys.ShortHelp()), raw: true}})
	return out
}

// header gives the two lines of the header: the command, the folder and
// the count of the changes; then the tabs and the values file.
func (m *Model) header() []line {
	sd := m.opt.Side
	left := " " + core.CmdName(sd.Mac, "tui") + " · "
	where := sd.Root
	if !sd.Mac {
		left += "repo · "
	} else {
		where = sd.Values
	}
	left += m.home(where)
	right := ""
	if n := m.changes(); n > 0 {
		right = plural(n, "change") + " "
	}
	// A long folder is cut, so the count of the changes shows.
	left = strings.TrimRight(fit(left, min(len([]rune(left)), screenW-len([]rune(right))-1)), " ")
	gap := max(1, screenW-len([]rune(left))-len([]rune(right)))
	first := line{styled(left, stHead), text(strings.Repeat(" ", gap)), styled(right, stAdd)}

	tabs := line{text(" ")}
	for i, t := range m.tabs {
		if i > 0 {
			tabs = append(tabs, styled(" | ", stDim))
		}
		if i == m.cur {
			tabs = append(tabs, styled("["+t.name+"]", stHead))
		} else {
			tabs = append(tabs, styled(t.name, stDim))
		}
	}
	file := strings.TrimPrefix(strings.TrimPrefix(sd.Values, strings.TrimSuffix(sd.Root, "/")), "/") + " "
	gap = max(1, screenW-tabs.width()-len([]rune(file)))
	tabs = append(tabs, text(strings.Repeat(" ", gap)), styled(file, stDim))
	return []line{clip(first, screenW), clip(tabs, screenW)}
}

// home gives <p> with the home folder as ~.
func (m *Model) home(p string) string {
	h := strings.TrimSuffix(m.opt.Home, "/")
	if h != "" && (p == h || strings.HasPrefix(p, h+"/")) {
		return "~" + strings.TrimPrefix(p, h)
	}
	return p
}
