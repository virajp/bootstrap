// Package macos is the Mac front-end of the sync engine: the subcommand
// `sync-engine macos`. It renders the selected tools of `macos/tools/` of a
// clone of bootstrap into $HOME, then runs `mise bootstrap`. It replaces the
// bash file sync-engine/setup-sync, with the same output, files and exit
// codes.
//
// Run reads the same environment as `setup_sync`: SYNC_ENGINE_CLONE,
// SYNC_ENGINE_CACHE, MACOS_SETUP_CONFIG_PATH, SYNC_ENGINE_VERBOSE, HOME, USER,
// TMPDIR, XDG_CACHE_HOME, PATH, PWD and the style variables FORCE_COLOR,
// NO_COLOR and TERM.
//
// Contract: docs/design/setup-sync.md.
package macos

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/virajp/bootstrap/sync-engine/internal/fail"
	"github.com/virajp/bootstrap/sync-engine/internal/manifest"
	"github.com/virajp/bootstrap/sync-engine/internal/misegen"
	"github.com/virajp/bootstrap/sync-engine/internal/report"
	"github.com/virajp/bootstrap/sync-engine/internal/selection"
	"github.com/virajp/bootstrap/sync-engine/internal/source"
	"github.com/virajp/bootstrap/sync-engine/internal/target"
	"github.com/virajp/bootstrap/sync-engine/internal/values"
)

// Prefix is the name at the start of each log line (_SE_PREFIX).
const Prefix = "setup-sync"

// The names of the Mac side in a message.
const (
	valuesName = "macos.toml"
	tilde      = "~"
	confTarget = ".config/mise/conf.d"
	// ownTool is the tool name of the own config of macos.toml.
	ownTool = "zz-macos"
	// valuesHeader is the first line of a new macos.toml.
	valuesHeader = "# The values of the Mac sync of bootstrap. See docs/setup/index.md."
)

// Usage is the help text of `_ss_usage`.
const Usage = `usage: setup_sync [--dry-run] [--no-bootstrap]

  --dry-run       show the changes without writing them
  --no-bootstrap  skip every mise bootstrap step
  -h, --help      show this help

exit codes: 0 done, 1 other failure, 2 usage or config error, 3 a tool or a
pack that macos/ does not have, 4 the init stage changed (open a new terminal,
then run mise run updateall:configs)
`

// genFile is one generated mise file: a line of _SS_GEN.
type genFile struct{ tool, file, target string }

// badTool is a tool whose mise config cannot generate: a line of _SS_MISE_BAD.
type badTool struct {
	tool string
	code fail.Code
}

// planItem is one file of the plan: a line of _SS_PLAN.
type planItem struct {
	tool   string
	path   string // Relative to the tool folder, or absolute for a generated file.
	target string // Relative to $HOME.
	mode   string // generated, seed or merge.
	format string // The `format` key, or "-".
}

// state is the state of one run, as the _SS_* variables of bash.
type state struct {
	stdout, stderr io.Writer
	r              *report.Report
	e              *target.Env
	getenv         func(string) string
	pwd            string

	dir      string // _SS_DIR: the clone.
	src      source.Source
	config   string // _SS_CONFIG
	home     string // _SS_HOME
	homeReal string // _SS_HOME_REAL
	work     string // _SS_WORK

	values  *values.Macos
	valFile string // _SS_VALUES: empty when there is no macos.toml.
	idx     *source.Index
	res     *selection.Resolved
	all     *source.Obj // _SS_ALL_VALUES
	ctx     map[string]any
	gen     []genFile
	miseBad []badTool
	plan    []planItem
	handled []string // _SS_HANDLED

	ok, total int
	failed    string
	rows      []manifest.Row
	oldRows   []manifest.Row
	changed   bool
	sshDone   bool
	show      string // _SS_SHOW
}

// Run runs `setup_sync <args>` and gives its exit code. The usage of -h goes
// to <stdout>; each other line goes to <stderr>.
func Run(args []string, stdout, stderr io.Writer) fail.Code {
	s := &state{stdout: stdout, stderr: stderr, getenv: os.Getenv}
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
	return s.run(args)
}

func (s *state) log(format string, args ...any) { s.r.Log(fmt.Sprintf(format, args...)) }

// logAll prints each message of <err> and gives its code.
func (s *state) logAll(err error) fail.Code { return source.LogAll(s.r, err) }

