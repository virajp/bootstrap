package mac

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

// ---- the tree changes and the case makers ---------------------------------

// expect is expect: a dry run that stops with <code> and <msg>.
func expect(cfg string, code int, msg string, change func(c *mcase)) func(c *mcase) {
	return func(c *mcase) {
		c.useConfig(cfg)
		if change != nil {
			c.caseTree(change)
		}
		c.runDry()
		c.exit(code)
		c.has(msg)
	}
}

// passTree is pass_tree: a dry run of the ok config that passes.
func passTree(change func(c *mcase)) func(c *mcase) {
	return func(c *mcase) {
		c.useConfig("ok")
		c.caseTree(change)
		c.runDry()
		c.exit(0)
	}
}

func packs(ls ...string) func(c *mcase)        { return func(c *mcase) { c.appendPacks(ls...) } }
func prepend(tool, line string) func(c *mcase) { return func(c *mcase) { c.prependKey(tool, line) } }
func appendTo(tool string, ls ...string) func(c *mcase) {
	return func(c *mcase) { c.appendLines(tool, ls...) }
}

func addBadTool(c *mcase) {
	if err := harness.CopyTree(filepath.Join(fixtures, "repo", "tools", "bad"), filepath.Join(c.work, "tools", "bad")); err != nil {
		c.t.Fatal(err)
	}
}
func noPacks(c *mcase) { c.remove(filepath.Join(c.work, "packs.toml")) }
func noBase(c *mcase) {
	c.write(filepath.Join(c.work, "packs.toml"), harness.Lines("[pack.extra]", "opt = {}"))
}
func badQuestion(c *mcase) { c.appendLines("opt", "[questions.level]", `ask = "Which level?"`) }

func newTool(name string) func(c *mcase) {
	return func(c *mcase) {
		c.write(filepath.Join(c.work, "tools", name, "tools.toml"), harness.Lines(`description = "x"`, `summary = "x"`, `category = "dev"`))
	}
}
func shipsConfD(c *mcase) {
	c.write(filepath.Join(c.work, "tools/core/.config/mise/conf.d/x.toml"), "[env]\n")
}
func twoToolsOneTarget(c *mcase) {
	c.write(filepath.Join(c.work, "tools/opt/.config/core/plain.txt"), "a\n")
}

// catalogKeys: the catalog keys of a Mac tool and the categories file.
func catalogKeys(c *mcase) {
	c.setKey("opt", "summary", `summary = "The opt tool of the fixture repo"`)
	c.setKey("opt", "category", `category = "dev"`)
	c.prependKey("opt", "selectable = true")
	c.write(filepath.Join(c.work, "categories.toml"), harness.Lines("[category.dev]", `title = "Development"`))
}

// morePacks: a nested pack, and a pack whose member init needs its member
// initdep.
func morePacks(c *mcase) {
	c.appendPacks("[pack.outer]", `"pack:extra" = {}`)
	c.appendPacks("[pack.dev]", "initdep = {}", "init = {}")
}

// migrateTree: the tool alfred of the pack essentials of the frozen copy,
// and merger in the new pack base, a member that is not required.
func migrateTree(c *mcase) {
	c.write(filepath.Join(c.work, "tools/alfred/tools.toml"), harness.Lines(`summary = "The alfred tool of the fixture repo"`, `category = "dev"`))
	p := filepath.Join(c.work, "packs.toml")
	var b strings.Builder
	b.WriteString("\n" + harness.Lines("[pack.base]", "core = { required = true }", "mise = { required = true }", "merger = {}"))
	in := false
	for _, l := range lines(p) {
		if strings.HasPrefix(l, "[pack.extra]") {
			in = true
		}
		if in {
			b.WriteString(l + "\n")
		}
	}
	c.write(p, b.String())
}

// ---- config -----------------------------------------------------------------

func caseConfigMissing(c *mcase) {
	c.remove(c.cfg)
	c.runReal()
	c.exit(2)
	c.has("MACOS_SETUP_REPO_PATH " + c.cfg + " is not a git repository")
	c.check("nothing written", !exists(c.home+"/.config"))
}

func caseConfigNotGit(c *mcase) {
	c.remove(c.cfg + "/.git")
	c.runReal()
	c.exit(2)
	c.has("MACOS_SETUP_REPO_PATH " + c.cfg + " is not a git repository")
	c.check("nothing written", !exists(c.home+"/.config"))
}

func caseConfigUnset(c *mcase) {
	c.setEnv("MACOS_SETUP_REPO_PATH")
	c.runReal()
	c.exit(2)
	c.has("MACOS_SETUP_REPO_PATH is not set")
	c.check("nothing written", !exists(c.home+"/.config"))
}

func caseUnknownFlag(c *mcase) {
	c.useConfig("ok")
	c.syncWith(nil, "--bogus")
	c.exit(2)
	c.has("sync has no flag --bogus")
	c.has("usage: bootstrap")
}

// ---- packs and selection ----------------------------------------------------

func caseToolsList(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.has("tools: core mise opt")
	c.check("merger (pack extra, not listed) is not rendered", !exists(c.home+"/.config/merger"))
}

func caseAddPack(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.commitConfig()
	c.mac("add", "pack", "--yes", "--no-bootstrap", "extra")
	c.exit(0)
	c.has("select merger")
	c.check("merger is selected", hasLine(c.values, "[tools.merger]"))
	c.check("opt has one table", countLine(c.values, "[tools.opt]") == 1)
	c.check("no [packs.*] table", !hasRe(c.values, `^\[packs\.`))
	c.check("merger is rendered", isFile(c.home+"/.config/merger/plain.txt"))
}

func caseAddPackExclude(c *mcase) {
	c.useConfig("mise-user")
	c.mac("add", "pack", "--yes", "--no-bootstrap", "extra", "--exclude", "merger")
	c.exit(0)
	c.check("opt is selected", hasLine(c.values, "[tools.opt]"))
	c.check("the excluded merger is not selected", !hasStr(c.values, "[tools.merger]"))
}

func caseAddPackNested(c *mcase) {
	c.useConfig("mise-user")
	c.caseTree(morePacks)
	c.mac("add", "pack", "--yes", "--no-bootstrap", "outer")
	c.exit(0)
	c.check("merger of the nested pack extra is selected", hasLine(c.values, "[tools.merger]"))
	c.check("opt of the nested pack extra is selected", hasLine(c.values, "[tools.opt]"))
}

