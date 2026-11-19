// Package macostools renders each Mac tool with the sample values file, and
// compares the shared mise values of the two sides. The cases came from
// test:macos-tools (.config/mise/tasks/test/macos-tools, lines 1 to 416).
//
// The seed format case: no [files] table of a tools.toml gives
// `mode = "seed"` to a target of format json, yaml or toml. The render case:
// the first real run exits 4 when a tool has `stage = "init"`, else 0, the
// second run exits 0, the example selects each tool of macos/tools/, and each
// .json, .yaml/.yml and .toml file of the temporary HOME parses. The missing
// question case: the example without the answer `config_path` of mise exits
// 2. The init case: `bootstrap mac migrate` migrates the old form of the
// example, then `bootstrap mac sync` runs. The git cases: the git tool needs
// one email only. The drift cases: a mise tool on both sides of repo/tools/
// and macos/tools/ has one version and one `depends`, a mise setting on both
// sides has one value, and a fixture proves that the check finds a
// difference.
package macostools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

var (
	// tree is a git copy of cli/ and macos/ of the working tree, so a run can
	// record its commit.
	tree string
	// example is the sample values file.
	example string
	// miseDirs are the real mise folders, so `mise x <tool>@latest`
	// resolves from the mise cache when a case changes HOME and XDG_*.
	miseDirs []string
)

func TestMain(m *testing.M) { harness.Main(m, setup, "macos", "repo") }

