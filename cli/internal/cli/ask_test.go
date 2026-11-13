package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/field"
	"github.com/virajp/bootstrap/cli/internal/source"
)

const askTOML = `
[questions.a]
question = "A?"
group = "g"
example = "pa"
note = "a note"
[questions.b]
question = "B?"
group = "g"
[questions.c]
question = "C?"
optional = true
[questions.d]
question = "D?"
example = "pd"
`

// askIndex gives the Mac index of the tool t with the questions of askTOML,
// and of the tool u that needs t.
func askIndex(t *testing.T, tools map[string]string) *source.Index {
	t.Helper()
	root := t.TempDir()
	for name, text := range tools {
		p := filepath.Join(root, "macos", "tools", name, "tools.toml")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	x, err := source.Source{Root: root, Side: source.Macos}.Open()
	if err != nil {
		t.Fatal(err)
	}
	return x
}

func userOf(kv map[string]any) *source.Obj {
	tt := source.NewObj()
	for k, v := range kv {
		tt.Set(k, v)
	}
	o := source.NewObj()
	o.Set("t", tt)
	return o
}

// asker is a prompt that gives the queued texts and records each field.
type asker struct {
	texts []string
	specs []field.Spec
}

func (k *asker) prompt(s field.Spec) (string, error) {
	k.specs = append(k.specs, s)
	if len(k.texts) == 0 {
		return "", errors.New("interrupted")
	}
	v := k.texts[0]
	k.texts = k.texts[1:]
	return v, nil
}

type askRun struct {
	terminal bool
	user     map[string]any
	answers  []string
	texts    []string
}

func runAsk(t *testing.T, r askRun) ([][3]string, fail.Code, *asker, string, string) {
	t.Helper()
	sel := &selected{idx: askIndex(t, map[string]string{"t": askTOML}), tools: []string{"t"}, user: userOf(r.user)}
	var stdout, stderr bytes.Buffer
	k := &asker{texts: r.texts}
	a := &app{stdout: &stdout, stderr: &stderr, getenv: func(string) string { return "" }, terminal: r.terminal, prompt: k.prompt}
	out, code := a.ask(true, sel, opts{answers: r.answers}, nil)
	return out, code, k, stdout.String(), stderr.String()
}

func TestAskFlagsFirst(t *testing.T) {
	out, code, k, _, _ := runAsk(t, askRun{terminal: true, answers: []string{"t.a=x", "t.c=", "t.d=y"}, texts: []string{"z"}})
	want := [][3]string{{"t", "a", "x"}, {"t", "b", "z"}, {"t", "c", ""}, {"t", "d", "y"}}
	if code != fail.OK || !reflect.DeepEqual(out, want) {
		t.Fatalf("code %d out %v", code, out)
	}
	if len(k.specs) != 1 || k.specs[0].Question != "B? (t.b)" || k.specs[0].Necessary {
		t.Errorf("specs = %+v", k.specs)
	}
}

func TestAskField(t *testing.T) {
	_, code, k, _, _ := runAsk(t, askRun{terminal: true, user: map[string]any{"b": "x"}, texts: []string{"1", "", "4"}})
	if code != fail.OK {
		t.Fatalf("code %d", code)
	}
	want := []field.Spec{
		{Question: "A? (t.a)", Note: "a note", Proposal: "pa"},
		{Question: "C? (t.c)"},
		{Question: "D? (t.d)", Proposal: "pd", Necessary: true},
	}
	if !reflect.DeepEqual(k.specs, want) {
		t.Errorf("specs = %+v, want %+v", k.specs, want)
	}
}

func TestAskEmptyGroupAgain(t *testing.T) {
	// The files give the group only empty answers: init asks both again, then
	// the whole group again while it is empty.
	out, code, k, stdout, _ := runAsk(t, askRun{terminal: true, user: map[string]any{"a": "", "b": "", "c": "", "d": "v"}, answers: []string{"t.b="},
		texts: []string{"", "", "", "", "e"}})
	if code != fail.OK {
		t.Fatalf("code %d", code)
	}
	if want := [][3]string{{"t", "a", ""}, {"t", "b", "e"}}; !reflect.DeepEqual(out, want) {
		t.Errorf("out = %v, want %v", out, want)
	}
	if len(k.specs) != 5 || strings.Count(stdout, "Give at least one answer: a, b\n") != 2 {
		t.Errorf("%d fields, stdout %q", len(k.specs), stdout)
	}
}

func TestAskInterrupt(t *testing.T) {
	out, code, _, _, stderr := runAsk(t, askRun{terminal: true, user: map[string]any{"a": "x", "c": ""}})
	if code != fail.Usage || out != nil || !strings.Contains(stderr, "the question t.b has no answer") {
		t.Errorf("code %d out %v stderr %q", code, out, stderr)
	}
}

func TestAskNoTerminal(t *testing.T) {
	_, code, _, _, stderr := runAsk(t, askRun{})
	for _, w := range []string{
		"the group t.g has no answer", "give at least one with --answer <tool>.<key>=<value>: a, b",
		"the question t.c has no answer", "give it with --answer t.c=<value>, or leave it empty with --answer t.c=",
		"the question t.d has no answer", "give it with --answer t.d=<value>",
	} {
		// The fix line starts with a capital letter.
		if !strings.Contains(strings.ToLower(stderr), strings.ToLower(w)) {
			t.Errorf("stderr lacks %q:\n%s", w, stderr)
		}
	}
	if strings.Contains(stderr, "t.a has no answer") || strings.Contains(stderr, "t.d=<value>, or") || code != fail.Usage {
		t.Errorf("code %d stderr:\n%s", code, stderr)
	}
	// A values file that init wrote for the group passes.
	_, code, _, _, stderr = runAsk(t, askRun{user: map[string]any{"a": "", "b": "x", "c": "", "d": "v"}})
	if code != fail.OK {
		t.Errorf("code %d stderr %q", code, stderr)
	}
}

func TestAskEmptyFlag(t *testing.T) {
	_, code, _, _, stderr := runAsk(t, askRun{answers: []string{"t.d="}})
	if code != fail.Usage || !strings.Contains(stderr, "--answer t.d=: an empty answer is only for an optional or a group question") {
		t.Errorf("code %d stderr %q", code, stderr)
	}
	// The last --answer of a question wins.
	out, code, _, _, _ := runAsk(t, askRun{user: map[string]any{"b": "", "c": ""}, answers: []string{"t.a=1", "t.a=2", "t.d=v"}})
	if code != fail.OK || out[0] != [3]string{"t", "a", "2"} {
		t.Errorf("code %d out %v", code, out)
	}
}

func TestRefuse(t *testing.T) {
	old := source.GOOS
	defer func() { source.GOOS = old }()
	source.GOOS = "linux"
	x := askIndex(t, map[string]string{
		"base": "selectable = false\n",
		"u":    "depends = [\"base\"]\n",
		"v":    "depends = [\"base\"]\n",
		"w":    "",
	})
	// A Mac tool has no hosts: a host refusal needs a repo tool.
	x.Manifests["w"].Raw.Set("hosts", []any{"macos"})
	x.Manifests["w"].Hosts = []string{"macos"}
	var stderr bytes.Buffer
	a := &app{stderr: &stderr}
	if code := a.refuse(x, []string{"u", "base", "w"}); code != fail.Usage {
		t.Errorf("code %d", code)
	}
	for _, w := range []string{"base is not selectable. Select u or v.", "w needs the host macos - this host is linux"} {
		if !strings.Contains(stderr.String(), w) {
			t.Errorf("stderr lacks %q: %s", w, stderr.String())
		}
	}
	if code := a.refuse(x, []string{"u", "v"}); code != fail.OK {
		t.Errorf("code %d", code)
	}
	if proposable(x.Manifests["base"]) || proposable(x.Manifests["w"]) || !proposable(x.Manifests["u"]) {
		t.Error("proposable")
	}
}

func runQuestions(t *testing.T, user map[string]any, json bool) string {
	t.Helper()
	sel := &selected{idx: askIndex(t, map[string]string{"t": askTOML}), tools: []string{"t"}, user: userOf(user)}
	var stdout, stderr bytes.Buffer
	a := &app{stdout: &stdout, stderr: &stderr, getenv: func(string) string { return "" }}
	if code := a.printQuestions(true, sel, opts{json: json}); code != fail.OK {
		t.Fatalf("code %d stderr %q", code, stderr.String())
	}
	return stdout.String()
}

func TestQuestionsJSON(t *testing.T) {
	var rows []map[string]any
	if err := json.Unmarshal([]byte(runQuestions(t, map[string]any{}, true)), &rows); err != nil {
		t.Fatal(err)
	}
	want := []map[string]any{
		{"key": "a", "group": "g", "note": "a note", "optional": false},
		{"key": "b", "group": "g", "note": "", "optional": false},
		{"key": "c", "group": "", "note": "", "optional": true},
		{"key": "d", "group": "", "note": "", "optional": false},
	}
	if len(rows) != len(want) {
		t.Fatalf("rows = %v", rows)
	}
	for i, w := range want {
		for k, v := range w {
			if rows[i][k] != v {
				t.Errorf("row %d: %s = %v, want %v", i, k, rows[i][k], v)
			}
		}
	}
}

func TestQuestionsText(t *testing.T) {
	// An answer "" shows the ok glyph alone (D16); an optional question and
	// a group question of a group with an answer are not without an answer.
	out := runQuestions(t, map[string]any{"b": "x", "d": ""}, false)
	for _, w := range []string{
		"  a   A?\n      [INFO] no answer (group g has an answer)\n",
		"  b   B?\n      [OK] x\n",
		"  c   C?\n      [INFO] no answer (optional)\n",
		"  d   D?\n      [OK]\n",
		"[INFO] 4 questions · 2 answered\n",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
	if strings.Contains(out, "without an answer") {
		t.Errorf("output:\n%s", out)
	}
	// An all-empty group: each question of it is without an answer.
	out = runQuestions(t, map[string]any{"a": "", "c": "", "d": "v"}, false)
	for _, w := range []string{
		"  a   A?\n      [ERROR] no answer\n",
		"  b   B?\n      [ERROR] no answer\n",
		"[INFO] 4 questions · 2 answered · 2 without an answer\n",
	} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q:\n%s", w, out)
		}
	}
}

func TestWarnNotSelectable(t *testing.T) {
	root := t.TempDir()
	for name, text := range map[string]string{"base": "selectable = false\n", "u": "depends = [\"base\"]\n"} {
		p := filepath.Join(root, "macos", "tools", name, "tools.toml")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	file := filepath.Join(root, "macos.toml")
	if err := os.WriteFile(file, []byte("[tools.base]\n[tools.u]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	a := &app{stderr: &stderr, pwd: root}
	a.warnNotSelectable(true, root, file)
	want := "[WARN] base is not selectable - remove [tools.base] from macos.toml; a tool that needs it selects it\n"
	if stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
}
