// Package detect tests `bootstrap detect`, the detect script of each tool,
// the projects, and the selection commands add, remove and replace, the exec
// commands of dprint.jsonc and the first catalog cases. The cases came from
// test:repo, lines 4120 to 4559.
package detect

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/virajp/bootstrap/e2e/internal/harness"
	rc "github.com/virajp/bootstrap/e2e/internal/repocase"
)

func TestMain(m *testing.M) { rc.Main(m) }

const devValues = `environments = [ "dev" ]`

// writeFiles writes each file of <files> (path=content) in the consumer, with
// a newline at the end.
func writeFiles(files ...string) func(c *rc.Case) {
	return func(c *rc.Case) {
		for _, f := range files {
			p, content, _ := strings.Cut(f, "=")
			harness.WriteFile(c.T, c.Path(p), content+"\n")
		}
	}
}

var (
	pnpmFiles       = writeFiles("package.json={}", "pnpm-workspace.yaml=packages: []")
	aubeFiles       = writeFiles("package.json={}", "aube-workspace.yaml=packages: []")
	bothFiles       = writeFiles("package.json={}", "pnpm-workspace.yaml=packages: []", "aube-workspace.yaml=packages: []")
	kotlinJavaFiles = writeFiles(`build.gradle.kts=plugins { kotlin("jvm") version "2.0.0" }`)
)

// engineCall gives a Call that syncs, commits, then runs the engine with
// <args>.
func engineCall(args ...string) func(c *rc.Case) int {
	return func(c *rc.Case) int {
		if code := c.Synced(); code != 0 {
			return code
		}
		return c.Engine(args...)
	}
}

var detectNo = engineCall("detect")

func notGrep(path, s string) bool { return !strings.Contains(harness.ReadFile(path), s) }

func checkDetectKept(c *rc.Case) {
	c.CheckThat("detect: the values file stays", notGrep(c.ToolsValues, "pnpm"))
}

func checkDetectDry(c *rc.Case) {
	c.CheckThat("detect --dry-run: the values file stays", notGrep(c.ToolsValues, "pnpm"))
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestDetect(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "detect: a tool that matches and is not selected", Code: 1, Values: devValues, Change: pnpmFiles,
		Call:    detectNo,
		Message: "pnpm is found, but it is not selected - run bootstrap add tool pnpm, or bootstrap detect --add"})
	rc.Run(t, rc.Spec{Name: "detect with no flag writes nothing", Code: 1, Values: devValues, Change: pnpmFiles,
		Call: detectNo, Check: checkDetectKept})
	rc.Run(t, rc.Spec{Name: "detect --add --yes adds the tool that matches", Values: devValues, Change: pnpmFiles,
		Call: engineCall("detect", "--add", "--yes"), Message: "Synced: ", Check: func(c *rc.Case) {
			c.CheckThat("detect --add --yes: the values file selects pnpm", c.TomlQ(`.tools | has("pnpm")`, ""))
			c.CheckThat("detect --add --yes: the mise config of pnpm", exists(c.Path(".config/mise/conf.d/pnpm/mise.dev.toml")))
		}})
	rc.Run(t, rc.Spec{Name: "detect --add --yes --dry-run", Values: devValues, Change: pnpmFiles,
		Call: engineCall("detect", "--add", "--yes", "--dry-run"), Message: "changes, nothing written", Check: checkDetectDry})
	rc.Run(t, rc.Spec{Name: "detect --dry-run without --add", Code: 1, Values: devValues, Change: pnpmFiles,
		Call: engineCall("detect", "--dry-run"), Message: "pnpm is found, but it is not selected", Check: checkDetectDry})
	rc.Run(t, rc.Spec{Name: "detect --add without --yes and without a terminal", Code: 2, Values: devValues, Change: pnpmFiles,
		Call: engineCall("detect", "--add"), Message: "changes need --yes", Check: checkDetectKept})
	rc.Run(t, rc.Spec{Name: "detect: each detected tool is selected", Values: devValues + "\n[tools.pnpm]", Change: pnpmFiles,
		Call: detectNo})
	rc.Run(t, rc.Spec{Name: "detect: a selected tool that finds nothing stays", Values: devValues, Call: detectNo})
	rc.Run(t, rc.Spec{Name: "detect: a mismatch for a selected tool", Code: 2, Values: devValues + "\n[tools.pnpm]", Change: bothFiles,
		Call: detectNo, Message: "pnpm is selected, but its detect script gives mismatch: aube-workspace.yaml exists"})
	rc.Run(t, rc.Spec{Name: "detect: a mismatch for a tool that is not selected", Values: devValues + "\n[tools.aube]", Change: bothFiles,
		Call: detectNo})
	rc.Run(t, rc.Spec{Name: "detect: a tool with the evidence of a selected tool", Values: devValues + "\n[tools.java]",
		Change: kotlinJavaFiles, Call: detectNo, Check: func(c *rc.Case) {
			c.CheckThat("detect: kotlin matches", c.LogHas("not offered"))
		}})
	// detect_flag: the sync stops the case on an error, as `synced` did.
	rc.Run(t, rc.Spec{Name: "detect --no: an unknown flag", Code: 2, Values: devValues,
		Call: engineCall("detect", "--no"), Message: "detect has no flag --no"})
	rc.Run(t, rc.Spec{Name: "detect --yes without --add", Code: 2, Values: devValues,
		Call: engineCall("detect", "--yes"), Message: "detect --yes needs --add - add --add"})
	rc.Run(t, rc.Spec{Name: "detect --answer without --add", Code: 2, Values: devValues,
		Call: engineCall("detect", "--answer", "go.flavor=x"), Message: "detect --answer needs --add - add --add"})
}

