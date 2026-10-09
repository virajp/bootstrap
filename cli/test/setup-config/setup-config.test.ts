import {
  requireSetUp,
  versionGuard,
} from "@/setup-config/guard";
import {
  parse,
  read,
  write,
} from "@/setup-config/io";
import {
  refreshFiles,
  type RewriteMode,
  rewrites,
} from "@/setup-config/refresh";
import { repair } from "@/setup-config/repair";
import type { SetupConfig } from "@/setup-config/schema";
import type { Running } from "@/setup-config/validity";
import {
  toolsByName,
  unremovableTools,
} from "@/tool/catalog";
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
  Runtime,
} from "effect";
import { stringify } from "yaml";

const running: Running = { version: "1.2.0", format: 1 };

/** Names of the unremovable tools, read from the catalog. */
const unremovable = unremovableTools.map(tool => tool.name);

const valid: SetupConfig = {
  format: 1,
  version: "1.2.0",
  values: {
    repo: "virajp/bootstrap",
    commit_scopes: ["cli", "site-docs"],
    merge_model: { develop: "direct", main: "pr" },
    tools: [...unremovable, "github"],
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
      ["an empty tools list", { values: { ...valid.values, tools: [] } }],
      ["a commit scope ending in -", {
        values: { ...valid.values, commit_scopes: ["a-"] },
      }],
      ["a commit scope with --", {
        values: { ...valid.values, commit_scopes: ["a--b"] },
      }],
      ["missing files", { files: undefined }],
      ["a kept key", { kept: [] }],
      ["a deleted key", { deleted: [] }],
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
      ["1.2.0-rc.a", "VersionTooNew"],
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

  it.effect("accepts the unremovable tools in values.tools", () =>
    Effect.gen(function*() {
      const config = yield* parse(text(), running);
      expect(config.values.tools).toEqual(expect.arrayContaining(unremovable));
    }));

  /** Parses `valid` with the given tools and returns the typed failure. */
  const toolsFailure = (tools: ReadonlyArray<string>) =>
    failure(text({ values: { ...valid.values, tools } }));

  /** A Validity failure: the tag, what, why and fix, and exit code 3. */
  const expectValidity = (
    error: { _tag: string; what: string; why: string; fix: string; },
    tag: string,
  ) => {
    expect(error._tag).toBe(tag);
    expectProblem(error);
    expect((error as unknown as Record<string, unknown>)[Runtime.errorExitCode]).toBe(3);
  };

  it.effect("refuses a tool whose required tool is not selected", () =>
    Effect.gen(function*() {
      const error = yield* toolsFailure([...unremovable, "taplo"]);
      expectValidity(error, "RequiredToolMissing");
      expect(error.what).toContain("taplo");
      expect(error.why).toContain("dprint");
    }));

  it.effect("checks Validity before the version guard: an older version with a Validity failure exits 3", () =>
    Effect.gen(function*() {
      const error = yield* failure(
        text({ version: "1.0.0", values: { ...valid.values, tools: [...unremovable, "taplo"] } }),
      );
      expectValidity(error, "RequiredToolMissing");
    }));

  it.effect("refuses a removed tool with exit code 3 and the fix of the entity doc", () =>
    Effect.gen(function*() {
      const error = yield* toolsFailure([...unremovable, "bitbucket"]);
      expectValidity(error, "ToolRemoved");
      expect(error.fix).toBe("remove `bitbucket` from values.tools");
    }));

  it.effect.each(
    [
      ["files", "/etc/passwd"],
      ["files", "C:\\repo\\a"],
      ["files", "../outside"],
      ["files", "a/../../b"],
      ["files", "*.md"],
      ["files", "src/?.ts"],
      ["files", ".config/[ab].toml"],
      ["files", ""],
      ["files", "a//b"],
      ["files", "./a"],
      ["files", "a/./b"],
      ["files", "a/."],
      ["files", ".config/"],
      ["files", ".git/config"],
      ["files", "sub/.git/hooks/pre-commit"],
      ["files", ".GIT/config"],
      ["files", "a/.Git/hooks/x"],
      ["files", ".git./config"],
      ["files", ".git /config"],
      ["files", ".git. . /config"],
      ["files", "GIT~1/config"],
      ["files", "a/git~1/hooks/x"],
      ["files", "a\u0000b"],
      ["files", "a\nb"],
      ["files", "a\u007fb"],
      ["files", ".git::$INDEX_ALLOCATION/config"],
      ["files", "a:b"],
      ["files", "GIT~2/config"],
      ["files", "a/git~10/x"],
      ["files", ".g\u200cit/config"],
      ["files", "\ufeff.git/config"],
      ["files", ".git\u200b/config"],
      ["files", ".GIT\u202e/x"],
      ["files", ".git\u2060./x"],
      ["files", ".. /x"],
      ["files", ".../x"],
      ["files", "a/. /b"],
      ["files", "a/ /b"],
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
        "a.../b",
        "git~x/y",
      ];
      const config = yield* parse(text({ files }), running);
      expect(config.files).toEqual(files);
    }));
});

