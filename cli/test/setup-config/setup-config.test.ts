import {
  parse,
  read,
  write,
} from "@/setup-config/io";
import type { SetupConfig } from "@/setup-config/schema";
import type { Running } from "@/setup-config/validity";
import type { RenderValues } from "@/tool/render";
import {
  describe,
  expect,
  it,
} from "@effect/vitest";
import {
  Effect,
  FileSystem,
  Option,
  Path,
} from "effect";
import { stringify } from "yaml";

const running: Running = { version: "1.2.0", format: 1 };

const valid: SetupConfig = {
  format: 1,
  version: "1.2.0",
  values: {
    repo: "virajp/bootstrap",
    commit_scopes: ["cli", "site-docs"],
    merge_model: { develop: "direct", main: "pr" },
    tools: ["github"],
  },
  files: [".gitignore", ".config/mise.toml"],
};

/** A setup config file text built from `valid` with the given top-level overrides. */
const text = (overrides: Record<string, unknown> = {}) =>
  stringify({ ...valid, ...overrides });

/** Parses the text and returns the typed failure. */
const failure = (source: string) => Effect.flip(parse(source, running));

/** Every error carries the what, why and fix text. */
const expectProblem = (error: { what: string; why: string; fix: string; }) => {
  expect(error.what).not.toBe("");
  expect(error.why).not.toBe("");
  expect(error.fix).not.toBe("");
};

describe("setup config schema", () => {
  it.effect("decodes a valid file", () =>
    Effect.gen(function*() {
      const config = yield* parse(text(), running);
      expect(config).toEqual(valid);
    }));

  it("types values so the renderer accepts them", () => {
    const values: RenderValues = valid.values;
    expect(values.repo).toBe("virajp/bootstrap");
  });

  it.effect.each(
    [
      ["an extra top-level key", { extra: true }],
      ["an extra values key", { values: { ...valid.values, members: [] } }],
      ["a format of 0", { format: 0 }],
      ["a non-integer format", { format: 1.5 }],
      ["a non-semver version", { version: "1.2" }],
      ["a repo with no owner", {
        values: { ...valid.values, repo: "bootstrap" },
      }],
      ["a commit scope not kebab-case", {
        values: { ...valid.values, commit_scopes: ["Cli"] },
      }],
      ["a duplicate commit scope", {
        values: { ...valid.values, commit_scopes: ["cli", "cli"] },
      }],
      ["a merge model outside direct | pr", {
        values: {
          ...valid.values,
          merge_model: { develop: "squash", main: "pr" },
        },
      }],
      ["a duplicate tool", {
        values: { ...valid.values, tools: ["github", "github"] },
      }],
      ["missing files", { files: undefined }],
      ["a duplicate kept path", { kept: ["a", "a"] }],
    ] satisfies Array<[string, Record<string, unknown>]>,
  )(
    "refuses %s as a schema violation",
    ([, overrides]) =>
      Effect.gen(function*() {
        const error = yield* failure(text(overrides));
        expect(error._tag).toBe("SchemaViolation");
        expectProblem(error);
      }),
  );

  it.effect.each([
    "o/{{exec(command='id')}}",
    "o/{% raw %}",
    "o/{n}",
    "o/n\"x",
    "o/n'x",
    "o/%n",
    "o/n#x",
    "o\\n/x",
    "o/n\u0007",
    "o/n:x",
    "./name",
    "o/..",
    "o/./n",
    "o//n",
  ])(
    "refuses repo %j: segments are letters, digits, . _ - and never . or ..",
    repo =>
      Effect.gen(function*() {
        const error = yield* failure(
          text({ values: { ...valid.values, repo } }),
        );
        expect(error._tag).toBe("SchemaViolation");
      }),
  );

  it.effect.each([
    "virajp/bootstrap",
    "group/sub.group/my_repo-1",
    "o/.dotfiles",
    "o/n..x",
  ])(
    "accepts repo %s",
    repo =>
      Effect.gen(function*() {
        const config = yield* parse(
          text({ values: { ...valid.values, repo } }),
          running,
        );
        expect(config.values.repo).toBe(repo);
      }),
  );

  it.effect("refuses text that is not YAML as unreadable", () =>
    Effect.gen(function*() {
      const error = yield* failure("format: [1\n");
      expect(error._tag).toBe("SetupConfigUnreadable");
      expectProblem(error);
    }));
});

