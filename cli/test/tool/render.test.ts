import { describe, expect, it } from "@effect/vitest";
import { NodeServices } from "@effect/platform-node";
import { Effect, FileSystem, Path } from "effect";
import { build } from "tsdown";
import tsdownConfig from "../../tsdown.config";
import { toolsByName, type Tool } from "@/tool/catalog";
import { render, type RenderValues } from "@/tool/render";
import { templatesFrom } from "@/tool/templates";

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

const values: RenderValues = {
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
        "# bootstrap\n\nOwner: virajp, repository: virajp/bootstrap.\n\nReaches main by pull request.\n\n",
      );
    }));

  it.effect("derives the owner path from every segment before the name", () =>
    Effect.gen(function*() {
      const rendered = yield* render(
        [tool("github")],
        { ...values, repo: "group/sub/name", merge_model: { develop: "pr", main: "direct" } },
        fixtures,
      );
      expect(rendered.get(".github/pull_request_template.md")).toBe(
        "# name\n\nOwner: group/sub, repository: group/sub/name.\n\n",
      );
    }));

  it.effect("returns one entry per catalog path of the selected tools", () =>
    Effect.gen(function*() {
      const selected = [tool("github"), tool("grype")];
      const rendered = yield* render(selected, values, fixtures);
      expect([...rendered.keys()].sort()).toEqual(selected.flatMap((t) => t.files).sort());
      expect(rendered.get(".config/grype.yaml")).toBe("fail-on-severity: high\n");
    }));

  it.effect("fails with a typed render error naming the path on an undefined variable", () =>
    Effect.gen(function*() {
      const error = yield* render([tool("grype")], values, new Map([[".config/grype.yaml", "<%= nope %>"]])).pipe(
        Effect.flip,
      );
      expect(error._tag).toBe("RenderError");
      expect(error.path).toBe(".config/grype.yaml");
    }));

  it.effect("fails with a typed error when a catalog path has no template", () =>
    Effect.gen(function*() {
      const error = yield* render([tool("grype")], values, new Map()).pipe(Effect.flip);
      expect(error._tag).toBe("TemplateMissing");
      expect(error.path).toBe(".config/grype.yaml");
    }));
});

describe("templatesFrom", () => {
  it("keys each template by its target path under the tool folder", () => {
    expect(templatesFrom("../t/", { "../t/mise/.config/mise.toml": "a", "../t/git/.gitignore": "b" })).toEqual(
      new Map([[".config/mise.toml", "a"], [".gitignore", "b"]]),
    );
  });
});


describe("build embedding", () => {
  it.effect("bundles every file under a templates root into the JS as text", () =>
    Effect.gen(function*() {
      const fs = yield* FileSystem.FileSystem;
      const path = yield* Path.Path;
      const root = yield* fs.makeTempDirectoryScoped();
      const template = "embedded <%= repo %>\n<% if tools %>{{ config_root }}<% endif %>\n";
      yield* fs.makeDirectory(path.join(root, "templates/demo/.config"), { recursive: true });
      yield* fs.writeFileString(path.join(root, "templates/demo/.config/demo.toml"), template);
      yield* fs.writeFileString(
        path.join(root, "entry.ts"),
        "export const t = import.meta.glob(\"./templates/**\", "
          + "{ query: \"?raw\", import: \"default\", eager: true, exhaustive: true });\n",
      );
      yield* Effect.promise(() =>
        build({
          ...tsdownConfig,
          config: false,
          cwd: root,
          entry: { bin: "entry.ts" },
          outDir: path.join(root, "dist"),
          logLevel: "silent",
        })
      );
      const bundle = yield* fs.readFileString(path.join(root, "dist/bin.mjs"));
      expect(bundle).toContain(JSON.stringify(template));
      expect(bundle).not.toContain("?raw");
    }).pipe(Effect.scoped, Effect.provide(NodeServices.layer)), 30_000);
});
