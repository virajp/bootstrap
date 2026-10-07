import {
  coreTools,
  toolsByName,
} from "@/tool/catalog";
import {
  type RenderValues,
  render,
} from "@/tool/render";
import { NodeServices } from "@effect/platform-node";
import {
  describe,
  expect,
  it,
} from "@effect/vitest";
import {
  Effect,
  FileSystem,
  Path,
} from "effect";
import { spawnSync } from "node:child_process";
import { parse } from "yaml";

/** The core tools whose templates this suite covers; `fnox` ships with the new templates. */
const owned = coreTools.filter(tool => tool.name !== "fnox");

const values: RenderValues = {
  repo: "acme/widgets",
  commit_scopes: ["api"],
  merge_model: { develop: "direct", main: "pr" },
  tools: [],
};

const rendered = () => render(owned, values);

/** Text this repo carries that no target repository may inherit. */
const forbidden: ReadonlyArray<RegExp> = [
  /virajp\/bootstrap/,
  // The repository's name; the CLI's own `bootstrap update` command may be named.
  /\bbootstrap\b(?! update\b)/i,
  /doppler/i,
  /vwf/i,
  /stackgen/i,
  /(^|[^\w-])site\//m,
  /claude-plugins/i,
  /flutter/i,
  /\bdart/i,
  /pubspec/i,
  // The composition vocabulary of the generator this repo's own tooling came from.
  /toolchain-(gate|manager)|\boverlays the\b|\bmaterialize\b|\bpacks?\b|hook-runner|\bcomponent (that|later)\b/i,
];

