import { describe, expect, it } from "@effect/vitest";
import { Effect } from "effect";
import { spawnSync } from "node:child_process";
import { parse } from "yaml";
import { toolsByName } from "@/tool/catalog";
import { render, type RenderValues } from "@/tool/render";

const values: RenderValues = {
  repo: "acme/widgets",
  commit_scopes: ["api"],
  merge_model: { develop: "direct", main: "pr" },
  tools: ["github", "gitlab"],
};

const renderTool = (name: string) => render([toolsByName.get(name)!], values);

/** Text this repo carries, and secret material, that no target repository may inherit. */
const forbidden: ReadonlyArray<RegExp> = [
  /virajp/i,
  /\bbootstrap\b/i,
  /doppler/i,
  /vwf/i,
  /stackgen/i,
  /askviraj/i,
  /token\s*=|password\s*=|api[_-]?key\s*=/i,
];

/**
 * No TOML parser is a dependency of the cli package; taplo (a dev tool of this repo) parses
 * TOML to JSON. Without it on PATH the TOML assertion is skipped.
 */
const taplo = spawnSync("taplo", ["--version"]).status === 0;
const parseToml = (text: string): unknown => {
  const result = spawnSync("taplo", ["get", "-o", "json"], { input: text, encoding: "utf8" });
  if (result.status !== 0) throw new Error(result.stderr);
  return JSON.parse(result.stdout);
};

describe("new templates", () => {
  for (const name of ["fnox", "github", "gitlab"]) {
    it.effect(`render exactly the catalog paths of ${name}`, () =>
      Effect.gen(function*() {
        const files = yield* renderTool(name);
        expect([...files.keys()].sort()).toEqual([...toolsByName.get(name)!.files].sort());
      }));

    it.effect(`carry nothing specific to this repository and no secret in ${name}`, () =>
      Effect.gen(function*() {
        const files = yield* renderTool(name);
        const leaks = [...files].flatMap(([path, text]) =>
          forbidden.filter((pattern) => pattern.test(text)).map((pattern) => `${path}: ${pattern}`)
        );
        expect(leaks).toEqual([]);
      }));
  }

  it.effect("turn off blank issues on GitHub", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("github");
      expect(parse(files.get(".github/ISSUE_TEMPLATE/config.yml")!).blank_issues_enabled).toBe(false);
    }));

  it.effect("give every GitHub issue form a name, a description and a body", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("github");
      for (const form of [".github/ISSUE_TEMPLATE/bug.yml", ".github/ISSUE_TEMPLATE/feature.yml"]) {
        const parsed = parse(files.get(form)!);
        expect(typeof parsed.name).toBe("string");
        expect(typeof parsed.description).toBe("string");
        expect(Array.isArray(parsed.body) && parsed.body.length > 0).toBe(true);
      }
    }));

  it.effect("render every .yml as YAML", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("github");
      const yml = [...files].filter(([path]) => path.endsWith(".yml"));
      expect(yml).toHaveLength(3);
      for (const [, text] of yml) expect(() => parse(text)).not.toThrow();
    }));

  it.effect.skipIf(!taplo)("render .config/fnox.toml as TOML with no provider and no secret", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("fnox");
      expect(parseToml(files.get(".config/fnox.toml")!)).toEqual({ providers: {}, secrets: {} });
    }));

  it.effect("set fnox up from the secrets task", () =>
    Effect.gen(function*() {
      const task = (yield* renderTool("fnox")).get(".config/mise/tasks/setup/secrets")!;
      expect(task.startsWith("#!")).toBe(true);
      expect(task).toMatch(/^\s*fnox check\b/m);
    }));

  it.effect("place the GitLab merge-request template at Default.md", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("gitlab");
      expect(files.has(".gitlab/merge_request_templates/Default.md")).toBe(true);
    }));
});
