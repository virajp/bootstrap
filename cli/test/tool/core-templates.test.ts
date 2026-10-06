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

  it.effect("start every task file with a shebang", () =>
    Effect.gen(function*() {
      const files = yield* rendered();
      const tasks = [...files].filter(([path]) => path.startsWith(".config/mise/tasks/"));
      expect(tasks.length).toBeGreaterThan(0);
      expect(tasks.filter(([, text]) => !text.startsWith("#!")).map(([path]) => path)).toEqual([]);
    }));
});
