package files

import (
	"path/filepath"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// outsideJSON is the file outside the repository. It is a secret, which must
// not land in the repository.
const outsideJSON = `{"secret": "outside-value"}
`

// mergeSpec writes a spec below files/ of the tool fixture.
func mergeSpec(c *rc.Case, name string, lines ...string) {
	harness.WriteFile(c.T, c.ClonePath("repo/tools/fixture/files/"+name), harness.Lines(lines...))
}

func mergeTool(c *rc.Case) {
	fixtureTool(c)
	mergeSpec(c, "m.json.merge.toml", `[[item]]`, `kind = "keys"`, `value = { scalar = "src", nested = { set = 1 } }`,
		`[[item]]`, `kind = "item"`, `path = [ "list" ]`, `value = "b"`,
		`[[item]]`, `kind = "key"`, `path = [ "gone" ]`, `"$mode" = "remove"`)
	fixtureToml(c, `[files.".config/merger/m.json"]`, `mode = "merge"`, `source = "m.json.merge.toml"`)
}

// live writes a live shared file, with no newline at the end.
func live(c *rc.Case, ext, content string) { harness.WriteFile(c.T, mFile(c, ext), content) }

func mFile(c *rc.Case, ext string) string { return c.Path(".config/merger/m." + ext) }

func manifestPath(c *rc.Case) string { return c.Path(manifestRel) }

func noCommitSet(c *rc.Case) bool { return fileLacks(c.RepoValues, "bootstrap_commit") }

func outside(c *rc.Case, rel string) string { return filepath.Join(c.Dir, "outside", rel) }

func mergeRequired(c *rc.Case) {
	fixtureTool(c)
	mergeSpec(c, "m.json.merge.toml", `[[item]]`, `kind = "key"`, `path = [ "a" ]`, `value = 1`)
	fixtureToml(c, `[files."m.json"]`, `mode = "merge"`, `source = "m.json.merge.toml"`, `requires = [ "need.txt" ]`)
	consumerFile(c, "m.json", `{"keep": 1}`)
}

// mergeLinkTool: a shared file and a file outside the repository. The tool
// has no generated file in the folder of the target, so the sync writes
// nothing else through a link.
func mergeLinkTool(c *rc.Case) {
	fixtureTool(c)
	mergeSpec(c, "m.json.merge.toml", `[[item]]`, `kind = "key"`, `path = [ "a" ]`, `value = 1`)
	fixtureToml(c, `[files.".config/merger/m.json"]`, `mode = "merge"`, `source = "m.json.merge.toml"`)
	harness.WriteFile(c.T, outside(c, "m.json"), outsideJSON)
}

func mergeLinkLive(c *rc.Case) {
	mergeLinkTool(c)
	mkdir(c, c.Path(".config/merger"))
	symlink(c, outside(c, "m.json"), mFile(c, "json"))
}

func mergeLinkFolder(c *rc.Case) {
	mergeLinkTool(c)
	symlink(c, filepath.Join(c.Dir, "outside"), c.Path(".config/merger"))
}

// checkLinkKept: the link, the outside file and the repository are as they
// were.
func checkLinkKept(c *rc.Case, name, link, target string) {
	c.CheckThat(name+": the link stays", linkIs(link, target))
	checkExact(c, name+": the outside file stays", outside(c, "m.json"), outsideJSON)
	c.CheckThat(name+": nothing new outside", outsideOnlyFile(c))
	c.CheckThat(name+": the secret is not in the repository", repoLacks(c, "outside-value"))
	c.CheckThat(name+": the secret is not in the output", c.LogLacks("outside-value"))
}

// Shared files (mode = "merge"). test:mac and the Go tests cover each kind of
// item; these cases cover the rules of the repo side: the manifest rows, the
// mode, requires, symlinks and the config errors.
func TestMerge(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "merge: a JSON target", Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeTool(c)
			live(c, "json", `{"keep": 1, "scalar": "app", "list": ["x"], "gone": true}`)
		},
		Check: func(c *rc.Case) {
			f := mFile(c, "json")
			c.CheckThat("merge: a found key keeps its value, new items are added, a removed key goes",
				jsonIs(c, f, `{"keep": 1, "scalar": "app", "list": ["x", "b"], "nested": {"set": 1}}`))
			c.CheckThat("merge: the manifest records an added item",
				rc.FileHas(manifestPath(c), "fixture"+tab+".config/merger/m.json"+tab+`{"kind":"item","path":["list"],"value":"b"}`))
			c.CheckThat("merge: the manifest does not record a found key", fileLacks(manifestPath(c), `"path":["scalar"]`))
			got := rerun(c)
			c.CheckThat("merge: a second run exits 0", got == 0)
			c.CheckThat("merge: a second run writes nothing", c.LogLacks("updated    "+f))
		}})
	rc.Run(t, rc.Spec{Name: "merge: an absent target gets the items", Values: fixtureValues, Args: realRun,
		Change: mergeTool,
		Check: func(c *rc.Case) {
			c.CheckThat("merge absent: the items are written",
				jsonIs(c, mFile(c, "json"), `{"scalar": "src", "nested": {"set": 1}, "list": ["b"]}`))
		}})
	// The spec drops an item that bootstrap added before: the next run removes
	// it.
	rc.Run(t, rc.Spec{Name: "merge: an item that no spec has now goes", Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeTool(c)
			manifestFile(c, "fixture"+tab+".config/merger/m.json"+tab+`{"kind":"item","path":["list"],"value":"b"}`,
				"fixture"+tab+".config/merger/m.json"+tab+`{"kind":"item","path":["list"],"value":"old"}`)
			live(c, "json", `{"list": ["x", "b", "old"]}`)
		},
		Check: func(c *rc.Case) {
			c.CheckThat("merge cleanup: the item that no spec has now goes",
				jsonIs(c, mFile(c, "json"), `{"list": ["x", "b"], "scalar": "src", "nested": {"set": 1}}`))
			c.CheckThat("merge cleanup: its row goes", fileLacks(manifestPath(c), `"value":"old"`))
		}})
	rc.Run(t, rc.Spec{Name: "merge: a live target that does not parse", Code: 2, Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeTool(c)
			live(c, "json", `{"a": `)
		}, Message: "m.json is not valid json - kept it as it is",
		Check: func(c *rc.Case) { checkExact(c, "merge live bad: the file stays", mFile(c, "json"), `{"a": `) }})
	rc.Run(t, rc.Spec{Name: "merge: a spec that cannot be used", Code: 2, Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeTool(c)
			mergeSpec(c, "m.json.merge.toml", "zzz = 1")
			live(c, "json", `{"keep": 1}`)
		}, Message: "repo/tools/fixture/files/m.json.merge.toml: zzz is not a key of a merge spec",
		Check: func(c *rc.Case) { checkExact(c, "merge spec bad: the file stays", mFile(c, "json"), `{"keep": 1}`) }})
	// An existing shared file keeps its mode, also on a run that changes
	// nothing.
	rc.Run(t, rc.Spec{Name: "merge: an existing target keeps its mode", Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeTool(c)
			chmod(c, c.ClonePath("repo/tools/fixture/files/m.json.merge.toml"), 0o755)
			live(c, "json", `{"keep": 1}`)
			chmod(c, mFile(c, "json"), 0o600)
		},
		Check: func(c *rc.Case) {
			f := mFile(c, "json")
			c.CheckThat("merge mode: updated", c.LogHas("updated    "+f))
			c.CheckThat("merge mode: 600 stays", rc.ModeIs(f, "600"))
			got := rerun(c)
			c.CheckThat("merge mode: a second run exits 0", got == 0)
			c.CheckThat("merge mode: a second run, no chmod", c.LogLacks("chmod 755  "+f))
			c.CheckThat("merge mode: a second run, 600 stays", rc.ModeIs(f, "600"))
		}})
	// A shared file with a condition.
	rc.Run(t, rc.Spec{Name: "merge: its requires absent", Values: fixtureValues, Args: realRun,
		Change: mergeRequired,
		Check: func(c *rc.Case) {
			c.CheckThat("merge, its requires absent: skip merge", c.LogHas("skip merge "+c.Path("m.json")+" (no need.txt)"))
			checkExact(c, "merge, its requires absent: stays", c.Path("m.json"), `{"keep": 1}`+"\n")
		}})
	rc.Run(t, rc.Spec{Name: "merge: its requires present", Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeRequired(c)
			consumerFile(c, "need.txt")
		},
		Check: func(c *rc.Case) {
			c.CheckThat("merge, its requires present: merged", jsonIs(c, c.Path("m.json"), `{"keep": 1, "a": 1}`))
		}})
	// A shared file that is a symlink, or is below a symlinked folder: the
	// repo side follows no symlink.
	rc.Run(t, rc.Spec{Name: "merge: a target that is a symlink", Code: 2, Values: fixtureValues, Args: realRun,
		Change: mergeLinkLive, Message: "m.json is a symlink - kept it as it is",
		Check: func(c *rc.Case) {
			c.CheckThat("merge link: the message", c.LogHas(mFile(c, "json")+" is a symlink - kept it as it is"))
			checkLinkKept(c, "merge link", mFile(c, "json"), outside(c, "m.json"))
			c.CheckThat("merge link: no bootstrap_commit", noCommitSet(c))
		}})
	rc.Run(t, rc.Spec{Name: "merge: a dry run, a target that is a symlink", Code: 2, Values: fixtureValues,
		Change: mergeLinkLive, Message: "m.json is a symlink - kept it as it is",
		Check: func(c *rc.Case) {
			c.CheckThat("merge link, dry run: the message", c.LogHas(mFile(c, "json")+" is a symlink - kept it as it is"))
			checkLinkKept(c, "merge link, dry run", mFile(c, "json"), outside(c, "m.json"))
			c.CheckThat("merge link, dry run: no bootstrap_commit", noCommitSet(c))
			c.CheckThat("merge link, dry run: would set nothing", c.LogLacks("would set bootstrap_commit"))
		}})
	rc.Run(t, rc.Spec{Name: "merge: a target that is a dangling symlink", Code: 2, Values: fixtureValues, Args: realRun,
		Change: func(c *rc.Case) {
			mergeLinkTool(c)
			mkdir(c, c.Path(".config/merger"))
			symlink(c, outside(c, "absent.json"), mFile(c, "json"))
		}, Message: "m.json is a symlink - kept it as it is",
		Check: func(c *rc.Case) {
			c.CheckThat("merge dangling link: the message", c.LogHas(mFile(c, "json")+" is a symlink - kept it as it is"))
			checkLinkKept(c, "merge dangling link", mFile(c, "json"), outside(c, "absent.json"))
			c.CheckThat("merge dangling link: no bootstrap_commit", noCommitSet(c))
		}})
	// The sync runs in the repository, so the engine shows both paths as
	// relative.
	rc.Run(t, rc.Spec{Name: "merge: a target below a symlinked folder", Code: 2, Values: fixtureValues,
		Change: mergeLinkFolder,
		Call: func(c *rc.Case) int {
			return c.Sh(`cd -- "$consumer" && "$bin" sync`)
		}, Message: ".config/merger/m.json: the folder .config/merger is a symlink - kept it as it is",
		Check: func(c *rc.Case) {
			checkLinkKept(c, "merge linked folder", c.Path(".config/merger"), filepath.Join(c.Dir, "outside"))
			c.CheckThat("merge linked folder: no bootstrap_commit", noCommitSet(c))
		}})
	rc.Run(t, rc.Spec{Name: "merge: a dry run, a target below a symlinked folder", Code: 2, Values: fixtureValues,
		Change: mergeLinkFolder, Message: "m.json: the folder",
		Check: func(c *rc.Case) {
			c.CheckThat("merge linked folder, dry run: the message",
				c.LogHas(mFile(c, "json")+": the folder "+c.Path(".config/merger")+" is a symlink - kept it as it is"))
			checkLinkKept(c, "merge linked folder, dry run", c.Path(".config/merger"), filepath.Join(c.Dir, "outside"))
			c.CheckThat("merge linked folder, dry run: no bootstrap_commit", noCommitSet(c))
			c.CheckThat("merge linked folder, dry run: would set nothing", c.LogLacks("would set bootstrap_commit"))
		}})
	// The config errors. The tool `fixture` is not selected.
	rc.Run(t, rc.Spec{Name: "merge: a source that is not a merge spec", Code: 2, Values: noToolValues,
		Change: func(c *rc.Case) {
			fixtureTool(c)
			mergeSpec(c, "m.json", "{}")
			fixtureToml(c, `[files."m.json"]`, `mode = "merge"`, `source = "m.json"`)
		}, Message: `fixture/tools.toml: [files."m.json"] the source of a shared file is a merge spec <name>.merge.toml`})
	rc.Run(t, rc.Spec{Name: "merge: the old key remove", Code: 2, Values: noToolValues,
		Change: func(c *rc.Case) {
			mergeTool(c)
			fixtureToml(c, `remove = [ [ "a" ] ]`)
		}, Message: `fixture/tools.toml: [files.".config/merger/m.json"] remove is not a key of a files entry`})
}