// run is setup_sync.
func (s *state) run(args []string) fail.Code {
	dryRun, bootstrap := false, true
	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--no-bootstrap":
			bootstrap = false
		case "-h", "--help":
			fmt.Fprint(s.stdout, Usage)
			return fail.OK
		default:
			s.log("unknown flag '%s'", arg)
			fmt.Fprint(s.stderr, Usage)
			return fail.Usage
		}
	}

	if code := s.checkMise(); code != fail.OK {
		return code
	}
	clone := s.getenv("SYNC_ENGINE_CLONE")
	if clone == "" || !isDir(clone+"/sync-engine") || !isDir(clone+"/macos/tools") {
		s.log("SYNC_ENGINE_CLONE is not a clone of bootstrap - run mise run updateall:configs")
		return fail.Usage
	}
	cfg := s.getenv("MACOS_SETUP_CONFIG_PATH")
	if cfg == "" {
		s.log("MACOS_SETUP_CONFIG_PATH is not set - set it to the folder of macos.toml")
		return fail.Usage
	}
	if !isDir(cfg) {
		s.log("the config folder %s does not exist - set MACOS_SETUP_CONFIG_PATH to the folder of macos.toml", cfg)
		return fail.Usage
	}

	// Each path is absolute, as for the tool wrappers of bash.
	s.dir = strings.TrimSuffix(s.abs(clone), "/")
	s.src = source.Source{Root: s.dir, Side: source.Macos}
	s.config = strings.TrimSuffix(s.abs(cfg), "/")
	s.home = strings.TrimSuffix(s.getenv("HOME"), "/")
	real, ok := manifest.Real(s.home)
	if !ok {
		s.log("cannot resolve %s - a symlink loop", s.home)
		return fail.Other
	}
	s.homeReal = real
	cache := s.getenv("SYNC_ENGINE_CACHE")
	if cache == "" {
		xdg := s.getenv("XDG_CACHE_HOME")
		if xdg == "" {
			xdg = s.getenv("HOME") + "/.cache"
		}
		cache = xdg + "/sync-engine"
	}
	manifestFile := cache + "/macos-generated.tsv"
	work, err := os.MkdirTemp(target.DryRunDir(s.getenv), "setup-sync.")
	if err != nil {
		s.log("cannot create a temp folder")
		return fail.Other
	}
	s.work = work
	s.e.Temps.Add(work)

	if code := s.readValues(); code != fail.OK {
		return code
	}
	idx, err := s.src.Open()
	if err != nil {
		return s.logAll(err)
	}
	s.idx = idx
	res, err := selection.ResolveMacos(idx, s.values)
	if err != nil {
		return s.logAll(err)
	}
	s.res = res
	if code := s.toolValues(); code != fail.OK {
		return code
	}
	g := &misegen.Generator{
		Side:        source.Macos,
		ToolsDir:    s.src.ToolsDir(),
		Shown:       s.src.ToolsShown(),
		Values:      s.valFile,
		ValuesShown: s.e.Show(s.valFile),
		All:         s.all,
		Getenv:      s.getenv,
	}
	defer g.Close()
	if s.valFile != "" {
		if err := g.Overrides(s.res.Tools, nil, false); err != nil {
			return s.logAll(err)
		}
	}
	s.miseGenerate(g)
	if code := s.makePlan(); code != fail.OK {
		return code
	}

	commit := gitHead(s.dir)
	s.r.Start(dryRun)
	s.e.DryRun = false
	if s.r.Verbose() {
		shown := commit
		if shown == "" {
			shown = "an unknown commit"
		}
		s.log("using bootstrap at %s", shown)
		s.log("tools: %s", strings.Join(s.res.Tools, " "))
	} else {
		short := commit
		if len(short) > 7 {
			short = short[:7]
		}
		if short == "" {
			short = "an unknown commit"
		}
		s.r.Info("bootstrap " + short + " · " + strconv.Itoa(len(s.res.Tools)) + " tools")
	}
	s.oldRows, _ = manifest.Read(s.e, manifest.Macos, manifestFile, "")

	if dryRun {
		s.e.DryRun = true
		worst := s.render("all")
		s.stale("all")
		if worst != fail.OK {
			s.log("FAILED:%s", s.failed)
			s.r.Summary(int(worst))
			return worst
		}
		if commit != "" {
			s.recordCommit(commit)
		}
		if s.r.Verbose() {
			s.log("dry run: %d/%d files checked, nothing written", s.ok, s.total)
		}
		s.r.Summary(0)
		return fail.OK
	}

	// The init stage: skipped when no init-stage tool is selected.
	if len(s.res.InitOrder) > 0 {
		before := s.loginShell()
		worst := s.render("init")
		if rc := s.stale("init"); rc != fail.OK {
			s.failed += " stale files"
			worst = max(worst, rc)
		}
		if worst != fail.OK {
			s.reportFailed()
			s.r.Summary(int(worst))
			return worst
		}
		if bootstrap && s.bootstrap() != fail.OK {
			s.r.Summary(1)
			return fail.Other
		}
		if after := s.loginShell(); before != after {
			s.changed = true
		}
		if s.changed {
			rows := append(s.oldRowsOf("main"), s.rows...)
			if manifest.Write(s.e, manifestFile, rows) != fail.OK {
				s.r.Summary(1)
				return fail.Other
			}
			s.log("the init stage changed")
			s.log("open a new terminal, then run: mise run updateall:configs")
			return fail.Terminal
		}
	}

	worst := s.render("main")
	if rc := s.stale("main"); rc != fail.OK {
		s.failed += " stale files"
		worst = max(worst, rc)
	}
	if worst != fail.OK {
		s.reportFailed()
		s.r.Summary(int(worst))
		return worst
	}
	if manifest.Write(s.e, manifestFile, s.rows) != fail.OK {
		s.r.Summary(1)
		return fail.Other
	}
	if bootstrap && s.bootstrap() != fail.OK {
		s.r.Summary(1)
		return fail.Other
	}
	if commit != "" && s.recordCommit(commit) != fail.OK {
		s.r.Summary(1)
		return fail.Other
	}
	if s.r.Verbose() {
		s.log("done: %d/%d files in sync", s.ok, s.total)
	}
	s.r.Summary(0)
	return fail.OK
}

