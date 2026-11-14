package core

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
)

// The keys that each tool needs, and the categories file of the clone.
const (
	catalogKeys = "summary = \"A tool\"\ncategory = \"core\"\n"
	catalogFile = "[category.core]\ntitle = \"Core\"\n"
)

// fixture is a clone with the tools <tools> and the packs file <packs> on
// the side <side>, and a folder with the values file <values>.
func fixture(t *testing.T, side source.Side, tools map[string]string, packs, values string) Side {
	t.Helper()
	clone, root := t.TempDir(), t.TempDir()
	write := func(p, text string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, text := range tools {
		write(filepath.Join(clone, string(side), "tools", name, "tools.toml"), catalogKeys+text)
	}
	write(filepath.Join(clone, string(side), "categories.toml"), catalogFile)
	if packs != "" {
		write(filepath.Join(clone, string(side), "packs.toml"), packs)
	}
	sd := Side{Mac: side == source.Macos, Root: root, Clone: clone, Pwd: root}
	if sd.Mac {
		sd.Values = filepath.Join(root, "macos.toml")
	} else {
		sd.Values = filepath.Join(root, ".config", "bootstrap.toml")
	}
	if values != "" {
		write(sd.Values, values)
	}
	return sd
}

// fakeSync replaces the sync of Plan and Apply: it gives the file <file>
// and <code>, and records the values file that each run sees.
type fakeSync struct {
	code  fail.Code
	file  string
	runs  int
	dry   []bool
	saw   []string
	value string // The values file of the side.
}

func (f *fakeSync) install(t *testing.T, sd Side) {
	t.Helper()
	old := frontEnd
	t.Cleanup(func() { frontEnd = old })
	f.value = sd.Values
	frontEnd = func(bool) FrontEnd {
		return func(args []string, stdout, stderr io.Writer) ([]target.Planned, fail.Code) {
			f.runs++
			dry := len(args) > 0 && args[0] == "--dry-run"
			f.dry = append(f.dry, dry)
			file := f.value
			if dry && sd.Mac {
				file = os.Getenv("MACOS_SETUP_REPO_PATH") + "/macos.toml"
			}
			b, _ := os.ReadFile(file)
			f.saw = append(f.saw, string(b))
			if !dry && f.code != fail.OK {
				// The sync fails after it wrote the values file.
				_ = os.WriteFile(f.value, []byte("broken\n"), 0o644)
				return nil, f.code
			}
			return []target.Planned{{Path: f.file, Kind: 1}}, fail.OK
		}
	}
}

const macTools = `
[pack.base]
a = { required = true }

[pack.dev]
b = {}
c = {}
"pack:extra" = {}

[pack.extra]
d = {}
h = {}
`

func macFixture(t *testing.T, values string) Side {
	return fixture(t, source.Macos, map[string]string{
		"a": "",
		"b": "depends = [ \"e\" ]\n",
		"c": "",
		"d": "",
		"e": "",
		"h": "selectable = false\n",
		"q": "[questions.k]\nquestion = \"K?\"\n",
	}, macTools, values)
}

func problems(err error) []string {
	var out []string
	var walk func(error)
	walk = func(e error) {
		if j, ok := e.(interface{ Unwrap() []error }); ok {
			for _, x := range j.Unwrap() {
				walk(x)
			}
			return
		}
		var p *Problem
		if errors.As(e, &p) {
			out = append(out, p.Error())
			return
		}
		out = append(out, e.Error())
	}
	if err != nil {
		walk(err)
	}
	return out
}

func TestPlanAddTool(t *testing.T) {
	sd := macFixture(t, "format = 2\n[tools.a]\n")
	f := &fakeSync{file: "/x/y"}
	f.install(t, sd)
	p, err := Plan(sd, List{Items: []Item{{Kind: AddTool, Names: []string{"b"}}}}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if !strings.Contains(string(p.Next), "[tools.b]") || !strings.Contains(string(p.Next), "[tools.e]") {
		t.Errorf("next:\n%s", p.Next)
	}
	if want := []string{"select e", "select b"}; !reflect.DeepEqual(p.Notes, want) {
		t.Errorf("notes = %v, want %v", p.Notes, want)
	}
	if !p.AskOnly["b"] || !p.AskOnly["e"] || len(p.AskOnly) != 2 {
		t.Errorf("askOnly = %v", p.AskOnly)
	}
	if len(p.Changes) != 1 || p.Changes[0].Path != sd.Values || len(p.Files) != 1 || p.Files[0].Path != "/x/y" {
		t.Errorf("changes %v files %v", p.Changes, p.Files)
	}
	// The dry run saw the new values file; the values file did not change.
	if f.runs != 1 || !f.dry[0] || !strings.Contains(f.saw[0], "[tools.b]") {
		t.Errorf("runs %d dry %v saw %q", f.runs, f.dry, f.saw)
	}
	if b, _ := os.ReadFile(sd.Values); string(b) != "format = 2\n[tools.a]\n" {
		t.Errorf("values file changed: %q", b)
	}
	// Again: b is already selected.
	sd2 := macFixture(t, "format = 2\n[tools.b]\n[tools.e]\n")
	f.install(t, sd2)
	p, err = Plan(sd2, List{Items: []Item{{Kind: AddTool, Names: []string{"b"}}}}, Options{})
	if err != nil || len(p.Changes) != 0 || !reflect.DeepEqual(p.Reasons, []string{"b is already selected"}) {
		t.Errorf("err %v changes %v reasons %v", err, p.Changes, p.Reasons)
	}
}

func TestPlanProblems(t *testing.T) {
	sd := macFixture(t, "format = 2\n[tools.b]\n[tools.e]\n")
	(&fakeSync{}).install(t, sd)
	cases := []struct {
		item Item
		code fail.Code
		want []string
	}{
		{Item{Kind: AddTool, Names: []string{"zz"}}, fail.Missing, []string{"zz is not a bootstrap tool"}},
		{Item{Kind: AddTool, Names: []string{"h"}}, fail.Usage, []string{"h is not selectable."}},
		{Item{Kind: RemoveTool, Names: []string{"a"}}, fail.Usage, []string{"a is required by the base pack, so mac remove cannot remove it"}},
		{Item{Kind: RemoveTool, Names: []string{"e"}}, fail.Usage, []string{"e is needed by b - remove b first"}},
		{Item{Kind: RemoveTool, Names: []string{"c"}}, fail.Usage, []string{"c is not selected"}},
		{Item{Kind: AddPack, Names: []string{"nope"}}, fail.Missing, []string{"nope is not a bootstrap pack"}},
		{Item{Kind: AddPack, Names: []string{"dev", "extra"}, Exclude: []string{"c"}}, fail.Usage, []string{"--exclude needs exactly one pack name - run bootstrap mac add pack once for each pack"}},
		{Item{Kind: AddPack, Names: []string{"dev"}, Exclude: []string{"x"}}, fail.Usage, []string{"x is not a member of the pack dev"}},
		{Item{Kind: AddPack, Names: []string{"base"}, Exclude: []string{"a"}}, fail.Usage, []string{"a is required in the pack base, so mac add pack cannot exclude it"}},
	}
	for _, c := range cases {
		_, err := Plan(sd, List{Items: []Item{c.item}}, Options{})
		if got := problems(err); fail.CodeOf(err) != c.code || !reflect.DeepEqual(got, c.want) {
			t.Errorf("%+v: code %d problems %q, want %d %q", c.item, fail.CodeOf(err), got, c.code, c.want)
		}
	}
}

func TestPlanRemoveTool(t *testing.T) {
	sd := macFixture(t, "format = 2\n[tools.a]\n[tools.b]\n[tools.e]\n")
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Items: []Item{{Kind: RemoveTool, Names: []string{"b"}}}}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if strings.Contains(string(p.Next), "tools.b") || strings.Contains(string(p.Next), "tools.e") {
		t.Errorf("next:\n%s", p.Next)
	}
	if want := []string{"remove b", "remove e (a dependency that no tool needs)"}; !reflect.DeepEqual(p.Notes, want) {
		t.Errorf("notes = %v", p.Notes)
	}
	if len(p.AskOnly) != 0 || p.AskOnly == nil {
		t.Errorf("askOnly = %v", p.AskOnly)
	}
}

func TestPlanAddPack(t *testing.T) {
	sd := macFixture(t, "format = 2\n[tools.a]\n[tools.c]\n")
	(&fakeSync{}).install(t, sd)
	// The nested pack extra adds d; h is not selectable; c is selected.
	p, err := Plan(sd, List{Items: []Item{{Kind: AddPack, Names: []string{"dev"}}}}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if want := []string{"select e", "select b", "select d"}; !reflect.DeepEqual(p.Notes, want) {
		t.Errorf("notes = %v, want %v", p.Notes, want)
	}
	if want := []Skip{{"h", "selectable"}}; !reflect.DeepEqual(p.Skipped, want) {
		t.Errorf("skipped = %v", p.Skipped)
	}
	// --exclude of a nested pack and of a member.
	p, err = Plan(sd, List{Items: []Item{{Kind: AddPack, Names: []string{"dev"}, Exclude: []string{"pack:extra"}}}}, Options{})
	if err != nil || !reflect.DeepEqual(p.Notes, []string{"select e", "select b"}) {
		t.Errorf("err %v notes %v", problems(err), p.Notes)
	}
	// An excluded member that another member needs.
	_, err = Plan(sd, List{Items: []Item{{Kind: AddPack, Names: []string{"dev"}, Exclude: []string{"pack:extra", "b"}}}}, Options{})
	if err != nil {
		t.Errorf("b needs nothing excluded: %v", problems(err))
	}
	sd2 := fixture(t, source.Macos, map[string]string{"a": "", "b": "depends = [ \"c\" ]\n", "c": ""},
		"[pack.base]\na = { required = true }\n[pack.dev]\nb = {}\nc = {}\n", "format = 2\n")
	(&fakeSync{}).install(t, sd2)
	_, err = Plan(sd2, List{Items: []Item{{Kind: AddPack, Names: []string{"dev"}, Exclude: []string{"c"}}}}, Options{})
	if got := problems(err); fail.CodeOf(err) != fail.Usage || !reflect.DeepEqual(got, []string{"c is needed by b - exclude b too, or keep c"}) {
		t.Errorf("problems %q", got)
	}
	// A pack with no member that is not selected.
	sd3 := macFixture(t, "format = 2\n[tools.d]\n")
	(&fakeSync{}).install(t, sd3)
	p, err = Plan(sd3, List{Items: []Item{{Kind: AddPack, Names: []string{"extra"}}}}, Options{})
	if err != nil || len(p.Changes) != 0 || !reflect.DeepEqual(p.Reasons, []string{"extra is already selected"}) {
		t.Errorf("err %v changes %v reasons %v", problems(err), p.Changes, p.Reasons)
	}
}

func TestPlanQuestions(t *testing.T) {
	sd := macFixture(t, "format = 2\n[tools.a]\n")
	f := &fakeSync{}
	f.install(t, sd)
	l := List{Items: []Item{{Kind: AddTool, Names: []string{"q"}}}}
	p, err := Plan(sd, l, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if len(p.Pending) != 1 || p.Pending[0].Name() != "q.k" || p.Files != nil || f.runs != 0 {
		t.Errorf("pending %v files %v runs %d", p.Pending, p.Files, f.runs)
	}
	// A flag that names no question, and an empty answer of a question
	// that needs one.
	l.Answers = []Answer{FlagAnswer("q.x=1"), FlagAnswer("q.k=")}
	_, err = Plan(sd, l, Options{})
	want := []string{
		"--answer q.x=1: no question q.x of the selected tools - the form is <tool>.<key>=<value>",
		"--answer q.k=: an empty answer is only for an optional or a group question - give it with --answer q.k=<value>",
	}
	if got := problems(err); !reflect.DeepEqual(got, want) {
		t.Errorf("problems %q", got)
	}
	l.Answers = []Answer{{Name: "q.k", Value: "v"}}
	p, err = Plan(sd, l, Options{})
	if err != nil || len(p.Pending) != 0 || len(p.Files) != 1 || !strings.Contains(string(p.Next), "k = \"v\"") {
		t.Errorf("err %v pending %v files %v next:\n%s", problems(err), p.Pending, p.Files, p.Next)
	}
	if want := []string{"select q", "answer q.k"}; !reflect.DeepEqual(p.Notes, want) {
		t.Errorf("notes = %v", p.Notes)
	}
}

// TestPlanRequiredTables: a command that writes the values file also adds
// the table of each required member of base that has none, so the sync of
// Apply writes no table and the values file is written one time (D21 d). A
// command that writes nothing leaves it to the sync (D14 d).
func TestPlanRequiredTables(t *testing.T) {
	sd := macFixture(t, "format = 2\n")
	f := &fakeSync{file: "/x/y"}
	f.install(t, sd)
	p, err := Plan(sd, List{Items: []Item{{Kind: AddTool, Names: []string{"c"}}}}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	if want := []string{"select c", "add [tools.a] (required by the base pack)"}; !reflect.DeepEqual(p.Notes, want) {
		t.Errorf("notes = %v, want %v", p.Notes, want)
	}
	if len(p.Changes) != 1 || !strings.Contains(string(p.Changes[0].Content), "[tools.a]") {
		t.Errorf("changes %v", p.Changes)
	}
	// The dry run of the sync saw the table, so it has none to add.
	if f.runs != 1 || !strings.Contains(f.saw[0], "[tools.a]") {
		t.Errorf("runs %d saw %q", f.runs, f.saw)
	}
	// An answer that gave the table keeps its keys.
	sd2 := fixture(t, source.Macos, map[string]string{"a": "[questions.k]\nquestion = \"K?\"\n", "c": ""},
		"[pack.base]\na = { required = true }\n[pack.dev]\nc = {}\n", "format = 2\n")
	(&fakeSync{}).install(t, sd2)
	p, err = Plan(sd2, List{Items: []Item{{Kind: AddTool, Names: []string{"c"}}}, Answers: []Answer{{Name: "a.k", Value: "v"}}}, Options{})
	if err != nil || !strings.Contains(string(p.Next), "k = \"v\"") {
		t.Errorf("err %v next:\n%s", problems(err), p.Next)
	}
	// Nothing to add: no write, the table stays for the sync.
	sd3 := macFixture(t, "format = 2\n[tools.c]\n")
	(&fakeSync{}).install(t, sd3)
	p, err = Plan(sd3, List{Items: []Item{{Kind: AddTool, Names: []string{"c"}}}}, Options{})
	if err != nil || len(p.Changes) != 0 || strings.Contains(string(p.Next), "[tools.a]") {
		t.Errorf("err %v changes %v next:\n%s", problems(err), p.Changes, p.Next)
	}
}

func TestPlanSyncFails(t *testing.T) {
	sd := macFixture(t, "format = 2\n")
	old := frontEnd
	t.Cleanup(func() { frontEnd = old })
	frontEnd = func(bool) FrontEnd {
		return func(_ []string, stdout, _ io.Writer) ([]target.Planned, fail.Code) {
			_, _ = io.WriteString(stdout, "the reason\n")
			return nil, fail.Missing
		}
	}
	var out strings.Builder
	_, err := Plan(sd, List{Items: []Item{{Kind: AddTool, Names: []string{"c"}}}}, Options{Stdout: &out})
	var p *Problem
	if !errors.As(err, &p) || p.Code != fail.Missing || p.Msg != "" || out.String() != "the reason\n" {
		t.Errorf("err %v output %q", err, out.String())
	}
}

func TestApply(t *testing.T) {
	const cur = "format = 2\n"
	sd := macFixture(t, cur)
	f := &fakeSync{}
	f.install(t, sd)
	p, err := Plan(sd, List{Items: []Item{{Kind: AddTool, Names: []string{"c"}}}}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	var steps []Step
	code, putBack, err := Apply(p, Options{Step: func(s Step) { steps = append(steps, s) }})
	if code != fail.OK || putBack || err != nil {
		t.Fatalf("code %d put back %v err %v", code, putBack, err)
	}
	if b, _ := os.ReadFile(sd.Values); string(b) != string(p.Next) || f.saw[len(f.saw)-1] != string(p.Next) {
		t.Errorf("values %q, sync saw %q", b, f.saw)
	}
	if len(steps) != 1 || steps[0].Kind != Wrote || steps[0].Path != sd.Values {
		t.Errorf("steps %v", steps)
	}

	// A sync that fails: the values file is back as it was before Apply.
	sd = macFixture(t, cur)
	f = &fakeSync{code: fail.Network}
	f.install(t, sd)
	if p, err = Plan(sd, List{Items: []Item{{Kind: AddTool, Names: []string{"c"}}}}, Options{}); err != nil {
		t.Fatal(problems(err))
	}
	code, putBack, err = Apply(p, Options{})
	if code != fail.Network || !putBack || err != nil {
		t.Fatalf("code %d put back %v err %v", code, putBack, err)
	}
	if b, _ := os.ReadFile(sd.Values); string(b) != cur {
		t.Errorf("values %q, want %q", b, cur)
	}
	if !strings.Contains(f.saw[len(f.saw)-1], "[tools.c]") {
		t.Errorf("the sync did not see the new values file: %q", f.saw)
	}

	// A values file that init creates goes when the sync fails.
	sd = macFixture(t, "")
	f.install(t, sd)
	if p, err = Plan(sd, List{Init: true}, Options{}); err != nil {
		t.Fatal(problems(err))
	}
	if code, putBack, err = Apply(p, Options{}); code != fail.Network || !putBack || err != nil {
		t.Fatalf("code %d put back %v err %v", code, putBack, err)
	}
	if _, err := os.Stat(sd.Values); !os.IsNotExist(err) {
		t.Errorf("values file stays: %v", err)
	}
}

func TestDetect(t *testing.T) {
	const repoKeys = "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"
	sd := fixture(t, source.Repo, map[string]string{
		"a": repoKeys,
		"b": repoKeys,
		"c": repoKeys,
		"d": repoKeys,
	}, "[pack.base]\na = { required = true }\n", "format = 2\nrepo_name = \"x\"\n")
	scripts := map[string]string{
		"a": `{"result":"match","evidence":["f"],"message":"m"}`,
		"b": `{"result":"match","evidence":["f"],"message":"m"}`,
		"c": `{"result":"match","evidence":["g","h"],"message":"n"}`,
		"d": `{"result":"none","evidence":[],"message":"no"}`,
	}
	for tool, out := range scripts {
		p := filepath.Join(sd.Clone, "repo", "tools", tool, "scripts", "detect")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("#!/bin/bash\necho '"+out+"'\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	d, err := Detect(sd)
	if err != nil {
		t.Fatal(problems(err))
	}
	if !reflect.DeepEqual(d.Tools, []string{"a", "b", "c", "d"}) || !reflect.DeepEqual(d.Offers, []string{"c"}) {
		t.Errorf("tools %v offers %v", d.Tools, d.Offers)
	}
	want := []Proposal{
		{Tool: "b", Shared: "f", Of: "a", Found: d.Found["b"]},
		{Tool: "c", Offered: true, Found: d.Found["c"]},
	}
	if !reflect.DeepEqual(d.Proposals, want) || d.Mismatch != nil {
		t.Errorf("proposals %+v mismatch %v", d.Proposals, d.Mismatch)
	}
	if items := d.Items([]string{"c"}, nil); len(items) != 1 || items[0].Kind != AddTool {
		t.Errorf("items %v", items)
	}
}

func TestPending(t *testing.T) {
	x := fixture(t, source.Macos, map[string]string{"t": `
[questions.a]
question = "A?"
group = "g"
[questions.b]
question = "B?"
group = "g"
[questions.d]
question = "D?"
`}, "", "format = 2\n[tools.t]\nd = \"v\"\n")
	sel, err := SelectionOf(x, x.Values, x.Values)
	if err != nil {
		t.Fatal(err)
	}
	names := func(given map[string]string, only map[string]bool) string {
		var out []string
		for _, q := range Pending(sel, given, only) {
			out = append(out, q.Name())
		}
		return strings.Join(out, " ")
	}
	if got := names(nil, nil); got != "t.a t.b" {
		t.Errorf("pending %q", got)
	}
	if got := names(map[string]string{"t.b": "x"}, nil); got != "t.a" {
		t.Errorf("pending %q", got)
	}
	if got := names(nil, map[string]bool{}); got != "" {
		t.Errorf("pending %q", got)
	}
}

// TestPlanInit: init on a new values file writes format = 2 and a table for
// each member of base (add pack base; D6, D16 c). init on a file with format
// = 2 adds no member of base that is not required again.
func TestPlanInit(t *testing.T) {
	tools := map[string]string{"a": "", "b": "depends = [ \"e\" ]\n", "e": "", "h": "selectable = false\n"}
	const packs = "[pack.base]\na = { required = true }\nb = {}\nh = {}\n"
	sd := fixture(t, source.Macos, tools, packs, "")
	(&fakeSync{}).install(t, sd)
	p, err := Plan(sd, List{Init: true}, Options{})
	if err != nil {
		t.Fatal(problems(err))
	}
	// a is selected as a required member already: the table of a required
	// member comes from requiredTables.
	if want := []string{"create the values file", "select e", "select b", "add [tools.a] (required by the base pack)"}; !reflect.DeepEqual(p.Notes, want) {
		t.Errorf("notes = %v, want %v", p.Notes, want)
	}
	next := string(p.Next)
	if !strings.HasPrefix(next, MacHeader+"\n") || !strings.Contains(next, "format = 2") {
		t.Errorf("next:\n%s", next)
	}
	for _, tbl := range []string{"[tools.a]", "[tools.b]", "[tools.e]"} {
		if !strings.Contains(next, tbl) {
			t.Errorf("next lacks %s:\n%s", tbl, next)
		}
	}
	if strings.Contains(next, "[tools.h]") || !reflect.DeepEqual(p.Skipped, []Skip{{"h", "selectable"}}) {
		t.Errorf("h: skipped %v next:\n%s", p.Skipped, next)
	}
	if p.AskOnly != nil {
		t.Errorf("askOnly = %v", p.AskOnly)
	}

	// The repo side: format = 2 next to the keys of a new file.
	rd := fixture(t, source.Repo, map[string]string{"a": "default_envs = [ \"dev\" ]\ndepends = []\nroot_only = false\n"},
		"[pack.base]\na = { required = true }\n", "")
	(&fakeSync{}).install(t, rd)
	if p, err = Plan(rd, List{Init: true}, Options{}); err != nil {
		t.Fatal(problems(err))
	}
	if next := string(p.Next); !strings.Contains(next, "format = 2") || !strings.Contains(next, "repo_name") || !strings.Contains(next, "[tools.a]") {
		t.Errorf("repo next:\n%s", next)
	}

	// A file with format = 2 without the member b: init does not add it.
	sd2 := fixture(t, source.Macos, tools, packs, "format = 2\n\n[tools.a]\n")
	(&fakeSync{}).install(t, sd2)
	if p, err = Plan(sd2, List{Init: true}, Options{}); err != nil {
		t.Fatal(problems(err))
	}
	if len(p.Changes) != 0 || strings.Contains(string(p.Next), "[tools.b]") {
		t.Errorf("changes %v next:\n%s", p.Changes, p.Next)
	}
}
