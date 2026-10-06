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
        selected.flatMap(t => t.files).sort(),
      );
      expect(rendered.get(".config/grype.yaml")).toBe(
        "fail-on-severity: high\n",
      );
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
      tools.flatMap(t => t.files.map(path => `${t.name}/${path}`)),
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
  it.effect("marks a rendered file executable exactly when its template file is", () =>
    Effect
      .gen(function*() {
        const fs = yield* FileSystem.FileSystem;
        const path = yield* Path.Path;
        const files = yield* renderFiles(tools, values);
        expect(files.size).toBe(templates.size);
        for (const t of tools) {
          for (const file of t.files) {
            const { mode } = yield* fs.stat(
              path.join(import.meta.dirname, "../../templates", t.name, file),
            );
            expect({ file, executable: files.get(file)!.executable }).toEqual({
              file,
              executable: (mode & 0o111) !== 0,
            });
          }
        }
      })
      .pipe(Effect.provide(NodeServices.layer)));

  it.effect("carries the rendered content with the executable flag", () =>
    Effect.gen(function*() {
      const files = yield* renderFiles([tool("grype")], values, fixtures);
      expect(files.get(".config/grype.yaml")).toEqual({
        content: "fail-on-severity: high\n",
        executable: false,
      });
    }));
});
