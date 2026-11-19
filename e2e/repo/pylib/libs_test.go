package pylib

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// checkCalls are the predicates of checks, with their arguments.
var checkCalls = [][]string{
	{"has_uncommitted_changes"}, {"has_untracked_files"},
	{"is_local_branch"}, {"is_local_branch", "main"}, {"is_local_branch", "local"}, {"is_local_branch", "ahead"},
	{"has_unpushed_commits"}, {"has_unpushed_commits", "main"}, {"has_unpushed_commits", "ahead"},
	{"has_unpushed_commits", "local"},
	{"branch_exists", "main"}, {"branch_exists", "local"}, {"branch_exists", "remoteonly"}, {"branch_exists", "nope"},
	{"is_git_worktree"},
}

// TestChecks: each predicate of checks.py gives the answer of the shell
// predicate, in the main checkout, in a linked worktree, in a dirty checkout
// and outside a repository.
func TestChecks(t *testing.T) {
	t.Parallel()
	dir, root := newCase(t)
	e := env(dir, root, "GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com")
	origin := filepath.Join(dir, "origin.git")
	work := filepath.Join(dir, "work")
	git(t, e, dir, "init", "--quiet", "--bare", "-b", "main", origin)
	git(t, e, dir, "init", "--quiet", "-b", "main", work)
	harness.WriteFile(t, filepath.Join(work, "file"), "one\n")
	git(t, e, work, "add", "file")
	git(t, e, work, "commit", "--quiet", "-m", "one")
	git(t, e, work, "remote", "add", "origin", origin)
	git(t, e, work, "push", "--quiet", "-u", "origin", "main")
	git(t, e, work, "branch", "local")
	git(t, e, work, "branch", "remoteonly")
	git(t, e, work, "push", "--quiet", "origin", "remoteonly")
	git(t, e, work, "branch", "-D", "remoteonly")
	git(t, e, work, "checkout", "--quiet", "-b", "ahead")
	git(t, e, work, "push", "--quiet", "-u", "origin", "ahead")
	harness.WriteFile(t, filepath.Join(work, "file"), "two\n")
	git(t, e, work, "commit", "--quiet", "-am", "two")
	git(t, e, work, "checkout", "--quiet", "main")
	linked := filepath.Join(dir, "linked")
	git(t, e, work, "worktree", "add", "--quiet", linked, "local")
	dirty := filepath.Join(dir, "dirty")
	git(t, e, dir, "clone", "--quiet", origin, dirty)
	harness.WriteFile(t, filepath.Join(dirty, "file"), "changed\n")
	harness.WriteFile(t, filepath.Join(dirty, "new"), "new\n")
	outside := filepath.Join(dir, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}

	var py strings.Builder
	py.WriteString("import checks  # noqa: E402\n\n")
	for _, c := range checkCalls {
		name := strings.Join(c, " ")
		args := make([]string, 0, len(c)-1)
		for _, a := range c[1:] {
			args = append(args, pyQuote(a))
		}
		fmt.Fprintf(&py, "print('%s=' + ('1' if checks.%s(%s) else '0'))\n", name, c[0], strings.Join(args, ", "))
	}
	py.WriteString("print('main=' + checks.get_main_worktree_path())\n")

	for _, where := range []struct{ name, dir string }{
		{"the main checkout", work}, {"a linked worktree", linked}, {"a dirty checkout", dirty}, {"outside a repository", outside},
	} {
		p := probe{dir: where.dir, env: harness.SetEnv(e, "GIT_CEILING_DIRECTORIES="+dir)}
		gotPy := p.py(t, dir, py.String())
		if where.dir == outside {
			// Outside a repository the shell predicates show the fatal lines of
			// git on stderr, and checks.py shows fewer: the case compares the
			// answers and the code only.
			got, want := subst(gotPy, dir, "@DIR@"), shell(t, "checks-"+slug(where.name))
			harness.Check(t, "checks, "+where.name+": the answers of the shell",
				got.out == want.out && got.code == want.code, "python:\n"+got.String()+"shell:\n"+want.String())
			continue
		}
		same(t, "checks, "+where.name+": the answers of the shell", "checks-"+slug(where.name), gotPy, dir)
	}
}