// addPackRefused: add pack stops with code 2, and macos.toml stays.
func addPackRefused(args, msg string) func(c *mcase) {
	return func(c *mcase) {
		c.useConfig("ok")
		c.caseTree(morePacks)
		before := catT(c.values)
		c.mac(append([]string{"add", "pack", "--yes", "--no-bootstrap"}, strings.Fields(args)...)...)
		c.exit(2)
		c.has(msg)
		c.check("macos.toml is unchanged", catT(c.values) == before)
	}
}

func caseAddPackTwo(c *mcase) {
	c.useConfig("mise-user")
	c.mac("add", "pack", "--yes", "--no-bootstrap", "extra", "base")
	c.exit(0)
	c.has("base is already selected")
	c.check("merger is selected", hasLine(c.values, "[tools.merger]"))
	c.check("opt is selected", hasLine(c.values, "[tools.opt]"))
}

// ---- the old form of macos.toml and its migration ---------------------------

func caseSyncNoFormat(c *mcase) {
	c.useConfig("old-packs")
	// mac sync runs mac migrate first; mac check does not migrate.
	c.runDry()
	c.exit(0)
	c.has("bootstrap mac migrate has a change to make: it runs before the sync")
	c.mac("check")
	c.exit(2)
	c.has("without format = 2 is the old form - run bootstrap mac migrate")
}

func caseInitMigratePacks(c *mcase) {
	c.useConfig("old-packs")
	c.caseTree(migrateTree)
	c.mac("init", "--yes", "--no-bootstrap")
	c.exit(0)
	c.has("[packs.essentials] removed")
	c.has("[tools.devtools] removed: bootstrap no longer has the tool")
	c.check("format = 2", hasLine(c.values, "format = 2"))
	c.check("no [packs.*] table", !hasRe(c.values, `^\[packs\.`))
	c.check("no [tools.devtools]", !hasStr(c.values, "[tools.devtools]"))
	c.check("alfred of the frozen pack essentials is selected", hasLine(c.values, "[tools.alfred]"))
	c.check("mise of the frozen pack base is selected", hasLine(c.values, "[tools.mise]"))
	c.check("core, a required member, has a table", hasLine(c.values, "[tools.core]"))
	c.check("merger, a member of the new base that is not required, has no table", !hasStr(c.values, "[tools.merger]"))
	c.check("the table of opt keeps its value", hasLine(c.values, `level = "migrated"`))
	c.commitConfig()
	c.runDry()
	c.exit(0)
}

// ---- the required members of the pack base ---------------------------------

func caseRequiredTable(c *mcase) {
	c.useConfig("-")
	c.write(c.values, harness.Lines("format = 2", "", "[tools.core]"))
	c.commitConfig()
	before := catT(c.values)
	c.runDry()
	c.exit(0)
	c.has("[tools.mise] would be added to config/macos.toml")
	c.check("a dry run writes no table", catT(c.values) == before)
	c.mac("check")
	c.exit(0)
	c.has("[tools.mise] would be added to")
	c.check("check writes no table", catT(c.values) == before)
	c.runReal()
	c.exit(0)
	c.has("[tools.mise] added to config/macos.toml")
	c.check("the table of mise is written", hasLine(c.values, "[tools.mise]"))
	c.runReal()
	c.exit(0)
	c.lacks("added to config/macos.toml")
}

// ---- mise (the check of the install) ----------------------------------------

// fakeMise: a mise in <dir> that is not ~/.local/bin/mise.
func (c *mcase) fakeMise(dir string) {
	c.write(dir+"/mise", harness.Lines("#!/bin/sh", "exit 0"))
	c.chmod(dir+"/mise", 0o755)
}

func caseMiseElsewhere(c *mcase) {
	c.useConfig("ok")
	c.fakeMise(c.dir + "/homebrew/bin")
	c.prependPath(c.dir + "/homebrew/bin")
	c.runReal()
	c.exit(2)
	c.has("mise must be installed with `curl https://mise.run | sh` at ~/.local/bin/mise, not from a package manager (found: " + c.dir + "/homebrew/bin/mise)")
	c.check("nothing written", !exists(c.home+"/.config"))
}

func caseMiseLink(c *mcase) {
	c.useConfig("ok")
	c.fakeMise(c.dir + "/homebrew/bin")
	c.mkdir(c.dir + "/link/.local/bin")
	c.symlink(c.dir+"/homebrew/bin/mise", c.dir+"/link/.local/bin/mise")
	c.prependPath(c.dir + "/link/.local/bin")
	c.runReal()
	c.exit(2)
	c.has("(found: " + c.dir + "/homebrew/bin/mise)")
}

// ---- the generated mise config ----------------------------------------------

const miseCoreHeader = "# Synced from bootstrap: macos/tools/core/tools.toml. Do not edit this file.\n" +
	"# The next sync writes it again, and each change in this file is lost."

func caseMiseTool(c *mcase) {
	file := c.home + "/.config/mise/conf.d/core/mise.toml"
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.file(file, miseCoreHeader+"\n\n[env]\nCORE_HOME = \"{{ env.HOME }}\"\nCORE_NAME = \"override-name\"")
	c.check("the manifest has its line", hasLine(c.manifest, "core\t.config/mise/conf.d/core/mise.toml"))
	c.check("a tool with no [mise.*] gets no file", !exists(c.home+"/.config/mise/conf.d/opt"))
}

func caseMiseUser(c *mcase) {
	file := c.home + "/.config/mise/conf.d/zz-macos/mise.toml"
	c.useConfig("mise-user")
	c.runReal()
	c.exit(0)
	c.file(file, "# Synced from bootstrap: macos.toml [mise]. Do not edit this file.\n"+
		"# The next sync writes it again, and each change in this file is lost.\n\n[shell_alias]\nfixture = \"echo fixture\"")
	c.check("the manifest has its line", hasLine(c.manifest, "zz-macos\t.config/mise/conf.d/zz-macos/mise.toml"))
}

func caseMiseOverride(c *mcase) {
	file := c.home + "/.config/mise/conf.d/core/mise.toml"
	c.useConfig("override-mise")
	c.runReal()
	c.exit(0)
	c.check("the key changes, with the comment", hasRe(file, `^ *CORE_NAME = "from-override"  # override: macos\.toml$`))
	c.check("the other key has no comment", hasRe(file, `^ *CORE_HOME = "\{\{ env\.HOME }}"$`))
}

func caseMiseDeselected(c *mcase) {
	file := c.home + "/.config/mise/conf.d/init/mise.toml"
	c.useConfig("init")
	c.runReal()
	c.exit(4)
	c.check("the mise config is written", isFile(file))
	c.useConfig("-")
	c.runReal()
	c.exit(0)
	c.check("the mise config of the tool is removed", !exists(file))
	c.check("and its folder", !exists(filepath.Dir(file)))
	c.check("the mise config of core stays", isFile(c.home+"/.config/mise/conf.d/core/mise.toml"))
}

