// Package fail holds the typed errors of the sync engine and their exit codes.
//
// A failure of one file does not stop the other files, and a run returns the
// highest code of all its failures. Contract: docs/design/cli.md ("The
// exit codes").
package fail

import (
	"errors"
	"fmt"
)

// Code is an exit code of the engine.
type Code int

// The exit codes. The design names the failures of each code.
const (
	OK       Code = 0 // Done, or nothing to change.
	Other    Code = 1 // Any other failure, for example a file system error.
	Usage    Code = 2 // A usage error, or a config that breaks a rule.
	Missing  Code = 3 // A name that the clone does not have.
	Terminal Code = 4 // The Mac init stage changed: open a new terminal.
	Network  Code = 5 // A failure of the network side or of the clone.
)

// Error is a failure with its exit code.
type Error struct {
	Code Code
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	switch {
	case e.Err == nil:
		return e.Msg
	case e.Msg == "":
		return e.Err.Error()
	default:
		return e.Msg + ": " + e.Err.Error()
	}
}

func (e *Error) Unwrap() error { return e.Err }

// New gives an error with <code> and a formatted message.
func New(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...)}
}

// Wrap gives an error with <code> that wraps <err>. An empty <format> keeps
// the message of <err> alone.
func Wrap(code Code, err error, format string, args ...any) *Error {
	return &Error{Code: code, Msg: fmt.Sprintf(format, args...), Err: err}
}

// CodeOf gives the exit code of <err>: 0 for nil, the code of the first
// *Error in the chain, else 1. For a joined error, it gives the highest code
// of its parts.
func CodeOf(err error) Code {
	if err == nil {
		return OK
	}
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		code := OK
		for _, e := range j.Unwrap() {
			code = max(code, CodeOf(e))
		}
		return code
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return Other
}

// Collector keeps the failures of a run. The highest code wins.
type Collector struct {
	code Code
	errs []error
}

// Add keeps <err>. A nil <err> changes nothing.
func (c *Collector) Add(err error) {
	if err == nil {
		return
	}
	c.errs = append(c.errs, err)
	c.code = max(c.code, CodeOf(err))
}

// Code gives the highest code of the kept failures, or 0.
func (c *Collector) Code() Code { return c.code }

// Err gives the kept failures as one error, or nil.
func (c *Collector) Err() error { return errors.Join(c.errs...) }
