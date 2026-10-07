import {
  Data,
  Effect,
  Runtime,
} from "effect";
import {
  CliError,
  CliOutput,
  Command,
} from "effect/cli";
import pkg from "../package.json" with { type: "json" };

/** The bare command: help is already on stdout, so the error only carries the exit code. */
export class NoCommand extends Data.TaggedError("NoCommand") {
  readonly [Runtime.errorExitCode]: number = 2;
  readonly [Runtime.errorReported] = false;
}

export const command = Command.make(
  "bootstrap",
  {},
  () =>
    Effect.fail(
      new CliError.ShowHelp({ commandPath: ["bootstrap"], errors: [] }),
    ),
);

/** `--version` prints the bare version, not the framework's `<name> v<version>`. */
const versionOnly = Effect.map(
  CliOutput.Formatter,
  (formatter): CliOutput.Formatter => ({
    ...formatter,
    formatVersion: (_name, version) => version,
  }),
);

/** Runs the root command; a bare invocation prints help and fails with exit code 2. */
export const run = (args: ReadonlyArray<string>) =>
  Command.runWith(command, { version: pkg.version })(args).pipe(
    Effect.catchTag(
      "ShowHelp",
      error => Effect.fail(error.errors.length === 0 ? new NoCommand() : error),
    ),
    Effect.provideServiceEffect(CliOutput.Formatter, versionOnly),
  );
