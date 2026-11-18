package mac

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// ---- the commands init and questions ----------------------------------------

// noToken removes each line with "token" from macos.toml, and commits it.
func (c *mcase) noToken() {
	var b strings.Builder
	for _, l := range lines(c.values) {
		if !strings.Contains(l, "token") {
			b.WriteString(l + "\n")
		}
	}
	c.write(c.values, b.String())
	c.commitConfig()
}

func caseQuestionMissing(c *mcase) {
	c.useConfig("contrib")
	c.noToken()
	c.runDry()
	c.exit(2)
	c.has("the question owner.token has no answer in macos.toml - run bootstrap mac migrate")
}

func caseQuestionsJSON(c *mcase) {
	c.useConfig("contrib")
	c.noToken()
	c.mac("questions", "--json")
	c.exit(0)
	c.has(`{"contract":1,"command":"mac questions","ok":true,"code":0,"data":[{`)
	c.has(`"tool":"owner"`)
	c.has(`"key":"token"`)
	c.has(`"question":"Which token does the owner use?"`)
	c.has(`"proposal":"token"`)
	c.has(`"source":"detect"`)
	c.has(`"answer":null`)
	c.useConfig("contrib")
	c.mac("questions", "owner")
	c.exit(0)
	c.has("token   Which token does the owner use?")
	c.has("[OK] token-1")
}

// caseMacJSONRefused: on the Mac side only mac questions takes --json. Each
// other --json is a usage error, with the envelope.
func caseMacJSONRefused(c *mcase) {
	c.useConfig("ok")
	c.mac("add", "tool", "--json", "--no-bootstrap", "merger")
	c.exit(2)
	c.has(`{"contract":1,"command":"mac add","ok":false,"code":2,`)
	c.has(`"message":"mac add has no flag --json"`)
	c.mac("check", "--json")
	c.exit(2)
	c.has(`{"contract":1,"command":"mac check","ok":false,"code":2,`)
	c.has(`"message":"mac check has no flag --json"`)
	c.mac("describe", "--json")
	c.exit(2)
	c.has(`{"contract":1,"command":"","ok":false,"code":2,`)
	c.has(`"message":"mac describe is not a command"`)
}

func caseInitAnswer(c *mcase) {
	c.useConfig("contrib")
	c.noToken()
	c.mac("init", "--yes", "--no-bootstrap", "--answer", "owner.token=abc")
	c.exit(0)
	c.check("the answer is in macos.toml", hasLine(c.values, `token = "abc"`))
	c.check("init synced the files", isFile(c.home+"/.config/owner/data.json"))
	c.check("init set bootstrap_commit", hasStr(c.values, commitLine()))
	c.runDry()
	c.exit(0)
}

func caseInitYesMissing(c *mcase) {
	c.useConfig("contrib")
	c.noToken()
	c.mac("init", "--yes", "--no-bootstrap")
	c.exit(2)
	c.has("the question owner.token has no answer - give it with --answer owner.token=<value>")
	c.check("nothing written", !exists(c.home+"/.config"))
}

func caseInitNoYes(c *mcase) {
	c.useConfig("contrib")
	c.mac("init", "--no-bootstrap")
	c.exit(2)
	c.has("changes need --yes")
	c.check("nothing written", !exists(c.home+"/.config"))
}

func caseInitDryRun(c *mcase) {
	c.useConfig("contrib")
	c.mac("init", "--dry-run")
	c.exit(0)
	c.has("new     " + c.home + "/.config/owner/data.json")
	c.has("changes, nothing written")
	c.check("nothing written", !exists(c.home+"/.config"))
}

// oldMacValues: a committed macos.toml with the old [tools.macos-setup].
func (c *mcase) oldMacValues() {
	c.useConfig("-")
	c.write(c.values, harness.Lines("# My Mac.", "", "[tools.macos-setup]", "# The repository of macos.toml.",
		`config_path = "~/macos-setup"`, "", "[tools.opt]"))
	c.git("-C", c.cfg, "add", "-A")
	c.git("-C", c.cfg, "commit", "--quiet", "--no-verify", "-m", "old")
}

