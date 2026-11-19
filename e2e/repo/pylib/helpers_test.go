package pylib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// printCall is one call of a print function: the shell form of the fixed
// result, and the Python form.
type printCall struct{ name, sh, py string }

var printCalls = []printCall{
	{"print_header", `print_header "Running the checks ..."`, `helpers.print_header("Running the checks ...")`},
	{"print_header words", `print_header Running two words`, `helpers.print_header("Running", "two", "words")`},
	{"print_subheader", `print_subheader "Installing the hooks ..."`, `helpers.print_subheader("Installing the hooks ...")`},
	{"print_success empty", `print_success`, `helpers.print_success()`},
	{"print_success", `print_success "Done"`, `helpers.print_success("Done")`},
	{"print_success bold", `print_success --bold "Done"`, `helpers.print_success("Done", bold=True)`},
	{"print_success bold empty", `print_success --bold`, `helpers.print_success(bold=True)`},
	{"print_warn empty", `print_warn`, `helpers.print_warn()`},
	{"print_warn", `print_warn "Careful" now`, `helpers.print_warn("Careful", "now")`},
	{"print_warn bold", `print_warn --bold "Careful"`, `helpers.print_warn("Careful", bold=True)`},
	{"print_error empty", `print_error`, `helpers.print_error()`},
	{"print_error", `print_error "Failed"`, `helpers.print_error("Failed")`},
	{"print_error bold", `print_error --bold "Failed"`, `helpers.print_error("Failed", bold=True)`},
	{"print_info empty", `print_info`, `helpers.print_info()`},
	{"print_info", `print_info "Note"`, `helpers.print_info("Note")`},
	{"print_info bold", `print_info --bold "Note"`, `helpers.print_info("Note", bold=True)`},
	{"print_wait", `print_wait "Checking docker ... "; print_success`,
		`helpers.print_wait("Checking docker ... "); helpers.print_success()`},
	{"print_wait error", `print_wait "Checking ... "; print_error`, `helpers.print_wait("Checking ... "); helpers.print_error()`},
	{"print_header_wait", `print_header_wait "Header wait"; print_newline`,
		`helpers.print_header_wait("Header wait"); helpers.print_newline()`},
	{"print_yellow", `print_yellow "Yellow line"`, `helpers.print_yellow("Yellow line")`},
	{"print_yellow empty", `print_yellow ""`, `helpers.print_yellow("")`},
	{"print_newline", `print_newline`, `helpers.print_newline()`},
	{"line_sep1", `line_sep1`, `helpers.line_sep1()`},
	{"line_sep2", `line_sep2`, `helpers.line_sep2()`},
	{"escapes", `print_info 'a\tb\nc \\ d'`, `helpers.print_info(r"a\tb\nc \\ d")`},
	{"unicode escapes", `print_info 'caf\u00e9 \U0001F600'`, `helpers.print_info(r"caf\u00e9 \U0001F600")`},
	{"octal and hex escapes", `print_info '\0101 \x41 \x4a'`, `helpers.print_info(r"\0101 \x41 \x4a")`},
	{"stop escape", `print_info 'a\cb'; print_info after`, `helpers.print_info(r"a\cb"); helpers.print_info("after")`},
	{"percent", `print_info '100% %s'`, `helpers.print_info("100% %s")`},
	{"print_warn to stderr", `print_warn "to stderr" >&2`, `helpers.print_warn("to stderr", file=sys.stderr)`},
	{"print_success to stderr", `print_success --bold "to stderr" >&2`,
		`helpers.print_success("to stderr", bold=True, file=sys.stderr)`},
	{"print_header to stderr", `print_header "to stderr" >&2`, `helpers.print_header("to stderr", file=sys.stderr)`},
}

// colorModes are the environments of the print cases: color forced on, color
// off with NO_COLOR, and a pipe with no color variable.
var colorModes = []struct {
	name string
	kv   []string
}{
	{"color on", []string{"FORCE_COLOR=1", "TERM=xterm-256color"}},
	{"NO_COLOR", []string{"NO_COLOR=1", "TERM=xterm-256color"}},
	{"no variable", []string{"TERM=xterm-256color"}},
}

