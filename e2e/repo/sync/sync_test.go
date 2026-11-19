// Package sync tests `bootstrap sync --dry-run` with the values of a
// repository: the base pack, packs, environments, members and the tool
// folders. The cases came from test:repo, lines 380 to 560.
package sync

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

func TestMain(m *testing.M) { rc.Main(m) }

const devValues = `environments = [ "dev" ]`

// The required tools of the pack base. _security_base arrives through the
// depends of the scanners.
var baseTools = []string{"mise", "git", "pre-commit", "dprint", "linter", "_security_base", "gitleaks", "grype", "osv-scanner"}

var entryRe = regexp.MustCompile(`^\[(files|tasks)\."(.*)"\]$`)

// checkBaseFiles: each tool of the pack base syncs a file, no conf.d/_base/,
// and the mise tool ships a dispatcher for each test port.
func checkBaseFiles(c *rc.Case) {
	log := c.LogText()
	var missing []string
	for _, tool := range baseTools {
		toml := harness.ReadFile(c.ClonePath("repo/tools/" + tool + "/tools.toml"))
		found, entries := false, false
		for _, l := range strings.Split(toml, "\n") {
			m := entryRe.FindStringSubmatch(l)
			if m == nil {
				continue
			}
			entries = true
			path := m[2]
			if m[1] == "tasks" {
				path = ".config/mise/tasks/" + path
			}
			if strings.Contains(log, "would update "+c.Consumer+"/"+path) {
				found = true
				break
			}
		}
		// A tool without an entry ships no file, for example _security_base.
		if !entries {
			found = true
		}
		if !found {
			missing = append(missing, tool)
		}
	}
	if len(missing) > 0 {
		harness.Fail(c.T, "base: no file listed for: "+strings.Join(missing, " "), "")
	} else {
		c.Pass("base: a file of each tool listed")
	}
	c.CheckThat("base: no conf.d/_base/", !strings.Contains(log, "conf.d/_base/"))
	for _, p := range []string{"test/_default", "test/e2e/_default", "test/golden/_default"} {
		if strings.Contains(log, "would update "+c.Consumer+"/.config/mise/tasks/"+p) {
			c.Pass("base: the dispatcher " + p)
		} else {
			c.Fail("base: no dispatcher " + p)
		}
	}
}

func appendClone(c *rc.Case, rel string, lines ...string) {
	harness.AppendFile(c.T, c.ClonePath(rel), harness.Lines(lines...))
}

// editNode replaces the line that matches <re> in the tools.toml of node.
func editNode(c *rc.Case, re, repl string) {
	p := c.ClonePath("repo/tools/node/tools.toml")
	harness.WriteFile(c.T, p, regexp.MustCompile("(?m)"+re).ReplaceAllString(harness.ReadFile(p), repl))
}

func lateToolFolder(name string) func(c *rc.Case) {
	return func(c *rc.Case) {
		harness.WriteFile(c.T, c.ClonePath("repo/tools/"+name+"/tools.toml"),
			harness.Lines(`default_envs = [ "dev" ]`, `depends = []`, `root_only = false`))
	}
}

func memberFolder(c *rc.Case) {
	if err := os.MkdirAll(c.Path("projects/web"), 0o755); err != nil {
		c.T.Fatal(err)
	}
}

// consumerSubmodule: the consumer as a git submodule, a gitlink at its path
// in a superproject at the folder of the case.
func consumerSubmodule(c *rc.Case) {
	harness.Git(c.T, c.Dir, "init", "--quiet")
	head := harness.Git(c.T, c.Consumer, "rev-parse", "HEAD")
	harness.Git(c.T, c.Dir, "update-index", "--add", "--cacheinfo", "160000,"+head+",consumer")
}

func checkNoVscode(c *rc.Case) {
	for _, l := range strings.Split(c.LogText(), "\n") {
		if strings.Contains(l, "bootstrap: tools:") && regexp.MustCompile(`\bvscode\b`).MatchString(l) {
			c.Fail("submodule: the tools include vscode")
			return
		}
	}
	c.Pass("submodule: the tools do not include vscode")
}

