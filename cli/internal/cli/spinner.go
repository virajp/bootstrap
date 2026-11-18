package cli

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/virajp/bootstrap/cli/internal/report"
)

// spinFrames are the frames of the spinner of a slow step.
const spinFrames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

// spinner shows `<frame> <text>` on the last line of a terminal while a
// slow step runs. Each write through one of its writers clears that line
// first, so a line of the step shows above the spinner.
type spinner struct {
	mu   sync.Mutex
	w    io.Writer // The terminal of the spinner line: stderr.
	text string
	stop chan struct{}
	done chan struct{}
}

// spinWriter is stdout or stderr while a spinner shows.
type spinWriter struct {
	s  *spinner
	to io.Writer
}

func (sw spinWriter) Write(p []byte) (int, error) {
	sw.s.mu.Lock()
	defer sw.s.mu.Unlock()
	_, _ = io.WriteString(sw.s.w, "\r\x1b[K")
	return sw.to.Write(p)
}

func (s *spinner) run() {
	defer close(s.done)
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()
	f := []rune(spinFrames)
	for i := 0; ; i++ {
		s.mu.Lock()
		_, _ = fmt.Fprintf(s.w, "\r\x1b[K%c %s", f[i%len(f)], s.text)
		s.mu.Unlock()
		select {
		case <-s.stop:
			return
		case <-t.C:
		}
	}
}

// spin shows a spinner with <text> on stderr until the call of the stop
// function that it gives, when stderr is a terminal and the command has no
// --json. The lines of stderr, and of stdout when it is a terminal too, show
// above the spinner, and stop clears the line of the spinner. A step that
// can ask the user for input does not spin.
func (a *app) spin(text string) (stop func()) {
	if !a.progress {
		return func() {}
	}
	origErr, origOut := a.stderr, a.stdout
	s := &spinner{w: origErr, text: text, stop: make(chan struct{}), done: make(chan struct{})}
	a.stderr = spinWriter{s, origErr}
	if f, ok := origOut.(*os.File); ok && report.IsTerminal(f) {
		a.stdout = spinWriter{s, origOut}
	}
	go s.run()
	return func() {
		close(s.stop)
		<-s.done
		_, _ = io.WriteString(origErr, "\r\x1b[K")
		a.stderr, a.stdout = origErr, origOut
	}
}
