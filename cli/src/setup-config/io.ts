/** Reads and writes `<repository root>/.config/bootstrap.yaml`. */
import {
  type Repaired,
  repair,
} from "@/setup-config/repair";
import {
  SetupConfig,
  setupConfigPath,
} from "@/setup-config/schema";
import {
  type Running,
  type SetupConfigError,
  type ValidSetupConfig,
  checkFormat,
  SchemaViolation,
  SetupConfigUnreadable,
  validate,
} from "@/setup-config/validity";
import {
  Effect,
  FileSystem,
  Option,
  Path,
  Schema,
} from "effect";
import * as YAML from "yaml";

const decode = Schema.decodeUnknownEffect(SetupConfig, {
  onExcessProperty: "error",
});

/** Parses setup config text, decodes it against the schema and applies the Validity rules. */
export const parse = (
  text: string,
  running: Running,
): Effect.Effect<ValidSetupConfig, SetupConfigError> =>
  Effect.gen(function*() {
    const raw: unknown = yield* Effect.try({
      try: () => YAML.parse(text),
      catch: cause =>
        new SetupConfigUnreadable({
          what: `${setupConfigPath} is not valid YAML`,
          why: String(cause),
          fix: `correct the YAML in ${setupConfigPath}`,
        }),
    });
    yield* checkFormat((raw as { format?: unknown; } | null)?.format, running);
    const config = yield* decode(raw).pipe(
      Effect.mapError(error =>
        new SchemaViolation({
          what: `${setupConfigPath} does not match the setup config schema`,
          why: error.message,
          fix: `correct the entry in ${setupConfigPath}`,
        })
      ),
    );
    return yield* validate(config, running);
  });

/**
 * Reads the setup config under the repository root, applies Validity and then Repair on read;
 * `None` when the file is absent.
 */
export const read = (
  root: string,
  running: Running,
): Effect.Effect<
  Option.Option<Repaired>,
  SetupConfigError,
  FileSystem.FileSystem | Path.Path
> =>
  Effect.gen(function*() {
    const fs = yield* FileSystem.FileSystem;
    const file = (yield* Path.Path).join(root, setupConfigPath);
    const text = yield* Effect
      .gen(function*() {
        return (yield* fs.exists(file))
          ? Option.some(yield* fs.readFileString(file))
          : Option.none();
      })
      .pipe(
        Effect.mapError(error =>
          new SetupConfigUnreadable({
            what: `${setupConfigPath} cannot be read`,
            why: error.message,
            fix: `check the permissions of ${setupConfigPath}`,
          })
        ),
      );
    return Option.isSome(text)
      ? Option.some(repair(yield* parse(text.value, running)))
      : Option.none();
  });

/**
 * The setup config file text per the Write format: keys in schema order, `values.tools`,
 * `values.commit_scopes` and `files` sorted, the running cli's `format` and `version`, no
 * comments, and the setup config itself never in `files`. Writing it to disk is the caller's.
 */
export const write = (config: SetupConfig, running: Running): string =>
  YAML.stringify({
    format: running.format,
    version: running.version,
    values: {
      repo: config.values.repo,
      commit_scopes: config.values.commit_scopes.toSorted(),
      merge_model: {
        develop: config.values.merge_model.develop,
        main: config.values.merge_model.main,
      },
      tools: config.values.tools.toSorted(),
    },
    files: config.files.filter(path => path !== setupConfigPath).toSorted(),
  });
