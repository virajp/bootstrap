import { describe, expect, it } from "@effect/vitest";
import { NodeServices } from "@effect/platform-node";
import { Cause, Console, Effect, Exit, Runtime } from "effect";
import { run } from "@/cli";
import pkg from "../package.json" with { type: "json" };

const runCaptured = (args: ReadonlyArray<string>) => {
  const stdout: Array<string> = [];
  const capture: Console.Console = {
    ...globalThis.console,
    log: (...parts: ReadonlyArray<unknown>) => {
      stdout.push(parts.join(" "));
    },
  };
  return run(args).pipe(
    Effect.provideService(Console.Console, capture),
    Effect.provide(NodeServices.layer),
    Effect.exit,
    Effect.map((exit) => ({ exit, stdout: stdout.join("\n") })),
  );
};

describe("bootstrap", () => {
  it.effect("--version prints the package version and succeeds", () =>
    Effect.gen(function*() {
      const { exit, stdout } = yield* runCaptured(["--version"]);
      expect(Exit.isSuccess(exit)).toBe(true);
      expect(stdout).toBe(pkg.version);
    }));

  it.effect("no arguments prints help and exits 2", () =>
    Effect.gen(function*() {
      const { exit, stdout } = yield* runCaptured([]);
      expect(stdout).toContain("USAGE");
      expect(stdout).toContain("bootstrap");
      expect(Exit.isFailure(exit)).toBe(true);
      if (Exit.isFailure(exit)) {
        expect(Runtime.getErrorExitCode(Cause.squash(exit.cause))).toBe(2);
      }
    }));
});