// detectCase: a consumer with the evidence of <tool>, and the match line.
func detectCase(t *testing.T, tool string, code int, files ...string) {
	re := regexp.MustCompile(`(` + tool + ` +\[OK\] found|\[NEW\] ` + tool + ` +offered)`)
	rc.Run(t, rc.Spec{Name: "detect script of " + tool, Code: code, Values: devValues, Change: writeFiles(files...),
		Call: detectNo, Check: func(c *rc.Case) {
			c.CheckThat("detect script of "+tool+": found", re.MatchString(c.LogText()))
		}})
}

const melange = "melange.yaml=package:\n  name: x\nenvironment:\n  contents: {}\npipeline: []"

func TestDetectScripts(t *testing.T) {
	detectCase(t, "node", 0, "package.json={}")
	detectCase(t, "pnpm", 1, "package.json={}", "pnpm-workspace.yaml=packages: []")
	detectCase(t, "aube", 1, "package.json={}", "aube-workspace.yaml=packages: []")
	detectCase(t, "bun", 1, "bunfig.toml=[install]")
	detectCase(t, "python", 0, "pyproject.toml=[project]")
	detectCase(t, "go", 1, "go.mod=module example.com/x")
	detectCase(t, "rust", 1, "Cargo.toml=[package]")
	detectCase(t, "flutter", 1, "pubspec.yaml=name: x\ndependencies:\n  flutter:\n    sdk: flutter")
	detectCase(t, "swift", 1, "Package.swift=// swift-tools-version:5.9")
	detectCase(t, "swift", 1, "App.xcodeproj/project.pbxproj={}")
	detectCase(t, "kotlin", 1, `build.gradle.kts=plugins { kotlin("jvm") }`)
	detectCase(t, "java", 1, `build.gradle=apply plugin: "java"`)
	detectCase(t, "android", 1, `app/build.gradle.kts=plugins { id("com.android.application") }`)
	detectCase(t, "docker", 1, "Dockerfile=FROM scratch")
	detectCase(t, "cloudflare", 1, `wrangler.toml=name = "x"`)
	detectCase(t, "firebase", 1, "firebase.json={}")
	detectCase(t, "pulumi", 1, "Pulumi.yaml=name: x")
	detectCase(t, "fnox", 1, ".config/fnox.toml=[secrets]")
	detectCase(t, "github", 1, ".github/CODEOWNERS=* @x")
	detectCase(t, "gitlab", 1, ".gitlab-ci.yml=stages: []")
	detectCase(t, "doppler", 1, "doppler.yaml=setup: []")
	detectCase(t, "gcloud", 1, ".gcloudignore=.git")
	detectCase(t, "gcloud", 1, "cloudbuild.yaml=steps: []")
	detectCase(t, "claude", 1, "CLAUDE.md=# x")
	detectCase(t, "claude", 1, ".claude/settings.json={}")
	detectCase(t, "apko", 1, "image/apko.yaml=contents:\n  packages: []\nentrypoint:\n  command: /bin/sh")
	detectCase(t, "melange", 1, melange)

	noApko := regexp.MustCompile(`(apko +\[OK\] found|\[NEW\] apko +offered)`)
	rc.Run(t, rc.Spec{Name: "detect script of apko: a melange config", Code: 1, Values: devValues, Change: writeFiles(melange),
		Call: detectNo, Check: func(c *rc.Case) {
			c.CheckThat("detect: apko does not match a melange config", !noApko.MatchString(c.LogText()))
		}})
	rc.Run(t, rc.Spec{Name: "detect script of flutter: Dart only", Values: devValues, Change: writeFiles("pubspec.yaml=name: x"),
		Call: detectNo, Check: func(c *rc.Case) {
			c.CheckThat("detect: flutter does not match Dart only", c.LogLacks("flutter  offered"))
		}})
}