func caseSyncOldForm(c *mcase) {
	c.oldMacValues()
	// mac sync runs mac migrate first.
	c.runDry()
	c.exit(0)
	c.has("bootstrap mac migrate has a change to make: it runs before the sync")
}

func caseInitMigrate(c *mcase) {
	c.oldMacValues()
	c.mac("init", "--yes", "--no-bootstrap")
	c.exit(0)
	c.has("[tools.macos-setup] to [tools.mise]")
	c.check("no [tools.macos-setup]", !hasStr(c.values, "macos-setup]"))
	c.check("config_path in [tools.mise]", hasLine(c.values, `config_path = "~/macos-setup"`))
	c.check("the [tools.mise] table", hasLine(c.values, "[tools.mise]"))
	c.check("the header comment stays", hasLine(c.values, "# My Mac."))
	c.check("init synced the files", isFile(c.home+"/.config/opt/opt.txt"))
	c.git("-C", c.cfg, "add", "-A")
	c.git("-C", c.cfg, "commit", "--quiet", "--no-verify", "-m", "migrated")
	c.runReal()
	c.exit(0)
}

func caseInitDirtyValues(c *mcase) {
	c.oldMacValues()
	harness.AppendFile(c.t, c.values, "# a change that is not committed\n")
	c.mac("init", "--yes", "--no-bootstrap")
	c.exit(2)
	c.has("macos.toml has changes that are not committed")
	c.check("the old form stays", hasStr(c.values, "[tools.macos-setup]"))
}

// caseInitNew: a new macos.toml has format = 2 and a table for each member
// of the pack base.
func caseInitNew(c *mcase) {
	c.useConfig("-")
	c.caseTree(migrateTree)
	c.mac("init", "--yes", "--no-bootstrap")
	c.exit(0)
	c.check("the header line", hasLine(c.values, "# The values of the Mac sync of bootstrap. See docs/setup/index.md."))
	c.check("format = 2", hasLine(c.values, "format = 2"))
	c.check("core is selected", hasLine(c.values, "[tools.core]"))
	c.check("mise is selected", hasLine(c.values, "[tools.mise]"))
	c.check("merger, a member of base that is not required, is selected", hasLine(c.values, "[tools.merger]"))
	c.check("opt is not selected", !hasStr(c.values, "[tools.opt]"))
	c.check("init synced the files", isFile(c.home+"/.config/core/values.txt"))
}

func caseInitNames(c *mcase) {
	c.useConfig("-")
	c.mac("init", "--yes", "--no-bootstrap", "opt")
	c.exit(2)
	c.has("mac migrate takes no tool names - use: bootstrap mac add tool opt")
	c.check("no macos.toml", !exists(c.values))
}

func caseInitNotGit(c *mcase) {
	c.useConfig("-")
	c.remove(c.cfg + "/.git")
	c.mac("init", "--yes", "--no-bootstrap")
	c.exit(2)
	c.has("is not a git repository - create one with git init, then run bootstrap mac migrate again")
}

func caseInitNothing(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.git("-C", c.cfg, "add", "-A")
	c.git("-C", c.cfg, "commit", "--quiet", "--no-verify", "-m", "synced")
	c.mac("init", "--no-bootstrap")
	c.exit(0)
	c.has("Nothing to change")
}

// ---- the selection commands of the Mac side ----------------------------------

// fakePackages: fake brew, mas and mise commands for the package scan of
// mac detect. The fake mise is at .local/bin/mise, and it runs the real mise
// for each command but ls. With <nobrew>, PATH has no brew and no mas.
func (c *mcase) fakePackages(nobrew bool) {
	fake := c.dir + "/fake/.local/bin"
	c.write(fake+"/mise", harness.Lines("#!/bin/sh",
		`if [ "$1" = ls ]; then echo '{"pkgmise": [], "loosetool": []}'; exit 0; fi`,
		"exec '"+harness.RealMise+`' "$@"`))
	c.chmod(fake+"/mise", 0o755)
	if nobrew {
		c.setEnv("PATH=" + fake + ":/usr/bin:/bin")
		return
	}
	c.write(fake+"/brew", harness.Lines("#!/bin/sh",
		`case "$1 $2" in leaves*) printf "pkgfmla\nloosefmla\n" ;;`,
		`"list --formula") printf "pkgfmla\nloosefmla\ndepfmla\n" ;; "list --cask") echo loosecask ;; esac`))
	c.write(fake+"/mas", harness.Lines("#!/bin/sh", `echo "123  Some App  (1.0)"`))
	c.chmod(fake+"/brew", 0o755)
	c.chmod(fake+"/mas", 0o755)
	c.prependPath(fake)
}

