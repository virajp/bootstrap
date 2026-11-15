package cli

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// spinFrames are the frames of the spinner of a slow step.
const spinFrames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

// spinner shows `<frame> <text>` on the last line of a terminal while a
// slow step runs. Each write through it clears that line first, so a line
// of the step shows above the spinner.
type spinner struct {
	mu   sync.Mutex
	w    io.Writer
	text string
	stop chan struct{}
	done chan struct{}
}

func (s *spinner) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = io.WriteString(s.w, "\r\x1b[K")
	return s.w.Write(p)
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
// function that it gives. The lines of stderr show above the spinner, and
// stop clears the line of the spinner.
func (a *app) spin(text string) (stop func()) {
	orig := a.stderr
	s := &spinner{w: orig, text: text, stop: make(chan struct{}), done: make(chan struct{})}
	a.stderr = s
	go s.run()
	return func() {
		close(s.stop)
		<-s.done
		_, _ = io.WriteString(orig, "\r\x1b[K")
		a.stderr = orig
	}
}
