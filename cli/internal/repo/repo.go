// Package repo is the repo front-end of the sync engine: the subcommand
// `bootstrap sync`. It copies the shared files of the selected tools of
// `repo/tools/` of a clone of bootstrap into the repository of
// MISE_PROJECT_ROOT. It replaced the bash engine, with the same output, files
// and exit codes.
//
// Run reads this environment: BOOTSTRAP_CLONE,
// MISE_PROJECT_ROOT, BOOTSTRAP_VALUES,
// BOOTSTRAP_VERBOSE, TMPDIR, PATH, PWD and the style variables FORCE_COLOR,
// NO_COLOR and TERM. A template reads the environment with env.Getenv.
//
// Contract: docs/design/repo.md.
package repo

import (
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/virajp/bootstrap/cli/internal/fail"
	"github.com/virajp/bootstrap/cli/internal/manifest"
	"github.com/virajp/bootstrap/cli/internal/misegen"
	"github.com/virajp/bootstrap/cli/internal/oldform"
	"github.com/virajp/bootstrap/cli/internal/report"
	"github.com/virajp/bootstrap/cli/internal/selection"
	"github.com/virajp/bootstrap/cli/internal/source"
	"github.com/virajp/bootstrap/cli/internal/target"
	"github.com/virajp/bootstrap/cli/internal/values"
)

// Prefix is the name at the start of each log line (_SE_PREFIX).
const Prefix = "bootstrap"

// Usage is the usage line of the repo commands. It is a log line, so it has
// the prefix.
const Usage = "usage: bootstrap sync [--dry-run] | bootstrap check"

// The names of the repo side.
const (
	toolsShown   = "repo/tools"
	manifestPath = ".config/repo-generated.tsv"
	commitNote   = "# The bootstrap commit of the last sync. The sync writes it."
)

// cond is one `requires` or `requires_found` item of a `[files]` table: a
// line of _RS_FILE_CONDS.
type cond struct{ key, kind, item string }

// pend is one target after its first step: a line of _RS_PENDING.
// confDir is the folder of the mise files of a tool in the target.
const confDir = ".config/mise/conf.d"

type pend struct {
	tool string
	f    source.File
	tmp  string
}

// state is the state of one run, as the _RS_* variables of bash.
type state struct {
	stdout, stderr io.Writer
	r              *report.Report
	e              *target.Env
	getenv         func(string) string
	pwd            string
	root           string // MISE_PROJECT_ROOT

	values string // The values file, absolute.
	show   string // The values file in a message.
	valDS  any    // The values file as gomplate reads it.
	valErr error

	dir  string // The clone copy.
	src  source.Source
	idx  *source.Index
	vals *values.Repo
	res  *selection.Resolved
	envs []string // _RS_ENVS

	fileModes   map[string]string   // _RS_FILE_MODES: <tool>/<target> to its mode.
	fileFormats map[string]string   // _RS_FILE_FORMATS
	fileConds   []cond              // _RS_FILE_CONDS
	spare       []string            // _RS_SPARE
	reqSkip     map[string]string   // _RS_REQ_SKIP: <tool>/<target> to its missing item.
	formatter   *source.Formatter   // The formatter of the run, or nil.
	oldRows     []manifest.Row      // The rows of the manifest of the last run.
	oldCode     fail.Code           // The code of the read of the manifest.
	items       map[string][]string // The item rows of each shared file, from oldRows.
	newItems    map[string][]string // The item rows of each shared file of this run.
	generated   map[string]bool     // _RS_GENERATED
	handled     []string            // _RS_HANDLED
	rows        []manifest.Row      // _RS_ROWS
	stage       string              // _RS_STAGE
	pending     []pend              // _RS_PENDING
	all         *source.Obj         // The values of the tools (_RS_VALUES_FILE).
	allJSON     string
	derived     map[string]*source.Obj // The derived mise tables of each tool.
	dirDone     map[string]bool        // The folders at their dir_perm.

	listChecks bool // check: print a line for each check script that passes.

	memberRows  []member
	membersCode fail.Code
	membersRead bool
}

// Run runs `bootstrap sync <args>` and gives its exit code. Each line goes to
// <stderr>; <stdout> gets only the output of the commands that bash lets
// through.
func Run(args []string, stdout, stderr io.Writer) fail.Code {
	_, code := RunWith(args, stdout, stderr)
	return code
}

