import type { Values } from "@/setup-config/schema";
import {
  type Tool,
  tools,
  toolsByName,
} from "@/tool/catalog";
import {
  render,
  renderFiles,
} from "@/tool/render";
import {
  templates,
  templatesFrom,
} from "@/tool/templates";
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
import { build } from "tsdown";
import { parse } from "yaml";
import tsdownConfig from "../../tsdown.config";

const fixtures = templatesFrom(
  "../fixtures/templates/",
  import.meta.glob<string>("../fixtures/templates/**", {
    query: "?raw",
    import: "default",
    eager: true,
    exhaustive: true,
  }),
);

const tool = (name: string): Tool => toolsByName.get(name)!;

const values: typeof Values.Type = {
  repo: "virajp/bootstrap",
  commit_scopes: ["cli"],
  merge_model: { develop: "direct", main: "pr" },
  tools: ["github"],
};

describe("render", () => {
  it.effect("renders <% if %> and <%= %> from the setup values and derived names", () =>
    Effect.gen(function*() {
      const rendered = yield* render([tool("github")], values, fixtures);
      expect(rendered.get(".github/pull_request_template.md")).toBe(
        "# bootstrap\n\nRepository: virajp/bootstrap.\n\n\n\nReaches main by pull request.\n\n\n",
      );
    }));

  it.effect("derives the repo name from the last segment of a nested repo path", () =>
    Effect.gen(function*() {
      const rendered = yield* render(
        [tool("github")],
        {
          ...values,
          repo: "group/sub/name",
          merge_model: { develop: "pr", main: "direct" },
        },
        fixtures,
      );
      expect(rendered.get(".github/pull_request_template.md")).toBe(
        "# name\n\nRepository: group/sub/name.\n\n\n",
      );
    }));

  it.effect("returns one entry per catalog path of the selected tools", () =>
    Effect.gen(function*() {
      const selected = [tool("github"), tool("grype")];
      const rendered = yield* render(selected, values, fixtures);
      expect([...rendered.keys()].sort()).toEqual(
        selected.flatMap(t => t.files.map(file => file.path)).sort(),
      );
      expect(rendered.get(".config/grype.yaml")).toBe(
        "fail-on-severity: high\n",
      );
    }));

  it.effect("fails with a typed render error naming the path on an unknown filter", () =>
    Effect.gen(function*() {
      const error = yield* render(
        [tool("grype")],
        values,
        new Map([["grype/.config/grype.yaml", "<%= repo | jsn %>"]]),
      )
        .pipe(Effect.flip);
      expect(error._tag).toBe("RenderError");
      expect(error.path).toBe(".config/grype.yaml");
    }));

  it.effect("fails with a typed render error naming the path on an undefined variable", () =>
    Effect.gen(function*() {
      const error = yield* render(
        [tool("grype")],
        values,
        new Map([["grype/.config/grype.yaml", "<%= nope %>"]]),
      )
        .pipe(
          Effect.flip,
        );
      expect(error._tag).toBe("RenderError");
      expect(error.path).toBe(".config/grype.yaml");
    }));

  it.effect("fails with a typed error when a catalog path has no template", () =>
    Effect.gen(function*() {
      const error = yield* render([tool("grype")], values, new Map()).pipe(
        Effect.flip,
      );
      expect(error._tag).toBe("TemplateMissing");
      expect(error.path).toBe(".config/grype.yaml");
    }));

  it.effect("does not serve a path from a template under another tool's folder", () =>
    Effect.gen(function*() {
      const error = yield* render(
        [tool("grype")],
        values,
        new Map([["github/.config/grype.yaml", "x"]]),
      )
        .pipe(
          Effect.flip,
        );
      expect(error._tag).toBe("TemplateMissing");
      expect(error.path).toBe(".config/grype.yaml");
    }));
});

describe("templatesFrom", () => {
  it("keys each template by its tool folder and target path", () => {
    expect(
      templatesFrom("../t/", {
        "../t/mise/.config/mise.toml": "a",
        "../t/git/.gitignore": "b",
      }),
    )
      .toEqual(
        new Map([["mise/.config/mise.toml", "a"], ["git/.gitignore", "b"]]),
      );
  });

  it("embeds every real template in the folder of the catalog tool that owns its path", () => {
    const owned = new Set(
      tools.flatMap(t => t.files.map(file => `${t.name}/${file.path}`)),
    );
    expect([...templates.keys()].filter(key => !owned.has(key))).toEqual([]);
  });
});