func caseMacDetect(c *mcase) {
	c.useConfig("ok")
	c.fakePackages(false)
	c.mac("detect", "--no-bootstrap")
	c.exit(1)
	c.has("pkgtool is found, but it is not selected")
	c.has("brew:pkgfmla, +1 more: installed")
	c.has("brew:loosefmla is installed, but no Mac tool has it - add it to [mise.bootstrap.packages] of macos.toml")
	c.has("brew-cask:loosecask is installed, but no Mac tool has it")
	c.has("mas:123 is installed, but no Mac tool has it")
	c.has("loosetool is installed, but no Mac tool has it - add it to [mise.tools] of macos.toml")
	c.lacks("pkgmise is installed")
	c.lacks("brew:pkgfmla is installed")
	c.lacks("brew:depfmla")
}

func caseMacDetectYes(c *mcase) {
	c.useConfig("ok")
	c.fakePackages(false)
	c.mac("detect", "--add", "--yes", "--no-bootstrap")
	c.exit(0)
	c.check("macos.toml selects pkgtool", hasLine(c.values, "[tools.pkgtool]"))
	c.check("macos.toml selects appcmd for the mas: app", hasLine(c.values, "[tools.appcmd]"))
	c.check("macos.toml has the formula", hasLine(c.values, `"brew:loosefmla" = "latest"`))
	c.check("macos.toml has the cask", hasLine(c.values, `"brew-cask:loosecask" = "latest"`))
	c.check("macos.toml has the app", hasLine(c.values, `"mas:123" = "latest"`))
	c.check("macos.toml has the mise tool", hasLine(c.values, "[mise.tools.loosetool]"))
	c.commitConfig()
	c.mac("detect", "--no-bootstrap")
	c.exit(0)
}

func caseMacDetectNobrew(c *mcase) {
	c.useConfig("ok")
	c.fakePackages(true)
	c.mac("detect", "--no-bootstrap")
	c.exit(1)
	c.has("brew is not installed, so Homebrew packages are not scanned")
	c.has("mas is not installed, so App Store apps are not scanned")
	c.has("loosetool is installed, but no Mac tool has it")
	c.lacks("brew:loosefmla")
}

func caseMacDetectNotSelectable(c *mcase) {
	c.useConfig("ok")
	c.caseTree(prepend("pkgtool", "selectable = false"))
	c.fakePackages(false)
	c.mac("detect", "--add", "--yes", "--dry-run", "--no-bootstrap")
	c.exit(0)
	c.lacks("select pkgtool")
	c.lacks("[NEW] pkgtool")
	c.check("macos.toml does not select pkgtool", !hasStr(c.values, "[tools.pkgtool]"))
}

// caseMacDetectFlags: --no is gone, --yes needs --add.
func caseMacDetectFlags(c *mcase) {
	c.useConfig("ok")
	c.mac("detect", "--no", "--no-bootstrap")
	c.exit(2)
	c.has("mac detect has no flag --no")
	c.mac("detect", "--yes", "--no-bootstrap")
	c.exit(2)
	c.has("mac detect --yes needs --add - add --add")
}

func caseMacAdd(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.commitConfig()
	c.mac("add", "tool", "--yes", "--no-bootstrap", "merger")
	c.exit(0)
	c.has("select merger")
	c.check("merger is selected", hasLine(c.values, "[tools.merger]"))
}

func caseMacRemove(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.check("opt synced", isFile(c.home+"/.config/opt/opt.txt"))
	c.commitConfig()
	c.mac("remove", "tool", "--yes", "--no-bootstrap", "opt")
	c.exit(0)
	c.check("opt is not selected", !hasStr(c.values, "[tools.opt]"))
	c.check("the file of opt is gone", !exists(c.home+"/.config/opt/opt.txt"))
}