// RunWith is Run, and it also gives the change list of a dry run: the
// absolute path of each file that the run would write, change or remove.
func RunWith(args []string, stdout, stderr io.Writer) ([]target.Planned, fail.Code) {
	s := &state{stdout: stdout, stderr: stderr, getenv: os.Getenv, derived: map[string]*source.Obj{}, dirDone: map[string]bool{}}
	isTerm := false
	if f, ok := stderr.(*os.File); ok {
		isTerm = report.IsTerminal(f)
	}
	s.r = report.NewReport(stderr, report.Options{
		Prefix:  Prefix,
		Style:   report.StyleFromEnv(s.getenv, isTerm),
		Verbose: report.VerboseFromEnv(s.getenv),
	})
	s.pwd = pwd(s.getenv)
	s.root = s.getenv("MISE_PROJECT_ROOT")
	s.e = &target.Env{Report: s.r, Stderr: stderr, Pwd: s.pwd}

	// The traps of common: remove the temp files, then exit 130 or 143.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan struct{})
	defer func() {
		signal.Stop(sig)
		close(done)
	}()
	go func() {
		select {
		case got := <-sig:
			s.e.Temps.Cleanup()
			if got == syscall.SIGINT {
				os.Exit(130)
			}
			os.Exit(143)
		case <-done:
		}
	}()
	defer s.e.Temps.Cleanup()
	code := s.run(args)
	return s.e.Planned, code
}

// log prints an error message: the problem and, after " - ", its fix.
func (s *state) log(msg string) { s.r.Error(msg) }

// logAll prints each message of <err> and gives its code.
func (s *state) logAll(err error) fail.Code { return source.LogAll(s.r, err) }

// path is _se_path.
func (s *state) path(p string) string { return s.e.Show(p) }

// run is repo_sync.
func (s *state) run(args []string) fail.Code {
	dryRun, check := false, false
	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--check":
			check = true
		default:
			s.r.Log(Usage)
			return fail.Usage
		}
	}
	if dryRun && check {
		s.r.Log(Usage)
		return fail.Usage
	}

	base := s.root
	if base == "" {
		base = s.pwd
	}
	valuesFile := s.getenv("BOOTSTRAP_VALUES")
	if valuesFile == "" {
		valuesFile = base + "/.config/bootstrap.toml"
	}
	// `.config/repo-config.toml` is the old name, from before this
	// repository was renamed to bootstrap.
	old := base + "/.config/repo-config.toml"
	if s.getenv("BOOTSTRAP_VALUES") == "" && !isFile(valuesFile) && isFile(old) {
		s.r.Warn(s.path(old) + " is the old name of the values file - rename it to .config/bootstrap.toml")
		valuesFile = old
	}
	if !isFile(valuesFile) {
		s.log("values file " + s.path(valuesFile) + " not found - it lists the tools to sync")
		return fail.Usage
	}
	if lookPath("mise", s.getenv("PATH")) == "" {
		s.log("mise is not available - install it from https://mise.run")
		return fail.Usage
	}
	if !strings.HasPrefix(valuesFile, "/") {
		valuesFile = s.pwd + "/" + valuesFile
	}
	s.values = valuesFile
	s.show = s.path(valuesFile)
	if check {
		return s.check()
	}
	return s.sync(dryRun)
}