func setup() error {
	example = filepath.Join(harness.Root, "macos", "examples", "macos.toml")
	tree = filepath.Join(harness.Tmp, "tree")
	for _, d := range []string{"cli", "macos"} {
		if err := harness.CopyTree(filepath.Join(harness.Root, d), filepath.Join(tree, d)); err != nil {
			return err
		}
	}
	for _, args := range [][]string{{"init", "--quiet", "-b", "main"}, {"add", "-A"},
		{"commit", "--quiet", "--no-verify", "-m", "tree"}} {
		if out, err := harness.GitErr(tree, args...); err != nil {
			return fmt.Errorf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	home := os.Getenv("HOME")
	data := os.Getenv("MISE_DATA_DIR")
	if data == "" {
		x := os.Getenv("XDG_DATA_HOME")
		if x == "" {
			x = filepath.Join(home, ".local", "share")
		}
		data = filepath.Join(x, "mise")
	}
	cache := os.Getenv("MISE_CACHE_DIR")
	if cache == "" {
		x := os.Getenv("XDG_CACHE_HOME")
		if x == "" {
			x = filepath.Join(home, ".cache")
		}
		cache = filepath.Join(x, "mise")
	}
	miseDirs = []string{"MISE_DATA_DIR=" + data, "MISE_CACHE_DIR=" + cache}
	return nil
}

// macCase is one isolated environment with a copy of the example.
type macCase struct {
	t      *testing.T
	name   string
	dir    string
	home   string
	config string
	env    []string
	runs   int
	log    string
	code   int
}

// newCase is setup_case: a new isolated environment with a copy of the
// example in a git config folder.
func newCase(t *testing.T, name string) *macCase {
	dir := harness.Dir(t)
	c := &macCase{t: t, name: name, dir: dir, home: filepath.Join(dir, "home"), config: filepath.Join(dir, "config")}
	for _, d := range []string{"home", "xdg/cache", "xdg/config", "xdg/state", "xdg/data", "tmp", "config"} {
		if err := os.MkdirAll(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	gitconfig := filepath.Join(dir, "gitconfig")
	harness.WriteFile(t, gitconfig, "")
	c.env = harness.SetEnv(harness.Env(), append([]string{
		"MISE_ENV", "BOOTSTRAP_URL", "BOOTSTRAP_CLONE",
		"HOME=" + c.home,
		"XDG_CACHE_HOME=" + filepath.Join(dir, "xdg/cache"),
		"XDG_CONFIG_HOME=" + filepath.Join(dir, "xdg/config"),
		"XDG_STATE_HOME=" + filepath.Join(dir, "xdg/state"),
		"XDG_DATA_HOME=" + filepath.Join(dir, "xdg/data"),
		"BOOTSTRAP_CACHE=" + filepath.Join(dir, "xdg/cache/bootstrap"),
		"MACOS_SETUP_REPO_PATH=" + c.config,
		"GIT_CONFIG_GLOBAL=" + gitconfig,
		"GIT_CONFIG_NOSYSTEM=1",
		"TMPDIR=" + filepath.Join(dir, "tmp"),
	}, miseDirs...)...)
	harness.WriteFile(t, filepath.Join(c.config, "macos.toml"), harness.ReadFile(example))
	c.git("init", "--quiet", c.config)
	c.commitConfig()
	return c
}

// git runs git without the config of the user, hooks, signing or identity.
func (c *macCase) git(args ...string) {
	c.t.Helper()
	all := append([]string{"-c", "user.name=test", "-c", "user.email=test@example.com",
		"-c", "commit.gpgsign=false", "-c", "core.hooksPath=/dev/null", "-c", "init.defaultBranch=main"}, args...)
	cmd := harness.Cmd{Dir: c.dir, Env: harness.SetEnv(c.env, "GIT_CONFIG_GLOBAL=/dev/null"), Args: append([]string{"git"}, all...)}
	if code, out := cmd.Run(); code != 0 {
		c.t.Fatalf("git %s: exit %d\n%s", strings.Join(args, " "), code, out)
	}
}

// commitConfig commits each file of the config folder.
func (c *macCase) commitConfig() {
	c.git("-C", c.config, "add", "-A")
	c.git("-C", c.config, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "config")
}

// engine runs the engine in <dir>, as its entry task does, with the output
// in the next log file.
func (c *macCase) engine(dir string, args ...string) {
	c.runs++
	c.log = filepath.Join(c.dir, fmt.Sprintf("run-%d.txt", c.runs))
	f, err := os.Create(c.log)
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	cmd := harness.Cmd{Dir: dir, Env: harness.SetEnv(c.env, "BOOTSTRAP_CLONE="+tree), Out: f,
		Args: append([]string{harness.Bin}, args...)}
	c.code, _ = cmd.Run()
}

// runReal is run_real: a real run of `bootstrap mac sync`.
func (c *macCase) runReal() {
	c.engine(c.dir, "mac", "sync", "--no-bootstrap")
}

func (c *macCase) logText() string { return harness.ReadFile(c.log) }

func (c *macCase) pass(msg string) { harness.Pass(c.t, c.name+": "+msg) }

func (c *macCase) fail(msg string) { harness.Fail(c.t, c.name+": "+msg, c.logText()) }

func (c *macCase) check(ok bool, passMsg, failMsg string) {
	if ok {
		c.pass(passMsg)
	} else {
		c.fail(failMsg)
	}
}

func (c *macCase) checkExit(want int) {
	c.check(c.code == want, fmt.Sprintf("exit %d", c.code), fmt.Sprintf("exit %d, expected %d", c.code, want))
}

// tool runs a tool through `mise x <tool>@latest --` from `/`, and gives its
// exit code and output.
func tool(name string, args ...string) (int, string) {
	cmd := harness.Cmd{Dir: "/", Args: append([]string{"mise", "x", name + "@latest", "--", name}, args...)}
	return cmd.Run()
}

// parallelMap runs <f> for each item of <items>, eight at a time, and gives
// the results in the order of <items>.
func parallelMap[T any](items []string, f func(string) T) []T {
	out := make([]T, len(items))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for i, it := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			out[i] = f(it)
		}()
	}
	wg.Wait()
	return out
}

func TestMacosTools(t *testing.T) {
	t.Run("cases", func(t *testing.T) {
		t.Run("seed format", testSeedFormat)
		t.Run("render", testRender)
		t.Run("missing question", testMissingQuestion)
		t.Run("init", testInit)
		t.Run("git email", testGitEmail)
		t.Run("drift", testDrift)
		t.Run("drift fixture", testDriftFixture)
	})
	if !t.Failed() {
		harness.Pass(t, "Every Mac tool renders with the sample values file")
	}
}

type seedRows struct {
	code int
	out  string
}

// testSeedFormat is case 0: no JSON, YAML or TOML file is a seed.
func testSeedFormat(t *testing.T) {
	t.Parallel()
	tools, err := filepath.Glob(filepath.Join(harness.Root, "macos", "tools", "*", "tools.toml"))
	if err != nil {
		t.Fatal(err)
	}
	rows := parallelMap(tools, func(f string) seedRows {
		code, out := tool("yq", "-p", "toml", "-o", "tsv",
			`.files // {} | to_entries | map(select(.value.mode == "seed")) | .[] | [.key, (.value.format // "")]`, f)
		return seedRows{code, out}
	})
	seeds, bad := 0, 0
	for i, r := range rows {
		rel := strings.TrimPrefix(tools[i], harness.Root+"/")
		if r.code != 0 {
			t.Fatalf("yq %s: exit %d\n%s", rel, r.code, r.out)
		}
		for _, line := range strings.Split(r.out, "\n") {
			target, format, _ := strings.Cut(line, "\t")
			if target == "" {
				continue
			}
			if format == "" {
				format = formatOf(target)
			}
			seeds++
			if format != "text" {
				bad++
				harness.Fail(t, fmt.Sprintf("seed format: %s: %s is a seed of format %s; use mode = \"merge\"", rel, target, format), "")
			}
		}
	}
	if bad == 0 {
		harness.Pass(t, fmt.Sprintf("seed format: no seed of format json, yaml or toml (%d seed files)", seeds))
	}
}

// formatOf gives the format of a target from its extension.
func formatOf(target string) string {
	switch {
	case strings.HasSuffix(target, ".json"):
		return "json"
	case strings.HasSuffix(target, ".yaml"), strings.HasSuffix(target, ".yml"):
		return "yaml"
	case strings.HasSuffix(target, ".toml"):
		return "toml"
	}
	return "text"
}

var toolsLine = regexp.MustCompile(`(?m)^bootstrap: tools:(.*)$`)

// testRender is case 1: each tool renders, and each structured file parses.
func testRender(t *testing.T) {
	t.Parallel()
	c := newCase(t, "render")
	c.runReal()
	c.checkExit(0)

	selected := ""
	if m := toolsLine.FindStringSubmatch(c.logText()); m != nil {
		selected = m[1]
	}
	selected = " " + selected + " "
	dirs, err := os.ReadDir(filepath.Join(harness.Root, "macos", "tools"))
	if err != nil {
		t.Fatal(err)
	}
	missing := ""
	for _, d := range dirs {
		if d.IsDir() && !strings.Contains(selected, " "+d.Name()+" ") {
			missing += " " + d.Name()
		}
	}
	c.check(missing == "", "macos/examples/macos.toml selects every tool of macos/tools/",
		"macos/examples/macos.toml does not select:"+missing)
	c.check(!strings.Contains(c.logText(), "] added to "),
		"macos/examples/macos.toml has a table for each required member of the pack base",
		"the sync writes a table of a required member into macos/examples/macos.toml")

	c.runReal()
	c.checkExit(0)

	// mise does not expand ~ in [env], so the answer of the example reaches
	// it with $HOME in its place.
	miseConfig := filepath.Join(c.home, ".config/mise/conf.d/mise/mise.toml")
	c.check(strings.Contains(harness.ReadFile(miseConfig),
		`MACOS_SETUP_REPO_PATH = "`+c.home+`/Projects/github.com/janedoe/macos-setup"`),
		"a ~ at the start of config_path becomes $HOME",
		miseConfig+" does not have config_path with $HOME in place of ~")

	var files []string
	err = filepath.WalkDir(c.home, func(p string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.Type().IsRegular() && formatOf(p) != "text" {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	results := parallelMap(files, func(f string) seedRows {
		var code int
		var out string
		switch formatOf(f) {
		case "json":
			code, out = tool("jq", "-e", ".", f)
		case "yaml":
			code, out = tool("yq", "-p", "yaml", "-e", ".", f)
		default:
			code, out = tool("yq", "-p", "toml", "-o", "json", ".", f)
		}
		return seedRows{code, out}
	})
	bad := 0
	for i, r := range results {
		if r.code != 0 {
			bad++
			harness.Fail(t, "render: "+strings.TrimPrefix(files[i], c.home+"/")+" does not parse", r.out)
		}
	}
	if bad == 0 {
		c.pass(fmt.Sprintf("all %d .json, .yaml/.yml and .toml files parse", len(files)))
	}
	if len(files) == 0 {
		c.fail("no .json, .yaml/.yml or .toml file was rendered")
	}
}

// testMissingQuestion is case 2: a missing answer.
func testMissingQuestion(t *testing.T) {
	t.Parallel()
	c := newCase(t, "missing question")
	var keep []string
	for _, l := range strings.SplitAfter(harness.ReadFile(example), "\n") {
		if !strings.HasPrefix(l, "config_path = ") {
			keep = append(keep, l)
		}
	}
	harness.WriteFile(t, filepath.Join(c.config, "macos.toml"), strings.Join(keep, ""))
	c.runReal()
	c.checkExit(2)
	c.check(strings.Contains(c.logText(), "the question mise.config_path has no answer in macos.toml"),
		"the message names the question", "the message does not name the question 'mise.config_path'")
}

// hasLine tells whether <text> has the line <line>.
func hasLine(text, line string) bool {
	for _, l := range strings.Split(text, "\n") {
		if l == line {
			return true
		}
	}
	return false
}

// testInit is case 4: mac migrate migrates the old form and, with no
// manifest, writes the first setup (--no-bootstrap: no mise bootstrap, exit
// 0); then mac sync runs.
func testInit(t *testing.T) {
	t.Parallel()
	c := newCase(t, "init")
	var out []string
	for _, l := range strings.SplitAfter(harness.ReadFile(example), "\n") {
		switch strings.TrimSuffix(l, "\n") {
		case "[tools.mise]":
			l = strings.Replace(l, "[tools.mise]", "[tools.macos-setup]", 1)
		case "format = 2":
			continue
		}
		out = append(out, l)
	}
	values := filepath.Join(c.config, "macos.toml")
	harness.WriteFile(t, values, strings.Join(out, ""))
	c.commitConfig()
	c.engine(c.config, "mac", "migrate", "--yes", "--no-bootstrap")
	c.checkExit(0)
	text := harness.ReadFile(values)
	c.check(hasLine(text, "[tools.mise]") && !strings.Contains(text, "macos-setup]"),
		"[tools.macos-setup] becomes [tools.mise]", "macos.toml still has the old form")
	c.check(hasLine(text, "format = 2"),
		"the migration writes format = 2", "macos.toml has no format = 2 after mac migrate")
	c.commitConfig()
	c.runReal()
	c.checkExit(0)
}

// answer gives the answer of <key> in the example, as sed -n
// 's/^<key> = "\(.*\)"$/\1/p' prints it.
func answer(t *testing.T, key string) string {
	m := regexp.MustCompile(`(?m)^` + key + ` = "(.*)"$`).FindStringSubmatch(harness.ReadFile(example))
	if m == nil {
		t.Fatalf("the example has no answer %s", key)
	}
	return m[1]
}

// emptyEmail writes the copy of the example, with "" as the answer of each
// <key> of [tools.git], and commits it.
func (c *macCase) emptyEmail(keys ...string) {
	text := harness.ReadFile(example)
	for _, k := range keys {
		text = regexp.MustCompile(`(?m)^`+k+` = .*$`).ReplaceAllString(text, k+` = ""`)
	}
	harness.WriteFile(c.t, filepath.Join(c.config, "macos.toml"), text)
	c.commitConfig()
}

// checkFile passes <message> when <ok>.
func (c *macCase) checkFile(message string, ok bool) {
	c.check(ok, message, "not: "+message)
}

var emailLine = regexp.MustCompile(`(?m)^[[:space:]]*email[[:space:]]*=`)

// oneSigner tells whether <text> has one signer line. The header comment
// lines and the empty lines do not count.
func oneSigner(text string) bool {
	n := 0
	for _, l := range strings.Split(strings.TrimSuffix(text, "\n"), "\n") {
		if l != "" && !strings.HasPrefix(l, "#") {
			n++
		}
	}
	return n == 1
}

// testGitEmail is case 5: the git tool needs one email only.
func testGitEmail(t *testing.T) {
	t.Parallel()
	github, gitlab := answer(t, "email_github"), answer(t, "email_gitlab")
	for _, tc := range []struct{ name, key, email, config string }{
		{"git: email_gitlab empty", "email_gitlab", github, "gitlab.config"},
		{"git: email_github empty", "email_github", gitlab, "github.config"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			c := newCase(t, tc.name)
			c.emptyEmail(tc.key)
			c.runReal()
			c.checkExit(0)
			c.runReal()
			c.checkExit(0)
			git := filepath.Join(c.home, ".config/git")
			c.checkFile("identity has the email "+tc.email,
				hasLine(harness.ReadFile(filepath.Join(git, "identity")), "\temail = "+tc.email))
			c.checkFile(tc.config+" has no email line",
				!emailLine.MatchString(harness.ReadFile(filepath.Join(git, tc.config))))
			c.checkFile("allowed_signers has one signer line",
				oneSigner(harness.ReadFile(filepath.Join(git, "allowed_signers"))))
		})
	}
	t.Run("both", func(t *testing.T) {
		t.Parallel()
		c := newCase(t, "git: both emails empty")
		c.emptyEmail("email_github", "email_gitlab")
		c.runReal()
		c.checkExit(2)
		c.check(strings.Contains(c.logText(), "the group git.email has no answer in macos.toml"),
			"the message names the group", "the message does not name the group 'git.email'")
	})
}

// driftAllowed are the keys that can differ between the two sides, for
// example `tool:python` or `depends:python`. Each entry needs a reason in a
// comment here.
var driftAllowed = map[string]bool{}

const miseRowsExpr = `(.mise.tools // {} | to_entries | .[] | "tool:" + .key + "=" + (.value.version // "")), (.mise.tools // {} | to_entries | .[] | "depends:" + .key + "=" + ((.value.depends // []) | sort | join(","))), (.mise.settings // {} | .. | select(kind == "scalar") | "setting:" + (path | .[2:] | join(".")) + "=" + (. | tostring))`

// miseRows gives the sorted rows of the mise tools and mise settings of a
// side: `tool:<name>=<version>`, `depends:<name>=<sorted list>` and
// `setting:<path>=<value>`. A mise tool with no `depends` has a row with an
// empty list, because the check compares only the keys of both sides.
func miseRows(t *testing.T, root, side string) []string {
	tools, err := filepath.Glob(filepath.Join(root, side, "tools", "*", "tools.toml"))
	if err != nil {
		t.Fatal(err)
	}
	outs := parallelMap(tools, func(f string) seedRows {
		code, out := tool("yq", "-p", "toml", "-o", "yaml", miseRowsExpr, f)
		return seedRows{code, out}
	})
	set := map[string]bool{}
	for i, o := range outs {
		if o.code != 0 {
			t.Fatalf("yq %s: exit %d\n%s", tools[i], o.code, o.out)
		}
		for _, l := range strings.Split(o.out, "\n") {
			if l != "" {
				set[l] = true
			}
		}
	}
	rows := make([]string, 0, len(set))
	for r := range set {
		rows = append(rows, r)
	}
	sort.Strings(rows)
	return rows
}

// miseDrift gives one line for each key that has two values, and the count
// of the shared keys.
func miseDrift(t *testing.T, root string) ([]string, int) {
	repo := map[string]string{}
	for _, r := range miseRows(t, root, "repo") {
		k, _, _ := strings.Cut(r, "=")
		repo[k] = r
	}
	var drift []string
	shared := 0
	for _, r := range miseRows(t, root, "macos") {
		k, _, _ := strings.Cut(r, "=")
		v, ok := repo[k]
		if !ok {
			continue
		}
		shared++
		if v != r && !driftAllowed[k] {
			drift = append(drift, "repo "+v+", macos "+r)
		}
	}
	return drift, shared
}

// testDrift is case 3: the two sides agree on a shared mise tool and setting.
func testDrift(t *testing.T) {
	t.Parallel()
	drift, shared := miseDrift(t, harness.Root)
	if len(drift) == 0 {
		harness.Pass(t, "drift: the two sides agree on "+strconv.Itoa(shared)+" shared mise versions, depends and settings")
	}
	for _, l := range drift {
		harness.Fail(t, "drift: two values: "+l, "")
	}
}

// testDriftFixture proves that the drift check finds a difference, or the
// pass of the drift case proves nothing.
func testDriftFixture(t *testing.T) {
	t.Parallel()
	fixture := harness.Dir(t)
	harness.WriteFile(t, filepath.Join(fixture, "repo/tools/a/tools.toml"), harness.Lines(
		"[mise.settings]", "a.b = true", `c = "same"`, "[mise.tools.x]", `version = "1"`, `depends = ["p"]`,
		"[mise.tools.y]", `version = "2"`, `depends = ["p", "q"]`))
	harness.WriteFile(t, filepath.Join(fixture, "macos/tools/a/tools.toml"), harness.Lines(
		"[mise.settings]", "a.b = false", `c = "same"`, "[mise.tools.x]", `version = "9"`,
		"[mise.tools.y]", `version = "2"`, `depends = ["q", "p"]`))
	lines, _ := miseDrift(t, fixture)
	drift := strings.Join(lines, "\n")
	ok := regexp.MustCompile(`(?s)depends:x=p.*depends:x=`).MatchString(drift) &&
		regexp.MustCompile(`(?s)tool:x=1.*tool:x=9`).MatchString(drift) &&
		regexp.MustCompile(`(?s)setting:a\.b=true.*setting:a\.b=false`).MatchString(drift) &&
		!strings.Contains(drift, ":y") && !strings.Contains(drift, "setting:c")
	harness.Check(t, "drift fixture: a different version, depends and setting are found, and an equal one is not", ok,
		"wrong result: "+drift)
}