describe("build embedding", () => {
  it.effect(
    "bundles every template under the templates root as text and leaves out .DS_Store",
    () =>
      Effect
        .gen(function*() {
          const fs = yield* FileSystem.FileSystem;
          const path = yield* Path.Path;
          const root = yield* fs.makeTempDirectoryScoped();
          const template =
            "embedded <%= repo %>\n<% if tools %>{{ config_root }}<% endif %>\n";
          const junk = "finder junk 7f3a";
          yield* fs.makeDirectory(path.join(root, "templates/demo/.config"), {
            recursive: true,
          });
          yield* fs.makeDirectory(path.join(root, "src/tool"), {
            recursive: true,
          });
          yield* fs.writeFileString(
            path.join(root, "templates/demo/.config/demo.toml"),
            template,
          );
          yield* fs.writeFileString(
            path.join(root, "templates/demo/.DS_Store"),
            junk,
          );
          // The real templates.ts, so its own glob (../../templates/**) is what the build expands.
          yield* fs.copyFile(
            path.join(import.meta.dirname, "../../src/tool/templates.ts"),
            path.join(root, "src/tool/templates.ts"),
          );
          yield* fs.writeFileString(
            path.join(root, "src/bin.ts"),
            "export { templates } from \"./tool/templates\";\n",
          );
          yield* Effect.promise(() =>
            build({
              ...tsdownConfig,
              config: false,
              cwd: root,
              entry: { bin: "src/bin.ts" },
              outDir: path.join(root, "dist"),
              logLevel: "silent",
            })
          );
          const bundle = yield* fs.readFileString(
            path.join(root, "dist/bin.mjs"),
          );
          expect(bundle).toContain(JSON.stringify(template));
          expect(bundle).not.toContain(junk);
          expect(bundle).not.toMatch(/\bimport\b[^;]*\?raw/);
        })
        .pipe(Effect.scoped, Effect.provide(NodeServices.layer)),
    30_000,
  );
});

describe("renderFiles", () => {
  const all = tools.map(t => t.name);
  const input = (
    selection: ReadonlyArray<string>,
    originHost?: string,
  ) => ({ values, selection, originHost });

  it.effect("carries the path, the rendered content, the executable flag and createOnly", () =>
    Effect.gen(function*() {
      const files = yield* renderFiles(input(["grype"]), fixtures);
      expect(files.get(".config/grype.yaml")).toEqual({
        path: ".config/grype.yaml",
        content: "fail-on-severity: high\n",
        executable: false,
        createOnly: false,
      });
    }));

  it.effect("renders the files of the selected tools only", () =>
    Effect.gen(function*() {
      const files = yield* renderFiles(input(["grype", "github"]), fixtures);
      expect([...files.keys()].sort()).toEqual(
        [tool("github"), tool("grype")]
          .flatMap(t => t.files.map(file => file.path))
          .sort(),
      );
    }));

  it.effect("carries createOnly from the catalog, so mempalace.yaml is create-only", () =>
    Effect.gen(function*() {
      const files = yield* renderFiles(input(all, "github.com"));
      expect(files.get("mempalace.yaml")!.createOnly).toBe(true);
      for (const t of tools) {
        for (const file of t.files) {
          expect([file.path, files.get(file.path)!.createOnly]).toEqual([
            file.path,
            file.createOnly,
          ]);
        }
      }
    }));

  it.effect("marks every file under .config/mise/tasks/ executable and no other", () =>
    Effect.gen(function*() {
      const files = yield* renderFiles(input(all, "github.com"));
      expect(files.size).toBe(templates.size);
      for (const [path, file] of files) {
        expect([path, file.executable]).toEqual([
          path,
          path.startsWith(".config/mise/tasks/"),
        ]);
      }
    }));

  it.effect("marks a rendered file executable exactly when its template file is", () =>
    Effect
      .gen(function*() {
        const fs = yield* FileSystem.FileSystem;
        const path = yield* Path.Path;
        const files = yield* renderFiles(input(all));
        for (const t of tools) {
          for (const file of t.files) {
            const { mode } = yield* fs.stat(
              path.join(
                import.meta.dirname,
                "../../templates",
                t.name,
                file.path,
              ),
            );
            expect({
              file: file.path,
              executable: files.get(file.path)!.executable,
            })
              .toEqual({ file: file.path, executable: (mode & 0o111) !== 0 });
          }
        }
      })
      .pipe(Effect.provide(NodeServices.layer)));
});

