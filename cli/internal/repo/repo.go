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
	"errors"
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
	manifestPath = values.Folder + "/" + values.ManifestName
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

	values    string // The values folder, absolute.
	show      string // The values folder in a message.
	showRepo  string // Its repo.toml in a message.
	showTools string // Its tools.toml in a message.
	valDS     any    // The values as gomplate reads them.
	valErr    error

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

	home *source.Obj // The [home] of this member that the run writes, or nil.

	// The stages of Check: the stages that ended, the collector of the
	// findings of the stage that runs, or nil, and the lines that a run
	// printed outside a stage.
	stages []Stage
	cur    *fail.Collector
	loose  []Finding
}

// Run runs `bootstrap sync <args>` and gives its exit code. Each line goes to
// <stderr>; <stdout> gets only the output of the commands that bash lets
// through.
func Run(args []string, stdout, stderr io.Writer) fail.Code {
	_, _, code := RunWith(args, stdout, stderr)
	return code
}

// RunWith is Run, and it also gives the change list of a dry run: the
// absolute path of each file that the run would write, change or remove; and
// the outcomes of the run: one for each call of report.Outcome, in order.
func RunWith(args []string, stdout, stderr io.Writer) ([]target.Planned, []report.Result, fail.Code) {
	s := newState(stdout, stderr)
	var outcomes []report.Result
	s.r.ObserveOutcome = func(o report.Result) { outcomes = append(outcomes, o) }
	code := s.trapped(args)
	return s.e.Planned, outcomes, code
}

// Check runs `bootstrap check` as Run does, with the same lines, and gives
// its stages in order, the error and warning lines that it printed outside a
// stage, and its exit code. A failure before the first stage gives no stage.
func Check(stdout, stderr io.Writer) ([]Stage, []Finding, fail.Code) {
	s := newState(stdout, stderr)
	s.r.Observe = s.observe
	code := s.trapped([]string{"--check"})
	return s.stages, s.loose, code
}

// newState gives the state of a run that writes to <stdout> and <stderr>.
func newState(stdout, stderr io.Writer) *state {
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
	s.e = &target.Env{Report: s.r, Stderr: stderr, Pwd: s.pwd, Moving: oldform.Moving}
	return s
}