describe("core templates", () => {
  it.effect("run the linter and the package.json sorter as mise-installed tools, never through pnpm dlx", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(
        [...files].filter(([, text]) => /pnpm dlx/.test(text)).map(([path]) =>
          path
        ),
      )
        .toEqual([]);
      const dev = files.get(".config/mise/conf.d/_base/mise.dev.toml")!;
      expect(dev).toMatch(/^\[tools\."npm:@askviraj\/linter"\]$/m);
      expect(dev).toMatch(/^\[tools\."npm:sort-package-json"\]$/m);
      expect(files.get(".config/mise/tasks/code/lint")).toMatch(/^linter\b/m);
      expect(files.get(".config/mise/tasks/code/format")).toMatch(
        /^\s*sort-package-json\b/m,
      );
    }));

  it.effect("tell a repo missing its hook config to restore it with bootstrap update", () =>
    Effect.gen(function*() {
      expect((yield* rendered()).get(".config/mise/tasks/setup/precommit"))
        .toMatch(/bootstrap update/);
    }));

  it("covers the twelve core tools other than fnox", () => {
    expect(owned.map(tool => tool.name)).toHaveLength(12);
  });

  it.effect("render exactly the catalog paths of the core tools", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect([...files.keys()].sort()).toEqual(
        owned.flatMap(tool => tool.files).sort(),
      );
    }));

  it.effect("carry nothing specific to this repository", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      const leaks = [...files].flatMap(([path, text]) =>
        forbidden.filter(pattern => pattern.test(text)).map(pattern =>
          `${path}: ${pattern}`
        )
      );
      expect(leaks).toEqual([]);
    }));

  it.effect("ignore backups in .gitignore", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(files.get(".gitignore")!.split("\n")).toContain("*.bak");
    }));

  it.effect("list the commit scopes in the conventional-commits config", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(
        parse(files.get(".config/git-conventional-commits.yaml")!)
          .convention
          .commitScopes,
      )
        .toEqual(["api"]);
    }));

  it.effect("write the scope list in the shipped formatter's bracket spacing", () =>
    Effect.gen(function*() {
      const line = (scopes: ReadonlyArray<string>) =>
        render(owned, { ...values, commit_scopes: scopes }).pipe(
          Effect.map(files =>
            files.get(".config/git-conventional-commits.yaml")!.match(
              /^ {2}commitScopes:.*$/m,
            )![0]
          ),
        );
      expect(yield* line(["api", "web"])).toBe(
        "  commitScopes: [ \"api\", \"web\" ]",
      );
      expect(yield* line([])).toBe("  commitScopes: []");
    }));

  it.effect("render an empty scope list as an empty sequence", () =>
    Effect.gen(function*() {
      const files = yield* render(owned, { ...values, commit_scopes: [] });
      expect(
        parse(files.get(".config/git-conventional-commits.yaml")!)
          .convention
          .commitScopes,
      )
        .toEqual([]);
    }));

  it.effect("set both merge models in the mise base config and no single one", () =>
    Effect.gen(function*() {
      const base = (yield* rendered()).get(
        ".config/mise/conf.d/_base/mise.toml",
      )!;
      expect(base).toMatch(/^MERGE_MODEL_DEVELOP\s*=\s*"direct"$/m);
      expect(base).toMatch(/^MERGE_MODEL_MAIN\s*=\s*"pr"$/m);
      expect(base).not.toMatch(/^MERGE_MODEL\s*=/m);
      expect(base).toMatch(/^MEMBERS\s*=\s*""$/m);
      expect(base).toMatch(/^REPO_NAME\s*=\s*"widgets"$/m);
    }));

  it.effect("dispatch each dependency verb to the tools' own tasks, never to a package manager", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      for (
        const verb of ["install", "outdated", "audit", "upgrade", "cleanup"]
      ) {
        const task = files.get(`.config/mise/tasks/setup/deps/${verb}`)!;
        expect(task).toMatch(
          new RegExp(`^run_tool_tasks setup:deps:${verb}\\b`, "m"),
        );
        expect(task).not.toMatch(/\bpnpm\b/);
      }
      const all = files.get(".config/mise/tasks/setup/deps/all")!;
      expect(all).not.toMatch(/\bpnpm\b/);
      expect(all).toMatch(/^run_tool_tasks setup:deps:all\b/m);
    }));

  it.effect("run each tool's setup task from setup:all only when it exists", () =>
    Effect.gen(function*() {
      const all = (yield* rendered()).get(".config/mise/tasks/setup/all")!;
      for (const task of ["setup:ai", "setup:secrets", "setup:precommit"]) {
        expect(all).toMatch(new RegExp(`^run_task_if_present ${task}$`, "m"));
        expect(all).not.toMatch(new RegExp(`^mise run ${task}$`, "m"));
      }
      expect(all).not.toMatch(/@askviraj\/linter|\bpnpm\b/);
    }));

  it.effect("keep the editor settings in the JSON form the shipped formatter writes", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(() => JSON.parse(files.get(".vscode/settings.json")!))
        .not
        .toThrow();
      const dprint = JSON.parse(files.get(".config/dprint.json")!);
      expect(dprint.json.jsonTrailingCommaFiles).not.toContain(
        ".vscode/settings.json",
      );
    }));

  it.effect("recommend each editor extension once", () =>
    Effect.gen(function*() {
      const { recommendations } = JSON.parse(
        (yield* rendered()).get(".vscode/extensions.json")!,
      );
      expect(recommendations).toEqual([...new Set(recommendations)]);
    }));

  it.effect("install fnox, the core secrets tool, through mise", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(files.get(".config/mise/conf.d/_base/mise.dev.toml")).toMatch(
        /^\[tools\.fnox\]$/m,
      );
    }));

  it.effect("render every core file the same whatever non-core tools are selected", () =>
    Effect.gen(function*() {
      const none = yield* rendered();
      for (const tools of [["github"], ["gitlab"], ["github", "gitlab"]]) {
        const some = yield* render(owned, { ...values, tools });
        expect(
          [...some].filter(([path, text]) => none.get(path) !== text).map((
            [path],
          ) => path),
        )
          .toEqual([]);
      }
    }));

  it.effect("leave the forge links out of the changelog config", () =>
    Effect.gen(function*() {
      const { changelog } = parse(
        (yield* rendered()).get(".config/git-conventional-commits.yaml")!,
      );
      for (const key of ["commitUrl", "commitRangeUrl", "issueUrl"]) {
        expect(changelog).not.toHaveProperty(key);
      }
    }));

  it.effect("keep number- and null-like values strings in every structured file", () =>
    Effect.gen(function*() {
      const files = yield* render(owned, {
        ...values,
        repo: "null/2048",
        commit_scopes: ["404"],
      });
      expect(parse(files.get("mempalace.yaml")!).wing).toBe("2048");
      expect(JSON.parse(files.get(".config/claude-status.json")!).projectName)
        .toBe("null/2048");
      expect(
        parse(files.get(".config/git-conventional-commits.yaml")!)
          .convention
          .commitScopes,
      )
        .toEqual(["404"]);
      expect(files.get(".config/mise/conf.d/_base/mise.toml")).toMatch(
        /^REPO_NAME\s*=\s*"2048"$/m,
      );
      expect(files.get(".config/mise/conf.d/ai/mise.dev.toml")).toMatch(
        /^MEMPALACE_PALACE_PATH\s*=\s*"~\/\.local\/share\/mempalace\/2048"$/m,
      );
    }));

  it.effect("format the whole tree from the root dprint config", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(files.get("dprint.json")).toContain(
        "\"extends\": \".config/dprint.json\"",
      );
      expect(files.get(".config/mise/tasks/code/format")).not.toMatch(
        /--config \.config\/dprint\.json/,
      );
    }));

  it.effect("start every task file with a shebang", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      const tasks = [...files].filter(([path]) =>
        path.startsWith(".config/mise/tasks/")
      );
      expect(tasks.length).toBeGreaterThan(0);
      expect(
        tasks.filter(([, text]) => !text.startsWith("#!")).map(([path]) =>
          path
        ),
      )
        .toEqual([]);
    }));
});