// TestPrintFunctions: each print function of helpers.py writes the bytes of
// the shell function, on the same stream (D9, D22, D32).
func TestPrintFunctions(t *testing.T) {
	var py strings.Builder
	for _, c := range printCalls {
		fmt.Fprintf(&py, "mark(%s)\n%s\n", pyQuote(c.name), c.py)
	}
	for _, m := range colorModes {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			dir, root := newCase(t)
			p := probe{dir: dir, env: env(dir, root, m.kv...)}
			gotPy := p.py(t, dir, py.String())
			wantOut := shellText(t, "print-"+slug(m.name)+".stdout")
			wantErr := shellText(t, "print-"+slug(m.name)+".stderr")
			if !harness.Check(t, m.name+": stdout of each print function", gotPy.out == wantOut,
				firstDiff(gotPy.out, wantOut)) {
				return
			}
			harness.Check(t, m.name+": stderr of each print function", gotPy.errOut == wantErr,
				firstDiff(gotPy.errOut, wantErr))
			harness.Check(t, m.name+": exit 0", gotPy.code == 0, gotPy.String())
		})
	}
}

// firstDiff names the first block of two outputs that differs. The blocks
// are separated by the mark lines.
func firstDiff(py, sh string) string {
	a, b := strings.Split(py, "@@ "), strings.Split(sh, "@@ ")
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y string
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			return fmt.Sprintf("python: %q\nshell:  %q", x, y)
		}
	}
	return ""
}

// TestPrintToStderr: a print function with file=sys.stderr writes to stderr
// only (D22).
func TestPrintToStderr(t *testing.T) {
	t.Parallel()
	dir, root := newCase(t)
	p := probe{dir: dir, env: env(dir, root, "NO_COLOR=1")}
	got := p.py(t, dir, `for f in (helpers.print_header, helpers.print_subheader, helpers.print_success, helpers.print_warn,
          helpers.print_info, helpers.print_wait, helpers.print_header_wait, helpers.print_yellow):
  f("x", file=sys.stderr)
helpers.print_newline(file=sys.stderr)
helpers.line_sep1(file=sys.stderr)
helpers.line_sep2(file=sys.stderr)
`)
	harness.Check(t, "file=sys.stderr: nothing on stdout", got.out == "" && got.code == 0, got.String())
	harness.Check(t, "file=sys.stderr: the lines on stderr", strings.Count(got.errOut, "x") >= 8, got.String())
	got = p.py(t, dir, "helpers.print_error('x')\n")
	harness.Check(t, "print_error: stderr by default", got.out == "" && got.errOut == "[ERROR] x\n", got.String())
}

// The stub engine of the bootstrap_get cases: it writes its arguments, its
// folder and MISE_PROJECT_ROOT to STUB_LOG, prints STUB_ERR on stderr and
// STUB_OUT on stdout, and exits with STUB_CODE.
const stubEngine = `#!/bin/bash
printf '%s|%s|%s|%s\n' "$*" "$PWD" "${MISE_PROJECT_ROOT:-}" "${BOOTSTRAP_VALUES-unset}" >> "$STUB_LOG"
[ -z "${STUB_ERR:-}" ] || printf '%s\n' "$STUB_ERR" >&2
[ -z "${STUB_OUT:-}" ] || printf '%s\n' "$STUB_OUT"
exit "${STUB_CODE:-0}"
`

// stubCase is a case folder with the stub engine in BOOTSTRAP_BIN.
func stubCase(t *testing.T, kv ...string) (string, string, probe) {
	t.Helper()
	dir, root := newCase(t)
	stub := filepath.Join(dir, "engine")
	writeExec(t, stub, stubEngine)
	e := env(dir, root, append([]string{"NO_COLOR=1", "BOOTSTRAP_BIN=" + stub, "STUB_LOG=" + filepath.Join(dir, "log")}, kv...)...)
	return dir, root, probe{dir: dir, env: e, merge: true}
}

const answerJSON = `{"key":"repo.core_branches","value":["main","develop"],"source":"file"}`