// ---- blockers ---------------------------------------------------------------

func caseBlockerHeader(c *mcase) {
	c.useConfig("ok")
	c.headerFile(c.home+"/.config/core", "old")
	c.runReal()
	c.exit(0)
	c.has("removed    home/.config/core (a folder replaces it)")
	c.check("values.txt written", isFile(c.home+"/.config/core/values.txt"))
}

func caseBlockerPlain(c *mcase) {
	c.useConfig("ok")
	c.write(c.home+"/.config/opt", "mine\n")
	c.runReal()
	c.exit(1)
	c.has("home/.config/opt is a file of this repo, but the sync needs a folder there - move it")
	c.file(c.home+"/.config/opt", "mine")
	c.check("the other targets written", isFile(c.home+"/.config/core/values.txt"))
}

// ---- values -----------------------------------------------------------------

func caseValues(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.file(c.home+"/.config/core/values.txt", synced("core/files/.config/core/values.txt.tmpl")+
		"\nname=override-name\nitems=a,b,x,\nleft=file-left\nright=default-right\ntools=core,mise,opt,\nopt_extra=extra-value\nhome="+c.home+
		"\npassthrough={{ mise }} [[ bash ]]")
	c.file(c.home+"/.config/opt/opt.txt", synced("opt/files/.config/opt/opt.txt.tmpl")+"\nlevel=default-level\nextra=extra-value")
}

func caseValuesDependency(c *mcase) {
	items := c.home + "/.config/initdep/items.txt"
	head := synced("initdep/files/.config/initdep/items.txt.tmpl")
	c.useConfig("dep")
	c.runReal()
	c.exit(4)
	c.file(items, head+"\nitems=a,b,y,x,")
	c.valuesFile("[tools.init]", "[tools.initdep]", `items = [ "x", "a" ]`)
	c.runReal()
	c.exit(4)
	c.file(items, head+"\nitems=a,b,x,")
}

func ownerConfig(ls ...string) string {
	return synced("owner/files/.config/owner/config.tmpl") + "\n" + strings.Join(ls, "\n")
}

func caseValuesOptIn(c *mcase) {
	c.useConfig("replace")
	c.runReal()
	c.exit(0)
	c.file(c.home+"/.config/owner/config", ownerConfig("Include ~/a", "Include ~/b", "Include ~/z", "", "Host *",
		"  IdentityAgent agent-a", "nested=default,", "# again: agent-a"))
}

func caseValuesOptInType(c *mcase) {
	c.useConfig("type-opt")
	c.runReal()
	c.exit(0)
	c.check("the number replaces the string", hasLine(c.home+"/.config/opt/opt.txt", "level=3"))
}

// ---- render -----------------------------------------------------------------

// caseRenderOrder: the class of a target is 1 init ~/.config/mise, 2 init
// rest, 3 main ~/.config/mise, 4 main rest.
func caseRenderOrder(c *mcase) {
	const t = "~"
	c.useConfig("init")
	c.runDry()
	c.exit(0)
	re := regexp.MustCompile(`^bootstrap: would update ([^ ]*)`)
	last, ok, n := 0, true, 0
	for _, l := range lines(c.log) {
		m := re.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		p, class := m[1], 4
		switch {
		case strings.HasPrefix(p, t+"/.config/init/") || strings.HasPrefix(p, t+"/.config/initdep/"):
			class = 2
		case p == t+"/.config/mise/conf.d/init/mise.toml":
			class = 1
		case strings.HasPrefix(p, t+"/.config/mise/"):
			class = 3
		}
		if class < last {
			ok = false
		}
		last = class
		n++
	}
	c.check(fmt.Sprintf("%s/.config/mise/** first in each stage, the init stage first (%d files)", t, n), ok && n > 5)
}

func addExecutables(c *mcase) {
	run := c.work + "/tools/core/.config/core/run"
	rendered := c.work + "/tools/core/.config/core/rendered.tmpl"
	c.write(run, "#!/bin/sh\n")
	c.write(rendered, harness.Lines("#!/bin/sh", "# <% .config.name %>"))
	c.chmod(run, 0o755)
	c.chmod(rendered, 0o755)
}

func caseRenderExact(c *mcase) {
	c.useConfig("ok")
	c.caseTree(addExecutables)
	c.runReal()
	c.exit(0)
	c.check("plain.txt equals its source", cmpFiles(fixtures+"/repo/tools/core/files/.config/core/plain.txt", c.home+"/.config/core/plain.txt"))
	c.file(c.home+"/.config/core/settings.json", `{"name":"override-name","list":[1,2]}`)
	c.check("a file keeps its mode", modeOf(c.home+"/.config/core/plain.txt") == "644")
	c.check("an executable keeps its mode", modeOf(c.home+"/.config/core/run") == "755")
	c.check("a rendered file takes the mode of its template", modeOf(c.home+"/.config/core/rendered") == "755")
}

func caseSeed(c *mcase) {
	seed := c.home + "/.config/core/seeded.toml"
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.has("skip seed  ~/.config/core/seeded.toml (no ~/.config/core-needed)")
	c.check("no seed while a requires target is missing", !exists(seed))
	c.touch(c.home + "/.config/core-needed")
	c.runReal()
	c.exit(0)
	c.file(seed, `name = "override-name"`)
	c.write(seed, "mine\n")
	c.runReal()
	c.exit(0)
	c.has("kept seed  ~/.config/core/seeded.toml")
	c.file(seed, "mine")
	c.check("no seed in the manifest", !hasStr(c.manifest, "seeded.toml"))
}

func caseOverwrite(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.write(c.home+"/.config/core/values.txt", "edited\n")
	c.runReal()
	c.exit(0)
	c.has("updated    ~/.config/core/values.txt")
	c.check("the edit is overwritten", hasLine(c.home+"/.config/core/values.txt", "name=override-name"))
}

func caseContext(c *mcase) {
	c.useConfig("context")
	c.runReal()
	c.exit(0)
	c.file(c.home+"/.config/context/context.txt", synced("context/files/.config/context/context.txt.tmpl")+
		"\ncli_dir="+tree+"/cli\nconfig="+c.cfg)
}

// ---- symlinks ---------------------------------------------------------------