// abs is the path <p> made absolute with $PWD, as the bash `$PWD/$p`.
func (s *state) abs(p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	return s.pwd + "/" + p
}

// readValues is `_ss_read_values`.
func (s *state) readValues() fail.Code {
	file := s.config + "/" + valuesName
	v, err := values.ReadMacos(file, s.e.Show(file))
	if err != nil {
		return s.logAll(err)
	}
	s.values = v
	if v.Exists {
		s.valFile = file
	}
	return fail.OK
}

// toolValues is `_ss_tool_values`: the values of each selected tool.
func (s *state) toolValues() fail.Code {
	m, err := values.Merge(values.Input{
		Index:       s.idx,
		Selected:    s.res.Tools,
		Keep:        s.res.Keep,
		User:        s.values.User,
		ShownRoot:   s.src.ToolsShown(),
		ShownValues: valuesName,
	})
	if err != nil {
		return s.logAll(err)
	}
	s.all = m.All()
	s.ctx = map[string]any{}
	return fail.OK
}

// miseGenerate is `_ss_mise_generate`: the mise config of each selected tool
// and the own config, into the work folder.
func (s *state) miseGenerate(g *misegen.Generator) {
	confd := s.work + "/conf.d"
	add := func(tool string, written []string, err error) {
		for _, w := range written {
			s.gen = append(s.gen, genFile{tool, confd + "/" + w, confTarget + "/" + w})
		}
		if err != nil {
			if code := s.logAll(err); code != fail.OK {
				s.miseBad = append(s.miseBad, badTool{tool, code})
			}
		}
	}
	for _, tool := range s.res.Tools {
		w, err := g.Render(tool, confd, nil)
		add(tool, w, err)
	}
	if s.valFile == "" {
		return
	}
	w, err := g.OwnConfig(confd, s.res.Tools)
	add(ownTool, w, err)
}

// inStage is `_ss_in_stage`: <tool> is in the stage init, main or all.
func (s *state) inStage(stage, tool string) bool {
	switch stage {
	case "init":
		return s.res.Init[tool]
	case "main":
		return !s.res.Init[tool]
	}
	return true
}

// render is `_ss_render`: it syncs each file of <stage> and gives the highest
// code.
func (s *state) render(stage string) fail.Code {
	worst := fail.OK
	for _, b := range s.miseBad {
		if !s.inStage(stage, b.tool) {
			continue
		}
		if b.tool == ownTool {
			s.failed += " " + valuesName + "[mise]"
		} else {
			s.failed += " " + s.src.ToolsShown() + "/" + b.tool + "/tools.toml[mise]"
		}
		worst = max(worst, b.code)
	}
	for _, it := range s.plan {
		if !s.inStage(stage, it.tool) {
			continue
		}
		s.total++
		rc := s.file(it)
		if rc == fail.OK {
			s.ok++
			continue
		}
		s.r.Outcome(report.Failed, s.show, "cannot write the target", "")
		if strings.HasPrefix(it.path, "/") {
			s.failed += " " + tilde + "/" + it.target
		} else {
			s.failed += " " + s.src.ToolsShown() + "/" + it.tool + "/" + it.path
		}
		worst = max(worst, rc)
	}
	return worst
}

// oldRowsOf is `_ss_old_rows`: the old rows whose tool is in <stage>.
func (s *state) oldRowsOf(stage string) []manifest.Row {
	var out []manifest.Row
	for _, row := range s.oldRows {
		if s.inStage(stage, row.Tool) {
			out = append(out, row)
		}
	}
	return out
}

// stale is `_ss_stale`: the stale rule for the old rows of <stage>. A removed
// file sets the change flag.
func (s *state) stale(stage string) fail.Code {
	rows := s.oldRowsOf(stage)
	var before []string
	for _, row := range rows {
		real, ok := manifest.Real(s.home + "/" + row.Path)
		if ok && hasHeader(real) {
			before = append(before, real)
		}
	}
	rc := manifest.Stale(s.e, manifest.Macos, rows, s.handled, s.home)
	for _, f := range before {
		if !exists(f) {
			s.changed = true
		}
	}
	return rc
}

// recordCommit is `_ss_record_commit`.
func (s *state) recordCommit(commit string) fail.Code {
	return s.e.RecordCommit(s.config+"/"+valuesName, "bootstrap_commit", commit, "", valuesHeader)
}

// reportFailed is `_ss_report_failed`.
func (s *state) reportFailed() {
	s.log("FAILED:%s", s.failed)
	if s.r.Verbose() {
		s.log("%d/%d files in sync; existing copies of the rest were left untouched", s.ok, s.total)
	}
}