func caseMacRemoveRefused(c *mcase) {
	c.useConfig("ok")
	c.mac("remove", "tool", "--yes", "--no-bootstrap", "mise")
	c.exit(2)
	c.has("mise is required by the base pack, so mac remove cannot remove it")
	c.mac("remove", "tool", "--yes", "--no-bootstrap", "merger")
	c.exit(2)
	c.has("merger is not selected")
	c.useConfig("dep")
	c.mac("remove", "tool", "--yes", "--no-bootstrap", "initdep")
	c.exit(2)
	c.has("initdep is needed by init - remove init first")
	c.check("the table of initdep stays", hasLine(c.values, "[tools.initdep]"))
}

// caseMacOldGrammar: the grammar without a kind word stops with code 2 and
// the hint of the new form.
func caseMacOldGrammar(c *mcase) {
	c.useConfig("ok")
	c.mac("add", "--yes", "--no-bootstrap", "merger")
	c.exit(2)
	c.has("mac add needs the kind word tool or pack before the names - use: bootstrap mac add tool merger")
	c.mac("remove", "--yes", "--no-bootstrap", "opt")
	c.exit(2)
	c.has("mac remove needs the kind word tool before the names - use: bootstrap mac remove tool opt")
	c.mac("replace", "--yes", "--no-bootstrap", "opt", "merger")
	c.exit(2)
	c.has("mac replace needs the kind word tool before the names - use: bootstrap mac replace tool opt merger")
	c.check("macos.toml is unchanged", cmpFiles(fixtures+"/config/ok/macos.toml", c.values))
}

// caseMacAddFails: a sync that fails after add tool puts macos.toml back.
func caseMacAddFails(c *mcase) {
	c.useConfig("mise-user")
	before := catT(c.values)
	c.touch(c.home + "/.after-fail")
	c.mac("add", "tool", "--yes", "--no-bootstrap", "scripted")
	c.exit(2)
	c.has("scripted/scripts/after failed with exit code 1")
	c.has("Put back macos.toml as it was before the command")
	c.check("macos.toml is as it was", catT(c.values) == before)
}

func caseMacReplaceRefused(c *mcase) {
	c.useConfig("ok")
	c.mac("replace", "tool", "--yes", "--no-bootstrap", "opt", "merger")
	c.exit(2)
	c.has("mac replace works only between two tools with the same parent")
}

// ---- the entry task updateall:configs ---------------------------------------

// entrySetup is entry_setup: the origin, and an empty macos.toml.
func (c *mcase) entrySetup() {
	c.setEnv("BOOTSTRAP_URL="+entryURL, "_UPDATEALL_CONFIGS_RUNNING", "BOOTSTRAP_CLONE")
	c.repo = c.cache + "/repo"
	c.lock = c.cache + "/lock"
	c.oldCache = c.xdgCache + "/macos-setup"
	c.valuesFile()
}

// entryRun runs the entry task with bash, from the case folder.
func (c *mcase) entryRun(args ...string) {
	c.execIn(c.dir, c.env, append([]string{"bash", entryTask}, args...)...)
}
func (c *mcase) entryReal() { c.entryRun("--no-bootstrap") }
func (c *mcase) entryDry()  { c.entryRun("--dry-run", "--no-bootstrap") }

// entryMise runs the same task under `mise run`, with the tasks of the tool
// in the temporary $XDG_CONFIG_HOME/mise/tasks/.
func (c *mcase) entryMise(args ...string) {
	c.execIn(c.dir, c.env, append([]string{"mise", "run", "updateall:configs"}, args...)...)
}

// entryOrigin is entry_origin: an origin of this case only, a commit of the
// entry tree after <change> changed it in c.work.
func (c *mcase) entryOrigin(change func(c *mcase)) {
	c.work = c.dir + "/src"
	c.git("clone", "--quiet", "--", entrySrc, c.work)
	if change != nil {
		change(c)
	}
	c.git("-C", c.work, "add", "-A")
	c.git("-C", c.work, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "case")
	c.git("clone", "--quiet", "--bare", "--", c.work, c.dir+"/origin.git")
	c.setEnv("BOOTSTRAP_URL=file://" + c.dir + "/origin.git")
}