/** Stub `mise`: `tasks ls` prints the listing (or fails when there is none); `run` logs its arguments. */
const stubMise = (listing: string | undefined) =>
  `#!/usr/bin/env bash
if [ "$1" = tasks ]; then
  ${listing === undefined ? "exit 3" : `printf '%s\\n' ${listing}; exit 0`}
fi
echo "$* frozen=\${DEPS_FROZEN:-}" >>"\${LOG}"
`;

/** Runs a rendered `setup/deps/<verb>` dispatcher against the stub; returns its status and the stub's run log. */
const dispatch = (
  verb: string,
  env: Record<string, string>,
  listing?: string,
) =>
  Effect
    .gen(function*() {
      const fs = yield* FileSystem.FileSystem;
      const path = yield* Path.Path;
      const files = yield* rendered();
      const root = yield* fs.makeTempDirectoryScoped();
      for (
        const file of [
          ".config/mise/tasks/_scripts/helpers",
          `.config/mise/tasks/setup/deps/${verb}`,
        ]
      ) {
        yield* fs.makeDirectory(path.dirname(path.join(root, file)), {
          recursive: true,
        });
        yield* fs.writeFileString(path.join(root, file), files.get(file)!);
      }
      yield* fs.makeDirectory(path.join(root, "bin"));
      yield* fs.writeFileString(path.join(root, "bin/mise"), stubMise(listing));
      yield* fs.chmod(path.join(root, "bin/mise"), 0o755);
      const log = path.join(root, "runs.log");
      yield* fs.writeFileString(log, "");
      const { status } = spawnSync("bash", [
        path.join(root, `.config/mise/tasks/setup/deps/${verb}`),
      ], {
        env: {
          ...process.env,
          ...env,
          PATH: `${path.join(root, "bin")}:${process.env.PATH}`,
          MISE_PROJECT_ROOT: root,
          LOG: log,
        },
      });
      return {
        status,
        runs: (yield* fs.readFileString(log))
          .split("\n")
          .filter(Boolean),
      };
    })
    .pipe(Effect.scoped, Effect.provide(NodeServices.layer));

