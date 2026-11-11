package source

import (
	"errors"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// Logger prints one message line. *report.Report has it. A Logger that also
// has Error(msg string) gets each message of LogAll as an error.
type Logger interface{ Log(msg string) }

// Problems joins the failures of one check. Each part is a *fail.Error whose
// message is one log line of the bash engine, and the code of the whole is the
// highest code of its parts. A nil list gives nil.
func Problems(list []*fail.Error) error {
	if len(list) == 0 {
		return nil
	}
	errs := make([]error, len(list))
	for i, e := range list {
		errs[i] = e
	}
	return errors.Join(errs...)
}

// Messages gives each message of <err> in order, as a flat list.
func Messages(err error) []*fail.Error {
	if err == nil {
		return nil
	}
	if j, ok := err.(interface{ Unwrap() []error }); ok {
		var out []*fail.Error
		for _, e := range j.Unwrap() {
			out = append(out, Messages(e)...)
		}
		return out
	}
	var e *fail.Error
	if errors.As(err, &e) {
		return []*fail.Error{e}
	}
	return []*fail.Error{fail.New(fail.Other, "%s", err.Error())}
}

// LogAll prints each message of <err> and gives its exit code.
func LogAll(l Logger, err error) fail.Code {
	for _, e := range Messages(err) {
		if e.Msg != "" {
			if el, ok := l.(interface{ Error(string) }); ok {
				el.Error(e.Msg)
			} else {
				l.Log(e.Msg)
			}
		}
	}
	return fail.CodeOf(err)
}
