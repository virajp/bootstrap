package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

// releaseValues: the project web, a tag kind prod with package.json, a
// promotion, and a task kind with args.
const releaseValues = devValues + `

[projects.web]
folder = "web"

[release.web.prod]
type = "tag"
tag = "web-v{version}"
version = "package.json"

[release.web.staging-copy]
type = "tag"
tag = "web-promoted-v{version}"
from = "prod"

[release.web.local]
type = "task"
task = "p:local"
args = ["--target", "this machine"]`

// releaseRepo writes the project folder, a local task, a signing key and a
// bare origin with main, then commits and pushes. The case removes the hook
// config, so the hook safety net of the merge does not run.
func releaseRepo(c *rc.Case) {
	_ = os.Remove(c.Path(".config/pre-commit-config.yaml"))
	writeConsumer(c, "web/package.json", "{", `  "name": "web",`, `  "version": "1.2.0"`, "}")
	writeConsumer(c, tasksPath+"/p/local.py", "#!/bin/sh", `echo "local-build-ran $1|$2"`)
	if err := os.Chmod(c.Path(tasksPath+"/p/local.py"), 0o755); err != nil {
		c.T.Fatal(err)
	}
	key := filepath.Join(c.Dir, "key")
	if code, out := (&harness.Cmd{Dir: c.Dir, Args: []string{"ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-f", key}}).Run(); code != 0 {
		c.T.Fatalf("ssh-keygen: %s", out)
	}
	git(c, c.Consumer, "config", "gpg.format", "ssh")
	git(c, c.Consumer, "config", "user.signingkey", key)
	// The global config of a machine can sign with another program.
	git(c, c.Consumer, "config", "gpg.ssh.program", "ssh-keygen")
	c.CommitConsumer()
	origin := filepath.Join(c.Dir, "origin.git")
	git(c, c.Dir, "init", "--quiet", "--bare", "-b", "main", origin)
	git(c, c.Consumer, "remote", "add", "origin", origin)
	git(c, c.Consumer, "push", "--quiet", "--no-verify", "-u", "origin", "main")
}

// setLevel writes the level of web to releases.toml, commits and pushes it.
func setLevel(c *rc.Case, level string) {
	writeConsumer(c, ".config/bootstrap/releases.toml", "format = 3", `web = "`+level+`"`)
	c.CommitConsumer()
	git(c, c.Consumer, "push", "--quiet", "--no-verify", "origin", "main")
}

// The release tasks: the sync writes git:tag and one release task for each
// [release.<project>.<kind>] table, and the stale rule removes the task of a
// table that is gone.
func TestReleaseTasksSynced(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "release: a task for each table", Values: releaseValues,
		Change: func(c *rc.Case) { writeConsumer(c, "web/package.json", "{}") },
		Call: synced(func(c *rc.Case) int {
			harness.WriteFile(c.T, c.RepoValues, strings.Replace(harness.ReadFile(c.RepoValues),
				"[release.web.local]\ntype = \"task\"\ntask = \"p:local\"\nargs = [\"--target\", \"this machine\"]\n", "", 1))
			return c.Synced()
		}),
		Check: func(c *rc.Case) {
			prod := harness.ReadFile(c.Path(tasksPath + "/release/web/prod.py"))
			c.CheckThat("release: git:tag is written", isFile(c.Path(tasksPath+"/git/tag.py")))
			c.CheckThat("release: the library is written", isFile(c.Path(tasksPath+"/_scripts/release.py")))
			c.CheckThat("release: prod calls the library", strings.Contains(prod, `release.main("web", "prod")`))
			c.CheckThat("release: a tag kind has --push", strings.Contains(prod, `# USAGE flag "--push"`))
			c.CheckThat("release: the promotion is written", isFile(c.Path(tasksPath+"/release/web/staging-copy.py")))
			c.CheckThat("release: the task of a removed table goes", rc.NoFile(c.Path(tasksPath+"/release/web/local.py")))
		}})
}

// git:tag tags HEAD with a signed tag, pushes only the tag, and refuses a tag
// that exists.
func TestGitTag(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "git:tag: a signed tag, pushed", Values: releaseValues,
		Call: synced(func(c *rc.Case) int { releaseRepo(c); return 0 }),
		Check: func(c *rc.Case) {
			origin := filepath.Join(c.Dir, "origin.git")
			c.CheckThat("git:tag: exits 0", gitTask(c, "git:tag", "v1.0.0", "--push") == 0)
			c.CheckThat("git:tag: the tag is signed", strings.Contains(plainGit(c.Consumer, "cat-file", "-p", "v1.0.0"), "BEGIN SSH SIGNATURE"))
			c.CheckThat("git:tag: the tag is on origin", plainGit(origin, "rev-parse", "v1.0.0^{commit}") == plainGit(c.Consumer, "rev-parse", "HEAD"))
			c.CheckThat("git:tag: the same tag again exits 1", gitTask(c, "git:tag", "v1.0.0") == 1)
			c.CheckThat("git:tag: the refusal names the tag", c.LogHas("The tag v1.0.0 already exists here."))
			git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "local only")
			c.CheckThat("git:tag --push of a commit that origin does not have exits 1", gitTask(c, "git:tag", "v1.0.1", "--push") == 1)
			c.CheckThat("git:tag: no tag after the refusal", plainGit(c.Consumer, "tag", "--list", "v1.0.1") == "")
		}})
}

