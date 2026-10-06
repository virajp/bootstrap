import { toolsByName } from "@/tool/catalog";
import {
  type RenderValues,
  render,
} from "@/tool/render";
import {
  describe,
  expect,
  it,
} from "@effect/vitest";
import { Effect } from "effect";
import { spawnSync } from "node:child_process";
import {
  mkdirSync,
  mkdtempSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import {
  dirname,
  join,
} from "node:path";
import { parse } from "yaml";

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
  const result = spawnSync("taplo", ["get", "-o", "json"], {
    input: text,
    encoding: "utf8",
  });
  if (result.status !== 0) {
    throw new Error(result.stderr);
  }
  return JSON.parse(result.stdout);
};

/** dprint, with the shipped config, checks formatting; without it on PATH that assertion is skipped. */
const dprint = spawnSync("dprint", ["--version"]).status === 0;

describe("new templates", () => {
  for (const name of ["fnox", "github", "gitlab"]) {
    it.effect(`render exactly the catalog paths of ${name}`, () =>
      Effect.gen(function*() {
        const files = yield* renderTool(name);
        expect([...files.keys()].sort()).toEqual(
          [...toolsByName.get(name)!.files].sort(),
        );
      }));

    it.effect(`carry nothing specific to this repository and no secret in ${name}`, () =>
      Effect.gen(function*() {
        const files = yield* renderTool(name);
        const leaks = [...files].flatMap(([path, text]) =>
          forbidden.filter(pattern => pattern.test(text)).map(pattern =>
            `${path}: ${pattern}`
          )
        );
        expect(leaks).toEqual([]);
      }));
  }

  it.effect("turn off blank issues on GitHub", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("github");
      expect(
        parse(files.get(".github/ISSUE_TEMPLATE/config.yml")!)
          .blank_issues_enabled,
      )
        .toBe(false);
    }));

  it.effect("give every GitHub issue form a name, a description and a body", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("github");
      for (
        const form of [
          ".github/ISSUE_TEMPLATE/bug.yml",
          ".github/ISSUE_TEMPLATE/feature.yml",
        ]
      ) {
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
      for (const [, text] of yml) {
        expect(() => parse(text)).not.toThrow();
      }
    }));

  it.effect.skipIf(!taplo)(
    "render .config/fnox.toml as TOML with no provider and no secret",
    () =>
      Effect.gen(function*() {
        const files = yield* renderTool("fnox");
        expect(parseToml(files.get(".config/fnox.toml")!)).toEqual({
          providers: {},
          secrets: {},
        });
      }),
  );

  it.effect("set fnox up from the secrets task", () =>
    Effect.gen(function*() {
      const task = (yield* renderTool("fnox")).get(
        ".config/mise/tasks/setup/secrets",
      )!;
      expect(task.startsWith("#!")).toBe(true);
      expect(task).toMatch(/^\s*fnox check\b/m);
    }));

  it.effect("skip with a hint to install the pinned tools, never to edit a managed mise config", () =>
    Effect.gen(function*() {
      const task = (yield* renderTool("fnox")).get(
        ".config/mise/tasks/setup/secrets",
      )!;
      expect(task).toMatch(/^\s*exit 0$/m);
      expect(task).toMatch(/mise install/);
      expect(task).not.toMatch(/mise use/);
    }));

  it.effect("place the GitLab merge-request template at Default.md", () =>
    Effect.gen(function*() {
      const files = yield* renderTool("gitlab");
      expect(files.has(".gitlab/merge_request_templates/Default.md")).toBe(
        true,
      );
    }));

  it.effect.skipIf(!dprint || !taplo)(
    "format to the shipped dprint config with no diff",
    () =>
      Effect.gen(function*() {
        const own = ["fnox", "github", "gitlab"].map(name =>
          toolsByName.get(name)!
        );
        const shipped = ["dprint", "taplo"].map(name => toolsByName.get(name)!);
        const files = yield* render([...shipped, ...own], values);
        const root = mkdtempSync(join(tmpdir(), "new-templates-"));
        try {
          for (const [path, text] of files) {
            mkdirSync(dirname(join(root, path)), { recursive: true });
            writeFileSync(join(root, path), text);
          }
          const paths = own.flatMap(tool => tool.files).filter(path =>
            /\.(md|yml|toml)$/.test(path)
          );
          const result = spawnSync("dprint", ["check", ...paths], {
            cwd: root,
            encoding: "utf8",
          });
          expect(result.stdout + result.stderr).toBe("");
          expect(result.status).toBe(0);
        }
        finally {
          rmSync(root, { recursive: true, force: true });
        }
      }),
  );
});
