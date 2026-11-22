// Package mac tests `bootstrap mac sync` and the other Mac commands against
// the fixtures in macos/fixtures/, and the entry task updateall:configs of
// macos/tools/mise/ against a local bare origin. The cases came from the task
// test:mac (.config/mise/tasks/test/mac, all 2,499 lines).
//
// Each case has its own temporary HOME, XDG_* folders, cache, TMPDIR, git
// config and folder of macos.toml ($XDG_CONFIG_HOME/bootstrap). MISE_DATA_DIR and
// MISE_CACHE_DIR stay the folders of the user, so a real `mise x` finds its
// tools with no download.
package mac

import (
	"fmt"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
)

func TestMain(m *testing.M) { harness.Main(m, setup, "macos") }

// The tilde at the start of a target in a message.
const tilde = "~"

func TestMac(t *testing.T) {
	type tc struct {
		name string
		body func(c *mcase)
	}
	cases := []tc{
		{"config: a missing config folder", caseConfigMissing},
		{"config: macos.toml below HOME", caseConfigHome},
		{"config: a synced Mac with no macos.toml", caseConfigLost},
		{"config: an unknown flag, with the usage", caseUnknownFlag},
		{"config: sync migrates an unknown tool in a [tools.<name>] table away", expect("unknown-tool", 0, "[tools.nope] removed: bootstrap no longer has the tool", nil)},
		{"config: an [overrides.mise.<path>] that no selected tool has", expect("unknown-override", 2, "no selected tool has", nil)},
		{"config: a packs list, the old format", expect("old-packs-list", 2, "must be [packs.<name>] tables", nil)},
		{"config: a tools key that is a string, the old format", expect("tools-string", 2, "tools = [...] is the old format", nil)},
		{"config: a tools item that is not a name, the old format", expect("tools-number", 2, "tools = [...] is the old format", nil)},
		{"config: a tools list, the old format", expect("tools-list", 2, "tools = [...] is the old format", nil)},
		{"config: an unknown top-level key", expect("bad-key", 2, "colour is not a key of xdg/config/bootstrap/macos.toml", nil)},
		{"config: an [overrides.<tool>] table", expect("bad-override", 2, "[overrides.opt] in xdg/config/bootstrap/macos.toml is not valid - write [tools.<name>] for template values and [overrides.mise.<path>] for mise config", nil)},
		{"config: a [tools.<name>] table with env", expect("env-opt", 2, "[tools.opt] env is for repo tools only", nil)},
		{"config: an [overrides.mise_env.<env>.<path>] table", expect("override-env", 2, "[overrides.mise_env] in xdg/config/bootstrap/macos.toml: the Mac side has no environments", nil)},
		{"config: a tool folder with no tools.toml", expect("ok", 2, "macos/tools/bad has no tools.toml", addBadTool)},
		{"config: a repo with no macos/packs.toml", expect("ok", 3, "base is not a bootstrap pack (named in macos/packs.toml)", noPacks)},
		{"config: a packs.toml with no [pack.base]", expect("ok", 3, "base is not a bootstrap pack (named in macos/packs.toml)", noBase)},
		{"config: required = true in a pack other than base", expect("ok", 2, "macos/packs.toml: [pack.more] opt: required = true is only for a tool of pack base", packs("[pack.more]", "opt = { required = true }"))},
		{"config: the old member key default", expect("ok", 2, "macos/packs.toml: [pack.more] opt has the entry key 'default', the old form - each member is a default: write opt = {}", packs("[pack.more]", "opt = { default = true }"))},
		{"config: an unknown key in a tools.toml", expect("ok", 2, "macos/tools/opt/tools.toml: depend is not a key of tools.toml - the keys are description, summary, category, selectable, stage, depends, parent, runtime, [defaults], [questions], [contributes], [files], [tasks] and [mise]", prepend("opt", `depend = [ "core" ]`))},
		{"config: a stage that is not init", expect("ok", 2, `macos/tools/opt/tools.toml: stage = "login" - the only stage is "init"`, prepend("opt", `stage = "login"`))},
		{"config: the old stage = shell", expect("ok", 2, `macos/tools/opt/tools.toml: stage = "shell" is the old name - write stage = "init"`, prepend("opt", `stage = "shell"`))},
		{"config: the old seed key", expect("ok", 2, `macos/tools/core/tools.toml: seed is the old key - write a [files."<target>"] table with mode = "seed"`, prepend("core", `seed = [ ".config/core/seeded.toml" ]`))},
		{"config: the old [seed_requires] table", expect("ok", 2, `macos/tools/core/tools.toml: [seed_requires] is the old table - write requires = [...]`, appendTo("core", "[seed_requires]", `".config/core/seeded.toml" = [ ".config/core-needed" ]`))},
		{"config: a [files] table with no source", expect("ok", 2, `macos/tools/core/tools.toml: [files.".config/core/nothing"] needs source, a path below files/ with no .. part`, appendTo("core", `[files.".config/core/nothing"]`, `mode = "seed"`))},
		{"config: an unknown key in a [files] table", expect("ok", 2, `macos/tools/opt/tools.toml: [files.".config/opt/opt.txt"] owner is not a key of a files entry`, appendTo("opt", `[files.".config/opt/opt.txt"]`, `owner = "core"`))},
		{"config: an unknown key in a [questions] table", expect("ok", 2, `macos/tools/opt/tools.toml: [questions.level] ask is not a key of a question - the keys are question, detect, example, group, note and optional`, badQuestion)},
		{"config: a [questions] table with no question", expect("ok", 2, `macos/tools/opt/tools.toml: [questions.level] needs a question string`, badQuestion)},
		{"config: a tool with default_envs", expect("ok", 2, "macos/tools/opt/tools.toml: default_envs is for repo tools only", prepend("opt", `default_envs = [ "dev" ]`))},
		{"config: a tool with [mise_env.dev.tools]", expect("ok", 2, "macos/tools/opt/tools.toml: mise_env is for repo tools only", appendTo("opt", "[mise_env.dev.tools]", `jq = "latest"`))},
		{"config: a [files] table with requires_found", expect("ok", 2, `macos/tools/opt/tools.toml: [files.".config/opt/opt.txt"] requires_found is for repo tools only`, appendTo("opt", `[files.".config/opt/opt.txt"]`, `requires_found = [ "x" ]`))},
		{"config: a tool folder zz-macos", expect("ok", 2, "tool folder macos/tools/zz-macos sorts after zz-macos or is named zz-macos", newTool("zz-macos"))},
		{"config: a tool folder that sorts after zz-macos", expect("ok", 2, "tool folder macos/tools/zzz sorts after zz-macos or is named zz-macos", newTool("zzz"))},
		{"config: a tool that ships a file in conf.d", expect("ok", 2, "macos/tools/core/files/.config/mise/conf.d/x.toml: the sync generates .config/mise/conf.d/ from [mise.*] of tools.toml", shipsConfD)},
		{"config: two tools that ship one target", expect("ok", 2, tilde+"/.config/core/plain.txt comes from more than one source", twoToolsOneTarget)},
		{"config: a pack name that is a tool folder", expect("ok", 2, "macos/packs.toml: pack 'core' is also a tool folder", packs("[pack.core]", "init = {}"))},
		{"config: a pack member that is not a tool", expect("ok", 3, "nope is not a bootstrap tool (named in pack more)", packs("[pack.more]", "nope = {}"))},
		{"config: a pack-level key", expect("ok", 2, "macos/packs.toml: [pack.more] depends is the old format", packs("[pack.more]", `depends = [ "opt" ]`))},
		{"config: a pack entry key other than default", expect("ok", 2, "macos/packs.toml: [pack.more] opt has the entry key 'tag'", packs("[pack.more]", "opt = { tag = true }"))},
		{"config: a pack entry that is not a table", expect("ok", 2, "macos/packs.toml: [pack.more] tools must be a table", packs("[pack.more]", `tools = "init"`))},
		{"catalog: summary, category and selectable", passTree(catalogKeys)},
		{"catalog: an unknown category", expect("ok", 2, "macos/tools/opt/tools.toml: category nope is not a category of macos/categories.toml", func(c *mcase) { c.setKey("opt", "category", `category = "nope"`) })},
		{"catalog: a category and no macos/categories.toml", expect("ok", 2, "macos/tools/opt/tools.toml: category dev needs the file macos/categories.toml", func(c *mcase) { c.remove(c.work + "/categories.toml") })},
		{"catalog: a key other than title in macos/categories.toml", expect("ok", 2, "macos/categories.toml: [category.dev] color is not a key of a category - the only key is title", func(c *mcase) {
			c.write(c.work+"/categories.toml", harness.Lines("[category.dev]", `color = "red"`, `title = "Development"`))
		})},
		{"catalog: a tool with no summary", expect("ok", 2, "macos/tools/opt/tools.toml: summary is missing", func(c *mcase) { c.dropKey("opt", "summary") })},
		{"catalog: a tool with no category", expect("ok", 2, "macos/tools/opt/tools.toml: category is missing", func(c *mcase) { c.dropKey("opt", "category") })},
		{`catalog: a tool with summary = ""`, expect("ok", 2, "macos/tools/opt/tools.toml: summary is missing", func(c *mcase) { c.setKey("opt", "summary", `summary = ""`) })},
		{"catalog: hosts on a Mac tool", expect("ok", 2, "macos/tools/opt/tools.toml: hosts is for repo tools only", prepend("opt", `hosts = [ "macos" ]`))},
		{"catalog: a pack description", passTree(packs("[pack.more]", `description = "More tools."`, "opt = {}"))},
		{"catalog: a pack description that is not a string", expect("ok", 2, "macos/packs.toml: [pack.more] description must be a string", packs("[pack.more]", "description = 1", "opt = {}"))},
		{"questions: a group with no answer", expect("ok", 2, "the group opt.email has no answer in macos.toml - run bootstrap mac migrate", appendTo("opt", "[questions.email_a]", `group = "email"`, `note = "One answer is enough."`, `question = "Which a?"`,
			"", "[questions.email_b]", `group = "email"`, `question = "Which b?"`))},
		{"config: a contribution to a name that is not a tool", expect("contrib-bad", 3, "macos/tools/contrib-bad/tools.toml: [contributes.nope] names a tool that bootstrap does not have", nil)},
		{"config: two contributors that set one scalar", expect("conflict", 2, "macos/tools/contrib-a/tools.toml and macos/tools/contrib-c/tools.toml set [contributes.owner] agent to different values", nil)},
		{"config: a scalar over a list of a base tool", expect("type-list", 2, "[tools.core] of macos.toml sets items to a string, but it is a list in the base or dependency tool core", nil)},
		{"config: a scalar over a table of a base tool", expect("type-table", 2, "[tools.core] of macos.toml sets table to a string, but it is a table in the base or dependency tool core", nil)},
		{"selection: a [tools.<name>] table selects one tool of a pack", caseToolsList},
		{"packs: add pack writes a table for each member not selected", caseAddPack},
		{"packs: add pack --exclude", caseAddPackExclude},
		{"packs: add pack of a nested pack", caseAddPackNested},
		{"packs: add pack of two packs", caseAddPackTwo},
		{"packs: --exclude of a required member", addPackRefused("base --exclude mise", "mise is required in the pack base, so mac add pack cannot exclude it")},
		{"packs: --exclude of a name that is not a member", addPackRefused("extra --exclude core", "core is not a member of the pack extra")},
		{"packs: --exclude with two packs", addPackRefused("extra outer --exclude opt", "--exclude needs exactly one pack name - run bootstrap mac add pack once for each pack")},
		{"packs: --exclude of a member that another member needs", addPackRefused("dev --exclude initdep", "initdep is needed by init - exclude init too, or keep initdep")},
		{"old form: sync migrates first, check stops on a file without format", caseSyncNoFormat},
		{"old form: init migrates [packs.*] with the frozen copy", caseInitMigratePacks},
		{"base: sync writes the table of a required member", caseRequiredTable},
		{"mise check: a mise that is not ~/.local/bin/mise", caseMiseElsewhere},
		{"mise check: a ~/.local/bin/mise link into another folder", caseMiseLink},
		{"mise: the [mise.env] of a tool, rendered", caseMiseTool},
		{"mise: the [mise.shell_alias] of macos.toml", caseMiseUser},
		{"mise: an [overrides.mise.<path>] changes a key", caseMiseOverride},
		{"mise: an override path that no selected tool has", expect("override-none", 2, "no selected tool has", nil)},
		{"mise: an [overrides.gitignore] table", expect("override-gitignore", 2, "[overrides.gitignore] in xdg/config/bootstrap/macos.toml is not valid", nil)},
		{"mise: a tool no longer selected loses its mise config", caseMiseDeselected},
		{"blockers: a file with the header where a folder must be", caseBlockerHeader},
		{"blockers: a file without the header where a folder must be", caseBlockerPlain},
		{"values: the merge order, a base tool list appends", caseValues},
		{"values: a dependency tool list appends", caseValuesDependency},
		{"values: an opt-in tool list appends", caseValuesOptIn},
		{"values: an opt-in tool value of another type replaces", caseValuesOptInType},
		{"render: ~/.config/mise/** first", caseRenderOrder},
		{"render: the output equals the source", caseRenderExact},
		{"render: a seed", caseSeed},
		{"render: a generated file overwritten", caseOverwrite},
		{"context: .cli_dir and .env", caseContext},
		{"backup: links the files into the clone", caseBackupLinks},
		{"backup: a dry run shows the links", caseBackupDryRun},
		{"backup: a conflict stops the sync", caseBackupConflict},
		{"backup: a missing clone", caseBackupMissing},
		{"symlinks: a link to a file", caseLinkFile},
		{"symlinks: a symlinked folder", caseLinkFolder},
		{"symlinks: a stale file through a link", caseLinkStale},
		{"stale: removed with its empty parents, kept without the header", caseStale},
		{"stale: a listed file with the old header", caseStaleOldHeader},
		{"stale: a tool no longer selected", caseStaleDeselected},
		{"manifest: its first line, the sorted rows, a second run", caseGenerated},
		{"first setup: migrate writes the init stage and runs mise bootstrap", caseFirstSetup},
		{"first setup: --no-bootstrap", caseFirstSetupNoBootstrap},
		{"first setup: a failed file", caseFirstSetupFail},
		{"first setup: a sync before the first setup", caseSyncBeforeFirstSetup},
		{"sync: a chmod-only change of an init tool", caseInitChmod},
		{"scripts: the derive values and mise table", caseScriptsDerive},
		{"scripts: a check that fails stops the run", caseScriptsCheck},
		{"scripts: an after script asks for a restart", caseScriptsAfter},
		{"scripts: a derive output that is not JSON", expect("scripted", 2, "macos/tools/scripted/scripts/derive did not print a JSON object", func(c *mcase) {
			c.write(c.work+"/tools/scripted/scripts/derive", harness.Lines("cat > /dev/null", "echo nope"))
		})},
		{"scripts: a file of scripts/ that is not a script", expect("scripted", 2, "scripts/build is not a script of a tool", func(c *mcase) {
			c.write(c.work+"/tools/scripted/scripts/build", "x\n")
		})},
		{"ssh folder: a new folder", caseSSHNew},
		{"ssh folder: an existing folder with another mode", caseSSHExisting},
		{"ssh folder: a linked file below the folder", caseSSHLink},
		{"ssh folder: a dry run", caseSSHDry},
		{"ssh folder: the folder is a link", caseSSHFolderLink},
		{"contributions: merged, placed twice, lists without duplicates, requires", caseContrib},
		{"contributions: a contributor deselected, a seed written once", caseContribDeselect},
		{"parse: a JSON, a YAML and a TOML target that do not parse", caseParseBad},
		{"parse: a value with a double quote and a newline stays one value", caseValueBreakout},
		{"macos.toml missing", caseValuesMissing},
		{"bootstrap_commit: after the header comments, a second run", caseCommit},
		{"bootstrap_commit: a run without a change keeps it", caseCommitKept},
		{"bootstrap_commit: a run with a change sets it", caseCommitMoved},
		{"bootstrap_commit: a symlinked macos.toml", caseCommitLink},
		{"bootstrap_commit: a dry run", caseCommitDry},
		{"--dry-run writes nothing", caseDryRun},
		{"the default output: the list and the table", caseDefaultOutput},
		{"merge: a JSON target", caseMergeJSON},
		{"merge: a YAML target keeps its comments", caseMergeYAML},
		{"merge: a TOML target keeps its comments", caseMergeTOML},
		{"merge: a text target by line", caseMergeText},
		{"merge: absent targets get the items", caseMergeAbsent},
		{"merge: empty targets count as absent", caseMergeEmpty},
		{"merge: live targets that do not parse", caseMergeLiveBad},
		{"merge: live targets that are not a table", caseMergeLiveKind},
		{"merge: a live YAML target with two documents", caseMergeLiveDocs},
		{"merge: a spec that cannot be used", caseMergeSpecBad},
		{"merge: the overrides of macos.toml", caseMergeOverrides},
		{"merge: targets behind symlinks", caseMergeLink},
		{"merge: a contribution comes and goes, then a deselected tool", caseMergeContrib},
		{"merge: a source that is not a merge spec", expect("merge", 2, `macos/tools/merger/tools.toml: [files.".config/merger/plain.txt"] the source of a shared file is a merge spec <name>.merge.toml`, appendTo("merger", `[files.".config/merger/plain.txt"]`, `mode = "merge"`))},
		{"merge: the old key remove", expect("merge", 2, `macos/tools/merger/tools.toml: [files.".config/merger/m.json"] remove is not a key of a files entry`, appendTo("merger", `[files.".config/merger/m.json"]`, `remove = [ [ "a" ] ]`))},

		{"questions: sync stops on a question without an answer", caseQuestionMissing},
		{"questions: mac questions, --json and text", caseQuestionsJSON},
		{"questions: --json on mac add, mac check and mac describe", caseMacJSONRefused},
		{"selection: mac detect offers the installed packages", caseMacDetect},
		{"selection: mac detect --add --yes adds the tools and the packages", caseMacDetectYes},
		{"selection: mac detect without brew and mas", caseMacDetectNobrew},
		{"selection: mac detect does not offer a tool with selectable = false", caseMacDetectNotSelectable},
		{"selection: mac detect --no and --yes without --add", caseMacDetectFlags},
		{"selection: mac add", caseMacAdd},
		{"selection: mac remove", caseMacRemove},
		{"selection: mac remove refuses a required tool, a needed tool and a tool that is not selected", caseMacRemoveRefused},
		{"selection: mac replace refuses tools without the same parent", caseMacReplaceRefused},
		{"selection: the old grammar without a kind word", caseMacOldGrammar},
		{"selection: a sync that fails after mac add tool puts macos.toml back", caseMacAddFails},
		{"init: --answer answers a question, then the sync", caseInitAnswer},
		{"init: --yes and a question without an answer", caseInitYesMissing},
		{"init: no --yes and no terminal", caseInitNoYes},
		{"init: --dry-run lists the changes", caseInitDryRun},
		{"init: sync migrates the old form first", caseSyncOldForm},
		{"init: migrates [tools.macos-setup], then sync", caseInitMigrate},
		{"init: a macos.toml with changes that are not committed", caseInitDirtyValues},
		{"init: a new macos.toml", caseInitNew},
		{"init: tool names", caseInitNames},
		{"init: from another git repository", caseInitOtherRepo},
		{"init: --backup writes the key and links", caseInitBackup},
		{"init: nothing to change", caseInitNothing},

		{"entry clone: a first clone", caseEntryCloneFirst},
		{"entry clone: a refresh after a new origin commit", caseEntryCloneRefresh},
		{"entry clone: a clone with another origin", caseEntryCloneOtherOrigin},
		{"entry clone: the SSH and https forms of one URL", caseEntryCloneSSHHTTPS},
		{"entry clone: a non-empty non-git folder", caseEntryCloneNotGit},
		{"entry clone: an origin that cannot be cloned", caseEntryCloneFails},
		{"entry flags: --verbose", caseEntryVerbose},
		{"entry lock: a run waits for a live holder, then takes over", caseEntryLockWait},
		{"entry lock: a lock whose pid is dead", caseEntryLockDead},
		{"entry engine: a VERSION that is not MAJOR.MINOR.PATCH", caseEntryBadVersion},
		{"entry flags: an unknown flag directly", caseEntryUnknownFlag},
		{"entry flags: --dry-run directly and under mise run", caseEntryMiseRun},
		{"entry: a nested run", caseEntryNested},
		{"entry legacy: the old list moves, the old cache goes", caseEntryLegacyMove},
		{"entry legacy: an existing manifest stays", caseEntryLegacyKeepManifest},
		{"entry legacy: a live old lock keeps the old cache", caseEntryLegacyLive},
		{"entry: no lock and no temp copy after each run", caseEntryLeftovers},
	}
	// The last line of the old task. The cleanup runs after the parallel
	// cases.
	t.Cleanup(func() {
		if !t.Failed() {
			harness.Pass(t, fmt.Sprintf("All %d bootstrap mac cases passed", len(cases)))
		}
	})
	for _, k := range cases {
		runCase(t, k.name, k.body)
	}
}