// TestPlaceholder: placeholder.py finds both forms of the marker, names a
// .py task without .py, and prints the notice of the shell version (D27).
func TestPlaceholder(t *testing.T) {
	t.Run("a .py task", func(t *testing.T) {
		t.Parallel()
		dir, root := newCase(t)
		tasks := filepath.Join(root, ".config", "mise", "tasks")
		writeExec(t, filepath.Join(tasks, "code", "lint", "70-x.py"), "#!/usr/bin/env -S uv run --script\n# PLACEHOLDER\n")
		writeExec(t, filepath.Join(tasks, "code", "format", "10-y"), "#!/usr/bin/env bash\n#PLACEHOLDER\n")
		writeExec(t, filepath.Join(tasks, "setup", "z"), "#!/usr/bin/env bash\necho z\n")
		p := probe{dir: dir, env: env(dir, root, "NO_COLOR=1"), merge: true}
		got := p.py(t, dir, `import placeholder  # noqa: E402

print(placeholder.placeholder_tasks())
placeholder.placeholder_notice("code:lint:70-x", "no linter")
`)
		want := "['code:format:10-y', 'code:lint:70-x']\n" + harness.Lines(
			"[WARN] code:lint:70-x is not configured yet — no linter", "",
			"Unconfigured in this repo: code:format:10-y code:lint:70-x", "",
			"To configure them:",
			"  /vwf:architecture   pin this repo's stack axes (do this first)",
			"  /stackgen           materialize the packs that fill them", "")
		harness.Check(t, "placeholder.py: the tasks and the notice", got.code == 0 && got.out == want, got.String())
	})
	for _, c := range []struct {
		name  string
		files map[string]string
	}{
		{"shell tasks", map[string]string{"code/format/10-y": "#PLACEHOLDER\n", "test/_default": "x #PLACEHOLDER\n",
			"setup/z": "echo z\n"}},
		{"no placeholder", map[string]string{"setup/z": "echo z\n"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, root := newCase(t)
			for name, body := range c.files {
				writeExec(t, filepath.Join(root, ".config", "mise", "tasks", name), "#!/usr/bin/env bash\n"+body)
			}
			p := probe{dir: dir, env: env(dir, root, "NO_COLOR=1"), merge: true}
			py := p.py(t, dir, "import placeholder  # noqa: E402\n\nplaceholder.placeholder_notice('code:lint', 'no stack')\n")
			same(t, "placeholder, "+c.name+": the notice of the shell", "placeholder-"+slug(c.name), py, dir)
		})
	}
}

// The rows of the page case. They have the HTML characters &<>" and paths
// below HOME, because the shell side ran with /bin/bash (D31).
var pageRows = []map[string]string{
	{"kind": "repo", "repo": "org/app", "r_role": "root", "r_host": "github", "r_glyph": "", "r_default": "main",
		"r_fetched": "2h", "r_pinned": ""},
	{"kind": "main_worktree", "repo": "org/app", "w_branch": "main", "w_commit": "abc1234", "w_age": "2h", "w_base": "",
		"w_ba": "", "w_bb": "", "w_pa": "", "w_pb": "", "w_origin": "tracked", "w_oa": "0", "w_ob": "0", "w_sync": "yes",
		"w_conflict": "0", "w_dirty": "0", "w_stash": "0", "w_merged": "develop", "w_default": "yes", "w_path": "/home/tester/src/app"},
	{"kind": "worktree", "repo": "org/app", "w_branch": `feat/<x>&"y"`, "w_commit": "def5678", "w_age": "5m",
		"w_base": "main", "w_ba": "2", "w_bb": "1", "w_pa": "0", "w_pb": "0", "w_origin": "gone", "w_sync": "",
		"w_conflict": "1", "w_dirty": "3", "w_dglyph": "±", "w_codes": "M2 ??1", "w_op": "rebase", "w_locked": "yes",
		"w_stash": "2", "w_merged": "no", "w_path": "/home/tester/src/app-feat & more"},
	{"kind": "worktree", "repo": "org/app", "w_branch": "", "w_detached": "yes", "w_commit": "0123abc", "w_age": "1d",
		"w_ba": "0", "w_bb": "0", "w_origin": "local", "w_sync": "yes", "w_conflict": "0", "w_dirty": "0",
		"w_notpin": "yes", "w_stash": "0", "w_merged": "both", "w_path": "/srv/other"},
	{"kind": "worktree", "repo": "org/app", "w_branch": "gone", "w_commit": "0000000", "w_age": "3d", "w_origin": "none",
		"w_conflict": "0", "w_dirty": "0", "w_missing": "yes", "w_stash": "0", "w_path": "/home/tester/gone"},
	{"kind": "branch", "repo": "org/app", "w_branch": "old", "w_commit": "1111111", "w_age": "9d", "w_ba": "0",
		"w_bb": "4", "w_origin": "tracked", "w_oa": "1", "w_ob": "0", "w_merged": "main", "w_conflict": "0", "w_dirty": "0",
		"w_stash": "0"},
	{"kind": "branch", "repo": "org/app", "w_branch": "wip", "w_commit": "2222222", "w_age": "1h", "w_origin": "local",
		"w_conflict": "0", "w_dirty": "0", "w_stash": "0", "w_merged": "no"},
	{"kind": "repo", "repo": "lib <one>", "r_role": "submodule", "r_host": "gitlab", "r_glyph": "\U000f0ba0",
		"r_default": "", "r_fetched": "", "r_pinned": "org/app", "r_note": "not on disk & \"missing\"",
		"r_fix": "git submodule update --init <lib>"},
}

