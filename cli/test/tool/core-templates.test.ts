import { describe, expect, it } from "@effect/vitest";
import { Effect } from "effect";
import { parse } from "yaml";
import { coreTools } from "@/tool/catalog";
import { render, type RenderValues } from "@/tool/render";

/** The core tools whose templates this suite covers; `fnox` ships with the new templates. */
const owned = coreTools.filter((tool) => tool.name !== "fnox");

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
  /\bbootstrap\b/i,
  /doppler/i,
  /vwf/i,
  /stackgen/i,
  /(^|[^\w-])site\//m,
  /claude-plugins/i,
  /flutter/i,
  /\bdart/i,
  /pubspec/i,
];

describe("core templates", () => {
  it("covers the twelve core tools other than fnox", () => {
    expect(owned.map((tool) => tool.name)).toHaveLength(12);
  });

  it.effect("render exactly the catalog paths of the core tools", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect([...files.keys()].sort()).toEqual(owned.flatMap((tool) => tool.files).sort());
    }));

  it.effect("carry nothing specific to this repository", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      const leaks = [...files].flatMap(([path, text]) =>
        forbidden.filter((pattern) => pattern.test(text)).map((pattern) => `${path}: ${pattern}`)
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
      expect(parse(files.get(".config/git-conventional-commits.yaml")!).convention.commitScopes).toEqual(["api"]);
    }));

  it.effect("write the scope list in the shipped formatter's bracket spacing", () =>
    Effect.gen(function*() {
      const line = (scopes: ReadonlyArray<string>) =>
        render(owned, { ...values, commit_scopes: scopes }).pipe(
          Effect.map((files) => files.get(".config/git-conventional-commits.yaml")!.match(/^ {2}commitScopes:.*$/m)![0]),
        );
      expect(yield* line(["api", "web"])).toBe("  commitScopes: [ \"api\", \"web\" ]");
      expect(yield* line([])).toBe("  commitScopes: []");
    }));

  it.effect("render an empty scope list as an empty sequence", () =>
    Effect.gen(function*() {
      const files = yield* render(owned, { ...values, commit_scopes: [] });
      expect(parse(files.get(".config/git-conventional-commits.yaml")!).convention.commitScopes).toEqual([]);
    }));

  it.effect("set both merge models in the mise base config and no single one", () =>
    Effect.gen(function*() {
      const base = (yield* rendered()).get(".config/mise/conf.d/_base/mise.toml")!;
      expect(base).toMatch(/^MERGE_MODEL_DEVELOP\s*=\s*"direct"$/m);
      expect(base).toMatch(/^MERGE_MODEL_MAIN\s*=\s*"pr"$/m);
      expect(base).not.toMatch(/^MERGE_MODEL\s*=/m);
      expect(base).toMatch(/^MEMBERS\s*=\s*""$/m);
      expect(base).toMatch(/^REPO_NAME\s*=\s*"widgets"$/m);
    }));

  it.effect("dispatch each dependency verb to the tools' own tasks, never to a package manager", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      for (const verb of ["install", "outdated", "audit", "upgrade", "cleanup"]) {
        const task = files.get(`.config/mise/tasks/setup/deps/${verb}`)!;
        expect(task).toMatch(new RegExp(`^run_tool_tasks setup:deps:${verb}\\b`, "m"));
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
      expect(() => JSON.parse(files.get(".vscode/settings.json")!)).not.toThrow();
      const dprint = JSON.parse(files.get(".config/dprint.json")!);
      expect(dprint.json.jsonTrailingCommaFiles).not.toContain(".vscode/settings.json");
    }));

  it.effect("recommend each editor extension once", () =>
    Effect.gen(function*() {
      const { recommendations } = JSON.parse((yield* rendered()).get(".vscode/extensions.json")!);
      expect(recommendations).toEqual([...new Set(recommendations)]);
    }));

  it.effect("install fnox, the core secrets tool, through mise", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(files.get(".config/mise/conf.d/_base/mise.dev.toml")).toMatch(/^\[tools\.fnox\]$/m);
    }));

  it.effect("point the changelog links at GitHub by default", () =>
    Effect.gen(function*() {
      const { changelog } = parse((yield* rendered()).get(".config/git-conventional-commits.yaml")!);
      expect(changelog.commitUrl).toBe("https://github.com/acme/widgets/commit/%commit%");
      expect(changelog.commitRangeUrl).toBe("https://github.com/acme/widgets/compare/%from%...%to%?diff=split");
      expect(changelog.issueUrl).toBe("https://github.com/acme/widgets/issues/%issue%");
    }));

  it.effect("point the changelog links at GitLab when gitlab is the selected forge", () =>
    Effect.gen(function*() {
      const files = yield* render(owned, { ...values, tools: ["gitlab"] });
      const { changelog } = parse(files.get(".config/git-conventional-commits.yaml")!);
      expect(changelog.commitUrl).toBe("https://gitlab.com/acme/widgets/-/commit/%commit%");
      expect(changelog.commitRangeUrl).toBe("https://gitlab.com/acme/widgets/-/compare/%from%...%to%");
      expect(changelog.issueUrl).toBe("https://gitlab.com/acme/widgets/-/issues/%issue%");
    }));

  it.effect("keep number- and null-like values strings in every structured file", () =>
    Effect.gen(function*() {
      const files = yield* render(owned, { ...values, repo: "null/2048", commit_scopes: ["404"] });
      expect(parse(files.get("mempalace.yaml")!).wing).toBe("2048");
      expect(JSON.parse(files.get(".config/claude-status.json")!).projectName).toBe("null/2048");
      expect(parse(files.get(".config/git-conventional-commits.yaml")!).convention.commitScopes).toEqual(["404"]);
      expect(files.get(".config/mise/conf.d/_base/mise.toml")).toMatch(/^REPO_NAME\s*=\s*"2048"$/m);
      expect(files.get(".config/mise/conf.d/ai/mise.dev.toml")).toMatch(
        /^MEMPALACE_PALACE_PATH\s*=\s*"~\/\.local\/share\/mempalace\/2048"$/m,
      );
    }));

  it.effect("format the whole tree from the root dprint config", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      expect(files.get("dprint.json")).toContain("\"extends\": \".config/dprint.json\"");
      expect(files.get(".config/mise/tasks/code/format")).not.toMatch(/--config \.config\/dprint\.json/);
    }));

  it.effect("start every task file with a shebang", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      const tasks = [...files].filter(([path]) => path.startsWith(".config/mise/tasks/"));
      expect(tasks.length).toBeGreaterThan(0);
      expect(tasks.filter(([, text]) => !text.startsWith("#!")).map(([path]) => path)).toEqual([]);
    }));
});
