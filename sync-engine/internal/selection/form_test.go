package selection

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/sync-engine/internal/source"
)

// The cases of the tool folder form and of `required` in the packs file. The
// bash engine never had these rules, so each case holds its own expected
// output, not a golden file of bash.

type formCase struct {
	name string
	side source.Side
	over tree
	want string
}

var formCases = []formCase{
	{name: "repo-pack-entry-keys", side: source.Repo,
		over: tree{"repo/packs.toml": "[pack.base]\na = { default = true, env = 1 }\n"},
		want: "rc=2\n--- stderr\nrepo-sync: repo/packs.toml: [pack.base] a has the entry key 'env' - the entry keys are default and required\n--- stdout\n"},
	{name: "repo-required-ok", side: source.Repo,
		over: tree{"repo/packs.toml": "[pack.base]\na = { default = true, required = true }\n"},
		want: "rc=0\n--- stderr\n--- stdout\ntools: a\nsel:\na - default\ndeps:\nvalues:\n{\"a\":{}}\n"},
	{name: "repo-required-not-base", side: source.Repo,
		over: tree{"repo/packs.toml": "[pack.base]\na = { default = true }\n[pack.p]\nb = { default = true, required = true }\n"},
		want: "rc=2\n--- stderr\nrepo-sync: repo/packs.toml: [pack.p] b: required = true is only for a member of pack base with default = true\n--- stdout\n"},
	{name: "repo-required-not-default", side: source.Repo,
		over: tree{"repo/packs.toml": "[pack.base]\na = { required = true }\n"},
		want: "rc=2\n--- stderr\nrepo-sync: repo/packs.toml: [pack.base] a: required = true is only for a member of pack base with default = true\n--- stdout\n"},
	{name: "repo-required-type", side: source.Repo,
		over: tree{"repo/packs.toml": "[pack.base]\na = { default = true, required = 1 }\n"},
		want: "rc=2\n--- stderr\nrepo-sync: repo/packs.toml: [pack.base] a: required must be true or false, not int64\n--- stdout\n"},
	{name: "repo-entries-ok", side: source.Repo,
		over: tree{
			"repo/tools/a/tools.toml": rt("parent = \"b\"\n[files.\"x.json\"]\nsource = \"x.json.merge.toml\"\nmode = \"merge\"\n" +
				"[files.\"y.txt\"]\nsource = \"y.txt.tmpl\"\nmode = \"seed\"\nrender = true\nrequires = [ \"go.mod\" ]\nrequires_found = [ \"pubspec.yaml\" ]\n" +
				"[files.\".config/d\"]\nsource = \"d\"\n" +
				"[tasks.\"code/lint/10-a\"]\nsource = \"code-lint-a.tmpl\"\nmode = \"template\"\nrequires_found = [ \"go.mod\" ]\n" +
				"[tasks.\"_scripts/lib\"]\nsource = \"lib\"\nlibrary = true\n"),
			"repo/tools/a/files/x.json.merge.toml": "",
			"repo/tools/a/files/y.txt.tmpl":        "y",
			"repo/tools/a/files/d/one":             "1",
			"repo/tools/a/tasks/code-lint-a.tmpl":  "a",
			"repo/tools/a/tasks/lib":               "l",
			"repo/tools/b/tools.toml":              rt(""),
			"repo/tools/c/tools.toml":              rt(""),
			"repo/tools/d/tools.toml":              rt(""),
			"repo/tools/a/files/d/sub/.hidden.txt": "h",
		},
		want: "rc=0\n--- stderr\n--- stdout\ntools: a\nsel:\na - default\ndeps:\nvalues:\n{\"a\":{}}\n"},
	{name: "repo-entries-bad", side: source.Repo,
		over: tree{
			"repo/tools/a/tools.toml": rt("parent = \"a\"\n[files.\"x.txt\"]\nmode = \"copy\"\nformat = \"ini\"\nzzz = 1\n" +
				"[files.\"y.txt\"]\nsource = \"y.txt\"\nmode = \"merge\"\n" +
				"[files.\"../z\"]\nsource = \"../z\"\nrequires = [ \"/abs\" ]\nrequires_found = [ \"a/b\" ]\n" +
				"[files.\"w.txt\"]\nsource = \"w.txt\"\nmode = \"template\"\nrender = true\n" +
				"[files.\"v.txt\"]\nsource = \"nope\"\n" +
				"[tasks.\"run/x\"]\nsource = \"run-x\"\nmode = \"seed\"\nlibrary = true\n" +
				"[tasks.\"_scripts/y\"]\nsource = \"y\"\n"),
			"repo/tools/a/files/y.txt": "y",
			"repo/tools/a/files/w.txt": "w",
			"repo/tools/a/files/u.txt": "u",
			"repo/tools/a/tasks/run-x": "r",
			"repo/tools/a/tasks/y":     "y",
			"repo/tools/a/old.txt":     "o",
		},
		want: "rc=2\n--- stderr\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"../z\"] must be a path relative to the target root, with no .. part\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"../z\"] needs source, a path below files/ with no .. part\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"../z\"] requires: '/abs' must be a path relative to the repository root, with no .. part\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"../z\"] requires_found: 'a/b' must be a file name, with no /\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"v.txt\"] source: files/nope is not there\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"w.txt\"] render works only with mode = \"seed\" - a template has mode = \"template\"\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"x.txt\"] zzz is not a key of a files entry - the keys are source, mode, render, format, requires, requires_found, when, perm, dir_perm\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"x.txt\"] needs source, a path below files/ with no .. part\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"x.txt\"] format must be \"text\", \"json\", \"jsonc\", \"yaml\" or \"toml\"\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"y.txt\"] the source of a shared file is a merge spec <name>.merge.toml\n" +
			"repo-sync: repo/tools/a/tools.toml: [tasks.\"_scripts/y\"] a file below _scripts/ is a library - write library = true\n" +
			"repo-sync: repo/tools/a/tools.toml: [tasks.\"run/x\"] mode must be \"copy\" or \"template\"\n" +
			"repo-sync: repo/tools/a/tools.toml: [tasks.\"run/x\"] library = true is for a library below _scripts/, and each file below _scripts/ is a library\n" +
			"repo-sync: repo/tools/a/tools.toml: files/u.txt has no entry - write a [files.\"<path>\"] table with source = \"u.txt\"\n" +
			"repo-sync: repo/tools/a/tools.toml: old.txt is not below files/ or tasks/ - move it, and write an entry for it (the old form of a tool folder is gone)\n" +
			"repo-sync: repo/tools/a/tools.toml: parent must be the name of another tool of repo/tools\n" +
			"--- stdout\n"},
	{name: "repo-mise-depends", side: source.Repo,
		over: tree{"repo/tools/a/tools.toml": rt("[mise.tools.ta]\nversion = \"1\"\ndepends = [ \"tb\" ]\n[mise_env.ci.tools.tc]\ndepends = [ \"ta\" ]\n")},
		want: "rc=2\n--- stderr\n" +
			"repo-sync: repo/tools/a/tools.toml: depends of mise tool ta in [mise.tools] - the sync computes it from depends and parent of the tool; remove it\n" +
			"repo-sync: repo/tools/a/tools.toml: depends of mise tool tc in [mise_env.ci.tools] - the sync computes it from depends and parent of the tool; remove it\n" +
			"--- stdout\n"},
	{name: "repo-new-keys-ok", side: source.Repo,
		over: tree{
			"repo/tools/a/tools.toml": rt("[files.\"x.txt\"]\nsource = \"x.txt\"\nwhen = \"!a.flag\"\nperm = \"600\"\ndir_perm = \"0700\"\n" +
				"[tasks.\"run/x\"]\nsource = \"run-x\"\nwhen = \"b.pair.on\"\n"),
			"repo/tools/a/files/x.txt":    "x",
			"repo/tools/a/tasks/run-x":    "r",
			"repo/tools/a/scripts/derive": "d",
			"repo/tools/a/scripts/check":  "c",
			"repo/tools/a/scripts/after":  "a",
		},
		want: "rc=0\n--- stderr\n--- stdout\ntools: a\nsel:\na - default\ndeps:\nvalues:\n{\"a\":{}}\n"},
	{name: "repo-new-keys-bad", side: source.Repo,
		over: tree{
			"repo/tools/a/tools.toml": rt("[files.\"x.txt\"]\nsource = \"x.txt\"\nwhen = \"a..b\"\nperm = \"800\"\ndir_perm = 700\n" +
				"[tasks.\"run/x\"]\nsource = \"run-x\"\nwhen = true\nperm = \"644\"\n"),
			"repo/tools/a/files/x.txt":   "x",
			"repo/tools/a/tasks/run-x":   "r",
			"repo/tools/a/scripts/build": "d",
		},
		want: "rc=2\n--- stderr\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"x.txt\"] when must be a value path, for example \"git.branch_pair\", with an optional ! at the start\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"x.txt\"] perm must be an octal mode string, for example \"700\"\n" +
			"repo-sync: repo/tools/a/tools.toml: [files.\"x.txt\"] dir_perm must be an octal mode string, for example \"700\"\n" +
			"repo-sync: repo/tools/a/tools.toml: [tasks.\"run/x\"] perm is not a key of a tasks entry - the keys are source, mode, requires, requires_found, library, when\n" +
			"repo-sync: repo/tools/a/tools.toml: [tasks.\"run/x\"] when must be a value path, for example \"git.branch_pair\", with an optional ! at the start\n" +
			"repo-sync: repo/tools/a/tools.toml: scripts/build is not a script of a tool - the scripts are scripts/derive, scripts/check, scripts/after, scripts/detect and scripts/migrate-<tool>\n" +
			"--- stdout\n"},
	{name: "macos-no-macos-setup", side: source.Macos,
		over: tree{"macos/packs.toml": "[pack.base]\nm1 = { default = true }\n"},
		want: "rc=0\n--- stderr\n--- stdout\ntools: m1\ninit:\nkeep:\nm1\nvalues:\n{\"m1\":{}}\n"},
	{name: "macos-required-ok", side: source.Macos,
		over: tree{"macos/packs.toml": "[pack.base]\nmacos-setup = { default = true, required = true }\nm1 = { default = true }\n"},
		want: "rc=0\n--- stderr\n--- stdout\ntools: m1 macos-setup\ninit:\nkeep:\nmacos-setup\nm1\nvalues:\n{\"m1\":{},\"macos-setup\":{}}\n"},
	{name: "macos-manifest-keys", side: source.Macos,
		over: tree{"macos/tools/m1/tools.toml": "description = 1\nparent = \"m2\"\nfoo = 1\n[tasks.\"x\"]\nsource = \"x\"\nrequires_found = [ \"a\" ]\n", "macos/tools/m1/tasks/x": "x"},
		want: "rc=2\n--- stderr\n" +
			"setup-sync: macos/tools/m1/tools.toml: description must be a string\n" +
			"setup-sync: macos/tools/m1/tools.toml: foo is not a key of tools.toml - the keys are description, stage, depends, parent, [defaults], [questions], [contributes], [files], [tasks] and [mise]\n" +
			"setup-sync: macos/tools/m1/tools.toml: [tasks.\"x\"] requires_found is for repo tools only\n" +
			"--- stdout\n"},
}

func TestFormCases(t *testing.T) {
	for _, c := range formCases {
		t.Run(c.name, func(t *testing.T) {
			base := repoBase
			if c.side == source.Macos {
				base = macosBase
			}
			clone := writeTree(t, base.with(c.over))
			dropOldKeys(t, clone)
			vdir := t.TempDir()
			values, shown := filepath.Join(vdir, "bootstrap.toml"), "bootstrap.toml"
			if c.side == source.Macos {
				// The Mac side reads <values>/macos.toml; this case has none.
				values, shown = vdir, "config/macos.toml"
			} else if err := os.WriteFile(values, []byte(repoValues), 0o644); err != nil {
				t.Fatal(err)
			}
			got := pcase{name: c.name, side: c.side, clone: clone, values: values, shown: shown}.golang()
			if got != c.want {
				t.Errorf("%s:\n--- got\n%s\n--- want\n%s", c.name, got, c.want)
			}
		})
	}
}