// TestBootstrapGet: bootstrap_get returns the dict of the answer; code 1 of
// the engine gives exit 1; code 2, another code, an answer that is not JSON
// and no engine give the error lines and exit 2 (D10, D22).
func TestBootstrapGet(t *testing.T) {
	const getPy = "answer = helpers.bootstrap_get('repo.core_branches')\nprint(json.dumps(answer, sort_keys=True))\n"

	t.Run("an answer", func(t *testing.T) {
		t.Parallel()
		dir, _, p := stubCase(t, "STUB_OUT="+answerJSON)
		got := p.py(t, dir, getPy)
		var gotMap, want map[string]any
		_ = json.Unmarshal([]byte(got.out), &gotMap)
		_ = json.Unmarshal([]byte(answerJSON), &want)
		harness.Check(t, "bootstrap_get: the dict of the answer", got.code == 0 && reflect.DeepEqual(gotMap, want), got.String())
		harness.Check(t, "bootstrap_get: the arguments of the engine",
			strings.HasPrefix(harness.ReadFile(filepath.Join(dir, "log")), "config get repo.core_branches --format json|"),
			harness.ReadFile(filepath.Join(dir, "log")))
	})
	t.Run("an answer of another repository", func(t *testing.T) {
		t.Parallel()
		dir, _, p := stubCase(t, "STUB_OUT="+answerJSON, "BOOTSTRAP_VALUES=/values")
		other := filepath.Join(dir, "other")
		if err := os.MkdirAll(other, 0o755); err != nil {
			t.Fatal(err)
		}
		got := p.py(t, dir, "helpers.bootstrap_get('repo.core_branches', "+pyQuote(other)+")\n")
		log := harness.ReadFile(filepath.Join(dir, "log"))
		real, _ := filepath.EvalSymlinks(other)
		want := "config get repo.core_branches --format json|" + real + "|" + other + "|unset\n"
		harness.Check(t, "bootstrap_get with a repo: the folder, the root and no BOOTSTRAP_VALUES",
			got.code == 0 && log == want, got.String()+"log: "+log+"want: "+want)
	})
	for _, c := range []struct {
		name string
		kv   []string
		code int
	}{
		{"engine code 1", []string{"STUB_CODE=1", "STUB_ERR=[ERROR] the stub read fails"}, 1},
		{"engine code 2", []string{"STUB_CODE=2", "STUB_ERR=[ERROR] the stub refuses the read"}, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, _, p := stubCase(t, c.kv...)
			py := p.py(t, dir, getPy)
			harness.Check(t, fmt.Sprintf("bootstrap_get, %s: exit %d", c.name, c.code), py.code == c.code, py.String())
			same(t, "bootstrap_get, "+c.name+": the output of the shell", "bootstrap_get-"+slug(c.name), py, dir)
		})
	}
	// The shell maps each code other than 0 and 1 to exit 2; the Python form
	// does the same, and also for an answer that is not JSON (D22).
	wantLines := harness.Lines(
		"[ERROR] bootstrap config get repo.core_branches failed.",
		"[WARN] - with no bootstrap, or one older than 4.0.0: mise use -g github:virajp/bootstrap",
		"[WARN] - else do the fix of the message above, for example run bootstrap migrate, or bootstrap catalog to fetch the clone")
	for _, c := range []struct {
		name string
		kv   []string
	}{
		{"engine code 3", []string{"STUB_CODE=3"}},
		{"an answer that is not JSON", []string{"STUB_OUT=not json"}},
		{"a missing BOOTSTRAP_BIN", []string{"BOOTSTRAP_BIN=/nonexistent/bootstrap"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, _, p := stubCase(t, c.kv...)
			got := p.py(t, dir, getPy)
			harness.Check(t, "bootstrap_get, "+c.name+": the error lines and exit 2", got.code == 2 && got.out == wantLines,
				got.String())
		})
	}
	t.Run("no engine", func(t *testing.T) {
		t.Parallel()
		dir, root := newCase(t)
		// PATH has the mise shim and the system folders only, so
		// `mise x -- which bootstrap` finds no engine.
		e := env(dir, root, "NO_COLOR=1", "PATH="+harness.ShimDir+":/usr/bin:/bin")
		p := probe{dir: dir, env: e, merge: true}
		py := p.py(t, dir, getPy)
		harness.Check(t, "bootstrap_get, no engine: the error lines and exit 2", py.code == 2 && py.out == wantLines, py.String())
		same(t, "bootstrap_get, no engine: the output of the shell", "bootstrap_get-no-engine", py, dir)
	})
}