/** A running cli newer than the recorded file, so write must restamp format and version. */
const newer: Running = { version: "1.3.0", format: 2 };

describe("setup config write", () => {
  it("never lists .config/bootstrap.yaml in files", () => {
    const out = write({
      ...valid,
      files: [...valid.files, ".config/bootstrap.yaml"],
    }, running);
    expect(out).not.toContain("bootstrap.yaml");
  });

  it("writes keys in schema order, the three lists sorted, and the running format and version", () => {
    const out = write({
      files: ["b", "a"],
      values: {
        tools: ["mise", "git", "github"],
        merge_model: { main: "pr", develop: "direct" },
        commit_scopes: ["site", "cli"],
        repo: "virajp/bootstrap",
      },
      version: "1.2.0",
      format: 1,
    }, newer);
    expect(out).toBe(
      [
        "format: 2",
        "version: 1.3.0",
        "values:",
        "  repo: virajp/bootstrap",
        "  commit_scopes:",
        "    - cli",
        "    - site",
        "  merge_model:",
        "    develop: direct",
        "    main: pr",
        "  tools:",
        "    - git",
        "    - github",
        "    - mise",
        "files:",
        "  - a",
        "  - b",
        "",
      ].join("\n"),
    );
    expect(out).not.toContain("#");
  });

  it.effect("round-trips through a read", () =>
    Effect.gen(function*() {
      const reread = yield* parse(write(valid, running), running);
      expect(reread).toEqual({
        ...valid,
        values: { ...valid.values, tools: valid.values.tools.toSorted() },
        files: [".config/mise.toml", ".gitignore"],
      });
    }));
});

describe("setup config repair on read", () => {
  it.effect("adds back a missing unremovable tool whose category holds no other tool", () =>
    Effect.gen(function*() {
      const [missing, ...rest] = unremovable;
      const config = yield* parse(
        text({ values: { ...valid.values, tools: [...rest, "github"] } }),
        running,
      );
      const repaired = repair(config);
      expect(repaired.config.values.tools).toEqual(expect.arrayContaining([...unremovable, "github"]));
      expect(repaired.added).toEqual([missing]);
      expect(repaired.warnings).toEqual([`added back unremovable tool \`${missing}\``]);
    }));

  it.effect("leaves a complete selection as is, with no warning", () =>
    Effect.gen(function*() {
      const config = yield* parse(text(), running);
      expect(repair(config)).toEqual({ config, added: [], warnings: [] });
    }));
});

describe("setup config guards", () => {
  it.effect("requireSetUp fails an absent file with exit 1 and 'not set up — run `bootstrap init`'", () =>
    Effect.gen(function*() {
      const error = yield* Effect.flip(requireSetUp(Option.none()));
      expect(error._tag).toBe("NotSetUp");
      expect(error.what).toBe("not set up — run `bootstrap init`");
      expect(error.fix).toBe("bootstrap init");
      expect(error[Runtime.errorExitCode]).toBe(1);
    }));

  it.effect("requireSetUp passes a present file", () =>
    Effect.gen(function*() {
      expect(yield* requireSetUp(Option.some("present"))).toBe("present");
    }));

  it.effect("versionGuard fails an older recorded version with exit 1 and next command `bootstrap init`", () =>
    Effect.gen(function*() {
      const config = yield* parse(text({ version: "1.1.9" }), running);
      const error = yield* Effect.flip(versionGuard(config, running));
      expect(error._tag).toBe("VersionTooOld");
      expect(error.what).toContain("1.1.9");
      expect(error.why).not.toBe("");
      expect(error.fix).toBe("bootstrap init");
      expect(error[Runtime.errorExitCode]).toBe(1);
    }));

  it.effect("versionGuard passes an equal version, build metadata ignored", () =>
    Effect.gen(function*() {
      const config = yield* parse(text({ version: "1.2.0+build" }), running);
      expect(yield* versionGuard(config, running)).toBe(config);
    }));

  it("versionGuard accepts only a config that passed Validity", () => {
    // @ts-expect-error -- a decoded but unvalidated config is not a ValidSetupConfig
    const guarded = () => versionGuard(valid, running);
    expect(guarded).toBeTypeOf("function");
  });
});

