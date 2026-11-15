package contract

import (
	"bytes"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
)

func TestWriteFieldOrder(t *testing.T) {
	var b bytes.Buffer
	code := Write(&b, Envelope{Command: "questions", Code: fail.OK, Data: []int{1}})
	if code != fail.OK {
		t.Errorf("code = %d, want 0", code)
	}
	want := `{"contract":1,"command":"questions","ok":true,"code":0,"data":[1],"errors":[],"warnings":[]}` + "\n"
	if b.String() != want {
		t.Errorf("got %q, want %q", b.String(), want)
	}
}

func TestWriteOkFollowsCode(t *testing.T) {
	var b bytes.Buffer
	code := Write(&b, Envelope{Command: "", Code: fail.Usage, Errors: []Item{{Code: fail.Usage, Message: "x is not a command"}}})
	if code != fail.Usage {
		t.Errorf("code = %d, want 2", code)
	}
	want := `{"contract":1,"command":"","ok":false,"code":2,"data":null,"errors":[{"code":2,"message":"x is not a command"}],"warnings":[]}` + "\n"
	if b.String() != want {
		t.Errorf("got %q, want %q", b.String(), want)
	}
}

func TestWritePath(t *testing.T) {
	var b bytes.Buffer
	Write(&b, Envelope{Command: "check", Code: fail.OK, Warnings: []Item{{Message: "w", Path: ".config/bootstrap/repo.toml"}}})
	if !strings.Contains(b.String(), `"warnings":[{"code":0,"message":"w","path":".config/bootstrap/repo.toml"}]`) {
		t.Errorf("got %q, want the path of the warning", b.String())
	}
}

func TestWriteNoHTMLEscape(t *testing.T) {
	var b bytes.Buffer
	Write(&b, Envelope{Command: "x", Data: "a <b> & c"})
	if !strings.Contains(b.String(), `"data":"a <b> & c"`) {
		t.Errorf("got %q, want the text as it is", b.String())
	}
}

func TestPorts(t *testing.T) {
	if len(Ports) != 14 {
		t.Errorf("len(Ports) = %d, want 14", len(Ports))
	}
	seen := map[string]bool{}
	for _, p := range Ports {
		if seen[p] {
			t.Errorf("port %s twice", p)
		}
		seen[p] = true
	}
	if len(Supported) != 1 || Supported[0] != Version {
		t.Errorf("Supported = %v, want [%d]", Supported, Version)
	}
}