// noLeftovers: no lock, and no bootstrap-* temporary copy in TMPDIR.
func (c *mcase) noLeftovers() bool {
	if exists(c.lock) {
		return false
	}
	ents, _ := os.ReadDir(c.tmpdir)
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), "bootstrap-") {
			return false
		}
	}
	return true
}

// startHolder starts `sleep 60`, a live process that holds a lock. The test
// kills it at its end.
func (c *mcase) startHolder() *exec.Cmd {
	h := exec.Command("sleep", "60")
	if err := h.Start(); err != nil {
		c.t.Fatal(err)
	}
	c.t.Cleanup(func() { _ = h.Process.Kill(); _ = h.Wait() })
	return h
}

func caseEntryCloneFirst(c *mcase) {
	c.entrySetup()
	c.entryReal()
	c.exit(0)
	c.has("Cloning into " + c.repo)
	head, _ := gitCmd(c.env, c.dir, "-C", c.repo, "rev-parse", "HEAD")
	c.check("the clone is at the origin commit", head == entryHead)
	c.check("no lock, no temp copy", c.noLeftovers())
}

func caseEntryCloneRefresh(c *mcase) {
	c.entrySetup()
	c.entryOrigin(nil)
	c.entryReal()
	c.exit(0)
	harness.AppendFile(c.t, c.repo+"/macos/packs.toml", "# local edit\n")
	c.write(c.repo+"/untracked.txt", "untracked\n")
	harness.AppendFile(c.t, c.work+"/macos/tools/core/files/.config/core/plain.txt", "new line\n")
	c.git("-C", c.work, "commit", "--quiet", "--no-verify", "-am", "new origin commit")
	c.git("-C", c.dir+"/origin.git", "fetch", "--quiet", "--", c.work, "main:main")
	c.entryReal()
	c.exit(0)
	head, _ := gitCmd(c.env, c.dir, "-C", c.repo, "rev-parse", "HEAD")
	want, _ := gitCmd(c.env, c.dir, "-C", c.work, "rev-parse", "HEAD")
	c.check("the clone is at the new commit", head == want)
	c.check("the local edit is reset", cmpFiles(c.repo+"/macos/packs.toml", c.work+"/macos/packs.toml"))
	c.check("the untracked file is removed", !exists(c.repo+"/untracked.txt"))
	c.check("the new commit is rendered", hasLine(c.home+"/.config/core/plain.txt", "new line"))
}

func caseEntryCloneOtherOrigin(c *mcase) {
	c.entrySetup()
	c.git("clone", "--quiet", "--", entryURL, c.repo)
	c.git("-C", c.repo, "config", "remote.origin.url", "file://"+c.dir+"/other.git")
	c.entryReal()
	c.exit(0)
	c.has("has the origin 'file://" + c.dir + "/other.git', not " + entryURL + " - cloning again")
	url, _ := gitCmd(c.env, c.dir, "-C", c.repo, "config", "--get", "remote.origin.url")
	c.check("the origin is the URL", url == entryURL)
}

// caseEntryCloneSSHHTTPS: insteadOf sends both forms to the local origin, so
// no case needs the network.
func caseEntryCloneSSHHTTPS(c *mcase) {
	c.entrySetup()
	c.git("clone", "--quiet", "--", entryURL, c.repo)
	c.git("-C", c.repo, "config", "remote.origin.url", "git@github.com:virajp/bootstrap.git")
	gc := harness.GetEnv(c.env, "GIT_CONFIG_GLOBAL")
	for _, from := range []string{"https://github.com/virajp/bootstrap.git", "git@github.com:virajp/bootstrap.git"} {
		c.git("config", "--file", gc, "--add", "url."+entryURL+".insteadOf", from)
	}
	c.setEnv("BOOTSTRAP_URL=https://github.com/virajp/bootstrap.git")
	c.entryReal()
	c.exit(0)
	c.lacks("cloning again")
}