describe("dependency dispatchers", () => {
  const listing =
    "setup:ai setup:deps:install setup:deps:install:demo setup:deps:install:demo:sub";

  it.effect("run each tool's install task with no flag it may not declare, passing --frozen as DEPS_FROZEN", () =>
    Effect.gen(function*() {
      expect(yield* dispatch("install", { usage_frozen: "true" }, listing))
        .toEqual({
          status: 0,
          runs: ["run setup:deps:install:demo frozen=1"],
        });
      expect((yield* dispatch("install", {}, listing)).runs).toEqual([
        "run setup:deps:install:demo frozen=",
      ]);
    }));

  it.effect("fail when the task list cannot be read", () =>
    Effect.gen(function*() {
      const { status, runs } = yield* dispatch("audit", {});
      expect(status).not.toBe(0);
      expect(runs).toEqual([]);
    }));
});

/** Whether a `gitleaks` binary is on PATH; the scan test needs the real scanner. */
const hasGitleaks = spawnSync("gitleaks", ["version"]).status === 0;

describe("secret scanning", () => {
  it.effect.skipIf(!hasGitleaks)(
    "finds well-known secret shapes with the rendered config",
    () =>
      Effect
        .gen(function*() {
          const fs = yield* FileSystem.FileSystem;
          const path = yield* Path.Path;
          const files = yield* render([toolsByName.get("gitleaks")!], values);
          const root = yield* fs.makeTempDirectoryScoped();
          const config = path.join(root, "gitleaks.toml");
          yield* fs.writeFileString(
            config,
            files.get(".config/gitleaks.toml")!,
          );
          const tree = path.join(root, "tree");
          yield* fs.makeDirectory(tree);
          // Built at run time so this file holds no secret-shaped literal.
          const aws = ["AKIA", "QYLPMN5HJRFPZAM2"].join("");
          const github = ["ghp", "_", "k9Xb2LmQ7vR4tW8yZ1cN5pD3fG6hJ0sA2eUq"]
            .join("");
          yield* fs.writeFileString(
            path.join(tree, "settings.env"),
            `AWS_KEY=${aws}\nGITHUB_TOKEN=${github}\n`,
          );
          const report = path.join(root, "report.json");
          const { status } = spawnSync("gitleaks", [
            "dir",
            tree,
            "--config",
            config,
            "--no-banner",
            "--report-format",
            "json",
            "--report-path",
            report,
          ]);
          expect(status).toBe(1);
          const found = JSON.parse(yield* fs.readFileString(report)).map((
            leak: { RuleID: string; },
          ) => leak.RuleID);
          expect(found).toEqual(
            expect.arrayContaining(["aws-access-token", "github-pat"]),
          );
        })
        .pipe(Effect.scoped, Effect.provide(NodeServices.layer)),
  );
});

describe("AI setup", () => {
  it.effect("runs with no plugins installed under the system bash", () =>
    Effect
      .gen(function*() {
        const fs = yield* FileSystem.FileSystem;
        const path = yield* Path.Path;
        const files = yield* render([
          toolsByName.get("mise")!,
          toolsByName.get("claude")!,
        ], values);
        const root = yield* fs.makeTempDirectoryScoped();
        for (
          const file of [
            ".config/mise/tasks/_scripts/helpers",
            ".config/mise/tasks/setup/ai",
          ]
        ) {
          yield* fs.makeDirectory(path.dirname(path.join(root, file)), {
            recursive: true,
          });
          yield* fs.writeFileString(path.join(root, file), files.get(file)!);
        }
        const bin = path.join(root, "bin");
        yield* fs.makeDirectory(bin);
        // `claude` lists no plugins and accepts every other call; `jq` turns that list into no lines.
        yield* fs.writeFileString(
          path.join(bin, "claude"),
          "#!/bin/sh\nexit 0\n",
        );
        yield* fs.writeFileString(
          path.join(bin, "jq"),
          "#!/bin/sh\ncat >/dev/null\n",
        );
        yield* fs.chmod(path.join(bin, "claude"), 0o755);
        yield* fs.chmod(path.join(bin, "jq"), 0o755);
        const { status, stderr } = spawnSync("/bin/bash", [
          path.join(root, ".config/mise/tasks/setup/ai"),
        ], {
          env: {
            ...process.env,
            PATH: `${bin}:${process.env.PATH}`,
            MISE_PROJECT_ROOT: root,
          },
        });
        expect(stderr.toString()).not.toMatch(/unbound variable/);
        expect(status).toBe(0);
      })
      .pipe(Effect.scoped, Effect.provide(NodeServices.layer)));
});