func caseLinkFile(c *mcase) {
	c.useConfig("ok")
	c.mkdir(c.home + "/.config/core")
	c.write(c.user+"/values.txt", "old\n")
	c.symlink(c.user+"/values.txt", c.home+"/.config/core/values.txt")
	c.symlink(c.user+"/missing/plain.txt", c.home+"/.config/core/plain.txt")
	c.runReal()
	c.exit(0)
	c.check("the link stays", isLink(c.home+"/.config/core/values.txt"))
	c.check("its source changes", hasLine(c.user+"/values.txt", "name=override-name"))
	c.check("a link with a missing source writes the source", isFile(c.user+"/missing/plain.txt") && isLink(c.home+"/.config/core/plain.txt"))
}

func caseLinkFolder(c *mcase) {
	c.useConfig("ok")
	c.mkdir(c.home + "/.config")
	c.mkdir(c.user + "/opt")
	c.symlink(c.user+"/opt", c.home+"/.config/opt")
	c.runReal()
	c.exit(0)
	c.check("the folder link stays", isLink(c.home+"/.config/opt"))
	c.check("the file is written in the linked folder", hasLine(c.user+"/opt/opt.txt", "level=default-level"))
}

func caseLinkStale(c *mcase) {
	c.useConfig("ok")
	c.mkdir(c.home + "/.config")
	c.mkdir(c.user + "/linked")
	c.symlink(c.user+"/linked", c.home+"/.config/linked")
	c.headerFile(c.user+"/linked/stale.txt", "stale")
	c.seedManifest("core\t.config/linked/stale.txt")
	c.runReal()
	c.exit(0)
	c.check("the source of the stale file is removed", !exists(c.user+"/linked/stale.txt"))
	c.check("the link stays", isLink(c.home+"/.config/linked") && isDir(c.user+"/linked"))
}

// ---- stale ------------------------------------------------------------------

func caseStale(c *mcase) {
	c.useConfig("ok")
	c.newHeaderFile(c.home+"/.config/old/deep/stale.txt", "stale")
	c.write(c.home+"/.config/nohead/kept.txt", "no header\n")
	c.seedManifest("core\t.config/old/deep/stale.txt", "core\t.config/nohead/kept.txt")
	c.runReal()
	c.exit(0)
	c.has("removed    home/.config/old/deep/stale.txt")
	c.check("the stale file is removed", !exists(c.home+"/.config/old/deep/stale.txt"))
	c.check("its empty parent folders are removed", !exists(c.home+"/.config/old"))
	c.check("a file without the header is kept", isFile(c.home+"/.config/nohead/kept.txt"))
}

func caseStaleOldHeader(c *mcase) {
	c.useConfig("ok")
	c.headerFile(c.home+"/.config/old/stale.txt", "stale")
	c.seedManifest("core\t.config/old/stale.txt")
	c.runReal()
	c.exit(0)
	c.check("the file with the old header is removed", !exists(c.home+"/.config/old/stale.txt"))
}

func caseStaleDeselected(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.check("opt is in the manifest", hasLine(c.manifest, "opt\t.config/opt/opt.txt"))
	c.useConfig("-")
	c.runReal()
	c.exit(0)
	c.check("the files of opt are removed", !exists(c.home+"/.config/opt"))
	c.check("core stays", isFile(c.home+"/.config/core/values.txt"))
}

func caseGenerated(c *mcase) {
	c.useConfig("ok")
	c.touch(c.home + "/.config/core-needed")
	c.runReal()
	c.exit(0)
	c.check("the manifest is in the cache folder", isFile(c.manifest))
	c.check("its first line is the header", firstLine(c.manifest) == "# Written by bootstrap. Do not edit this file.")
	var rows []string
	if l := lines(c.manifest); len(l) > 0 {
		rows = l[1:]
	}
	row := regexp.MustCompile("^[a-z][a-z0-9-]*\t[^/\t~][^\t]*$")
	good := true
	for _, r := range rows {
		if !row.MatchString(r) {
			good = false
		}
	}
	c.check("each other line is <tool><TAB><path relative to HOME>", good)
	c.check("the other lines are sorted", isFile(c.manifest) && sort.StringsAreSorted(rows))
	c.check("no path names HOME", !hasStr(c.manifest, c.home))
	c.check("core values.txt is listed", hasLine(c.manifest, "core\t.config/core/values.txt"))
	c.check("no seed in the manifest", !hasStr(c.manifest, "seeded.toml"))
	before := catT(c.manifest)
	c.runReal()
	c.exit(0)
	c.check("the manifest is the same after a second run", catT(c.manifest) == before)
}

// ---- the scripts of a tool --------------------------------------------------

func caseScriptsDerive(c *mcase) {
	c.useConfig("scripted")
	c.runReal()
	c.exit(0)
	c.file(c.home+"/.config/scripted/values.txt", "word=user derived=from-default home=yes")
	c.check("the derived mise table is in the mise file", hasStr(c.home+"/.config/mise/conf.d/scripted/mise.toml", `scripted = "echo scripted"`))
}

func caseScriptsCheck(c *mcase) {
	c.useConfig("scripted")
	c.touch(c.home + "/.check-fail")
	c.runDry()
	c.exit(2)
	c.has("scripted: the check failed")
	c.runReal()
	c.exit(2)
	c.check("a failed check writes nothing", !exists(c.home+"/.config/scripted/values.txt"))
}

func caseScriptsAfter(c *mcase) {
	c.useConfig("scripted")
	c.touch(c.home + "/.restart")
	c.runDry()
	c.exit(0)
	c.runReal()
	c.exit(4)
	c.has("scripted: a new shell")
	c.has("open a new terminal, then run: mise run updateall:configs")
	c.check("the files are written before the stop", isFile(c.home+"/.config/scripted/values.txt"))
	c.check("no commit after a stop", !hasStr(c.values, "bootstrap_commit"))
	c.remove(c.home + "/.restart")
	c.runReal()
	c.exit(0)
}

// ---- init stage -------------------------------------------------------------

func caseInitStage(c *mcase) {
	c.seedManifest("opt\t.config/opt/opt.txt")
	c.useConfig("init")
	c.runReal()
	c.exit(4)
	c.has("open a new terminal, then run: mise run updateall:configs")
	c.check("the init files are written", isFile(c.home+"/.config/init/initrc"))
	c.check("the init stage mise file is written", isFile(c.home+"/.config/mise/conf.d/init/mise.toml"))
	c.check("a dependency of the init tool renders in the init stage", isFile(c.home+"/.config/initdep/dep.txt"))
	c.check("the core targets are not written", !exists(c.home+"/.config/core/values.txt"))
	c.check("the lines of other tools stay", hasLine(c.manifest, "opt\t.config/opt/opt.txt"))
	c.check("the init lines are written", hasLine(c.manifest, "init\t.config/init/initrc"))
	c.check("the dependency lines are written", hasLine(c.manifest, "initdep\t.config/initdep/dep.txt"))
	c.check("no commit after a stop", !hasStr(c.values, "bootstrap_commit"))
	c.runReal()
	c.exit(0)
	c.lacks("bootstrap: open a new terminal")
	c.check("the second run writes the core targets", isFile(c.home+"/.config/core/values.txt"))
}