// trapped runs <args> with the traps of common: on a signal, it removes the
// temp files, then exits 130 or 143.
func (s *state) trapped(args []string) fail.Code {
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
	return s.run(args)
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
	valuesDir := s.getenv("BOOTSTRAP_VALUES")
	if valuesDir == "" {
		valuesDir = values.FolderOf(base)
	}
	valuesDir = strings.TrimSuffix(valuesDir, "/")
	if !exists(valuesDir) {
		s.log("values folder " + s.path(valuesDir) + " not found - run bootstrap init")
		return fail.Usage
	}
	missing := false
	for _, f := range []string{values.RepoFile(valuesDir), values.ToolsFile(valuesDir)} {
		if !isFile(f) {
			s.log(s.path(f) + " not found - run bootstrap init")
			missing = true
		}
	}
	if missing {
		return fail.Usage
	}
	if lookPath("mise", s.getenv("PATH")) == "" {
		s.log("mise is not available - install it from https://mise.run")
		return fail.Usage
	}
	if !strings.HasPrefix(valuesDir, "/") {
		valuesDir = s.pwd + "/" + valuesDir
	}
	s.values = valuesDir
	s.show = s.path(valuesDir)
	s.showRepo = s.path(values.RepoFile(valuesDir))
	s.showTools = s.path(values.ToolsFile(valuesDir))
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
	s.home = s.homeLink()
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
		Values:      values.ToolsFile(s.values),
		ValuesShown: s.showTools,
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
	// shared file. They are in the values folder, so the plan of a
	// migration reads the old manifest from its temp folder (D94).
	s.oldRows, s.oldCode = manifest.Read(s.e, manifest.Repo, values.ManifestFile(s.values), s.root)
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
		if s.writeHome() != fail.OK {
			s.r.Summary(1)
			return fail.Other
		}
		// Only a sync without a failure records its commit: then each file
		// is at that commit, and an agent can compare it with main of
		// bootstrap.
		if commit != "" && s.e.RecordCommit(values.RepoFile(s.values), "bootstrap_commit", commit, commitNote, "") != fail.OK {
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

// ownTasks gives a problem for each own task of the repo at the place of a
// dispatcher (D4, D7).
func (s *state) ownTasks() error {
	return oldform.TaskProblems(s.root, values.ManifestFile(s.values), s.path)
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

// Stage is one stage of `bootstrap check`: the kebab-case name of what it
// examines, its status (pass, fail or skipped) and the error and warning
// lines that it printed.
type Stage struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Findings []Finding `json:"findings"`
}

// Finding is one error or warning line of a check. Code is the code of the
// stage for an error, and 0 for a warning. Path is the file of the values
// folder that the line names, relative to the repository, or "".
type Finding struct {
	Code     fail.Code `json:"code"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	Path     string    `json:"path"`
}

// The statuses of a stage, and the severities of a finding.
const (
	StagePass    = "pass"
	StageFail    = "fail"
	StageSkipped = "skipped"
	SevError     = "error"
	SevWarning   = "warning"
)

// findingErr is a Finding as an error, so that the fail.Collector of a stage
// keeps it.
type findingErr struct{ f Finding }

func (e *findingErr) Error() string { return e.f.Message }

// observe is the Observe of the report of Check: it adds each error and
// warning line to the collector of the stage that runs, else to loose. A
// warning has the code 0, and an error the code 1 until its stage ends.
func (s *state) observe(warn bool, msg string) {
	f := Finding{Severity: SevError, Message: msg, Path: s.findingPath(msg)}
	code := fail.Other
	if warn {
		f.Severity = SevWarning
		code = fail.OK
	}
	if s.cur == nil {
		s.loose = append(s.loose, f)
		return
	}
	s.cur.Add(fail.Wrap(code, &findingErr{f}, ""))
}

// findingPath gives the file of the values folder that <msg> names first,
// relative to the repository, or "".
func (s *state) findingPath(msg string) string {
	if s.values == "" {
		return ""
	}
	best, at := "", -1
	for _, f := range []string{values.RepoFile(s.values), values.ToolsFile(s.values), values.ReleasesFile(s.values)} {
		if i := strings.Index(msg, s.path(f)); i >= 0 && (at < 0 || i < at) {
			best, at = f, i
		}
	}
	base := s.root
	if base == "" {
		base = s.pwd
	}
	return strings.TrimPrefix(best, base+"/")
}

// checkStage runs <f> as the stage <name> of check and gives its code. With
// <run> false, it keeps the stage as skipped and gives 0: its input failed,
// and that failure is already reported.
func (s *state) checkStage(name string, run bool, f func() fail.Code) fail.Code {
	st := Stage{Name: name, Status: StageSkipped, Findings: []Finding{}}
	if !run {
		s.stages = append(s.stages, st)
		return fail.OK
	}
	var c fail.Collector
	s.cur = &c
	code := f()
	s.cur = nil
	st.Status = StagePass
	if code != fail.OK {
		st.Status = StageFail
	}
	// The code of an error finding is the code of its stage.
	if j, ok := c.Err().(interface{ Unwrap() []error }); ok {
		for _, err := range j.Unwrap() {
			var fe *findingErr
			if !errors.As(err, &fe) {
				continue
			}
			if fe.f.Severity == SevError {
				fe.f.Code = code
			}
			st.Findings = append(st.Findings, fe.f)
		}
	}
	s.stages = append(s.stages, st)
	return code
}

// check is `_rs_check`: each check of `repo_sync --check`, in the order of
// the sync. A check whose input failed is skipped: its cause is already
// reported. It gives the highest code, or 0 after a line that says the file
// is valid. Each check is a stage of Check.
func (s *state) check() fail.Code {
	rc := fail.OK
	add := func(code fail.Code) bool {
		rc = max(rc, code)
		return code == fail.OK
	}
	envsOK := add(s.checkStage("envs", true, s.readEnvs))
	var sel []values.Item
	selOK := add(s.checkStage("values", true, func() fail.Code {
		var err error
		sel, err = s.readSel()
		return s.logAll(err)
	}))
	selOK = add(s.checkStage("old-forms", selOK, func() fail.Code {
		// A file in a dispatcher folder that the dispatcher does not run is
		// a warning (D15).
		for _, p := range oldform.UnnamedParts(s.root) {
			s.r.Warn(s.path(s.root+"/"+p) + " has no NN- name, so its dispatcher does not run it - rename it to NN-<name>")
		}
		return s.logAll(errors.Join(oldform.Repo(s.root, s.vals.Raw, s.show), s.ownTasks()))
	})) && selOK
	add(s.checkStage("env-rules", envsOK && selOK, func() fail.Code { return s.envRules(sel) }))
	// The line of core_tools prints as an error but never changes the
	// code: its stage always passes, and the line is a warning finding.
	add(s.checkStage("core-tools", true, func() fail.Code {
		obs := s.r.Observe
		if obs != nil {
			s.r.Observe = func(_ bool, msg string) { obs(true, msg) }
		}
		s.coreTools()
		s.r.Observe = obs
		return fail.OK
	}))
	g := &misegen.Generator{
		Side:        source.Repo,
		ToolsDir:    "/" + toolsShown,
		Shown:       toolsShown,
		Values:      values.ToolsFile(s.values),
		ValuesShown: s.showTools,
		ProjectRoot: s.root,
		Getenv:      s.getenv,
	}
	defer func() { _ = g.Close() }()
	add(s.checkStage("overrides", true, func() fail.Code { return s.overrides(g, true) }))
	add(s.checkStage("members", true, func() fail.Code {
		_, code := s.memberList(strings.TrimPrefix(values.RepoFile(s.values), s.root+"/"))
		return code
	}))
	add(s.checkStage("releases", true, s.releases))
	valid := rc == fail.OK
	if valid {
		s.r.Success(s.show + " is valid")
		s.listChecks = true
	}
	add(s.checkStage("scripts", valid, s.checkScripts))
	return rc
}