// TestMembers: members returns the submodule and sibling members only, with
// the defaults and the order of the shell version (D10).
func TestMembers(t *testing.T) {
	const membersPy = "for m in helpers.members():\n  print('\\t'.join(m))\nprint(json.dumps(helpers.members()))\n"
	for _, c := range []struct {
		name, answer, want string
	}{
		{"three kinds",
			`{"key":"repo.members","value":{"app":{"type":"submodule","folder":"apps/app","url":"https://example.com/app.git"},` +
				`"docs":{"type":"folder","folder":"docs"},"web":{"type":"sibling","folder":"../web","label":"Web"},` +
				`"api":{"type":"sibling"}},"source":"file"}`,
			`[["app", "apps/app", "app", "https://example.com/app.git"], ["web", "../web", "Web", ""], ["api", "api", "api", ""]]`},
		{"no members", `{"key":"repo.members","value":null,"source":"default"}`, `[]`},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, _, p := stubCase(t, "STUB_OUT="+c.answer)
			p.merge = false
			py := p.py(t, dir, membersPy)
			sh := shell(t, "members-"+slug(c.name))
			lines := strings.SplitAfter(py.out, "\n")
			last := strings.TrimSpace(lines[len(lines)-2])
			tsv := strings.Join(lines[:len(lines)-2], "")
			harness.Check(t, "members, "+c.name+": the tuples", py.code == 0 && last == c.want, py.String())
			harness.Check(t, "members, "+c.name+": the lines of the shell", sh.code == 0 && tsv == sh.out,
				"python:\n"+tsv+"shell:\n"+sh.String())
		})
	}
	t.Run("engine code 2", func(t *testing.T) {
		t.Parallel()
		dir, _, p := stubCase(t, "STUB_CODE=2", "STUB_ERR=[ERROR] no values")
		py := p.py(t, dir, membersPy)
		harness.Check(t, "members, engine code 2: exit 2", py.code == 2, py.String())
		same(t, "members, engine code 2: the output of the shell", "members-engine-code-2", py, dir)
	})
}