// A release of a tag kind: the level bumps the version file, one commit with
// the file and the cleared level lands on main with a merge, and the merge
// commit gets the tag. Then the promotion, the resume rule and NONE. With
// branch.autoSetupMerge=always, the release branch tracks no upstream, so the
// merge finds no unpushed commits.
func TestReleaseTag(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "release: bump, land, tag, promote", Values: releaseValues,
		Call: synced(func(c *rc.Case) int { releaseRepo(c); setLevel(c, "MINOR"); return 0 }),
		Check: func(c *rc.Case) {
			origin := filepath.Join(c.Dir, "origin.git")
			git(c, c.Consumer, "config", "branch.autoSetupMerge", "always")
			c.CheckThat("release --dry-run: exits 0", gitTask(c, "release:web:prod", "--dry-run") == 0)
			c.CheckThat("release --dry-run: names the tag", c.LogHas("web: 1.2.0 -> 1.3.0 (MINOR), tag web-v1.3.0"))
			c.CheckThat("release --dry-run: no tag", plainGit(c.Consumer, "tag", "--list") == "")

			c.CheckThat("release: exits 0", gitTask(c, "release:web:prod", "--push") == 0)
			c.CheckThat("release: the release branch has no upstream", c.LogLacks("Unpushed commits exist"))
			pkg := harness.ReadFile(c.Path("web/package.json"))
			c.CheckThat("release: the version file is bumped", strings.Contains(pkg, `"version": "1.3.0"`))
			c.CheckThat("release: the level is cleared", !strings.Contains(harness.ReadFile(c.Path(".config/bootstrap/releases.toml")), "web"))
			c.CheckThat("release: the tag is on the merge commit", plainGit(c.Consumer, "rev-parse", "web-v1.3.0^{commit}") == plainGit(c.Consumer, "rev-parse", "main"))
			c.CheckThat("release: main has the merge", plainGit(c.Consumer, "log", "-1", "--format=%s", "main^2") == "ops: release web-v1.3.0")
			c.CheckThat("release: the tag is on origin", plainGit(origin, "rev-parse", "web-v1.3.0^{commit}") != "")
			c.CheckThat("release: the release branch is gone", plainGit(c.Consumer, "branch", "--list", "release/web-v1.3.0") == "")

			c.CheckThat("release: NONE exits 0", gitTask(c, "release:web:prod") == 0)
			c.CheckThat("release: NONE prints one line", c.LogHas("Nothing to release: the level of web is NONE"))

			c.CheckThat("promotion: exits 0", gitTask(c, "release:web:staging-copy") == 0)
			c.CheckThat("promotion: the same version", plainGit(c.Consumer, "rev-parse", "web-promoted-v1.3.0^{commit}") == plainGit(c.Consumer, "rev-parse", "main"))

			// A release commit without its tag: the next run tags it.
			git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "ops: release web-v1.3.1")
			git(c, c.Consumer, "commit", "--quiet", "--no-verify", "--allow-empty", "-m", "later work")
			git(c, c.Consumer, "push", "--quiet", "--no-verify", "origin", "main")
			c.CheckThat("resume: exits 0", gitTask(c, "release:web:prod") == 0)
			c.CheckThat("resume: the release commit gets the tag", plainGit(c.Consumer, "log", "-1", "--format=%s", "web-v1.3.1") == "ops: release web-v1.3.1")
			c.CheckThat("resume: the branch comes back", plainGit(c.Consumer, "rev-parse", "--abbrev-ref", "HEAD") == "main")
		}})
	rc.Run(t, rc.Spec{Name: "release: a task kind runs its task with its args", Values: releaseValues,
		Call: synced(func(c *rc.Case) int { releaseRepo(c); return 0 }),
		Check: func(c *rc.Case) {
			c.CheckThat("task --dry-run: exits 0", gitTask(c, "release:web:local", "--dry-run") == 0)
			c.CheckThat("task --dry-run: prints the command", c.LogHas("Would run: mise run p:local --target 'this machine'"))
			c.CheckThat("task --dry-run: does not run the task", !c.LogHas("local-build-ran"))
			c.CheckThat("task: exits 0", gitTask(c, "release:web:local") == 0)
			c.CheckThat("task: the task ran with its args", c.LogHas("local-build-ran --target|this machine"))
		}})
}