func caseInitChmod(c *mcase) {
	c.useConfig("init")
	c.runReal()
	c.exit(4)
	c.runReal()
	c.exit(0)
	c.chmod(c.home+"/.config/init/initrc", 0o600)
	c.runReal()
	c.exit(4)
	c.has("chmod 644  ~/.config/init/initrc")
	c.has("open a new terminal, then run: mise run updateall:configs")
}

func caseInitFail(c *mcase) {
	c.useConfig("init")
	c.mkdir(c.home + "/.config/init/initrc")
	c.runReal()
	c.exit(1)
	c.lacks("bootstrap: open a new terminal")
	c.check("no later stage", !exists(c.home+"/.config/core/values.txt"))
	c.check("no manifest write", !exists(c.manifest))
}

// ---- dir_perm and perm: the ssh folder --------------------------------------

func caseSSHNew(c *mcase) {
	c.useConfig("sshdir")
	c.runReal()
	c.exit(0)
	c.check("the folder has mode 700", modeOf(c.home+"/.ssh") == "700")
	c.check("the file is written in it", isFile(c.home+"/.ssh/config"))
	c.check("perm gives the file its mode", modeOf(c.home+"/.ssh/config.identity") == "600")
	c.check("a file without perm has the mode of its source", modeOf(c.home+"/.ssh/config") == "644")
	c.runReal()
	c.exit(0)
	c.lacks("chmod 700")
}

func (c *mcase) sshFolder755() {
	c.mkdir(c.home + "/.ssh")
	c.chmod(c.home+"/.ssh", 0o755)
}

func caseSSHExisting(c *mcase) {
	c.useConfig("sshdir")
	c.sshFolder755()
	c.runReal()
	c.exit(0)
	c.has("chmod 700  ~/.ssh")
	c.check("the folder has mode 700", modeOf(c.home+"/.ssh") == "700")
}

func caseSSHLink(c *mcase) {
	c.useConfig("sshdir")
	c.sshFolder755()
	c.touch(c.user + "/ssh-config")
	c.symlink(c.user+"/ssh-config", c.home+"/.ssh/config")
	c.runReal()
	c.exit(0)
	c.check("the folder has mode 700", modeOf(c.home+"/.ssh") == "700")
	c.check("the link stays", isLink(c.home+"/.ssh/config"))
}

func caseSSHDry(c *mcase) {
	c.useConfig("sshdir")
	c.sshFolder755()
	c.runDry()
	c.exit(0)
	c.has("would chmod 700 ~/.ssh")
	c.check("one line for the two targets", countRe(c.log, `would chmod 700 ~/\.ssh$`) == 1)
	c.check("the mode stays 755", modeOf(c.home+"/.ssh") == "755")
	c.remove(c.home + "/.ssh")
	c.runDry()
	c.exit(0)
	c.check("no folder is made", !exists(c.home+"/.ssh"))
	c.lacks("stat:")
	c.lacks("chmod 700")
}

func caseSSHFolderLink(c *mcase) {
	c.useConfig("sshdir")
	c.mkdir(c.user + "/ssh")
	c.chmod(c.user+"/ssh", 0o755)
	c.symlink(c.user+"/ssh", c.home+"/.ssh")
	c.runReal()
	c.exit(0)
	c.check("the target of the link has mode 700", modeOf(c.user+"/ssh/.") == "700")
	c.runReal()
	c.exit(0)
	c.lacks("chmod 700")
}

// ---- macos.toml and bootstrap_commit ----------------------------------------

func commitLine() string { return `bootstrap_commit = "` + treeCommit + `"` }

func caseValuesMissing(c *mcase) {
	c.useConfig("-")
	c.runReal()
	c.exit(0)
	c.check("macos.toml starts with the header", firstLine(c.values) == "# The values of the Mac sync of bootstrap. See docs/setup/index.md.")
	c.check("format = 2", hasLine(c.values, "format = 2"))
	c.check("the tables of the required members", hasLine(c.values, "[tools.mise]"))
	c.check("it has the commit line", hasLine(c.values, commitLine()))
}

func caseCommit(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	at, table := "", ""
	var rest strings.Builder
	for i, l := range lines(c.values) {
		if l == commitLine() {
			if at == "" {
				at = strconv.Itoa(i + 1)
			}
			continue
		}
		if table == "" && strings.HasPrefix(l, "[") {
			table = strconv.Itoa(i + 1)
		}
		rest.WriteString(l + "\n")
	}
	num := func(s string, def int) int {
		if n, err := strconv.Atoi(s); err == nil {
			return n
		}
		return def
	}
	c.check(fmt.Sprintf("the commit line is after the header comments (%s)", at), at != "" && num(at, 0) > 2)
	c.check(fmt.Sprintf("and before the first table (%s)", table), num(at, 99) < num(table, 0))
	c.check("no other line changed", harness.ReadFile(fixtures+"/config/ok/macos.toml") == rest.String())
	after := catT(c.values)
	c.runReal()
	c.exit(0)
	c.has("unchanged  ~/.config/core/values.txt")
	c.check("a second run reports only unchanged", !c.logHasRe(`^bootstrap: (updated|chmod|removed|wrote|set |would)`))
	c.check("macos.toml is unchanged", catT(c.values) == after)
}

// caseCommitKept: a run that changes no file keeps a commit that is set, so
// it adds no commit to the repository of macos.toml.
func caseCommitKept(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	old := `bootstrap_commit = "0000000000000000000000000000000000000000"`
	c.write(c.values, strings.Replace(catT(c.values), commitLine(), old, 1))
	c.runReal()
	c.exit(0)
	c.check("the old commit stays", hasLine(c.values, old))
	c.check("no commit line in the log", !c.logHasRe(`bootstrap_commit`))
}

// caseCommitMoved: a run that changes a file sets the commit of the clone.
func caseCommitMoved(c *mcase) {
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	old := `bootstrap_commit = "0000000000000000000000000000000000000000"`
	c.write(c.values, strings.Replace(catT(c.values), commitLine(), old, 1))
	c.check("the manifest is removed", os.Remove(c.manifest) == nil)
	c.runReal()
	c.exit(0)
	c.check("the commit of the clone is set", hasLine(c.values, commitLine()))
}