// sync is the sync of repo_sync, after the checks of its arguments.
func (s *state) sync(dryRun bool) fail.Code {
	clone := s.getenv("BOOTSTRAP_CLONE")
	if clone == "" {
		s.log("BOOTSTRAP_CLONE is not set - run mise run setup:configs, which clones bootstrap")
		return fail.Usage
	}
	dir, err := os.MkdirTemp(target.DryRunDir(s.getenv), "bootstrap.")
	if err != nil {
		s.log("cannot create a temp directory")
		return fail.Other
	}
	s.e.Temps.Add(dir)
	// A clone that `setup:configs` already made, so that this code and the
	// files come from the same commit. A private copy, because the sync
	// writes the generated mise config into its clone.
	if !exists(clone+"/.git") || !s.copyClone(clone, dir) {
		s.log("cannot use the clone " + clone + ": nothing synced, existing files were left untouched")
		return fail.Network
	}
	s.dir = dir
	s.src = source.Source{Root: dir, Side: source.Repo}
	rev, _ := command("", s.stderr, "git", "-C", dir, "rev-parse", "--short", "HEAD")
	s.r.Start(dryRun)
	s.e.DryRun = dryRun
	if s.r.Verbose() {
		s.r.Log("using bootstrap at " + rev)
	}
	// The full id, for `bootstrap_commit` after a sync without a failure.
	commit, good := command("", s.stderr, "git", "-C", dir, "rev-parse", "HEAD")
	if !good {
		commit = ""
	}

	cleanup := func(code fail.Code) fail.Code {
		_ = os.RemoveAll(dir)
		return code
	}
	if code := s.readEnvs(); code != fail.OK {
		return cleanup(code)
	}
	idx, err := s.src.Open()
	s.idx = idx
	if err != nil {
		return cleanup(s.logAll(err))
	}
	if code := s.resolve(); code != fail.OK {
		return cleanup(code)
	}
	if code := s.requiredTables(false); code != fail.OK {
		return cleanup(code)
	}
	tools := s.res.Tools
	if s.r.Verbose() {
		s.r.Log("tools: " + strings.Join(tools, " "))
		s.r.Log("environments: " + strings.Join(s.envs, " "))
	} else {
		s.r.Clone("bootstrap " + rev + " · " + strconv.Itoa(len(tools)) + " tools · env " + strings.Join(s.envs, ", "))
	}
	s.fileSkips()
	s.coreTools()

	worst := fail.OK
	var failed []string
	s.generated = map[string]bool{}
	envs, err := misegen.NewEnvs(source.Repo, s.envs, s.res, toolsShown, s.hasFile)
	if err != nil {
		return cleanup(s.logAll(err))
	}
	g := &misegen.Generator{
		Side:        source.Repo,
		ToolsDir:    s.src.ToolsDir(),
		Shown:       toolsShown,
		Values:      s.values,
		ValuesShown: s.show,
		ProjectRoot: s.root,
		Envs:        envs,
		Getenv:      s.getenv,
		Index:       s.idx,
		Tools:       s.res.Tools,
		Ships: func(tool, path string) bool {
			for _, f := range s.idx.Files(tool) {
				if f.Target == path {
					return true
				}
			}
			return false
		},
	}
	defer func() { _ = g.Close() }()
	for _, step := range []func() fail.Code{
		func() fail.Code { return s.overrides(g, false) },
		func() fail.Code { _, code := s.memberList("template context"); return code },
		s.toolValues,
		s.checks,
	} {
		if code := step(); code != fail.OK {
			return cleanup(code)
		}
	}
	g.All = s.all
	g.Derived = s.derived
	// The rows of the last run: the items that bootstrap added to each
	// shared file.
	s.oldRows, s.oldCode = manifest.Read(s.e, manifest.Repo, s.root+"/"+manifestPath, s.root)
	s.items = manifest.Items(s.oldRows)
	s.newItems = map[string][]string{}
	// The files of each tool: the files of its entries, and the mise files
	// that the generator writes into the tool folder of the clone copy.
	toolFiles := map[string][]source.File{}
	for _, tool := range tools {
		var files []source.File
		for _, f := range s.idx.Files(tool) {
			// An entry whose `when` is false is not shipped: the stale rule
			// removes a synced copy.
			if s.when(f) {
				files = append(files, f)
			}
		}
		written, err := g.Render(tool, s.src.ToolDir(tool)+"/"+confDir)
		for _, w := range written {
			s.generated[tool+"/"+confDir+"/"+w] = true
			files = append(files, source.File{Path: confDir + "/" + w, Target: confDir + "/" + w})
		}
		sort.SliceStable(files, func(i, j int) bool { return files[i].Target < files[j].Target })
		toolFiles[tool] = files
		if code := s.logAll(err); code != fail.OK {
			failed = append(failed, toolsShown+"/"+tool+"/tools.toml [mise]")
			worst = max(worst, code)
		}
	}

	// The context is built once, and only when a template needs it. Nil
	// tells file that it could not be built, so each template fails and
	// every plain file still syncs.
	var ctx *contexts
	if s.idx.AnyTemplate() {
		ctx = s.contexts()
		if ctx == nil {
			s.log("cannot build the template context")
		}
	}

	// The stage of the formatter, with a copy of its config files of this
	// repo. Without the command of the formatter, the raw render stays.
	if f := s.formatter; f != nil && run("", "mise", "x", "--", "which", strings.Fields(f.Command)[0]) {
		if stage, err := os.MkdirTemp(target.DryRunDir(s.getenv), "bootstrap-stage."); err == nil {
			s.stage = stage
			s.e.Temps.Add(stage)
			for _, name := range f.Configs {
				if isFile(s.root + "/" + name) {
					p := s.stage + "/" + name
					if os.MkdirAll(target.Dirname(p), 0o777) == nil {
						copyFile(s.root+"/"+name, p)
					}
				}
			}
		}
	}

	ok, total := 0, 0
	for _, tool := range tools {
		for _, f := range toolFiles[tool] {
			total++
			rc := fail.OK
			deferred := false
			out := f.Target
			switch {
			case s.inReqSkip(tool + "/" + out):
				s.skipRequired(tool, out)
			default:
				rc, deferred = s.file(tool, f, ctx)
			}
			if rc != fail.OK {
				s.r.Outcome(report.Failed, s.path(s.root+"/"+out), "cannot write the target", "")
				failed = append(failed, toolsShown+"/"+tool+"/"+f.Path)
				worst = max(worst, rc)
			} else if !deferred {
				ok++
			}
		}
	}

	s.format()
	// The mise files first: a later file can need the tools of the new mise
	// config.
	sort.SliceStable(s.pending, func(i, j int) bool {
		return strings.HasPrefix(s.pending[i].f.Target, confDir+"/") && !strings.HasPrefix(s.pending[j].f.Target, confDir+"/")
	})
	for _, p := range s.pending {
		rc := s.fileFinish(p.tool, p.f, p.tmp)
		if rc == fail.OK {
			ok++
			continue
		}
		s.r.Outcome(report.Failed, s.path(s.root+"/"+p.f.Target), "cannot write the target", "")
		failed = append(failed, toolsShown+"/"+p.tool+"/"+p.f.Path)
		worst = max(worst, rc)
	}
	if s.stage != "" {
		_ = os.RemoveAll(s.stage)
	}
	_ = os.RemoveAll(dir)

	// Only a sync without a failure removes stale files and writes the
	// manifest, so a failure here leaves the old manifest for the next clean
	// run.
	if len(failed) == 0 {
		file := s.root + "/" + manifestPath
		step := manifestPath
		oldRows, rc := s.oldRows, s.oldCode
		if rc == fail.OK {
			step = "stale files"
			rc = manifest.Stale(s.e, manifest.Repo, oldRows, s.handled, s.root)
		}
		if rc == fail.OK {
			step = manifestPath
			rc = manifest.Write(s.e, file, append(s.rows, s.keptItems()...))
		}
		if rc != fail.OK {
			failed = append(failed, step)
			worst = max(worst, rc)
		}
	}

	if len(failed) == 0 {
		// Only a sync without a failure records its commit: then each file
		// is at that commit, and an agent can compare it with main of
		// bootstrap.
		if commit != "" && s.e.RecordCommit(s.values, "bootstrap_commit", commit, commitNote, "") != fail.OK {
			s.r.Summary(1)
			return fail.Other
		}
		if s.r.Verbose() {
			s.r.Log("done: " + strconv.Itoa(ok) + "/" + strconv.Itoa(total) + " files in sync")
		}
		s.r.Summary(0)
		return fail.OK
	}
	s.r.Info("FAILED: " + strings.Join(failed, " "))
	if s.r.Verbose() {
		s.r.Log(strconv.Itoa(ok) + "/" + strconv.Itoa(total) + " files in sync; existing copies of the rest were left untouched")
	}
	s.r.Summary(int(worst))
	return worst
}