// TestOwnTasks: own_tasks and has_own_task name a .py part as mise does
// (D18, D22), the same as the shell functions and `mise tasks ls`.
func TestOwnTasks(t *testing.T) {
	t.Parallel()
	dir, root := newCase(t)
	grp := filepath.Join(root, ".config", "mise", "tasks", "grp")
	for _, name := range []string{"10-a", "20-b.py", "30-c", "30-c.py", "40-d", "e", "e.py", "_default.py", "60-dir/x"} {
		writeExec(t, filepath.Join(grp, name), "#!/bin/bash\necho "+name+"\n")
	}
	if err := os.Chmod(filepath.Join(grp, "40-d"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeExec(t, filepath.Join(root, ".config", "mise", "tasks", "solo.py"), "#!/bin/bash\necho solo\n")
	// A folder whose _default is not executable.
	harness.WriteFile(t, filepath.Join(root, ".config", "mise", "tasks", "off", "_default"), "#!/bin/bash\n")
	names := []string{"grp", "grp:10-a", "grp:20-b", "grp:20-b.py", "grp:30-c", "grp:30-c.py", "grp:40-d", "grp:e",
		"grp:e.py", "grp:60-dir", "grp:60-dir:x", "grp:zz", "solo", "off", "none"}
	var py strings.Builder
	py.WriteString("for n in helpers.own_tasks('grp') + helpers.own_tasks('none'):\n  print(n)\n")
	for _, n := range names {
		fmt.Fprintf(&py, "print(%s + ('=1' if helpers.has_own_task(%s) else '=0'))\n", pyQuote(n), pyQuote(n))
	}
	p := probe{dir: dir, env: env(dir, root)}
	gotPy := p.py(t, dir, py.String())
	same(t, "own_tasks and has_own_task: the answers of the shell", "own_tasks", gotPy, dir)
	want := []string{"grp:10-a", "grp:20-b", "grp:30-c", "grp:30-c.py"}
	gotParts := strings.Split(gotPy.out, "\n")[:4]
	harness.Check(t, "own_tasks: the names of D22", reflect.DeepEqual(gotParts, want), gotPy.String())

	// mise names the same parts. The global tasks of the user can also show,
	// so only the parts of grp count.
	git(t, p.env, root, "init", "--quiet")
	cmd := probe{dir: root, env: harness.SetEnv(p.env, "PATH="+os.Getenv("PATH"))}
	ls := cmd.run(t, harness.RealMise, "tasks", "ls", "--hidden", "--json")
	var tasks []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(ls.out), &tasks); err != nil {
		t.Fatalf("mise tasks ls: %v\n%s", err, ls.String())
	}
	var miseParts []string
	for _, task := range tasks {
		if rest, ok := strings.CutPrefix(task.Name, "grp:"); ok && len(rest) > 3 && rest[2] == '-' &&
			rest[0] >= '0' && rest[0] <= '9' && rest[1] >= '0' && rest[1] <= '9' && !strings.Contains(rest, ":") {
			miseParts = append(miseParts, task.Name)
		}
	}
	harness.Check(t, "own_tasks: the parts of mise tasks ls", reflect.DeepEqual(miseParts, want),
		fmt.Sprintf("mise: %v", miseParts))
}

// gateCase is a scratch git repository that is also the project root, with a
// gate file outside it.
func gateCase(t *testing.T) (string, string, string, probe) {
	t.Helper()
	dir := harness.Dir(t)
	root := filepath.Join(dir, "repo")
	installLibs(t, root)
	gate := filepath.Join(dir, "gate")
	harness.WriteFile(t, gate, harness.Lines("#!/usr/bin/env bash", "", `#USAGE flag "--all"`,
		`#USAGE arg "[files]" var=#true`, `# FILES include="*.md"`, "", "set -e"))
	e := env(dir, root, "NO_COLOR=1")
	git(t, e, dir, "init", "--quiet", "-b", "main", root)
	harness.WriteFile(t, filepath.Join(root, ".gitignore"), ".config/\n")
	for _, f := range []string{"a.md", "b.txt", "c d.md", "sub/e.md", "sub/f.go"} {
		harness.WriteFile(t, filepath.Join(root, f), "x\n")
	}
	return dir, root, gate, probe{dir: root, env: e}
}

const gatePy = `list = helpers.gate_files(os.environ["GATE"], *json.loads(os.environ["GATE_ARGS"]))
print(open(list).read(), end="")
os.unlink(list)
`

// TestGateFiles: gate_files runs the real gate_files.py and returns the path
// of the list; the list is the list of the shell gate_files (D12).
func TestGateFiles(t *testing.T) {
	for _, c := range []struct {
		name string
		args []string
		want string
	}{
		{"--all", []string{"--all"}, "a.md\nc d.md\nsub/e.md\n"},
		{"the uncommitted files", nil, "a.md\nc d.md\nsub/e.md\n"},
		{"given files", []string{"sub/e.md", "b.txt", "c d.md"}, "sub/e.md\nc d.md\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, _, gate, p := gateCase(t)
			args, _ := json.Marshal(append([]string{}, c.args...))
			p.env = harness.SetEnv(p.env, "GATE="+gate, "GATE_ARGS="+string(args))
			py := p.py(t, dir, gatePy)
			harness.Check(t, "gate_files, "+c.name+": the list", py.code == 0 && sortLines(py.out) == sortLines(c.want), py.String())
			same(t, "gate_files, "+c.name+": the list of the shell", "gate_files-"+slug(c.name), py, dir)
		})
	}
	t.Run("a failure", func(t *testing.T) {
		t.Parallel()
		dir, _, _, p := gateCase(t)
		p.merge = true
		p.env = harness.SetEnv(p.env, "GATE="+filepath.Join(dir, "no-such-gate"), `GATE_ARGS=["--all"]`)
		py := p.py(t, dir, gatePy)
		sh := shell(t, "gate_files-a-failure")
		harness.Check(t, "gate_files, a failure: the error line and the code of the script",
			py.code != 0 && strings.Contains(py.out, "[ERROR] gate_files failed for "+filepath.Join(dir, "no-such-gate")+" with code "),
			py.String())
		harness.Check(t, "gate_files, a failure: the code of the shell", py.code == sh.code, py.String()+sh.String())
	})
}

// TestUsageFiles: usage_files keeps the quotes of a name with a space (D26),
// and gate_files gets the same list as the shell gate with eval.
func TestUsageFiles(t *testing.T) {
	t.Parallel()
	dir, _, gate, p := gateCase(t)
	usage := `'c d.md' a.md "sub/e.md"`
	p.env = harness.SetEnv(p.env, "usage_files="+usage, "GATE="+gate)
	py := p.py(t, dir, "print(json.dumps(helpers.usage_files()))\n"+
		"list = helpers.gate_files(os.environ['GATE'], *helpers.usage_files())\nprint(open(list).read(), end='')\nos.unlink(list)\n")
	harness.Check(t, "usage_files: the words", strings.HasPrefix(py.out, `["c d.md", "a.md", "sub/e.md"]`+"\n"), py.String())
	same(t, "usage_files: the list of the shell gate", "usage_files", py, dir)
}

// sortLines gives the lines of <s> in order, so two lists compare as sets.
func sortLines(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := 1; i < len(lines); i++ {
		for j := i; j > 0 && lines[j] < lines[j-1]; j-- {
			lines[j], lines[j-1] = lines[j-1], lines[j]
		}
	}
	return strings.Join(lines, "\n")
}