describe("setup config validity", () => {
  it.effect("refuses a newer format, before the schema", () =>
    Effect.gen(function*() {
      const error = yield* failure(text({ format: 2, renamed: true }));
      expect(error._tag).toBe("FormatTooNew");
      expectProblem(error);
    }));

  it.effect.each(["1.2.1", "1.3.0", "2.0.0"])(
    "refuses a newer version %s with the fix 'upgrade bootstrap'",
    version =>
      Effect.gen(function*() {
        const error = yield* failure(text({ version }));
        expect(error._tag).toBe("VersionTooNew");
        expect(error.fix).toContain("upgrade bootstrap");
        expectProblem(error);
      }),
  );

  it.effect.each(["1.2.0", "1.1.9", "0.9.0", "1.2.0-rc.1", "1.2.0+build"])(
    "accepts the same or an older version %s",
    version =>
      Effect.gen(function*() {
        const config = yield* parse(text({ version }), running);
        expect(config.version).toBe(version);
      }),
  );

  it.effect("refuses a prerelease newer than the running prerelease", () =>
    Effect.gen(function*() {
      const error = yield* Effect.flip(
        parse(text({ version: "1.2.0-rc.2" }), {
          version: "1.2.0-rc.1",
          format: 1,
        }),
      );
      expect(error._tag).toBe("VersionTooNew");
    }));

  it.effect.each(
    [
      ["1.2.0-rc.1.1", "VersionTooNew"],
      ["1.2.0-rd", "VersionTooNew"],
      ["1.2.0-rc.1", "ok"],
      ["1.2.0-rc", "ok"],
      ["1.2.0-1", "ok"],
      ["1.2.0-alpha", "ok"],
    ] satisfies Array<[string, string]>,
  )(
    "orders prerelease %s against 1.2.0-rc.1 by semver: %s",
    ([version, outcome]) =>
      Effect.gen(function*() {
        const result = yield* parse(text({ version }), {
          version: "1.2.0-rc.1",
          format: 1,
        })
          .pipe(
            Effect.map(() => "ok"),
            Effect.catch(error => Effect.succeed(error._tag)),
          );
        expect(result).toBe(outcome);
      }),
  );

  it.effect("refuses a tool this version does not ship as removed", () =>
    Effect.gen(function*() {
      const error = yield* failure(
        text({ values: { ...valid.values, tools: ["bitbucket"] } }),
      );
      expect(error._tag).toBe("ToolRemoved");
      expect(error.what).toContain("bitbucket");
      expect(error.fix).toContain("remove");
      expectProblem(error);
    }));

  it.effect("refuses a core tool name as invalid", () =>
    Effect.gen(function*() {
      const error = yield* failure(
        text({ values: { ...valid.values, tools: ["mise"] } }),
      );
      expect(error._tag).toBe("CoreToolListed");
      expect(error.what).toContain("mise");
      expectProblem(error);
    }));

  it.effect("refuses a path in both kept and deleted", () =>
    Effect.gen(function*() {
      const error = yield* failure(
        text({ kept: [".gitignore"], deleted: [".gitignore"] }),
      );
      expect(error._tag).toBe("PathKeptAndDeleted");
      expect(error.what).toContain(".gitignore");
      expect(error.fix).toContain("one");
      expectProblem(error);
    }));

  it.effect.each(
    [
      ["files", "/etc/passwd"],
      ["files", "C:\\repo\\a"],
      ["kept", "../outside"],
      ["kept", "a/../../b"],
      ["deleted", "*.md"],
      ["deleted", "src/?.ts"],
      ["files", ".config/[ab].toml"],
      ["files", ""],
      ["files", "a//b"],
      ["kept", "./a"],
      ["kept", "a/./b"],
      ["kept", "a/."],
      ["deleted", ".config/"],
      ["files", ".git/config"],
      ["kept", "sub/.git/hooks/pre-commit"],
      ["files", ".GIT/config"],
      ["kept", "a/.Git/hooks/x"],
      ["files", ".git./config"],
      ["files", ".git /config"],
      ["deleted", ".git. . /config"],
      ["files", "GIT~1/config"],
      ["kept", "a/git~1/hooks/x"],
      ["files", "a\u0000b"],
      ["deleted", "a\nb"],
      ["files", "a\u007fb"],
    ] satisfies Array<[string, string]>,
  )(
    "refuses a %s path %s that is not literal and repository-relative",
    ([list, path]) =>
      Effect.gen(function*() {
        const error = yield* failure(text({ [list]: [path] }));
        expect(error._tag).toBe("InvalidPath");
        expect(error.what).toContain(JSON.stringify(path));
        expect(error.what).toContain(list);
        expectProblem(error);
      }),
  );

  it.effect("accepts dot-named files and folders that are not .git", () =>
    Effect.gen(function*() {
      const files = [
        ".github/pull_request_template.md",
        ".gitignore",
        ".GITHUB/x",
        "a/b.c/.d",
        "git/x",
        "git~2/x",
      ];
      const config = yield* parse(text({ files }), running);
      expect(config.files).toEqual(files);
    }));
});

