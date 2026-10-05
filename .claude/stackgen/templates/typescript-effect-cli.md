---
slug: typescript-effect-cli
axis: project
kind: language-bundle
components:
  - language/typescript@0.3.1
  - package-manager/pnpm@0.6.1
  - toolchain-gate/tsconfig@0.2.2
  - toolchain-gate/eslint@0.3.4
  - framework/effect@0.1.0
platforms:
  - cli
languages:
  - typescript
  - javascript
language_facts:
  typescript:
    lsp: typescript-language-server, run through mise; declared by this pack as `typescript-lsp` and provisioned via the generated local plugin
    mise_tool: node
    manifest: package.json
  javascript:
    lsp: typescript-language-server, run through mise; the same `typescript-lsp` server serves both tokens
    mise_tool: node
    manifest: package.json
optional_languages: []
frameworks:
  - effect
dependencies: []
capabilities: []
artifact: n/a
package_manager: pnpm
harness: n/a
materialized: "2026-10-05"
---
# cli — TypeScript · Effect CLI

A **shipped command-line tool**: TypeScript ·
[`@effect/cli`](https://effect.website/docs/guides/cli) on
`@effect/platform-node`, published to a package registry and run by people, not
by a deploy target. The project's registry `platforms:` is `[ cli ]` — a
terminal surface, so it has no screens, no `<platform>.md` flow files and no
canvas; what it *does* need is the design system's **Terminal UX** section,
which fixes output shape, color semantics, error format, and exit codes.

`cli` sits under the `frontend` role because a CLI runs on the user's machine,
like an app — not because it has a GUI. It is a platform but **not a screen
platform**, which is what exempts it from the design system's screen mandates
while still requiring Terminal UX. Internal dev scripts are not this: a repo's
own task runner is tooling, and gets no project entry.

## Stack

- **`@effect/cli`** — commands are declarative values, not imperative parsing. A
  command pairs a name with its typed flags and positional arguments and a
  handler returning an `Effect`; subcommands compose onto a root command.
  `--help`, `--version`, shell completions, and the interactive wizard come from
  the framework, so they never drift from the command definitions.
- **`@effect/platform` + `@effect/platform-node`** — filesystem, process, and
  terminal access as Effect services, with `NodeContext.layer` provided at the
  entrypoint and `NodeRuntime.runMain` as the process boundary. Nothing calls
  `node:fs` or `process.exit` directly; that is what keeps handlers testable.
- **The same DI as the backend templates** — integrations and side effects are
  `Layer`-provided services, so a handler under test runs against a stub layer
  with no process, network, or disk.
- **Tooling**: Vitest + `@effect/vitest` for unit tests and one end-to-end pass
  that executes the built binary and asserts on stdout/stderr/exit code; the
  repo axis supplies the package manager, formatter, and lint config.

> **Effect v4 renames these.** `@effect/cli` folds into `effect/unstable/cli`,
> with `Options` → `Flag` and `Args` → `Argument`. The shape above is unchanged;
> only the import paths and two module names move. Check the installed major
> before writing code.

## Contract surfaces

- **Commands are the flow contract.** A CLI flow's `index.md` describes the
  journey — trigger, actors, steps, acceptance — and its commands, flags, and
  exit codes are that contract's observable surface. There is no Screens section
  to carry them.
- **Exit codes and stderr are the error contract**, the way status codes and the
  error envelope are for a `service`. Pin them in the design system's Terminal
  UX section and hold the code to them; the execute code reviewer checks
  conformance there, since no UX reviewer renders a terminal.
- **A machine-readable mode** (`--json` or `--porcelain`) is a contract, not a
  convenience — anything scripting the tool depends on its stability.

## Distribution

Ships through a **package registry**, not a deploy target: pair this with the
`npm-package` bundle on the `deploy` axis (`stacks/bundles/npm-package.md`),
which carries the publish pipeline, provenance, and versioning rules.

# TypeScript — conventions

The Node/TypeScript baseline. New code is TypeScript with `strict` on;
JavaScript files get the same standards minus the type-level rules and are
migration candidates, never an excuse to relax them.

**Errors are values at the boundary and exceptions in the middle.** One mapping
home turns internal failures into the product's coded responses — never
scattered `try`/`catch` that each invent their own shape.

**`async`/`await` throughout; never block the event loop.** CPU-bound work moves
off the main thread rather than being awaited around.

**Tests are Vitest**, colocated, with the shared config and v8 coverage.

**The `@/` path alias, no deep relative chains, and a clean→check→build
pipeline.** Barrels are for public surfaces only.

**Config is read once at the composition root** — names-not-values, catalogued
in `docs/blueprint/environment.md`, never `process.env` reads scattered through
the code.

**Telemetry is OTLP.** The product emits it and never imports a vendor SDK.

Full judgment: the `typescript` skill's references.

# pnpm — conventions

pnpm is the only package manager. A repo with two lockfiles has two dependency
graphs and resolves differently depending on who ran what.

**The lockfile is committed and authoritative.** CI installs frozen and fails on
drift rather than resolving something new — an install that can resolve
differently in CI than locally is not a gate.

**A publish cooldown guards the supply chain**, so neither a routine install nor
an automated update adopts a release published minutes ago.

**In a workspace, internal dependencies are linked, not versioned**, and shared
versions live in a catalog so one bump moves every package.

**Two settings ship as `.npmrc` at the repo root**, which is the one path the
manager reads them from: `ignore-scripts=true`, so an install never executes a
dependency's install-time code, and `fund=false`, so it never prints a banner
over what it did. A dependency that genuinely has to build is allowed by name
in `pnpm-workspace.yaml` (`allowBuilds`, or `onlyBuiltDependencies` before
pnpm 10.26) — the exception is a reviewable line, not a switch.
Beside it, the pack's template
`templates/.config/mise/conf.d/pnpm/mise.dev.toml`, rendered into the repo by
`stackgen:tool-config`, aliases `npx` to `pnpm dlx` in dev, so a one-off
package runs through this manager's store, resolver and registry settings
rather than another tool's.

**An agent's `npm`/`npx` command is rewritten before it runs.** This pack
ships `hooks/npm-normalize.sh`, which lands at `.claude/hooks/npm-normalize.sh`
and — once its `hooks.yaml` entry is accepted into `.claude/settings.json` —
resolves the repo's manager from its lockfile and rewrites the command to it.
Declining the settings entry leaves the script landed and inert, which is
safe: the hook only rewrites a command that was going to run the wrong manager
anyway. It allows exactly two managers, pnpm and bun (`npx` → `pnpm dlx` or
`bunx`, `npm ci` → `<pm> install --frozen-lockfile`, any other `npm` → `<pm>`,
flags after `npx` kept verbatim), and resolves which one by walking up from the
working directory: a lockfile first — `bun.lock`/`bun.lockb` or
`pnpm-lock.yaml`, the ground truth, since bun reuses npm's `workspaces` field
and nothing else tells them apart — then `package_manager: bun` in
`.config/vwf.yaml`, for a project scaffolded but not yet installed, then pnpm,
because the hook fires in every repo, including ones that never heard of vwf.
Its `sed` stays BSD-compatible: no `\s`, no `\b`.

## The task library this pack owns

This pack ships a `config/.config/mise/tasks/` tree — the `code/format/pnpm`
subtask and the five `setup/deps/<verb>/pnpm` subtasks — landing at the
repo's own `.config/mise/tasks/` behind the materializer's config consent
line. Every file is named for the pack, so no other component writes the same
path: the repo's `code:format:all` and `setup:deps:<verb>:all`, which
tool-config renders, call each subtask by name.

**`code:format:pnpm` sorts every `package.json`, and only that.** dprint runs
beside it as the universal `code:format:dprint` subtask, so this file carries
no formatter step of its own. The sorter, `npm:sort-package-json`, is a mise
pin in the dev environment only, in the same `mise.dev.toml` template —
resolved by its mise path, and the step is skipped where it is not installed,
as in CI.

**The subtask takes an optional file list, and the empty case is the whole
tree.** That is the whole pre-commit story for this pack: it ships **no
fragment**, because the gate config's `format` and `lint` hooks call
`code:format:all` and `code:lint:all` with the staged files. The sorter
narrows to the `package.json` files it is given. Linting is the house linter's,
through the universal `code:lint:house`, which runs the whole tree either way —
its rules are cross-file — and every exclusion it needs lives in
`.config/linter.yaml`, which `stackgen:tool-config` lands.

**The `setup/deps/*` verbs are `install`, `outdated`, `audit`, `upgrade` and
`cleanup` — all five slots.** `install` is `pnpm install --recursive`, because a
workspace install that stops at the root leaves the repo half resolved.
`cleanup` deletes `dist`, `node_modules` and `*.tsbuildinfo`, then prunes the
store — a store left behind makes the next install look clean when it is
replaying — and deliberately leaves the lockfile alone: the lockfile is an input
a human reviews, and moving it forward is `upgrade`'s job. The optional verbs
are **probed by name**, so a missing file is itself the answer: a manager that
ships no `upgrade` has no such verb, not a choice still pending.

`install --frozen` is the contract's name for "the lockfile is the input, not
the output" — what a fresh worktree and CI want, turning a stale lockfile into
a failure rather than a silent rewrite. `audit` is advisory and never a gate:
`pnpm audit` reads a registry feed that moves without any lockfile change, and
the blocking supply-chain check is `code:sec`, which runs pinned tools.
`outdated` swallows its exit status, since `pnpm outdated` fails whenever it
finds anything — a healthy repo's normal state. `upgrade` updates pnpm itself
first, because a resolver a major version behind writes a lockfile the current
one then rewrites, and passes `--latest` on purpose: the ranges say what still
works, and this task is where a person decides something newer should — the
diff is the review surface. `cleanup` prunes the store's `.pnpm` link farm
rather than deleting it, which would re-download every unchanged package. The
verbs print no header of their own; `setup:deps:all` frames each.

`code:format:pnpm`'s sorter pair is the inverse of dprint's — sorting is its
default, `--check` its read-only mode — and the sorter, like the house linter,
runs by its mise path so a package in `node_modules/.bin` cannot shadow the
pin.

Full judgment: the `pnpm` skill's references.

# tsconfig — conventions

**`strict` is on, everywhere, and is not negotiated per project.** The
type-level rules in the TypeScript baseline assume it; without it they are
suggestions.

**One shared base config, extended per project.** A per-project config that
restates the base has already drifted from it.

**The `@/` path alias** replaces deep relative chains, and the build resolves it
the same way the editor does.

**A separate emit variant for builds**, so type checking and emitting are
distinct operations — `tsc --noEmit` is the checker, and nothing about a check
should depend on output settings.

## What this pack writes

No file. This pack's conventions and skill guide the agent when it writes the
`tsconfig*.json` files, which are per-project and belong to the project —
written where the project is, not laid down from here.

Full judgment: the `tsconfig` skill.

# ESLint — conventions

The **correctness** gate for TypeScript and JavaScript. Topic 10 of the language
bundle, deliberately not a repo gate: a linter meaningful for exactly one
toolchain belongs to that toolchain's bundle, or a polyglot repo acquires one
per language.

**Flat config only.**

**Zero formatting rules.** The formatter owns layout — a rule a formatter can
satisfy must never be able to fail a lint run. The dprint config
`stackgen:tool-config` lands is the other half of that split.

**Overrides are scoped by `files` glob**, never disabled globally. A rule turned
off everywhere because one file could not satisfy it is a rule the repo no
longer has.

**One lint command, wired through the task library**, so local and CI run the
identical gate.

## What this pack writes

No file: its skill and these conventions. ESLint runs only inside the house
linter, `@askviraj/linter`, which `stackgen:tool-config` pins in every repo
and runs through the universal `code:lint:house` subtask — so this pack ships
no `code/lint/eslint` subtask and **no pre-commit fragment**: the gate
config's `lint` hook calls `mise run code:lint:all --fix`, which runs
`code:lint:house` with every other lint subtask. The house linter reads the
whole tree whatever list it is given — its rules are cross-file.

`.config/linter.yaml`, the linter's own config, is **not** this pack's either:
`stackgen:tool-config` lands it with the other gate configs, because the house
linter reads it in every repo, not on this stack alone. It lands **empty of
overrides** — the linter is zero-config without it, so the file exists to give
a misfiring default one obvious place to be answered — with an `ignores:` list
of the generated trees the stack packs produce. The `eslint` skill still guides
every edit to it.

**A disable comment goes on its own line above the offending one**, line style,
so the decision is visible, with its reason written by hand.

Full judgment: the `eslint` skill.

# Effect-TS — conventions

Effect layers **on top of** the TypeScript baseline rather than replacing it:
plain TypeScript gets the baseline alone, an Effect project gets both.

**TypeScript with `strict` is a hard requirement.** Effect never applies to
JavaScript.

**Failures are in the type.** This is the reason to adopt it — the error channel
makes the failure modes of a call visible to the compiler, which is exactly what
plain TypeScript cannot do (see the language pack's error semantics reference,
which describes working without this).

**Dependencies are in the type too**, provided at the composition root, which is
the same information-hiding rule the baseline states with a stronger enforcement
mechanism.

**The test runner is unchanged — only the assertions differ.** Where code under
test returns an `Effect`, read this skill's testing reference alongside the
language pack's.

Full judgment: the `effect` skill's references.