// projectsJQ tells whether the value at <path> (keys below the root) of the
// JSON envelope in the log is equal to the JSON <want> (projects_jq).
func projectsJQ(c *rc.Case, want string, path ...string) bool {
	var got any
	if err := json.Unmarshal([]byte(c.LogText()), &got); err != nil {
		return false
	}
	for _, k := range path {
		m, ok := got.(map[string]any)
		if !ok {
			return false
		}
		got = m[k]
	}
	var w any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		c.T.Fatal(err)
	}
	return reflect.DeepEqual(got, w)
}

// detectJSON: sync, then `bootstrap detect --json`, with only its stdout in
// the log (2> /dev/null).
func detectJSON(c *rc.Case) int {
	if code := c.Synced(); code != 0 {
		return code
	}
	cmd := exec.Command(harness.Bin, "detect", "--json")
	cmd.Dir = harness.Root
	cmd.Env = c.Env
	out, err := cmd.Output()
	harness.AppendFile(c.T, c.Log, string(out))
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		return 127
	}
	return 0
}

// checkLogLacks is check_log_lacks: the check name is "<name>: no <text>".
func checkLogLacks(c *rc.Case, name, text string) {
	if c.LogHas(text) {
		c.Fail(name + ": the output has: " + text)
	} else {
		c.Pass(name + ": no " + text)
	}
}

// The K1 shape: a pnpm workspace with apps/*, and a Go module in
// services/api. apps/docs holds no package.json, so it is not a project.
var k1Files = writeFiles("package.json={}", "pnpm-workspace.yaml=packages:\n  - apps/*", "apps/web/package.json={}",
	"apps/docs/README.md=# docs", "services/api/go.mod=module example.com/api")