describe("setup config refresh of files", () => {
  const recorded = (files: ReadonlyArray<string>): SetupConfig => ({ ...valid, files: [...files] });
  const removedTool = toolsByName.get("taplo")!;
  const removedPath = removedTool.files[0]!.path;

  it("full mode lists every rendered path, also one absent on disk", () => {
    const next = refreshFiles(recorded(["a"]), { mode: "full" }, new Set(["a", "b"]), new Set(["a"]));
    expect(next.config.files).toEqual(["a", "b"]);
    expect(next.orphaned).toEqual([]);
  });

  it("remove mode drops the removed tools' paths and adds no rendered path", () => {
    const next = refreshFiles(
      recorded(["a", removedPath]),
      { mode: "remove", removed: [removedTool.name], added: [] },
      new Set(["a", "b"]),
      new Set(["a", removedPath]),
    );
    expect(next.config.files).toEqual(["a"]);
    expect(next.orphaned).toEqual([]);
  });

  it("remove mode adds the paths of a tool that Repair on read added back", () => {
    const [missing] = unremovableTools;
    const paths = missing!.files.map(file => file.path);
    const next = refreshFiles(
      recorded(["a"]),
      { mode: "remove", removed: [], added: [missing!.name] },
      new Set(["a", "b", ...paths]),
      new Set(["a"]),
    );
    expect(next.config.files.toSorted()).toEqual(["a", ...paths].toSorted());
  });

  it.each([
    ["full", { mode: "full" } as const],
    ["remove", { mode: "remove", removed: [], added: [] } as const],
  ])("%s mode keeps a recorded path not rendered and on disk as orphaned, and drops one absent", (_, mode) => {
    const next = refreshFiles(
      recorded(["a", "kept", "gone", ".config/bootstrap.yaml"]),
      mode,
      new Set(["a", ".config/bootstrap.yaml"]),
      new Set(["a", "kept", ".config/bootstrap.yaml"]),
    );
    expect(next.config.files).toEqual(["a", "kept"]);
    expect(next.orphaned).toEqual(["kept"]);
  });
});

describe("setup config rewrite rule", () => {
  it.each([
    ["init", "a: 1\n", "a: 1\n", false, false],
    ["init", "# note\na: 1\n", "a: 1\n", false, true],
    ["init", "version: 1.1.0\n", "version: 1.2.0\n", false, true],
    ["tui-apply", "b: 2\na: 1\n", "a: 1\nb: 2\n", false, true],
    ["tui-apply", "a: 1\n", "a: 1\n", false, false],
    ["add", "a: 1\n", "a: 1\n", true, true],
    ["add", "# note\na: 1\n", "a: 1\n", false, false],
    ["remove", "version: 1.1.0\n", "version: 1.2.0\n", false, false],
    ["remove", "a: 1\n", "a: 2\n", true, true],
  ] satisfies Array<[RewriteMode, string, string, boolean, boolean]>)(
    "%s with recorded %j and render %j, tool change or repair %s: rewrite %s",
    (mode, recordedText, nextText, changed, expected) => {
      expect(rewrites(mode, recordedText, nextText, changed)).toBe(expected);
    },
  );
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
        expect(config).toEqual(Option.some({ config: valid, added: [], warnings: [] }));
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
        expect(error._tag).toBe("RequiredToolMissing");
      })
      .pipe(
        Effect.provide(fileSystem({
          "/repo/.config/bootstrap.yaml": text({
            version: "1.0.0",
            values: { ...valid.values, tools: [...unremovable, "taplo"] },
          }),
        })),
        Effect.provide(Path.layer),
      ));

  it.effect("repairs a read file and returns the warnings beside it", () =>
    Effect
      .gen(function*() {
        const [missing] = unremovable;
        const { config, added, warnings } = Option.getOrThrow(yield* read("/repo", running));
        expect(config.values.tools).toContain(missing);
        expect(added).toEqual([missing]);
        expect(warnings).toEqual([`added back unremovable tool \`${missing}\``]);
      })
      .pipe(
        Effect.provide(fileSystem({
          "/repo/.config/bootstrap.yaml": text({
            values: { ...valid.values, tools: [...unremovable.slice(1), "github"] },
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