describe("setup config write", () => {
  it("omits empty kept and deleted", () => {
    const out = write({ ...valid, kept: [], deleted: [] });
    expect(out).not.toContain("kept");
    expect(out).not.toContain("deleted");
  });

  it("never lists .config/bootstrap.yaml in files", () => {
    const out = write({
      ...valid,
      files: [...valid.files, ".config/bootstrap.yaml"],
    });
    expect(out).not.toContain("bootstrap.yaml");
  });

  it("sorts every path list by path", () => {
    const out = write({ ...valid, kept: ["b", "a"], deleted: ["d", "c"] });
    const config = Effect.runSync(parse(out, running));
    expect(config.files).toEqual([".config/mise.toml", ".gitignore"]);
    expect(config.kept).toEqual(["a", "b"]);
    expect(config.deleted).toEqual(["c", "d"]);
  });

  it.effect("round-trips through a read", () =>
    Effect.gen(function*() {
      const config = { ...valid, kept: [".gitignore"] };
      const reread = yield* parse(write(config), running);
      expect(reread).toEqual({
        ...config,
        files: [".config/mise.toml", ".gitignore"],
      });
    }));
});

describe("setup config read", () => {
  const fileSystem = (files: Record<string, string>) =>
    FileSystem.layerNoop({
      exists: path => Effect.succeed(path in files),
      readFileString: path => Effect.succeed(files[path]!),
    });

  it.effect("reads <root>/.config/bootstrap.yaml", () =>
    Effect
      .gen(function*() {
        const config = yield* read("/repo", running);
        expect(config).toEqual(Option.some(valid));
      })
      .pipe(
        Effect.provide(fileSystem({ "/repo/.config/bootstrap.yaml": text() })),
        Effect.provide(Path.layer),
      ));

  it.effect("returns absent, not an error, when the file does not exist", () =>
    Effect
      .gen(function*() {
        const config = yield* read("/repo", running);
        expect(config).toEqual(Option.none());
      })
      .pipe(Effect.provide(fileSystem({})), Effect.provide(Path.layer)));

  it.effect("fails validity on a read file", () =>
    Effect
      .gen(function*() {
        const error = yield* Effect.flip(read("/repo", running));
        expect(error._tag).toBe("CoreToolListed");
      })
      .pipe(
        Effect.provide(fileSystem({
          "/repo/.config/bootstrap.yaml": text({
            values: { ...valid.values, tools: ["git"] },
          }),
        })),
        Effect.provide(Path.layer),
      ));

  it.effect("reports a file it cannot read as unreadable", () =>
    Effect
      .gen(function*() {
        const error = yield* Effect.flip(read("/repo", running));
        expect(error._tag).toBe("SetupConfigUnreadable");
        expectProblem(error);
      })
      .pipe(
        Effect.provide(
          FileSystem.layerNoop({ exists: () => Effect.succeed(true) }),
        ),
        Effect.provide(Path.layer),
      ));
});