func TestBase(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "no [tools], environments dev", Values: devValues, Check: checkBaseFiles})
	rc.Run(t, rc.Spec{Name: "bootstrap_commit: a dry run names the commit of the clone", Values: devValues,
		Message: "would set bootstrap_commit " + rc.BaseCommit})
	// A sync that changes no file keeps a commit that is set; a sync that
	// changes a file sets the commit of the clone.
	for _, cs := range []struct {
		name, change, want string
	}{
		{"bootstrap_commit: a sync without a change keeps it", "", `bootstrap_commit = "old"`},
		{"bootstrap_commit: a sync with a change sets it", `rm "$values/repo-generated.tsv"`, `bootstrap_commit = "` + rc.BaseCommit + `"`},
	} {
		rc.Run(t, rc.Spec{Name: cs.name, Values: devValues,
			Call: func(c *rc.Case) int {
				if code := c.Engine("sync"); code != 0 {
					return code
				}
				if code := c.Sh(`sed -i.bak 's/^bootstrap_commit = .*/bootstrap_commit = "old"/' "$values/repo.toml"
rm "$values/repo.toml.bak"
` + cs.change); code != 0 {
					return code
				}
				return c.Engine("sync")
			},
			Check: func(c *rc.Case) {
				c.CheckThat("repo.toml has "+cs.want, strings.Contains(harness.ReadFile(c.RepoValues), cs.want+"\n"))
			}})
	}
	rc.Run(t, rc.Spec{Name: "[tools.node] env dev, ci", Values: `environments = [ "dev", "ci" ]
[tools.node]
env = [ "dev", "ci" ]`})
	rc.Run(t, rc.Spec{Name: `tools = ["node"], the old list`, Code: 2, Values: `environments = [ "dev" ]
tools = [ "node" ]`, Message: "is a list, the old format"})
	rc.Run(t, rc.Spec{Name: "a pack loop a -> b -> a", Code: 2, Values: devValues,
		Change: func(c *rc.Case) {
			appendClone(c, "repo/packs.toml", "", "[pack.a]", `"pack:b" = {}`, "", "[pack.b]", `"pack:a" = {}`)
		},
		Message: "a pack loop: a -> b -> a"})
	rc.Run(t, rc.Spec{Name: "a pack named like a tool folder", Code: 2, Values: devValues,
		Change:  func(c *rc.Case) { appendClone(c, "repo/packs.toml", "", "[pack.node]", "node = {}") },
		Message: "pack 'node' is also a tool folder"})
	rc.Run(t, rc.Spec{Name: "[branches.main], the merge config of the repo", Values: `environments = [ "dev" ]
[branches.main]
landing = "pr"
merge_never = [ "develop" ]`})
	rc.Run(t, rc.Spec{Name: "[tools.git.branch.main], the old place of the merge config", Code: 2, Values: `environments = [ "dev" ]
[tools.git.branch.main]
merge_model = "pr"`, Message: ".config/bootstrap/tools.toml is a key of repo.toml now - write it as branches.<branch> in repo.toml"})
}