func caseCommitLink(c *mcase) {
	c.useConfig("-")
	c.write(c.user+"/real.toml", harness.Lines("format = 2", "[tools.core]", "[tools.mise]", "[tools.opt]"))
	c.symlink(c.user+"/real.toml", c.values)
	c.runReal()
	c.exit(0)
	c.check("the link stays", isLink(c.values))
	c.check("the real file has the commit line", hasLine(c.user+"/real.toml", commitLine()))
}

func caseCommitDry(c *mcase) {
	c.useConfig("ok")
	before := catT(c.values)
	c.runDry()
	c.exit(0)
	c.has("would set bootstrap_commit " + treeCommit)
	c.check("macos.toml is unchanged", catT(c.values) == before)
}

// ---- --dry-run --------------------------------------------------------------

func caseDryRun(c *mcase) {
	c.useConfig("init")
	c.mkdir(c.home + "/.config/core")
	c.write(c.user+"/values.txt", "old\n")
	c.symlink(c.user+"/values.txt", c.home+"/.config/core/values.txt")
	c.headerFile(c.home+"/.config/old/stale.txt", "stale")
	c.seedManifest("core\t.config/old/stale.txt")
	before := snapshot(c.home, c.user, c.cfg, c.manifest)
	c.runDry()
	c.exit(0)
	c.has("would update ~/.config/core/values.txt (-> " + c.user + "/values.txt)")
	c.has("would remove home/.config/old/stale.txt")
	c.has("would update ~/.config/core/settings.json")
	c.lacks("bootstrap: open a new terminal")
	after := snapshot(c.home, c.user, c.cfg, c.manifest)
	c.check("nothing written in HOME, a link source, the config folder or the manifest", before == after)
}

// caseDefaultOutput: the default output of the engine, with no terminal.
func caseDefaultOutput(c *mcase) {
	c.setEnv("BOOTSTRAP_VERBOSE", "FORCE_COLOR", "NO_COLOR")
	c.useConfig("ok")
	c.runReal()
	c.exit(0)
	c.has("new ~/.config/core/plain.txt")
	c.check("the first run has the last line", c.logHasRe(`^\[OK\] Synced: [0-9]+ files changed( · .*)?$`))
	c.check("the first run has no table", !hasStr(c.log, "| Total"))
	c.write(c.home+"/.config/core/values.txt", "edited\n")
	c.remove(c.home + "/.config/core/plain.txt")
	c.runReal()
	c.exit(0)
	c.has("changed ~/.config/core/values.txt")
	c.has("new ~/.config/core/plain.txt")
	c.lacks(tilde + "/.config/opt/opt.txt")
	c.lacks("unchanged  ")
	c.lacks("using bootstrap")
	c.check("the info line", c.logHasRe(`^\[INFO\] bootstrap `+treeCommit[:7]+` · [0-9]+ tools`))
	c.check("one last line, with no table", countRe(c.log, `^\| Total `) == 0)
	c.check("the last line: 2 files changed, 1 new, no failed or seeded or removed",
		c.logHasRe(`^\[OK\] Synced: 2 files changed · 1 new( · [0-9]+ (unchanged|kept seeds|skipped))*$`))
	c.syncWith([]string{"FORCE_COLOR=1"}, "--no-bootstrap")
	c.exit(0)
	c.has("↻ bootstrap ")
	c.has("✔")
	c.has("Synced: no file changed")
	c.lacks("┌───────────────┬───────┐")
}

// ---- contributions ----------------------------------------------------------

func caseContrib(c *mcase) {
	o := c.home + "/.config/owner"
	c.useConfig("contrib")
	c.runReal()
	c.exit(0)
	c.file(o+"/config", ownerConfig("Include ~/a", "Include ~/b", "Include ~/c", "", "Host *",
		"  IdentityAgent agent-a", "nested=default,from-b", "# again: agent-a"))
	c.file(o+"/data.json", `{"agent": "agent-a", "includes": ["~/a","~/b","~/c"]}`)
	c.file(o+"/data.toml", synced("owner/files/.config/owner/data.toml.tmpl")+"\nagent = \"agent-a\"\nincludes = [\"~/a\",\"~/b\",\"~/c\"]")
	c.check("the yaml target is written", hasLine(o+"/data.yaml", `  - "~/c"`))
	c.check("a format = toml target with another extension is written", isFile(o+"/settings.conf"))
	c.check("a format = text .json target is copied unparsed",
		cmpFiles(fixtures+"/repo/tools/owner/files/.config/owner/lenient.json.tmpl", o+"/lenient.json"))
	c.check("the contribution to a tool that is not selected is ignored", !exists(c.home+"/.config/opt"))
	c.has("skip file  ~/.config/owner/needs.txt (no ~/.config/owner-needed)")
	c.check("a generated file waits for its requires", !exists(o+"/needs.txt"))
	c.touch(c.home + "/.config/owner-needed")
	c.touch(c.home + "/.config/core-needed")
	c.runReal()
	c.exit(0)
	c.check("a generated file is written once its requires exist", isFile(o+"/needs.txt"))
	c.check("needs.txt is in the manifest", hasLine(c.manifest, "owner\t.config/owner/needs.txt"))
	c.runReal()
	c.exit(0)
	c.has("unchanged  ~/.config/owner/config")
	c.check("a run with unchanged inputs reports only unchanged",
		!c.logHasRe(`^bootstrap: (updated|chmod|removed|wrote|set |skip|would)`))
}

func caseContribDeselect(c *mcase) {
	o := c.home + "/.config/owner"
	c.useConfig("contrib")
	c.runReal()
	c.exit(0)
	c.file(o+"/seed.txt", "agent=agent-a")
	c.valuesFile("[tools.contrib-b]", "[tools.owner]", `token = "token-1"`)
	c.runReal()
	c.exit(0)
	c.file(o+"/config", ownerConfig("Include ~/a", "Include ~/c", "Include ~/b", "", "Host *",
		"  IdentityAgent default-agent", "nested=default,from-b", "# again: default-agent"))
	c.has("kept seed  ~/.config/owner/seed.txt")
	c.check("the seed keeps the contribution of its first write", fileIs(o+"/seed.txt", "agent=agent-a"))
}

// ---- parse checks -----------------------------------------------------------

// noTempFile tells whether no file below <dir> has a name that starts
// with ".bootstrap.".
func noTempFile(dir string) bool {
	found := false
	_ = filepath.WalkDir(dir, func(p string, e fs.DirEntry, err error) error {
		if err == nil && strings.HasPrefix(e.Name(), ".bootstrap.") {
			found = true
		}
		return nil
	})
	return !found
}

