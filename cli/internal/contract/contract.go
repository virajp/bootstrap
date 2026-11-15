// Package contract is the JSON contract of the command line: the envelope
// that each command prints on stdout with --json, its version, and the names
// of the task ports.
//
// Contract: docs/design/cli.md.
package contract

import (
	"encoding/json"
	"io"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

// Version is the version of the contract. A new field, command or enum value
// keeps it; a removed or renamed field, or a changed type or meaning, raises
// it.
const Version = 1

// Supported lists the versions of the contract that this engine prints.
var Supported = []int{Version}

// Ports are the names of the task ports. <branch> and <id> stand for one
// segment without a colon.
var Ports = []string{
	"code:format",
	"code:lint",
	"code:sec",
	"code:precommit",
	"git:merge:<branch>",
	"git:worktree:create",
	"git:worktree:setup",
	"git:worktree:delete",
	"code:lsp:check",
	"setup",
	"test",
	"test:e2e",
	"test:golden",
	"p:<id>:deploy",
}

// Item is an error or a warning of the envelope. Code is the exit code of an
// error, and 0 for a warning. Path is the file, when one is known.
type Item struct {
	Code    fail.Code `json:"code"`
	Message string    `json:"message"`
	Path    string    `json:"path,omitempty"`
}

// Envelope is the one object that a command prints on stdout with --json.
// Command is the full command words, for example "config get", or "" when no
// command was parsed. Data is the result of the command, or nil when it
// failed before it had one.
type Envelope struct {
	Command  string
	Code     fail.Code
	Data     any
	Errors   []Item
	Warnings []Item
}

// wire is the envelope in the order of its JSON fields.
type wire struct {
	Contract int       `json:"contract"`
	Command  string    `json:"command"`
	OK       bool      `json:"ok"`
	Code     fail.Code `json:"code"`
	Data     any       `json:"data"`
	Errors   []Item    `json:"errors"`
	Warnings []Item    `json:"warnings"`
}

// Write prints <e> on <w> as one line of JSON, and gives its code: the exit
// code of the command. ok is true only when the code is 0, and an empty list
// is [], never null.
func Write(w io.Writer, e Envelope) fail.Code {
	out := wire{
		Contract: Version, Command: e.Command, OK: e.Code == fail.OK, Code: e.Code,
		Data: e.Data, Errors: e.Errors, Warnings: e.Warnings,
	}
	if out.Errors == nil {
		out.Errors = []Item{}
	}
	if out.Warnings == nil {
		out.Warnings = []Item{}
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		// Data that does not encode is a bug of the command; the envelope
		// still reaches the caller.
		out.Data, out.OK = nil, false
		if out.Code == fail.OK {
			out.Code = fail.Other
		}
		out.Errors = append(out.Errors, Item{Code: fail.Other, Message: "cannot write the data as JSON: " + err.Error()})
		_ = enc.Encode(out)
	}
	return out.Code
}