func TestMembers(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "[members.web], a folder below the root", Values: `environments = [ "dev" ]
[members.web]
folder = "projects/web"
type = "folder"`, Change: memberFolder})
	rc.Run(t, rc.Spec{Name: "[members.nope], not a submodule or a folder", Code: 2, Values: `environments = [ "dev" ]
[members.nope]
type = "folder"`, Message: ".config/bootstrap/repo.toml is not a submodule or a folder of the repo"})
	rc.Run(t, rc.Spec{Name: "[members.nope], no type", Code: 2, Values: `environments = [ "dev" ]
[members.nope]
folder = "projects/web"`, Change: memberFolder, Message: ".config/bootstrap/repo.toml has no type - set type to submodule, folder or sibling"})
	rc.Run(t, rc.Spec{Name: "[members.self], the folder of the repo itself", Code: 2, Values: `environments = [ "dev" ]
[members.self]
folder = "."
type = "folder"`, Message: "has the folder of the repo itself"})
	rc.Run(t, rc.Spec{Name: "[members.web], type submodule without a submodule", Code: 2, Values: `environments = [ "dev" ]
[members.web]
folder = "projects/web"
type = "submodule"`, Change: memberFolder, Message: "has type submodule, but projects/web is not in .gitmodules"})
	rc.Run(t, rc.Spec{Name: "[members.web], an unknown type", Code: 2, Values: `environments = [ "dev" ]
[members.web]
folder = "projects/web"
type = "repo"`, Change: memberFolder, Message: `members.web.type = "repo" in `})
	rc.Run(t, rc.Spec{Name: "[members.api], a sibling without a url", Code: 2, Values: `environments = [ "dev" ]
[members.api]
folder = "../api"
type = "sibling"`, Message: ".config/bootstrap/repo.toml has type sibling, but no url - add the url of its repo"})
	rc.Run(t, rc.Spec{Name: "[members.api], a sibling that is not there", Values: `environments = [ "dev" ]
[members.api]
folder = "../api"
type = "sibling"
url = "https://example.com/api"`, Message: "member api skipped: ../api is not there - clone https://example.com/api into it"})
	rc.Run(t, rc.Spec{Name: "[member.web], the old table", Code: 2, Values: `environments = [ "dev" ]
[member.web]
folder = "projects/web"`, Change: memberFolder, Message: "[ERROR] member in "})
	rc.Run(t, rc.Spec{Name: "a submodule skips vscode, a root_only tool of the pack base", Values: devValues,
		Change: consumerSubmodule, Check: checkNoVscode})
	rc.Run(t, rc.Spec{Name: "a submodule with [tools.vscode], a root_only tool", Code: 2, Values: `environments = [ "dev" ]
[tools.vscode]`, Change: consumerSubmodule, Message: "vscode is for the root repository only"})
}

func TestEnvironments(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "env ci on a tool whose mise tool allows dev only", Code: 2, Values: `environments = [ "dev", "ci" ]
[tools.pulumi]
env = [ "dev", "ci" ]`, Message: "mise tool 'pulumi' of repo/tools/pulumi allows only dev"})
	rc.Run(t, rc.Spec{Name: "env ci not in environments dev", Code: 2, Values: `environments = [ "dev" ]
[tools.node]
env = [ "ci" ]`, Message: "is not in environments: dev"})
	rc.Run(t, rc.Spec{Name: "a dependency without the environment of its parent", Code: 2, Values: `environments = [ "dev", "ci" ]
[tools.node]
env = [ "dev", "ci" ]`, Change: func(c *rc.Case) { editNode(c, `^depends = \[\]$`, `depends = [ "pulumi" ]`) },
		Message: "repo/tools/node needs repo/tools/pulumi in 'ci', but its mise tool 'pulumi' allows only dev"})
	rc.Run(t, rc.Spec{Name: "env ci without dev", Code: 2, Values: `environments = [ "dev", "ci" ]
[tools.node]
env = [ "ci" ]`, Message: "has no dev - each tool must be in dev, for example env = "})
	rc.Run(t, rc.Spec{Name: "default_envs ci without dev", Code: 2, Values: `environments = [ "dev", "ci" ]
[tools.node]`, Change: func(c *rc.Case) { editNode(c, `^default_envs = .*$`, `default_envs = [ "ci" ]`) },
		Message: "default_envs of repo/tools/node/tools.toml has no dev"})
	rc.Run(t, rc.Spec{Name: "default_envs with prod, outside dev and ci", Code: 2, Values: `environments = [ "dev" ]
[tools.node]`, Change: func(c *rc.Case) { editNode(c, `^default_envs = .*$`, `default_envs = [ "dev", "prod" ]`) },
		Message: "default_envs of repo/tools/node/tools.toml: 'prod' is not one of dev, ci"})
	rc.Run(t, rc.Spec{Name: "a tool folder that sorts after zz-repo", Code: 2, Values: devValues,
		Change: lateToolFolder("zzz-late"), Message: "tool folder repo/tools/zzz-late sorts after zz-repo"})
	rc.Run(t, rc.Spec{Name: "a tool folder named zz-repo", Code: 2, Values: devValues,
		Change: lateToolFolder("zz-repo"), Message: "tool folder repo/tools/zz-repo sorts after zz-repo or is named zz-repo"})
	rc.Run(t, rc.Spec{Name: "default_envs empty", Code: 2, Values: devValues,
		Change:  func(c *rc.Case) { editNode(c, `^default_envs = .*$`, `default_envs = []`) },
		Message: "default_envs of repo/tools/node/tools.toml has no dev"})
}