describe("selection-dependent files", () => {
  const baseDev = ".config/mise/conf.d/_base/mise.dev.toml";
  const fixed = ["node", "pnpm", "python", "uv", "jq", "yq"];
  const entries = (text: string) =>
    [...text.matchAll(/^\[tools\.(.+)\]$/gm)].map(([, name]) => name);
  const dev = (selection: ReadonlyArray<string>) =>
    renderFiles({ values, selection, originHost: undefined }).pipe(
      Effect.map(files => entries(files.get(baseDev)!.content)),
    );

  it.effect("_base/mise.dev.toml holds the fixed tools and every selected base tool", () =>
    Effect.gen(function*() {
      expect(yield* dev(tools.map(t => t.name))).toEqual([
        ...fixed,
        "\"pre-commit\"",
        "dprint",
        "taplo",
        "\"npm:@askviraj/linter\"",
        "\"npm:sort-package-json\"",
      ]);
    }));

  it.effect("_base/mise.dev.toml holds only the fixed tools when no base tool with entries is selected", () =>
    Effect.gen(function*() {
      expect(yield* dev(["mise"])).toEqual(fixed);
    }));

  it.effect("_base/mise.dev.toml renders no dprint entry with dprint not selected", () =>
    Effect.gen(function*() {
      const got = yield* dev(["mise", "pre-commit", "virajp-linter"]);
      expect(got).toEqual([
        ...fixed,
        "\"pre-commit\"",
        "\"npm:@askviraj/linter\"",
      ]);
      expect(got.join(" ")).not.toMatch(/dprint|sort-package-json|taplo/);
    }));

  const changelog = (selection: ReadonlyArray<string>, originHost?: string) =>
    renderFiles({ values, selection, originHost }).pipe(
      Effect.map(files =>
        parse(files.get(".config/git-conventional-commits.yaml")!.content)
          .changelog
      ),
    );
  const github = {
    commitUrl: "https://github.com/virajp/bootstrap/commit/%commit%",
    commitRangeUrl:
      "https://github.com/virajp/bootstrap/compare/%from%...%to%?diff=split",
    issueUrl: "https://github.com/virajp/bootstrap/issues/%issue%",
  };
  const gitlab = {
    commitUrl: "https://gitlab.com/virajp/bootstrap/-/commit/%commit%",
    commitRangeUrl:
      "https://gitlab.com/virajp/bootstrap/-/compare/%from%...%to%",
    issueUrl: "https://gitlab.com/virajp/bootstrap/-/issues/%issue%",
  };

  it.effect("links the forge whose hosts hold the origin host", () =>
    Effect.gen(function*() {
      expect(
        yield* changelog(["pre-commit", "github", "gitlab"], "gitlab.com"),
      )
        .toMatchObject(gitlab);
      expect(
        yield* changelog(["pre-commit", "github", "gitlab"], "github.com"),
      )
        .toMatchObject(github);
    }));

  it.effect("links the first selected forge in catalog order when no host matches", () =>
    Effect.gen(function*() {
      expect(
        yield* changelog(["pre-commit", "gitlab", "github"], "example.org"),
      )
        .toMatchObject(github);
      expect(yield* changelog(["pre-commit", "gitlab", "github"]))
        .toMatchObject(github);
      expect(yield* changelog(["pre-commit", "gitlab"], "github.com"))
        .toMatchObject(gitlab);
    }));

  it.effect("has no forge links with no forge selected", () =>
    Effect.gen(function*() {
      const got = yield* changelog(["pre-commit"], "github.com");
      for (const key of ["commitUrl", "commitRangeUrl", "issueUrl"]) {
        expect(got).not.toHaveProperty(key);
      }
      expect(got.issueRegexPattern).toBe("#[0-9]+");
    }));
});