var pageAttention = []map[string]string{
	{"a_rank": "2", "a_glyph": "", "a_color": "red", "a_repo": "org/app", "a_branch": `feat/<x>&"y"`,
		"a_msg": "a conflict & a stop", "a_cmd": "git rebase --continue"},
	{"a_rank": "1", "a_glyph": "", "a_color": "yellow", "a_repo": "lib <one>", "a_branch": "",
		"a_msg": "not on disk", "a_cmd": "git submodule update --init"},
	{"a_rank": "3", "a_glyph": "", "a_color": "yellow", "a_repo": "org/app", "a_branch": "main",
		"a_msg": "stash", "a_cmd": "git stash list"},
}

// pagePython gives the Python probe of the page.
func pagePython(folder string, rows, attention []map[string]string) string {
	enc := func(list []map[string]string) string {
		var b strings.Builder
		b.WriteString("[")
		for _, r := range list {
			b.WriteString("{")
			for k, v := range r {
				b.WriteString(pyQuote(k) + ": " + pyQuote(v) + ", ")
			}
			b.WriteString("}, ")
		}
		return b.String() + "]"
	}
	return "import worktrees_page  # noqa: E402\n\nresult = worktrees_page.write_worktrees_page(" + enc(rows) + ", " +
		enc(attention) + ", '2026-01-01 00:00', folder=" + pyQuote(folder) + ")\nprint(f'returned={result}')\n"
}

var pageDate = regexp.MustCompile(`on [0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}\.</p>`)

// TestWorktreesPage: write_worktrees_page writes the fixed HTML bytes, with
// the date line fixed (D23, D31). The bytes came from the shell function run
// by /bin/bash; the Merged column came after it.
func TestWorktreesPage(t *testing.T) {
	for _, c := range []struct {
		name            string
		rows, attention []map[string]string
	}{
		{"rows and attention", pageRows, pageAttention},
		{"nothing needs attention", pageRows[:2], nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, root := newCase(t)
			pyDir := filepath.Join(dir, "py-page")
			p := probe{dir: dir, env: env(dir, root, "HOME=/home/tester")}
			py := p.py(t, dir, pagePython(pyDir, c.rows, c.attention))
			name := strings.ReplaceAll(c.rows[0]["repo"], "/", "_") + ".html"
			pyFile := filepath.Join(pyDir, name)
			harness.Check(t, c.name+": the path, printed and returned",
				py.code == 0 && py.out == pyFile+"\nreturned="+pyFile+"\n", py.String())
			fixed := "on DATE.</p>"
			pyHTML := pageDate.ReplaceAllString(harness.ReadFile(pyFile), fixed)
			shHTML := shellText(t, "worktrees_page-"+slug(c.name)+".html")
			harness.Check(t, c.name+": the fixed HTML bytes", pyHTML != "" && pyHTML == shHTML,
				firstHTMLDiff(pyHTML, shHTML))
			harness.Check(t, c.name+": the date line is fixed", strings.Count(pyHTML, fixed) == 1, "")
		})
	}
	t.Run("a folder that is a symlink", func(t *testing.T) {
		t.Parallel()
		dir, root := newCase(t)
		real := filepath.Join(dir, "real")
		link := filepath.Join(dir, "link")
		if err := os.MkdirAll(real, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		p := probe{dir: dir, env: env(dir, root), merge: true}
		got := p.py(t, dir, pagePython(link, pageRows[:2], nil))
		harness.Check(t, "a symlink folder: no output, None", got.code == 0 && got.out == "returned=None\n", got.String())
	})
}