/** `git` in `root` with a throwaway identity, so a test can commit. */
const git = (root: string, ...args: ReadonlyArray<string>) =>
  spawnSync("git", [
    "-c",
    "user.name=t",
    "-c",
    "user.email=t@example.com",
    "-c",
    "commit.gpgsign=false",
    ...args,
  ], {
    cwd: root,
  });

/**
 * Writes the rendered `helpers` and `task` into a fresh git repository, puts `stubs` (name → shell body) first on
 * PATH, lets `prepare` lay out the tree, then runs the task under bash from the repository root.
 */
const runTask = (
  tools: ReadonlyArray<string>,
  task: string,
  stubs: Readonly<Record<string, string>>,
  prepare: (
    root: string,
  ) => Effect.Effect<void, unknown, FileSystem.FileSystem | Path.Path>,
  env: Readonly<Record<string, string>> = {},
) =>
  Effect
    .gen(function*() {
      const fs = yield* FileSystem.FileSystem;
      const path = yield* Path.Path;
      const files = yield* render(
        [...tools, "mise"].map(name => toolsByName.get(name)!),
        values,
      );
      const root = yield* fs.makeTempDirectoryScoped();
      git(root, "init", "-q");
      for (const file of [".config/mise/tasks/_scripts/helpers", task]) {
        yield* fs.makeDirectory(path.dirname(path.join(root, file)), {
          recursive: true,
        });
        yield* fs.writeFileString(path.join(root, file), files.get(file)!);
      }
      const bin = yield* fs.makeTempDirectoryScoped();
      const log = path.join(bin, "calls.log");
      yield* fs.writeFileString(log, "");
      for (const [name, body] of Object.entries(stubs)) {
        yield* fs.writeFileString(
          path.join(bin, name),
          `#!/bin/sh\necho "${name} $*" >>"${log}"\n${body}\n`,
        );
        yield* fs.chmod(path.join(bin, name), 0o755);
      }
      yield* prepare(root);
      const { status, stdout, stderr } = spawnSync("/bin/bash", [
        path.join(root, task),
      ], {
        cwd: root,
        env: {
          ...process.env,
          ...env,
          PATH: `${bin}:${process.env.PATH}`,
          MISE_PROJECT_ROOT: root,
        },
      });
      return {
        status,
        output: `${stdout}${stderr}`,
        calls: (yield* fs.readFileString(log)).split("\n").filter(Boolean),
      };
    })
    .pipe(Effect.scoped, Effect.provide(NodeServices.layer));

describe("code:format", () => {
  const format = ".config/mise/tasks/code/format";
  const stubs = { dprint: "exit 0", "sort-package-json": "exit 0" };

  it.effect("skips the package.json sorter in a repository that tracks none", () =>
    Effect.gen(function*() {
      const { status, calls } = yield* runTask(
        ["dprint"],
        format,
        stubs,
        () => Effect.void,
      );
      expect(status).toBe(0);
      expect(calls.filter(call => call.startsWith("sort-package-json")))
        .toEqual([]);
    }));

  it.effect("sorts exactly the tracked package.json files", () =>
    Effect.gen(function*() {
      const { status, calls } = yield* runTask(
        ["dprint"],
        format,
        stubs,
        root =>
          Effect.gen(function*() {
            const fs = yield* FileSystem.FileSystem;
            const path = yield* Path.Path;
            yield* fs.makeDirectory(path.join(root, "app"));
            for (
              const file of [
                "package.json",
                "app/package.json",
                "untracked.json",
              ]
            ) {
              yield* fs.writeFileString(path.join(root, file), "{}\n");
            }
            yield* fs.makeDirectory(path.join(root, "loose"));
            yield* fs.writeFileString(
              path.join(root, "loose/package.json"),
              "{}\n",
            );
            git(root, "add", "package.json", "app/package.json");
            // A tracked manifest since deleted from disk, and one under a non-ASCII path git would quote.
            for (const dir of ["gone", "ünï"]) {
              yield* fs.makeDirectory(path.join(root, dir));
              yield* fs.writeFileString(
                path.join(root, dir, "package.json"),
                "{}\n",
              );
            }
            git(root, "add", "gone/package.json", "ünï/package.json");
            yield* fs.remove(path.join(root, "gone/package.json"));
          }),
      );
      expect(status).toBe(0);
      expect(calls.filter(call => call.startsWith("sort-package-json")))
        .toEqual([
          "sort-package-json --check app/package.json package.json ünï/package.json",
        ]);
    }));
});