func caseEntryCloneNotGit(c *mcase) {
	c.entrySetup()
	c.write(c.repo+"/keep.txt", "keep\n")
	c.entryReal()
	c.exit(2)
	c.has("is not a git clone and not empty")
	ents, _ := os.ReadDir(c.repo)
	var names []string
	for _, e := range ents {
		names = append(names, e.Name())
	}
	c.check("the folder is unchanged", strings.Join(names, "\n") == "keep.txt")
	c.file(c.repo+"/keep.txt", "keep")
}

func caseEntryCloneFails(c *mcase) {
	c.entrySetup()
	c.setEnv("BOOTSTRAP_URL=file://" + c.dir + "/missing.git")
	c.entryReal()
	c.exit(5)
	c.has("Cannot clone")
}

// caseEntryVerbose: without --verbose only the changes print. The entry task
// consumes the flag: it exports BOOTSTRAP_VERBOSE=1 and does not pass the
// flag to the engine.
func caseEntryVerbose(c *mcase) {
	c.entrySetup()
	c.setEnv("BOOTSTRAP_VERBOSE", "FORCE_COLOR", "NO_COLOR")
	c.entryReal()
	c.exit(0)
	c.has("new ~/.config/core/plain.txt")
	c.lacks("unchanged  ")
	c.entryRun("--verbose", "--no-bootstrap")
	c.exit(0)
	c.has("unchanged  ~/.config/core/plain.txt")
	c.lacks("unknown")
	c.has("Synced: no file changed")
	c.lacks("| Total ")
}