// keptItems gives the item rows of the last run for each shared file that
// this run did not merge, so a later run can still remove those items.
func (s *state) keptItems() []manifest.Row {
	var out []manifest.Row
	for _, r := range s.oldRows {
		if _, merged := s.newItems[r.Path]; r.Item != "" && !merged {
			out = append(out, r)
		}
	}
	return out
}

// copyClone is `cp -R -- <clone>/. <dir>`. The error lines of cp go to
// stderr, as in bash.
func (s *state) copyClone(clone, dir string) bool {
	_, ok := command("", s.stderr, "cp", "-R", "--", clone+"/.", dir)
	return ok
}

// check is `_rs_check`: each check of `repo_sync --check`, in the order of
// the sync. A check whose input failed is skipped: its cause is already
// reported. It gives the highest code, or 0 after a line that says the file
// is valid.
func (s *state) check() fail.Code {
	rc := fail.OK
	add := func(code fail.Code) bool {
		rc = max(rc, code)
		return code == fail.OK
	}
	envsOK := add(s.readEnvs())
	sel, err := s.readSel()
	selOK := add(s.logAll(err))
	if selOK {
		selOK = add(s.logAll(oldform.Repo(s.root, s.vals.Raw, s.show)))
	}
	if envsOK && selOK {
		add(s.envRules(sel))
	}
	s.coreTools()
	g := &misegen.Generator{
		Side:        source.Repo,
		ToolsDir:    "/" + toolsShown,
		Shown:       toolsShown,
		Values:      s.values,
		ValuesShown: s.show,
		ProjectRoot: s.root,
		Getenv:      s.getenv,
	}
	defer func() { _ = g.Close() }()
	add(s.overrides(g, true))
	_, code := s.memberList(strings.TrimPrefix(s.values, s.root+"/"))
	add(code)
	if rc == fail.OK {
		s.r.Success(s.show + " is valid")
		s.listChecks = true
		add(s.checkScripts())
	}
	return rc
}