func TestProjects(t *testing.T) {
	rc.Run(t, rc.Spec{Name: "projects: detect --json of a pnpm and Go monorepo", Code: 1, Values: devValues, Change: k1Files,
		Call: detectJSON, Message: `"ok":false,"code":1`, Check: func(c *rc.Case) {
			c.CheckThat("projects, a pnpm and Go monorepo: the projects",
				projectsJQ(c, `{"web":{"folder":"apps/web"},"api":{"folder":"services/api"}}`, "data", "projects"))
			c.CheckThat("projects, a pnpm and Go monorepo: the links",
				projectsJQ(c, `{"go":["api"],"node":["web"],"pnpm":["web"]}`, "data", "links"))
		}})
	rc.Run(t, rc.Spec{Name: "projects: detect --add --yes of a pnpm and Go monorepo, then check", Values: devValues, Change: k1Files,
		Call: func(c *rc.Case) int {
			if code := c.Synced(); code != 0 {
				return code
			}
			add := &harness.Cmd{Env: c.Env, Args: []string{harness.Bin, "detect", "--add", "--yes"}}
			if code, _ := add.Run(); code != 0 {
				return code
			}
			c.CommitConsumer()
			return c.Engine("check")
		}, Message: ".config/bootstrap is valid", Check: func(c *rc.Case) {
			c.CheckThat("projects, detect --add: the tables in repo.toml",
				c.TomlQ(`.projects.web.folder == "apps/web" and .projects.api.folder == "services/api" and (.projects | length) == 2`,
					c.RepoValues))
			c.CheckThat("projects, detect --add: the projects of each tool", c.TomlQ(`(.tools.go.projects | join(",")) == "api" and
    (.tools.node.projects | join(",")) == "web" and (.tools.pnpm.projects | join(",")) == "web"`, ""))
			checkLogLacks(c, "projects, detect --add: check gives no link warning", "gives no match")
		}})
	rc.Run(t, rc.Spec{Name: "projects: detect --json of a repo with one package", Values: devValues, Change: writeFiles("package.json={}"),
		Call: detectJSON, Message: `"ok":true,"code":0`, Check: func(c *rc.Case) {
			c.CheckThat("projects, one package: the root is the project repo_name",
				projectsJQ(c, `{"test":{"folder":"."}}`, "data", "projects"))
			c.CheckThat("projects, one package: the link of node", projectsJQ(c, `{"node":["test"]}`, "data", "links"))
		}})
	rc.Run(t, rc.Spec{Name: "projects: detect with no flag only lists the projects", Values: devValues, Change: writeFiles("package.json={}"),
		Call: detectNo, Message: "run bootstrap detect --add to write the projects", Check: func(c *rc.Case) {
			c.CheckThat("projects, detect with no flag: repo.toml stays", notGrep(c.RepoValues, "projects"))
		}})
	rc.Run(t, rc.Spec{Name: "projects: two folders with the same basename", Code: 1, Values: devValues,
		Change: writeFiles("a/api/go.mod=module example.com/a", "b/api/go.mod=module example.com/b"),
		Call:   detectJSON, Check: func(c *rc.Case) {
			c.CheckThat("projects, two folders named api: the path form of each name",
				projectsJQ(c, `{"a-api":{"folder":"a/api"},"b-api":{"folder":"b/api"}}`, "data", "projects"))
		}})
	rc.Run(t, rc.Spec{Name: "projects: a gitignored folder is not a project", Code: 1, Values: devValues,
		Change: writeFiles("svc/go.mod=module example.com/svc", "ignored/go.mod=module example.com/ignored", ".gitignore=/ignored/"),
		Call:   detectJSON, Check: func(c *rc.Case) {
			c.CheckThat("projects, a gitignored folder: it is skipped", projectsJQ(c, `{"svc":{"folder":"svc"}}`, "data", "projects"))
		}})
	rc.Run(t, rc.Spec{Name: "projects: check of a project whose folder is gone", Code: 2, Values: devValues + `
[projects.api]
folder = "services/api"

[tools.go]
projects = [ "api" ]`, Call: engineCall("check"), Message: "[projects.api].folder = services/api in ",
		Check: func(c *rc.Case) { checkLogLacks(c, "projects, a deleted folder: no link warning", "gives no match") }})
	rc.Run(t, rc.Spec{Name: "projects: config unset of a project that a tool lists", Code: 2, Values: devValues + `
[projects.web]
folder = "apps/web"

[tools.node]
projects = [ "web" ]`, Call: func(c *rc.Case) int {
		harness.WriteFile(c.T, filepath.Join(c.Dir, "repo.before"), harness.ReadFile(c.RepoValues))
		return c.Engine("config", "unset", "repo.projects.web")
	}, Message: "the project web is in the projects of the tools node, so config unset cannot remove [projects.web]",
		Check: func(c *rc.Case) {
			before, err := os.ReadFile(filepath.Join(c.Dir, "repo.before"))
			after, err2 := os.ReadFile(c.RepoValues)
			c.CheckThat("a refused set: repo.toml stays", err == nil && err2 == nil && string(before) == string(after))
		}})
}
