/** The version guard of docs/blueprint/entities/setup-config/index.md: both refusals exit 1. */
import { setupConfigPath as file } from "@/setup-config/schema";
import {
  compareVersions,
  type Problem,
  type Running,
  type ValidSetupConfig,
} from "@/setup-config/validity";
import {
  Data,
  Effect,
  Option,
  Runtime,
} from "effect";

/** The next command of both refusals. */
const init = "bootstrap init";

/** The repository has no setup config. */
export class NotSetUp extends Data.TaggedError("NotSetUp")<Problem> {
  readonly [Runtime.errorExitCode]: number = 1;
}

/** The setup config was written by an older bootstrap. */
export class VersionTooOld extends Data.TaggedError("VersionTooOld")<Problem> {
  readonly [Runtime.errorExitCode]: number = 1;
}

/** Fails when the setup config is absent; `add`, `remove` and `show` call it. */
export const requireSetUp = <A>(found: Option.Option<A>): Effect.Effect<A, NotSetUp> =>
  Option.match(found, {
    onNone: () =>
      Effect.fail(
        new NotSetUp({
          what: `not set up — run \`${init}\``,
          why: `${file} is absent`,
          fix: init,
        }),
      ),
    onSome: Effect.succeed,
  });

/**
 * Fails when the recorded `version` is older than the running cli; `add`, `remove` and `tui`
 * call it. It takes only a config that passed Validity, so Validity (exit 3) runs first.
 */
export const versionGuard = (
  config: ValidSetupConfig,
  running: Running,
): Effect.Effect<ValidSetupConfig, VersionTooOld> =>
  compareVersions(config.version, running.version) < 0
    ? Effect.fail(
      new VersionTooOld({
        what: `${file} was written by bootstrap ${config.version}`,
        why: `this is bootstrap ${running.version}, which is newer`,
        fix: init,
      }),
    )
    : Effect.succeed(config);
