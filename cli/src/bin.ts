import { NodeRuntime, NodeServices } from "@effect/platform-node";
import { Effect, Stdio } from "effect";
import { run } from "@/cli";

// The process boundary: argv comes from Stdio, and runMain's teardown turns a
// failure's `Runtime.errorExitCode` into the process exit code.
Effect.gen(function*() {
  const stdio = yield* Stdio.Stdio;
  return yield* run(yield* stdio.args);
}).pipe(Effect.provide(NodeServices.layer), NodeRuntime.runMain);