func caseParseBad(c *mcase) {
	d := c.home + "/.config/badparse"
	c.useConfig("badparse")
	for _, f := range []string{"json", "yaml", "toml"} {
		c.write(d+"/bad."+f, "old "+f+"\n")
	}
	c.runReal()
	c.exit(2)
	for _, f := range []string{"json", "yaml", "toml"} {
		c.has(tilde + "/.config/badparse/bad." + f + ": the result does not parse as " + f + " - the current file stays")
		c.file(d+"/bad."+f, "old "+f)
	}
	c.check("no temp file is left", noTempFile(d))
}

// targetJSON gives the parsed target <p> of format <format> as compact JSON
// with sorted keys. yq of mise parses it, offline when it is installed, as
// the tool function of the old task did; Go sorts the keys (jq -cS).
func (c *mcase) targetJSON(p, format string) string {
	script := `cd / || exit 1
if [[ -n $(MISE_OFFLINE=1 mise ls --installed yq 2> /dev/null < /dev/null) ]]; then
  MISE_OFFLINE=1 mise x yq -- yq -p "$2" -o=json -I=0 '.' "$1"
else
  mise x yq@latest -- yq -p "$2" -o=json -I=0 '.' "$1"
fi`
	cmd := exec.Command("bash", "-c", script, "_", p, format)
	cmd.Env = c.env
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	var v any
	if err := json.Unmarshal(out, &v); err != nil {
		return ""
	}
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return ""
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// caseValueBreakout: each target parses to the same keys as with a plain
// value. The double quote and the newlines of the value stay in the one
// string of its key.
func caseValueBreakout(c *mcase) {
	agent := `"x\"\nadded = true\n[added_table]\nkey = \"1"`
	o := c.home + "/.config/owner"
	c.useConfig("breakout")
	c.runReal()
	c.exit(0)
	for _, f := range []string{"json", "yaml", "toml"} {
		c.check("data."+f+" holds the value as one string, with no added key or table",
			c.targetJSON(o+"/data."+f, f) == `{"agent":`+agent+`,"includes":["~/a"]}`)
	}
	c.check("settings.conf holds the value as one string, with no added key or table",
		c.targetJSON(o+"/settings.conf", "toml") == `{"owner":{"agent":`+agent+`}}`)
}

// ---- shared files (mode = "merge") ------------------------------------------

func (c *mcase) mDir() string { return c.home + "/.config/merger" }

// live writes a live shared file m.<ext> at test time.
func (c *mcase) live(ext, content string) { c.write(c.mDir()+"/m."+ext, content) }

// hdr is the shared header of the file m.<ext>.
func hdr(ext string) string {
	return "# Shared with bootstrap: macos/tools/merger/files/.config/merger/m." + ext + ".merge.toml.\n" +
		"# A change in this file stays. A removed bootstrap item comes back at the next\n" +
		"# sync, unless macos.toml has an override for it."
}

func caseMergeJSON(c *mcase) {
	c.useConfig("merge")
	f := c.mDir() + "/m.json"
	c.live("json", `{
    "keep": "app",
    "scalar": "app",
    "nested": {"set": 0, "other": 2},
    "kind": {"was": "table"},
    "list": ["x"],
    "maps": [{"a": 1, "b": 9, "v": 5}],
    "gone": true
}
`)
	c.chmod(f, 0o600)
	c.runReal()
	c.exit(0)
	c.has("updated    ~/.config/merger/m.json")
	c.has("~/.config/merger/m.json: kind is a table, not a list - written again")
	c.file(f, `{
  "keep": "app",
  "scalar": "app",
  "nested": {
    "set": 0,
    "other": 2,
    "deeper": {
      "x": true
    }
  },
  "kind": [
    1
  ],
  "list": [
    "x",
    "b"
  ],
  "maps": [
    {
      "a": 1,
      "b": 2,
      "v": 5
    }
  ],
  "items": [
    "a"
  ]
}`)
	c.check("the permissions are kept", modeOf(f) == "600")
	c.check("the manifest records an added item", hasStr(c.manifest, "merger\t.config/merger/m.json\t{\"kind\":\"item\",\"path\":[\"list\"],\"value\":\"b\"}"))
	c.check("the manifest does not record a found key", !hasStr(c.manifest, "m.json\t{\"kind\":\"key\",\"path\":[\"scalar\"]}"))
	c.check("the manifest has no file row of a shared file", !hasLine(c.manifest, "merger\t.config/merger/m.json"))
	c.runReal()
	c.exit(0)
	c.has("unchanged  ~/.config/merger/m.json")
	c.check("a second run writes nothing", !c.logHasRe(`^bootstrap: (updated|chmod).*merger/m\.`))
	c.write(f, "{\"keep\": \"edited\"}\n")
	before := catT(f)
	c.runDry()
	c.exit(0)
	c.has("would update ~/.config/merger/m.json")
	c.has(`+  "scalar": "src",`)
	c.file(f, before)
}

func caseMergeYAML(c *mcase) {
	c.useConfig("merge")
	c.live("yaml", `# top comment
keep: app # after keep
scalar: app
nested:
  # own line
  set: 0
list:
  - x
gone: 1
`)
	c.runReal()
	c.exit(0)
	c.file(c.mDir()+"/m.yaml", hdr("yaml")+`
# top comment
keep: app # after keep
scalar: app
nested:
  # own line
  set: 0
list:
  - x
  - b`)
}

func caseMergeTOML(c *mcase) {
	c.useConfig("merge")
	f := c.mDir() + "/m.toml"
	c.live("toml", `# head comment
keep = "app" # after keep
scalar = "app"
gone = 1

# own line
[nested]
set = 0
list = ["x"]
`)
	c.runReal()
	c.exit(0)
	c.check("the header is on top", hasLine(f, "# Shared with bootstrap: macos/tools/merger/files/.config/merger/m.toml.merge.toml."))
	c.check("the head comment stays", hasLine(f, "# head comment"))
	c.check("the own-line comment stays", hasLine(f, "# own line"))
	c.check("the after-key comment stays", hasRe(f, `^keep = "app" +# after keep$`))
	c.check("a found key keeps its value", hasLine(f, `scalar = "app"`))
	c.check("a found nested key keeps its value", hasLine(f, "set = 0"))
	c.check("a block is written again", hasLine(f, `list = ["x", "y"]`))
	c.check("a removed key", !hasRe(f, `^gone`))
}

func caseMergeText(c *mcase) {
	c.useConfig("merge")
	c.live("txt", "# mine\ntwo=old\nown\n")
	c.runReal()
	c.exit(0)
	c.file(c.mDir()+"/m.txt", hdr("txt")+"\n# mine\ntwo=old\nown\nline-one")
}

func caseMergeAbsent(c *mcase) {
	c.useConfig("merge")
	c.runReal()
	c.exit(0)
	c.has("updated    ~/.config/merger/m.json")
	c.file(c.mDir()+"/m.json", `{
  "scalar": "src",
  "nested": {
    "set": 1,
    "deeper": {
      "x": true
    }
  },
  "kind": [
    1
  ],
  "list": [
    "b"
  ],
  "maps": [
    {
      "a": 1,
      "b": 2,
      "v": 0
    }
  ],
  "items": [
    "a"
  ]
}`)
	c.file(c.mDir()+"/m.yaml", hdr("yaml")+"\nscalar: src\nnested:\n  set: 1\nlist:\n  - b")
	c.file(c.mDir()+"/m.txt", hdr("txt")+"\nline-one\ntwo=2")
	c.check("the toml block is written", hasLine(c.mDir()+"/m.toml", `list = ["x", "y"]`))
	c.check("a new target takes the mode of its source", modeOf(c.mDir()+"/m.json") == "644")
}

func caseMergeEmpty(c *mcase) {
	c.useConfig("merge")
	c.live("json", "")
	c.live("yaml", "  \n# only a comment\n\n")
	c.live("toml", "\n\t \n")
	c.chmod(c.mDir()+"/m.json", 0o600)
	c.runReal()
	c.exit(0)
	c.has("updated    ~/.config/merger/m.json")
	c.has("updated    ~/.config/merger/m.yaml")
	c.has("updated    ~/.config/merger/m.toml")
	c.check("the json items are written", hasStr(c.mDir()+"/m.json", `"scalar": "src"`))
	c.check("the permissions of the empty file are kept", modeOf(c.mDir()+"/m.json") == "600")
	c.check("the yaml items are written", hasLine(c.mDir()+"/m.yaml", "scalar: src"))
	c.check("the toml items are written", hasLine(c.mDir()+"/m.toml", `scalar = "src"`))
}

func caseMergeLiveBad(c *mcase) {
	j, y, t := `{"a": `, "a: [1\n", "a = [\n"
	c.useConfig("merge")
	c.live("json", j)
	c.live("yaml", y)
	c.live("toml", t)
	c.runReal()
	c.exit(2)
	c.has("/.config/merger/m.json is not valid json - kept it as it is")
	c.has("/.config/merger/m.yaml is not valid yaml - kept it as it is")
	c.has("/.config/merger/m.toml is not valid toml - kept it as it is")
	c.same(c.mDir()+"/m.json", j)
	c.same(c.mDir()+"/m.yaml", y)
	c.same(c.mDir()+"/m.toml", t)
	c.check("no temp file is left", noTempFile(c.mDir()))
}

func caseMergeLiveKind(c *mcase) {
	j, y := "[1]\n", "- 1\n"
	c.useConfig("merge")
	c.live("json", j)
	c.live("yaml", y)
	c.runReal()
	c.exit(2)
	c.has("/.config/merger/m.json holds a list, not a table - kept it as it is")
	c.has("/.config/merger/m.yaml holds a list, not a table - kept it as it is")
	c.same(c.mDir()+"/m.json", j)
	c.same(c.mDir()+"/m.yaml", y)
}

func caseMergeLiveDocs(c *mcase) {
	y := "a: 1\n---\nb: 2\n"
	c.useConfig("merge")
	c.live("yaml", y)
	c.runReal()
	c.exit(2)
	c.has("/.config/merger/m.yaml has more than one YAML document - kept it as it is")
	c.same(c.mDir()+"/m.yaml", y)
}

// caseMergeSpecBad: a spec that the merge cannot use keeps the live file.
func caseMergeSpecBad(c *mcase) {
	j := "{\"keep\": 1}\n"
	c.useConfig("merge")
	c.caseTree(func(c *mcase) {
		c.write(c.work+"/tools/merger/files/.config/merger/m.json.merge.toml", "zzz = 1\n")
	})
	c.live("json", j)
	c.runReal()
	c.exit(2)
	c.has("macos/tools/merger/files/.config/merger/m.json.merge.toml: zzz is not a key of a merge spec")
	c.same(c.mDir()+"/m.json", j)
}

func caseMergeOverrides(c *mcase) {
	c.useConfig("merge")
	c.valuesFile("[tools.merger]",
		`[tools.merger.files.".config/merger/m.json"]`, `scalar = "own"`, `kind = { "$skip" = true }`,
		`[tools.merger.files.".config/merger/m.json".list]`, `b = { "$mode" = "remove" }`)
	c.live("json", `{"list": ["b", "x"]}`)
	c.runReal()
	c.exit(0)
	c.check("an own value is written", hasStr(c.mDir()+"/m.json", `"scalar": "own"`))
	c.check("a skipped key is not added", !hasStr(c.mDir()+"/m.json", `"kind"`))
	c.check("a removed item goes", listOf(c.mDir()+"/m.json", "list") == `["x"]`)
	c.valuesFile("[tools.merger]", `[tools.merger.files.".config/merger/m.json"]`, "nope = 1")
	c.runReal()
	c.exit(2)
	c.has("the override nope has no item of bootstrap - remove it")
}

func caseMergeLink(c *mcase) {
	c.useConfig("merge")
	c.mkdir(c.mDir())
	c.write(c.user+"/m.json", "{\"keep\": \"app\"}\n")
	c.symlink(c.user+"/m.json", c.mDir()+"/m.json")
	c.symlink(c.user+"/missing/m.yaml", c.mDir()+"/m.yaml")
	c.runReal()
	c.exit(0)
	c.check("the link stays", isLink(c.mDir()+"/m.json"))
	c.check("its source keeps the app key", hasStr(c.user+"/m.json", `"keep": "app"`))
	c.check("its source gets the items", hasStr(c.user+"/m.json", `"scalar": "src"`))
	c.check("a link with a missing source writes the items", isFile(c.user+"/missing/m.yaml") && isLink(c.mDir()+"/m.yaml"))
}

func caseMergeContrib(c *mcase) {
	c.useConfig("merge")
	c.valuesFile("[tools.merger]", "[tools.merge-contrib]")
	c.runReal()
	c.exit(0)
	c.check("the contribution reaches the shared file", listOf(c.mDir()+"/m.json", "items") == `["a","z"]`)
	c.valuesFile("[tools.merger]")
	c.runReal()
	c.exit(0)
	c.check("the item of a tool that went is removed", listOf(c.mDir()+"/m.json", "items") == `["a"]`)
	c.useConfig("-")
	c.runReal()
	c.exit(0)
	c.check("the shared files of a deselected tool stay", isFile(c.mDir()+"/m.json") && isFile(c.mDir()+"/m.yaml"))
}