// firstHTMLDiff shows the text around the first byte where two pages differ.
func firstHTMLDiff(a, b string) string {
	i := 0
	for i < len(a) && i < len(b) && a[i] == b[i] {
		i++
	}
	from := max(0, i-80)
	return fmt.Sprintf("python (%d bytes): %q\nshell  (%d bytes): %q", len(a), a[from:min(len(a), i+80)], len(b),
		b[from:min(len(b), i+80)])
}

// mergeRepo makes a work repository with a bare origin: main, and feature
// with one commit more, both pushed. The libraries are below .config/, which
// .git/info/exclude hides, so the merge sees a clean tree.
func mergeRepo(t testing.TB, e []string, base string) string {
	t.Helper()
	origin := filepath.Join(base, "origin.git")
	work := filepath.Join(base, "work")
	git(t, e, base, "init", "--quiet", "--bare", "-b", "main", origin)
	git(t, e, base, "init", "--quiet", "-b", "main", work)
	installLibs(t, work)
	harness.AppendFile(t, filepath.Join(work, ".git", "info", "exclude"), "/.config/\n")
	harness.WriteFile(t, filepath.Join(work, "file"), "one\n")
	git(t, e, work, "add", "file")
	git(t, e, work, "commit", "--quiet", "-m", "one")
	git(t, e, work, "remote", "add", "origin", origin)
	git(t, e, work, "push", "--quiet", "-u", "origin", "main")
	git(t, e, work, "checkout", "--quiet", "-b", "feature")
	harness.WriteFile(t, filepath.Join(work, "feature"), "feature\n")
	git(t, e, work, "add", "feature")
	git(t, e, work, "commit", "--quiet", "-m", "feature")
	git(t, e, work, "push", "--quiet", "-u", "origin", "feature")
	return work
}

// TestMerge: merge_to_destination_branch lands feature on main in the landing
// mode direct-merge with the git log, hashes included, and the output of the
// shell function (D30).
func TestMerge(t *testing.T) {
	t.Parallel()
	dir := harness.Dir(t)
	stub := filepath.Join(dir, "engine")
	writeExec(t, stub, stubEngine)
	answer := `{"key":"repo.branches.\"main\"","value":{"landing":"direct-merge"},"source":"file"}`
	fixed := []string{"GIT_AUTHOR_NAME=Test Author", "GIT_AUTHOR_EMAIL=author@example.com",
		"GIT_COMMITTER_NAME=Test Committer", "GIT_COMMITTER_EMAIL=committer@example.com",
		"GIT_AUTHOR_DATE=2026-01-02T03:04:05Z", "GIT_COMMITTER_DATE=2026-01-02T03:04:05Z",
		"NO_COLOR=1", "BOOTSTRAP_BIN=" + stub, "STUB_OUT=" + answer, "STUB_LOG=" + filepath.Join(dir, "log")}
	base := filepath.Join(dir, "python")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	e := env(base, filepath.Join(base, "work"), fixed...)
	work := mergeRepo(t, e, base)
	p := probe{dir: work, env: e, merge: true}
	py := p.py(t, base, "import merge  # noqa: E402\n\nmerge.merge_to_destination_branch('main', 'feature')\n")
	log := git(t, e, work, "log", "--format=%H %P %s", "--all") + "\n" +
		git(t, e, filepath.Join(base, "origin.git"), "log", "--format=%H %P %s", "--all") + "\n" +
		git(t, e, work, "rev-parse", "--abbrev-ref", "HEAD")
	harness.Check(t, "merge: exit 0 and the success line", py.code == 0 &&
		strings.Contains(py.out, "Landing feature on main (direct-merge) ...") &&
		strings.HasSuffix(py.out, "[OK] Merge completed successfully!\n"), py.String())
	got, want := subst(py, base, "@BASE@", dir, "@DIR@"), shell(t, "merge.out")
	harness.Check(t, "merge: the output of the shell", got == want, "python:\n"+got.String()+"shell:\n"+want.String())
	wantLog := shellText(t, "merge.log")
	harness.Check(t, "merge: the git log of the shell, hashes included", log == wantLog,
		"python:\n"+log+"\nshell:\n"+wantLog)
	harness.Check(t, "merge: a merge commit on main of origin",
		strings.Contains(log, "Merge branch 'feature'") && strings.HasSuffix(log, "\nfeature"), log)
}

