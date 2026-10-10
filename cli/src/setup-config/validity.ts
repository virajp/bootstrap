/** The Validity rules of docs/blueprint/entities/setup-config/index.md beyond the schema. */
import {
  type SetupConfig,
  setupConfigPath as file,
} from "@/setup-config/schema";
import {
  checkRequires,
  toolsByName,
  unremovableTools,
} from "@/tool/catalog";
import { categoriesByName } from "@/tool-category/catalog";
import {
  type Brand,
  Data,
  Effect,
  Runtime,
} from "effect";

/** The running cli: its version and the newest setup-config format it reads. */
export interface Running {
  readonly version: string;
  readonly format: number;
}

/** What went wrong, why, and the fix — the shape of the conventions' error report. */
export interface Problem {
  readonly what: string;
  readonly why: string;
  readonly fix: string;
}

/** A tagged error that fails a Validity check: exit code 3. */
const ValidityError = <Tag extends string>(tag: Tag) =>
  class extends Data.TaggedError(tag)<Problem> {
    readonly [Runtime.errorExitCode]: number = 3;
  };

/** The file could not be read or is not YAML. */
export class SetupConfigUnreadable
  extends ValidityError("SetupConfigUnreadable")
{}
/** The file does not conform to schema.yaml. */
export class SchemaViolation extends ValidityError("SchemaViolation") {}
/** `format` is newer than the running cli understands. */
export class FormatTooNew extends ValidityError("FormatTooNew") {}
/** `version` is newer than the running cli. */
export class VersionTooNew extends ValidityError("VersionTooNew") {}
/** A name in `values.tools` is not a tool the running cli ships. */
export class ToolRemoved extends ValidityError("ToolRemoved") {}
/** Two tools in `values.tools` share a `max: one` category. */
export class MaxOneExceeded extends ValidityError("MaxOneExceeded") {}
/** A `max: one` category misses its unremovable tool and holds another tool in its place. */
export class UnremovableReplaced extends ValidityError("UnremovableReplaced") {}
/** A tool in `values.tools` misses a tool it requires. */
export class RequiredToolMissing extends ValidityError("RequiredToolMissing") {}
/** A path in `files` is not a literal repository-relative file path. */
export class InvalidPath extends ValidityError("InvalidPath") {}

export type SetupConfigError =
  | SetupConfigUnreadable
  | SchemaViolation
  | FormatTooNew
  | VersionTooNew
  | ToolRemoved
  | MaxOneExceeded
  | UnremovableReplaced
  | RequiredToolMissing
  | InvalidPath;

/** A setup config that passed Validity; only `validate` makes one. */
export type ValidSetupConfig = SetupConfig & Brand.Brand<"ValidSetupConfig">;

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

/** The `max: one` rule and the missing-unremovable rule (row 16) over a selection of known tools. */
const categoryProblem = (
  names: ReadonlyArray<string>,
): MaxOneExceeded | UnremovableReplaced | undefined => {
  const selected = names.map(name => toolsByName.get(name)!);
  const isMaxOne = (category: string) =>
    categoriesByName.get(category)?.max === "one";
  for (const [index, tool] of selected.entries()) {
    const other = selected
      .slice(index + 1)
      .find(next => next.category === tool.category);
    if (other !== undefined && isMaxOne(tool.category)) {
      return new MaxOneExceeded({
        what:
          `tools \`${tool.name}\` and \`${other.name}\` in values.tools of ${file} share the category \`${tool.category}\``,
        why: `the category \`${tool.category}\` holds one tool at most`,
        fix: `remove \`${tool.name}\` or \`${other.name}\` from values.tools`,
      });
    }
  }
  for (const tool of unremovableTools) {
    const other = selected.find(next => next.category === tool.category);
    if (
      !names.includes(tool.name) && other !== undefined
      && isMaxOne(tool.category)
    ) {
      return new UnremovableReplaced({
        what:
          `values.tools of ${file} holds \`${other.name}\` in place of the unremovable tool \`${tool.name}\``,
        why:
          `\`${tool.name}\` cannot be removed and the category \`${tool.category}\` holds one tool at most`,
        fix: `replace \`${other.name}\` with \`${tool.name}\` in values.tools`,
      });
    }
  }
  return undefined;
};

/** Applies the Validity rules beyond the schema to a decoded setup config; fails on the first violation. */
export const validate = (
  config: SetupConfig,
  running: Running,
): Effect.Effect<ValidSetupConfig, SetupConfigError> =>
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
          what: `tool \`${name}\` in values.tools of ${file}`,
          why:
            `bootstrap ${running.version} does not ship \`${name}\`: it is removed in this version`,
          fix: `remove \`${name}\` from values.tools`,
        });
      }
    }
    const category = categoryProblem(config.values.tools);
    if (category !== undefined) {
      return yield* category;
    }
    const [missing] = checkRequires(config.values.tools);
    if (missing !== undefined) {
      return yield* new RequiredToolMissing({
        what: `tool \`${missing.tool}\` in values.tools of ${file}`,
        why:
          `\`${missing.tool}\` requires \`${missing.missing}\`, which is not in values.tools`,
        fix:
          `add \`${missing.missing}\` to values.tools, or remove \`${missing.tool}\``,
      });
    }
    for (const path of config.files) {
      const why = pathProblem(path);
      if (why !== undefined) {
        return yield* new InvalidPath({
          what: `${JSON.stringify(path)} in files of ${file}`,
          why:
            `a path must be a literal file path relative to the repository root; ${why}`,
          fix: `correct or remove ${JSON.stringify(path)} in files`,
        });
      }
    }
    return config as ValidSetupConfig;
  });