describe("code:sec", () => {
  const sec = ".config/mise/tasks/code/sec";
  // Built at run time so this file holds no secret-shaped literal.
  const token = ["ghp", "_", "k9Xb2LmQ7vR4tW8yZ1cN5pD3fG6hJ0sA2eUq"].join("");
  /** Lays out a repository whose ignored `.env` holds a token; `commit` also commits one in `leak.txt`. */
  const tree = (commit: boolean) => (root: string) =>
    Effect.gen(function*() {
      const fs = yield* FileSystem.FileSystem;
      const path = yield* Path.Path;
      const files = yield* render([toolsByName.get("gitleaks")!], values);
      yield* fs.writeFileString(
        path.join(root, ".config/gitleaks.toml"),
        files.get(".config/gitleaks.toml")!,
      );
      yield* fs.writeFileString(path.join(root, ".gitignore"), ".env\n");
      yield* fs.writeFileString(
        path.join(root, ".env"),
        `GITHUB_TOKEN=${token}\n`,
      );
      if (commit) {
        yield* fs.writeFileString(
          path.join(root, "leak.txt"),
          `GITHUB_TOKEN=${token}\n`,
        );
      }
      git(root, "add", "-A");
      git(root, "commit", "-q", "-m", "init");
    });

  it.effect.skipIf(!hasGitleaks)(
    "ignores a secret in a file git ignores",
    () =>
      Effect.gen(function*() {
        const { status, output } = yield* runTask(["gitleaks"], sec, {
          grype: "exit 0",
        }, tree(false));
        expect(output).not.toMatch(/leaks found: [1-9]/);
        expect(status).toBe(0);
      }),
  );

  it.effect.skipIf(!hasGitleaks)(
    "fails on a committed secret",
    () =>
      Effect.gen(function*() {
        const { status, output } = yield* runTask(["gitleaks"], sec, {
          grype: "exit 0",
        }, tree(true));
        expect(status)
          .not
          .toBe(0);
        expect(output).toMatch(
          /Fingerprint: [0-9a-f]{40}:leak\.txt:github-pat:1/,
        );
        expect(output).toMatch(/\.gitleaksignore/);
        const runs = Array.from({ length: token.length - 7 }, (_, at) =>
          token.slice(at, at + 8));
        expect(runs
          .filter(run =>
            output.includes(run)
          ))
          .toEqual([]);
      }),
  );

  it.effect.skipIf(!hasGitleaks)(
    "passes once a rotated secret's fingerprint is in .gitleaksignore",
    () =>
      Effect.gen(function*() {
        const { status } = yield* runTask(["gitleaks"], sec, {
          grype: "exit 0",
        }, root =>
          Effect.gen(function*() {
            const fs = yield* FileSystem.FileSystem;
            const path = yield* Path.Path;
            yield* tree(true)(root);
            const commit = git(root, "rev-parse", "HEAD")
              .stdout
              .toString()
              .trim();
            yield* fs.writeFileString(
              path.join(root, ".gitleaksignore"),
              `${commit}:leak.txt:github-pat:1\n`,
            );
          }));
        expect(status).toBe(0);
      }),
  );
});
