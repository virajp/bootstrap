/** The Validity rules of docs/blueprint/entities/setup-config/index.md beyond the schema. */
import {
  type SetupConfig,
  setupConfigPath as file,
} from "@/setup-config/schema";
import { toolsByName } from "@/tool/catalog";
import {
  Data,
  Effect,
} from "effect";

/** The running cli: its version and the newest setup-config format it reads. */
export interface Running {
  readonly version: string;
  readonly format: number;
}

/** What went wrong, why, and the fix — the shape of the conventions' error report. */
interface Problem {
  readonly what: string;
  readonly why: string;
  readonly fix: string;
}

/** The file could not be read or is not YAML. */
export class SetupConfigUnreadable
  extends Data.TaggedError("SetupConfigUnreadable")<Problem>
{}
/** The file does not conform to schema.yaml. */
export class SchemaViolation
  extends Data.TaggedError("SchemaViolation")<Problem>
{}
/** `format` is newer than the running cli understands. */
export class FormatTooNew extends Data.TaggedError("FormatTooNew")<Problem> {}
/** `version` is newer than the running cli. */
export class VersionTooNew extends Data.TaggedError("VersionTooNew")<Problem> {}
/** A name in `values.tools` is not a tool the running cli ships. */
export class ToolRemoved extends Data.TaggedError("ToolRemoved")<Problem> {}
/** A path is in both `kept` and `deleted`. */
export class PathKeptAndDeleted
  extends Data.TaggedError("PathKeptAndDeleted")<Problem>
{}
/** A path in `files`, `kept` or `deleted` is not a literal repository-relative file path. */
export class InvalidPath extends Data.TaggedError("InvalidPath")<Problem> {}

export type SetupConfigError =
  | SetupConfigUnreadable
  | SchemaViolation
  | FormatTooNew
  | VersionTooNew
  | ToolRemoved
  | PathKeptAndDeleted
  | InvalidPath;

/** Fails when `format` is newer than the running cli reads; applied before the schema, whose shape may differ. */
export const checkFormat = (
  format: unknown,
  running: Running,
): Effect.Effect<void, FormatTooNew> =>
  Number.isInteger(format) && (format as number) > running.format
    ? Effect.fail(
      new FormatTooNew({
        what: `${file} has format ${format}`,
        why: `this bootstrap reads format ${running.format} or older`,
        fix: "upgrade bootstrap",
      }),
    )
    : Effect.void;

/** Semver 2.0 precedence of a prerelease identifier pair. */
const compareIdentifier = (a: string, b: string): number => {
  const numeric = /^\d+$/;
  if (numeric.test(a) && numeric.test(b)) {
    return Number(a) - Number(b);
  }
  if (numeric.test(a) !== numeric.test(b)) {
    return numeric.test(a) ? -1 : 1;
  }
  return a < b ? -1 : a > b ? 1 : 0;
};

/** Semver 2.0 precedence: negative when `a` is older than `b`; build metadata is ignored. */
export const compareVersions = (a: string, b: string): number => {
  const split = (version: string) => {
    const [core = "", pre] = version.split("+")[0]!.split(/-(.*)/s);
    return {
      core: core.split(".").map(Number),
      pre: pre === undefined ? [] : pre.split("."),
    };
  };
  const left = split(a);
  const right = split(b);
  for (let i = 0; i < 3; i++) {
    const diff = left.core[i]! - right.core[i]!;
    if (diff !== 0) {
      return diff;
    }
  }
  // A version without a prerelease outranks one with.
  if (left.pre.length === 0 || right.pre.length === 0) {
    return right.pre.length - left.pre.length;
  }
  for (let i = 0; i < Math.min(left.pre.length, right.pre.length); i++) {
    const diff = compareIdentifier(left.pre[i]!, right.pre[i]!);
    if (diff !== 0) {
      return diff;
    }
  }
  return left.pre.length - right.pre.length;
};

// eslint-disable-next-line no-control-regex -- control characters are what this refuses
const hasControl = (path: string) => /[\u0000-\u001f\u007f]/.test(path);
const isAbsolute = (path: string) => /^([/\\]|[A-Za-z]:)/.test(path);
/**
 * A path segment as a case-insensitive, HFS+ or Windows file system resolves it: Unicode
 * default-ignorable code points removed, trailing dots and spaces stripped, lower-cased.
 */
const normalise = (segment: string) =>
  segment
    .replace(/\p{Default_Ignorable_Code_Point}/gu, "")
    .replace(/[. ]+$/, "")
    .toLowerCase();
const segments = (path: string) => path.split(/[/\\]/);
/** A segment that resolves to `.git`, or to an 8.3 short name of the `git~<n>` form. */
const isGitSegment = (segment: string) =>
  /^(\.git|git~\d+)$/.test(normalise(segment));
const hasWildcard = (path: string) => /[*?[\]{}]/.test(path);

/** Why a path is not a literal repository-relative file path, or undefined when it is. */
const pathProblem = (path: string): string | undefined =>
  path === ""
    ? "it is empty"
    : hasControl(path)
    ? "it has a control character"
    : isAbsolute(path)
    ? "it is absolute"
    : /[/\\]$/.test(path)
    ? "it ends in `/`, so it names a directory"
    : path.includes(":")
    ? "it has a `:`, which names a drive or a file stream"
    : segments(path).some(segment =>
        normalise(segment) === "" && segment.startsWith("..")
      )
    ? "it has a `..` segment"
    : segments(path).some(segment => normalise(segment) === "")
    ? "it has an empty or `.` segment"
    : segments(path).some(isGitSegment)
    ? "it is inside a `.git` folder"
    : hasWildcard(path)
    ? "it has a wildcard"
    : undefined;

/** Applies the Validity rules beyond the schema to a decoded setup config; fails on the first violation. */
export const validate = (
  config: SetupConfig,
  running: Running,
): Effect.Effect<SetupConfig, SetupConfigError> =>
  Effect.gen(function*() {
    yield* checkFormat(config.format, running);
    if (compareVersions(config.version, running.version) > 0) {
      return yield* new VersionTooNew({
        what: `${file} was written by bootstrap ${config.version}`,
        why: `this is bootstrap ${running.version}, which is older`,
        fix: "upgrade bootstrap",
      });
    }
    for (const name of config.values.tools) {
      if (!toolsByName.has(name)) {
        return yield* new ToolRemoved({
          what: `tool "${name}" in values.tools of ${file}`,
          why:
            `bootstrap ${running.version} does not ship "${name}": it is removed in this version`,
          fix: `remove "${name}" from values.tools in ${file}`,
        });
      }
    }
    const deleted = new Set(config.deleted);
    const both = config.kept?.find(path => deleted.has(path));
    if (both !== undefined) {
      return yield* new PathKeptAndDeleted({
        what: `"${both}" is in both kept and deleted of ${file}`,
        why: "a path is either kept or deleted, never both",
        fix: `remove "${both}" from one list`,
      });
    }
    const lists = {
      files: config.files,
      kept: config.kept ?? [],
      deleted: config.deleted ?? [],
    };
    for (const [list, paths] of Object.entries(lists)) {
      for (const path of paths) {
        const why = pathProblem(path);
        if (why !== undefined) {
          return yield* new InvalidPath({
            what: `${JSON.stringify(path)} in ${list} of ${file}`,
            why:
              `a path must be a literal file path relative to the repository root; ${why}`,
            fix: `correct or remove ${JSON.stringify(path)} in ${list}`,
          });
        }
      }
    }
    return config;
  });