// TestXcode: find_xcodeproj returns the one project, or stops with the lines
// and the code of the shell; require_xcode returns the values that the shell
// exports, and sets them in os.environ (D19).
func TestXcode(t *testing.T) {
	const findPy = "import xcode  # noqa: E402\n\nprint(f'project={xcode.find_xcodeproj(\"swift:build\")}')\n"
	for _, c := range []struct {
		name     string
		projects []string
		code     int
	}{
		{"no project", nil, 1},
		{"one project", []string{"App.xcodeproj"}, 0},
		{"two projects", []string{"App.xcodeproj", "Other.xcodeproj"}, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, root := newCase(t)
			app := filepath.Join(dir, "app")
			for _, pr := range append([]string{""}, c.projects...) {
				if err := os.MkdirAll(filepath.Join(app, pr), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			// A file with the name of a project is not a project.
			harness.WriteFile(t, filepath.Join(app, "File.xcodeproj"), "")
			p := probe{dir: app, env: env(dir, root, "NO_COLOR=1")}
			py := p.py(t, dir, findPy)
			harness.Check(t, fmt.Sprintf("find_xcodeproj, %s: exit %d", c.name, c.code), py.code == c.code, py.String())
			if c.code == 0 {
				harness.Check(t, "find_xcodeproj, one project: the path", py.out == "project=App.xcodeproj\n" && py.errOut == "",
					py.String())
			}
			same(t, "find_xcodeproj, "+c.name+": the output of the shell", "find_xcodeproj-"+slug(c.name), py, dir)
		})
	}
	// require_xcode runs a stub xcodebuild and a stub xcrun, so the values
	// are the same on each machine.
	t.Run("require_xcode", func(t *testing.T) {
		t.Parallel()
		dir, root := newCase(t)
		bin := filepath.Join(dir, "bin")
		writeExec(t, filepath.Join(bin, "xcodebuild"), stubXcodebuild)
		writeExec(t, filepath.Join(bin, "xcrun"), stubXcrun)
		p := probe{dir: dir, env: env(dir, root, "NO_COLOR=1", "XCODE_VERSION", "SIMULATOR_PLATFORM", "SIMULATOR_OS",
			"SIMULATOR_DEVICE", "PATH="+bin+":/usr/bin:/bin")}
		const names = "XCODE_VERSION SIMULATOR_PLATFORM SIMULATOR_OS SIMULATOR_DEVICE"
		py := p.py(t, dir, `import xcode  # noqa: E402

values = xcode.require_xcode("swift:build")
for name in "`+names+`".split():
  print(f"{name}={values.get(name, '')}")
print("environ=" + str(all(os.environ[k] == v for k, v in values.items())))
`)
		same(t, "require_xcode: the values of the shell", "require_xcode", py, dir)
		harness.Check(t, "require_xcode: a dict with XCODE_VERSION, set in os.environ",
			py.code == 0 && strings.Contains(py.out, "XCODE_VERSION=26.1\n") && strings.HasSuffix(py.out, "environ=True\n"),
			py.String())
	})
}

// The stub xcodebuild of require_xcode: the selected Xcode is 26.1.
const stubXcodebuild = "#!/bin/sh\nprintf 'Xcode 26.1\\nBuild version 17B55\\n'\n"

// The stub xcrun of require_xcode: the available simulators of two iOS
// runtimes and a watchOS runtime. The newest iOS runtime is 26.1, and its
// iPhone with the highest model number and no other word is iPhone 17.
const stubXcrun = `#!/bin/sh
cat <<'END'
== Devices ==
-- iOS 18.2 --
    iPhone 16 (00000000-0000-0000-0000-000000000001) (Shutdown)
    iPhone 16 Pro (00000000-0000-0000-0000-000000000002) (Shutdown)
-- iOS 26.1 --
    iPhone 9 (00000000-0000-0000-0000-000000000003) (Shutdown)
    iPhone 17 (00000000-0000-0000-0000-000000000004) (Shutdown)
    iPhone 17 Pro Max (00000000-0000-0000-0000-000000000005) (Booted)
    iPhone Air (00000000-0000-0000-0000-000000000006) (Shutdown)
-- watchOS 26.0 --
    Apple Watch Series 11 (46mm) (00000000-0000-0000-0000-000000000007) (Shutdown)
END
`