func caseEntryLockWait(c *mcase) {
	c.entrySetup()
	c.mkdir(c.lock)
	holder := c.startHolder()
	pid := strconv.Itoa(holder.Process.Pid)
	c.write(c.lock+"/pid", pid+"\n")
	line := "Waiting for another sync (pid " + pid + ")"
	c.runs++
	c.log = filepath.Join(c.dir, fmt.Sprintf("run-%d.txt", c.runs))
	f, err := os.Create(c.log)
	if err != nil {
		c.t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	runner := exec.Command("bash", entryTask, "--no-bootstrap")
	runner.Dir = c.dir
	runner.Env = c.env
	runner.Stdout, runner.Stderr = f, f
	if err := runner.Start(); err != nil {
		c.t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- runner.Wait() }()
	for i := 0; i < 30 && !strings.Contains(harness.ReadFile(c.log), line); i++ {
		time.Sleep(time.Second)
	}
	c.has(line)
	var werr error
	alive := true
	select {
	case werr = <-done:
		alive = false
	default:
	}
	c.check("the run waits while the holder lives", alive)
	c.check("nothing written while waiting", !exists(c.home+"/.config"))
	_ = holder.Process.Kill()
	_ = holder.Wait()
	if alive {
		select {
		case werr = <-done:
		case <-time.After(60 * time.Second):
			_ = runner.Process.Kill()
			werr = <-done
		}
	}
	c.rc = 0
	if werr != nil {
		c.rc = 1
		if ee, ok := werr.(*exec.ExitError); ok {
			c.rc = ee.ExitCode()
		}
	}
	c.exit(0)
	c.check("values.txt written after the takeover", isFile(c.home+"/.config/core/values.txt"))
}

func caseEntryLockDead(c *mcase) {
	c.entrySetup()
	dead := exec.Command("true")
	if err := dead.Run(); err != nil {
		c.t.Fatal(err)
	}
	c.write(c.lock+"/pid", strconv.Itoa(dead.Process.Pid)+"\n")
	c.entryReal()
	c.exit(0)
	c.lacks("Waiting for another sync")
}

// caseEntryBadVersion: the engine reads cli/VERSION of the clone, its
// minimum version. A value that is not MAJOR.MINOR.PATCH stops it.
func caseEntryBadVersion(c *mcase) {
	c.entrySetup()
	c.entryOrigin(func(c *mcase) { c.write(c.work+"/cli/VERSION", "v1.2\n") })
	c.entryReal()
	c.exit(5)
	c.has("cli/VERSION of the clone is not MAJOR.MINOR.PATCH ('v1.2')")
	c.check("no lock, no temp copy", c.noLeftovers())
}

func caseEntryUnknownFlag(c *mcase) {
	c.entrySetup()
	c.entryRun("--bogus")
	c.exit(2)
	c.has("unknown flag '--bogus'")
	c.check("no clone made", !exists(c.repo))
}

// wouldUpdate gives the would-update lines of the log.
func (c *mcase) wouldUpdate() string {
	var out []string
	for _, l := range lines(c.log) {
		if strings.HasPrefix(l, "bootstrap: would update ") {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

// caseEntryMiseRun: the same dry run under bash and under `mise run`. Only
// the would-update lines: `mise run` starts the task in another folder, and
// the line of bootstrap_commit shows macos.toml relative to that folder.
func caseEntryMiseRun(c *mcase) {
	tasks := c.xdgCfg + "/mise/tasks"
	c.entrySetup()
	c.entryDry()
	c.exit(0)
	a := c.wouldUpdate()
	if err := copyFile(entryTool+"/tasks/helpers", tasks+"/_scripts/helpers"); err != nil {
		c.t.Fatal(err)
	}
	if err := copyFile(entryTask, tasks+"/updateall/configs/_default"); err != nil {
		c.t.Fatal(err)
	}
	c.entryMise("--dry-run", "--no-bootstrap")
	c.exit(0)
	b := c.wouldUpdate()
	c.check("mise run prints the same would-update lines", a != "" && a == b)
	c.entryMise("--bogus")
	c.check(fmt.Sprintf("mise run rejects an unknown flag itself (exit %d)", c.rc), c.rc != 0)
}

func caseEntryConfigUnset(c *mcase) {
	c.entrySetup()
	c.setEnv("MACOS_SETUP_REPO_PATH")
	c.entryReal()
	c.exit(2)
	c.has("MACOS_SETUP_REPO_PATH is not set")
	c.check("no clone made", !exists(c.repo))
}

func caseEntryNested(c *mcase) {
	c.entrySetup()
	c.setEnv("_UPDATEALL_CONFIGS_RUNNING=1")
	c.entryReal()
	c.exit(2)
	c.has("a nested run is refused")
	c.check("no clone made", !exists(c.repo))
}

// caseEntryLegacyMove: the list of the old cache names a stale file, so its
// removal proves the move.
func caseEntryLegacyMove(c *mcase) {
	c.entrySetup()
	c.newHeaderFile(c.home+"/.config/old/stale.txt", "stale")
	c.write(c.oldCache+"/generated.txt", "core\t.config/old/stale.txt\n")
	c.entryReal()
	c.exit(0)
	c.check("the old Mac cache is removed", !exists(c.oldCache))
	c.check("the moved list is the manifest: its stale file is removed", !exists(c.home+"/.config/old/stale.txt"))
	c.check("the manifest exists", isFile(c.manifest))
}

// caseEntryLegacyKeepManifest: a dry run does not write the manifest, so its
// content is the kept one.
func caseEntryLegacyKeepManifest(c *mcase) {
	c.entrySetup()
	c.write(c.oldCache+"/generated.txt", "core\t.config/from-old\n")
	c.seedManifest("core\t.config/core/values.txt")
	c.entryDry()
	c.exit(0)
	c.file(c.manifest, "core\t.config/core/values.txt")
	c.check("the old Mac cache is removed", !exists(c.oldCache))
}

func caseEntryLegacyLive(c *mcase) {
	c.entrySetup()
	holder := c.startHolder()
	pid := strconv.Itoa(holder.Process.Pid)
	c.write(c.oldCache+"/lock/pid", pid+"\n")
	c.entryReal()
	c.exit(0)
	c.has("the old Mac sync runs (pid " + pid + "): keeping " + c.oldCache)
	c.check("the old Mac cache stays", isFile(c.oldCache+"/lock/pid"))
}

func caseEntryLeftovers(c *mcase) {
	c.entrySetup()
	c.entryReal()
	c.exit(0)
	c.check("after a real run: no lock, no temp copy", c.noLeftovers())
	c.entryDry()
	c.exit(0)
	c.check("after a dry run: no lock, no temp copy", c.noLeftovers())
	c.write(c.values, `colour = "blue"`+"\n"+catT(c.values)+"\n")
	c.entryReal()
	c.exit(2)
	c.has("colour is not a key of")
	c.check("after a bad macos.toml: no lock, no temp copy", c.noLeftovers())
}
