---
type: vwf-plan
title: Tool catalog, tool categories and setup config
requires: []
backlog: []
backlog_pieces: []
covers:
  - docs/blueprint/entities/tool/index.md
  - docs/blueprint/entities/tool-category/index.md
  - docs/blueprint/entities/setup-config/index.md
---

# Plan — Tool catalog, tool categories and setup config (2026-10-08)

## Status

**RUNNING**

RUNNING since 2026-10-09T10:00 in /Users/virajpatel/Projects/github.com/virajp/bootstrap/.worktrees/all-2026-10-09-0954

## Consent

| Action                                            | Granted                       |
| ------------------------------------------------- | ----------------------------- |
| Merge to the integration branch and push on green | yes                           |
| After landing: archive the plan folder            | run                           |
| After landing: `mise run code:graph`              | run                           |
| Release cli publicly                              | none (stays `0.0.1`; no bump) |
| LSP typescript                                    | installed                     |
| End an `all` run after landing                    | no                            |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt, and `run` is the only mode. `/vwf:execute` asks nothing at run
time. The release intent is `none`, so no release step exists.

## Goal

After this plan lands, the `cli` code implements the three entity contracts, and
the three docs under `covers:` read `implementation: complete`. The code then
has the 13-category catalog, the 15-tool catalog with every field, the blueprint
target paths and dispatchers, the renders that change with the selection, and
every setup-config rule that does not write to disk: validity, repair on read,
write format, version guard and the refresh of `files`.

The blueprint rescope (`dd2ae45`) replaced the `core` / `non-core` model that
the earlier plan `2026-10-06-2332-tool-setup-config` built. This plan makes the
code agree with the rescoped contract. No standing decision is reversed.

## Slice

Entity cycle — one chain element of three docs that link each other in their
Relationships tables:

- [Tool](../../blueprint/entities/tool/index.md) — `implementation: partial`
- [Tool category](../../blueprint/entities/tool-category/index.md) —
  `implementation: none`
- [Setup config](../../blueprint/entities/setup-config/index.md) —
  `implementation: partial`

Plan 1 of 2 — requires nothing; required by the plan for
[Set up a repository](../../blueprint/flows/cli/110-setup-repository/index.md)
(flow 110), which is planned next.

## Facts the survey established

- **Revision (2026-10-09):** the blueprint pass `f7a96b6` changed the three
  entity contracts after the first approval. No file under `cli/` changed since
  `24674b1`, so the survey facts below stay valid. Rows 13–21 and the new
  wording of row 9 follow the changed contracts.
- **Backlog:** unreadable — no backlog project `bootstrap` exists under
  `virajp`. No backlog id applies.
- **Blind spots:** none. The repo is a monorepo; `site/` does not exist and is
  out of scope.
- **Stack gate:** `/vwf:doctor cli` found no blocking finding. The TypeScript
  LSP is installed. Known drift (not acted on): `config_format` 22 vs 23, the
  repo is not shaped by `/vwf:init` (declined on purpose), no branch protection.
- **Tool catalog** — `cli/src/tool/catalog.ts:3-144`. It has 15 tools with
  `name`, `kind` (`core` | `non-core`), `purpose` and `files` (plain strings).
  Absent: category, base, removable, replaceable, default (on | off | origin),
  `origin_hosts`, requires, per-file `create_only`. The order differs from the
  blueprint (code: gitleaks, grype, virajp-linter, claude, graphify, mempalace,
  fnox; blueprint: virajp-linter, gitleaks, grype, fnox, claude, graphify,
  mempalace). Every `purpose` text differs from the blueprint table.
  `toolsByName` and `toolByPath` exist (`:139-144`); `toolByPath` overwrites a
  duplicate path and does not fail.
- **Tool category:** absent. No type, no catalog, no invariant.
- **Old-model callers:** only `cli/src/setup-config/validity.ts:183`
  (`tool.kind === "core"`) and `cli/test/tool/catalog.test.ts` use `kind`,
  `coreTools` or `nonCoreToolNames`.
- **Renderer** — `cli/src/tool/render.ts:30-114`. LiquidJS with `<%` delimiters,
  strict variables, the `<%=` output swap (G6). `renderFiles` returns
  `RenderedFile { content, executable }`; `executable` is a path-prefix rule on
  `.config/mise/tasks/` (tool invariant 5 holds). No `create_only`; no input for
  the selection or the `origin` host.
- **Templates** — `cli/templates/<tool>/<target path>`, embedded at build with
  the `?raw` loader in `cli/tsdown.config.ts` (G7). Differences from the
  blueprint Target paths:
  - `mise`: absent `tasks/code/{format,lint,sec,graph}/_default`,
    `tasks/setup/secrets/_default`, `tasks/setup/ai/_default`; the five
    `tasks/setup/deps/<verb>` are plain files and must become
    `tasks/setup/deps/<verb>/_default`.
  - `dprint`: has the extra `.config/dprint.json`; `tasks/code/format` must
    become `tasks/code/format/dprint`.
  - `virajp-linter`: `tasks/code/lint` → `tasks/code/lint/virajp-linter`.
  - `gitleaks`: `tasks/code/sec` → `tasks/code/sec/gitleaks`; absent
    `conf.d/gitleaks/mise.dev.toml`.
  - `grype`: absent `conf.d/grype/mise.dev.toml`, `tasks/code/sec/grype`.
  - `fnox`: `tasks/setup/secrets` → `tasks/setup/secrets/fnox`; absent
    `conf.d/fnox/mise.dev.toml`.
  - `claude`: has the extra `conf.d/ai/mise.dev.toml`; `tasks/setup/ai` →
    `tasks/setup/ai/claude`.
  - `graphify`: `tasks/code/graph` → `tasks/code/graph/graphify`; absent
    `conf.d/graphify/mise.dev.toml`.
  - `mempalace`: absent `conf.d/mempalace/mise.dev.toml`; `mempalace.yaml` is
    not `create_only`.
  - `cli/templates/mise/.config/mise/conf.d/_base/mise.dev.toml:15-49` puts
    grype, gitleaks and fnox in the base file and does not change with the
    selection. `cli/templates/pre-commit/.config/git-conventional-commits.yaml`
    has no forge-link logic.
- **Setup config** — `cli/src/setup-config/{schema,validity,io}.ts`.
  - Present: path `.config/bootstrap.yaml`; `format` and `version`; the
    newer-format and newer-version refusals with "upgrade bootstrap"
    (`validity.ts:61-115`); semver compare (`compareVersions`, `:88`); the
    unknown-tool fix "remove `<name>` from values.tools" (`:173-182`); the path
    rules on `files` (`:136-157,201-218`); `read` (absent → `None`), `parse`,
    `write` (returns text, sorts `files`, drops the config's own path) in
    `io.ts:28-101`; typed errors with what / why / fix (`validity.ts:20-58`).
  - `values.repo` is stricter than `schema.yaml` (G13 hardening). Keep it:
    stricter is not a contradiction.
  - Partial: `values.tools` has no `minItems: 1`; `commit_scopes` pattern
    `^[a-z][a-z0-9-]*$` is looser than the blueprint
    `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`; `write` does not sort `values.tools` and
    `values.commit_scopes`; key order follows object spread.
  - Absent: the max-one and requires validity rules; repair on read and the
    `warnings` channel; the version guard; the pure computation of setup-config
    invariants 2, 4, 5 and 7 (refresh `files`, keep and report `orphaned`, drop
    paths absent on disk, rewrite when the text differs).
- **Earlier gaps (plan `2026-10-06-2332-tool-setup-config`)** — the blueprint
  now answers G8, G9, G10, G15 and G18. G1 (no coverage provider) is open and
  this plan closes it. G6, G7, G11, G12, G16 stay as they are. G2 and G3 belong
  to plan 2. G4 and G17 are upstream stack prose.
- **Harness:** `mise run p:cli:test` (`pnpm exec vitest run "$@"` in `cli/`),
  `p:cli:check` (`tsc --noEmit`), `p:cli:build` (`tsdown`).
  `cli/vitest.config.ts` names the v8 coverage provider, but
  `@vitest/coverage-v8` is not in `cli/package.json`. `cli/node_modules` is not
  installed in the main checkout; run `pnpm install --frozen-lockfile` first.
  `harness:` in `.config/vwf.yaml` is all `false`; entities have no Acceptance
  block, so no e2e harness is necessary for this plan.
- **Stack conventions:** `project/typescript-effect-cli` (effect 4.x:
  `effect/cli`, `NodeServices.layer`; Vitest + `@effect/vitest`, colocated
  config, v8 coverage; `@/` alias; errors are values), `repo/pnpm-workspace`,
  `deploy/npm-package`. No `node:fs` and no `process.exit` in handlers.
- **Commit types** allowed by `.config/git-conventional-commits.yaml`: `ops`,
  `docs`, `merge`, `feat`, `fix`, `refactor`. No scopes.
- **Docs:** `cli/` has no README and no CHANGELOG. The docs unit runs
  `vwf:docs-sync`.

## Assumed decisions — confirm or override at review

| #  | Decision                                                                                                                                   | Ruling                                                                                                                                                                                                                                                                                                                          | Rejected                                      | Unit       |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------- | ---------- |
| 1  | Drift: code has `kind: core \| non-core`                                                                                                   | Change the code to `category`, `base`, `removable`, `replaceable`, `default`, `requires` and `origin_hosts`, as the Tool catalog says                                                                                                                                                                                           | Amend the blueprint to keep `core`            | U2         |
| 2  | Drift: code refuses unremovable tools in `values.tools` (`CoreToolListed`)                                                                 | `values.tools` lists every selected tool, base and unremovable tools included (setup-config invariant 3)                                                                                                                                                                                                                        | Keep core tools implicit                      | U4         |
| 3  | Drift: config has `kept` / `deleted` keys and `PathKeptAndDeleted`                                                                         | Remove them; the schema is closed over `format`, `version`, `values`, `files`                                                                                                                                                                                                                                                   | Keep them as extensions                       | U4         |
| 4  | Drift: plain task files `code/format`, `code/lint`, `code/sec`, `code/graph`, `setup/secrets`, `setup/ai`, `setup/deps/<verb>`             | `_default` dispatchers owned by `mise`, plus per-tool subtasks, as Tool Target paths and Task dispatchers say                                                                                                                                                                                                                   | Keep single-tool task files                   | U3         |
| 5  | Version guard in plan 1                                                                                                                    | Build it now (user, 2026-10-08)                                                                                                                                                                                                                                                                                                 | Defer to slice 2                              | U4         |
| 6  | Catalog integrity invariants                                                                                                               | Tests over the static catalog data enforce tool invariants 1–12 and category invariants 5–6                                                                                                                                                                                                                                     | Checks at runtime                             | U2         |
| 7  | Exit codes of entity errors                                                                                                                | Typed errors carry their exit code, as `NoCommand` does (`cli/src/cli.ts:14-17`): Validity → 3, version guard → 1                                                                                                                                                                                                               | Map the codes only in the CLI                 | U4         |
| 8  | Renderer inputs                                                                                                                            | The selection and the `origin` host are inputs to the render; plan 2 reads git and supplies them                                                                                                                                                                                                                                | Read git in the renderer                      | U5         |
| 9  | Setup-config invariant 1 (written last; a safety target; a dirty setup file is refused only when its render differs from the working copy) | Lands with plan 2's atomic write; plan 1 has no disk writer                                                                                                                                                                                                                                                                     | A disk writer in plan 1                       | — (plan 2) |
| 10 | Template content                                                                                                                           | Templates conform to the catalog; this repo's own `.config/` files are sources, not targets, and stay untouched                                                                                                                                                                                                                 | Edit this repo's `.config/`                   | U3, U5     |
| 11 | G1: no coverage provider                                                                                                                   | Add `@vitest/coverage-v8` so the coverage gate measures                                                                                                                                                                                                                                                                         | Continue with `COVERAGE: n/a`                 | U1         |
| 12 | Order of catalog fields that break callers                                                                                                 | U2 makes only the edits that follow the new types in `render.ts` and `validity.ts:183`, so every wave compiles; U4 and U5 then do the behaviour                                                                                                                                                                                 | Leave the build broken between waves          | U2         |
| 13 | Plan constraint: catalog-driven code (user, 2026-10-08)                                                                                    | No tool name or category name appears as a literal in `cli/src/{tool,tool-category,setup-config}` outside `cli/src/tool/catalog.ts` and `cli/src/tool-category/catalog.ts`. Every rule reads catalog fields. A test scans these folders and fails on a literal                                                                  | Enforce it in the review only                 | U2, U4, U5 |
| 14 | Check order on read                                                                                                                        | `read` applies Validity (exit 3) first. The version guard (exit 1) accepts only a valid config. An older `version` plus a Validity failure exits 3                                                                                                                                                                              | Guard first                                   | U4         |
| 15 | The absent file and the older version                                                                                                      | Two functions. `requireSetUp`: an absent file exits 1 "not set up — run `bootstrap init`"; `add`, `remove` and `show` call it. `versionGuard`: an older `version` exits 1, next command `bootstrap init`; `add`, `remove` and `tui` call it. The tui first run calls neither                                                    | One guard with a command parameter            | U4         |
| 16 | Validity: a `max: one` category that misses its unremovable tool and holds another tool                                                    | Fails Validity, exit 3. The check runs before Repair on read                                                                                                                                                                                                                                                                    | Repair it                                     | U4         |
| 17 | Refresh of `files` (setup-config invariants 2, 4, 5, 6)                                                                                    | Full mode (init, add, tui apply): the rendered set. Remove mode: the recorded paths minus the paths of the removed tools, plus the paths of a repaired tool. In both modes a path that is not rendered stays as `orphaned` when it is on disk and drops when it is absent; a rendered path stays also when it is absent on disk | One mode for all commands                     | U4         |
| 18 | Rewrite rule (setup-config invariant 7)                                                                                                    | A pure function. Init and the tui apply rewrite when the text differs from its render. Add and remove rewrite only when the run has a tool change or a repair                                                                                                                                                                   | Leave it to each flow                         | U4         |
| 19 | Tool `default` (first-run selection)                                                                                                       | `defaultSelection(originHost)` selects every `on` tool, plus each `origin` tool whose `origin_hosts` holds the host, and never an `off` tool. Init `-y` and the tui first run use it                                                                                                                                            | Build it in plan 2                            | U2         |
| 20 | Tool `requires`                                                                                                                            | `checkRequires(selection)` returns each pair (tool, missing required tool). Validity uses it over `values.tools` (exit 3). The flows use it over the selection after the request and name the missing tool (init, add) or the dependent tool (remove, replace) with exit 2. `unremovableTools` is part of the catalog module    | A separate check for each command             | U2         |
| 21 | Forge links without a category literal                                                                                                     | The renderer finds the selected tools that have `origin_hosts`, in catalog order, and takes the tool whose hosts hold the `origin` host, else the first. The template holds the link text of each tool                                                                                                                          | Find forge tools by the category name `forge` | U5         |

## New dependencies

- `@vitest/coverage-v8` (devDependency of `cli`) — the v8 coverage provider that
  `cli/vitest.config.ts` already names; preferred over no measured coverage
  (G1). No installed package does this. Added by U1.

## Units

| Id | Wave | Unit file                                    | Kind   | Owns                                                                                                                                                                                                                                      | Depends on         | Status  | Commit |
| -- | ---- | -------------------------------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------ | ------- | ------ |
| U1 | 1    | [01-coverage.md](01-coverage.md)             | code   | `cli/package.json`, `pnpm-lock.yaml`, `cli/vitest.config.ts`                                                                                                                                                                              | —                  | green | 1014fb1 |
| U2 | 2    | [02-catalogs.md](02-catalogs.md)             | code   | `cli/src/tool/catalog.ts`, `cli/src/tool-category/catalog.ts`, `cli/test/tool/catalog.test.ts`, `cli/test/tool-category/catalog.test.ts`, `cli/test/catalog-driven.test.ts`, `cli/src/tool/render.ts`, `cli/src/setup-config/validity.ts` | U1                 | green | 84016e8 |
| U3 | 3    | [03-templates.md](03-templates.md)           | code   | `cli/templates/**`, `cli/test/tool/core-templates.test.ts`, `cli/test/tool/new-templates.test.ts`, `cli/test/fixtures/**`                                                                                                                 | U2                 | green | 5ed611b |
| U4 | 3    | [04-setup-config.md](04-setup-config.md)     | code   | `cli/src/setup-config/**`, `cli/test/setup-config/**`                                                                                                                                                                                     | U2                 | green | 82fcd61 |
| U5 | 4    | [05-renderer.md](05-renderer.md)             | code   | `cli/src/tool/render.ts`, `cli/test/tool/render.test.ts`, `cli/templates/mise/.config/mise/conf.d/_base/mise.dev.toml`, `cli/templates/pre-commit/.config/git-conventional-commits.yaml`                                                  | U3, U4             | green | 4331d5e |
| U6 | 5    | [06-review.md](06-review.md)                 | review | —                                                                                                                                                                                                                                         | U1, U2, U3, U4, U5 | running |  |
| U7 | 6    | [07-docs.md](07-docs.md)                     | edit   | the repo's docs (README, CLAUDE.md, `docs/**` outside `docs/blueprint/` and `docs/plans/`)                                                                                                                                                | all                | pending |        |
| U8 | 7    | [08-gates-and-bump.md](08-gates-and-bump.md) | edit   | — (no version bump; no generated file)                                                                                                                                                                                                    | U7                 | pending |        |

## Shared-file rule

| File                                                                                                                           | Why it collides                                         | Owner                                      |
| ------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------- | ------------------------------------------ |
| `cli/package.json`                                                                                                             | the version and the dependency list                     | U1 (dependency); no unit bumps the version |
| `pnpm-lock.yaml`                                                                                                               | generated by `pnpm install`                             | U1 only                                    |
| `cli/src/tool/render.ts`                                                                                                       | U2 follows the types; U5 adds behaviour                 | U2 in wave 2, U5 in wave 4                 |
| `cli/src/setup-config/validity.ts`                                                                                             | U2 removes the `kind` check; U4 adds the rules          | U2 in wave 2, U4 in wave 3                 |
| `cli/templates/mise/.config/mise/conf.d/_base/mise.dev.toml`, `cli/templates/pre-commit/.config/git-conventional-commits.yaml` | U3 moves entries out; U5 makes them selection-dependent | U3 in wave 3, U5 in wave 4                 |
| every human-facing doc                                                                                                         | n units editing one doc                                 | U7 only                                    |

## Waves

- **Wave 1:** U1 — the coverage harness, before any unit whose gate reads
  coverage.
- **Wave 2:** U2 — the catalog data model that every later unit reads.
- **Wave 3:** U3 and U4 — disjoint paths (templates vs setup config); both stand
  on U2 only.
- **Wave 4:** U5 — the renderer stands on the template layout (U3) and the
  setup-config `Values` type (U4).
- **Wave 5:** U6 — the one review row, over U1–U5.
- **Wave 6:** U7 — docs.
- **Wave 7:** U8 — the final gate.

Waves are ordering only. `execute` runs the units serially.

## Wave gate

```text
pnpm install --frozen-lockfile
mise x -- mise run p:cli:check
mise x -- mise run p:cli:test
mise x -- mise run p:cli:build
```

Plus the wave review, plus every report read for `UNRESOLVED:`.
`mise run p:cli:build` replaces `pnpm --filter cli build`, which matches no
package (G14).

## After landing

| Step                    | Mode | Notes                                                                |
| ----------------------- | ---- | -------------------------------------------------------------------- |
| archive the plan folder | run  | `plan-management archive` moves the folder to `docs/plans/archived/` |
| `mise run code:graph`   | run  | refreshes `graphify-out/` for the merged code                        |

## Gates the orchestrator keeps

- `node cli/dist/bin.mjs --version` prints `0.0.1` after `mise run p:cli:build`.
- A render of every catalog tool (all 15 selected, `origin` host `github.com`)
  into an empty temporary directory writes exactly the Target paths of the Tool
  catalog, and every file under `.config/mise/tasks/` has mode `0755`. Pass: the
  path list equals the catalog's path list, and no task file lacks the
  executable bit.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this file, its owned
paths plus "touch nothing outside this list", the facts section, the shared-file
rule, and the return block below. A unit never bumps a version, never runs a
generator, never edits a doc, never adds a dependency this file does not list,
never commits. A unit deletes with plain `rm`, never `git rm` — it stages
nothing.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    GAP: <what the plan left unspecified and the assumption taken>   (or none)
    UNRESOLVED: <the ruling needed>         (or none)

A `GAP:` is a hole in the plan the unit could proceed past on a stated
assumption; it is recorded and the run continues. An `UNRESOLVED:` is a ruling
the unit could not proceed without; it blocks the unit and its dependents.

## Out of scope

- The `bootstrap init` command, its flags, prompts and output — plan 2 (flow
  110).
- Mapping error exit codes at the process boundary, usage errors (exit 2) —
  plan 2.
- Git and disk I/O: preflight, `origin` parsing, target checks, the atomic
  write, restore from `HEAD` — plan 2.
- The `add`, `remove` and `tui` commands — slices 2–4.
- The `site` project — slice 5.
- The `show` error "missing unremovable tool" (exit 1) and the orphan output —
  the plan for flow 170. Plan 1 gives the detection, the repair function and the
  `orphaned` list that `show` uses.
- The replacement consent (`--replace`, tool-category invariant 3) — plan 2 and
  the plans for `add` and `tui`.

## Parked

- The earlier plan `docs/plans/2026-10-06-2332-tool-setup-config` is `COMPLETE`
  and stays live because of its open gaps. This plan closes G1, and the
  blueprint answers G8, G9, G10, G15 and G18. Archive it by hand after this plan
  lands.
- G2 (effect/cli exit code on an unknown flag) and G3 (`-v` taken by
  `--version`) — plan 2, which wires the flags.
- G4 and G17 — corrections to the upstream stack prose; not this repo.

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |
| 0 | preflight | — | 1 | pass | format 25=25 silent; doctor cli: no blocking (LSP typescript present; drift: repo shape not stackgen-shaped (declined), graph 20 min stale, mempalace.yaml keyword overlap); conventions read from .claude/stackgen/templates/{typescript-effect-cli,pnpm-workspace,npm-package}.md; wave gate green on 7ffcf60 (install, check, test 164 pass, build); coverage_target 100 and review_round_cap 4 (defaults) | — |
| 0 | override | — | — | — | override: shared worktree: all-2026-10-09-0954 | — |
| 1 | U1 coverage | opus | 1 | pass | code: @vitest/coverage-v8 ^5.0.3 added, vitest.config.ts unchanged; 164 tests pass; coverage 97.79% lines (<100, gap recorded; pre-existing lines outside Owns); DECIDED caret range; GAP none (closes G1) | 1014fb1 |
| 1 | R1 | opus | 1 | pass | wave review: FINDINGS 0, CONTRACT clean, RULINGS clean; wave gate green (install, check, test 164, build) | — |
| 2 | U2 catalogs | opus | 1 | pass | code: 13 categories, 15 rescoped tools, unremovableTools/defaultSelection/checkRequires/pathIndex (throws on duplicate), catalog-driven literal test; render.ts and validity.ts type-follow only; own tests green, 16 failing tests + 11 type errors in U3/U4/U5-owned test files (UNRESOLVED ruled GAP by orchestrator, see gaps); coverage 96.68% (<100: render.ts:108-111 pending U5, pre-existing bin.ts/cli.ts/validity.ts:77); DECIDED checkRequires name pairs, pathIndex exported | 84016e8 |
| 2 | R2 | opus | 1 | pass | wave review: FINDINGS 0, CONTRACT clean, RULINGS clean; wave gate red-by-expectation (install ok, build ok; check 11 TS errors and test 16 failures, all in core-templates/new-templates (U3), setup-config.test (U4), render.test (U5) — the U2 gap); no unit failed | — |
| 3 | U3 templates | opus | 1 | pass | code:_default dispatchers (code/format,lint,sec,graph; setup/secrets,ai; setup/deps/<verb>/_default) + per-tool subtasks; conf.d/<tool>/mise.dev.toml for gitleaks, grype, fnox, graphify, mempalace; grype/gitleaks/fnox out of_base; dprint .config/dprint.json and claude conf.d/ai removed; owned tests 76/76 green, suite 209 pass / 5 fail (U4, U5 files); coverage not measured (suite red; no src change); GAP U3/gap/1; orchestrator set +x on 8 new task files (pre-commit shebang check), matching the existing 100755 task templates | 917c7f8 |
| 3 | U4 setup-config | opus | 1 | pass | code: closed schema, tools minItems 1, blueprint commit_scopes pattern; Validity max-one / row-16 / requires (exit 3), ValidSetupConfig brand; repair on read; requireSetUp + versionGuard (exit 1); refreshFiles full/remove + rewrites; write format; setup-config/**coverage 100%; suite 241 pass / 3 fail (render.test.ts, U5); GAP U4/gap/1 | 82fcd61 |
| 3 | R3 | opus | 1 | findings(1) | wave review: cli/templates/mise/.config/mise/tasks/code/all:14 [U3] completeness — comment still says scanners pinned in_base/mise.dev.toml; CONTRACT clean, RULINGS clean; wave gate red-by-expectation (check 2 TS errors, test 3 failures, all in render.test.ts, U5) | — |
| 3 | U3 templates | opus | 2 | pass | code (R3 loop-back): code/all comment now names dprint in _base and each scanner in conf.d/<tool>; owned tests 76/76 | 5ed611b |
| 3 | R3 | opus | 2 | pass | wave review round 2: FINDINGS 0; wave gate red-by-expectation (render.test.ts only, U5) | — |
| 4 | U5 renderer | opus | 1 | pass | code: renderFiles({values, selection, originHost}) renders the selection in catalog order with createOnly + executable; origin_tool per #21; _base/mise.dev.toml base-tool blocks selection-dependent; forge links via origin_tool; suite 254 pass / 0 fail, check + build green, --version 0.0.1; coverage 98.5% (render.ts 100%; residual bin.ts/cli.ts pre-existing); GAP U5/gap/1; mechanical re-dispatch after pre-commit check-yaml + house linter failed on duplicate keys in the raw git-conventional-commits.yaml template (one key per link, forge chosen inside the string) | 4331d5e |
| 4 | R4 | opus | 1 | findings(1) | wave review: docs/memory/decisions/tool-setup-config-run.md:16 [U5] rule 5 — recorded decision says core files must not depend on the selection (forge URLs dropped); U5 brings forge links back per #21, reversal not noted. The path is U7's Owns (docs/**), so it does not loop to U5: handed to U7 as a DOCS FALSIFIED line. CONTRACT clean, RULINGS clean; wave gate GREEN (install, check, test 254/254, build) | — |

## Acceptance criteria (from blueprint)

none — no flow touched

## Gaps surfaced during execution

- U1 coverage (code-unit coverage residual, non-blocking): 97.79% lines vs target 100 — uncovered cli/src/bin.ts:13-16, cli/src/cli.ts:42 (branch), cli/src/setup-config/validity.ts:82 (branch); pre-existing, bin.ts and cli.ts are owned by no unit of this plan (validity.ts is U2/U4's). Assumption: left as is; re-measured after U4.
- Run artifact: `vitest --coverage` writes `cli/coverage/`, which no .gitignore entry covers; no unit owns .gitignore. Assumption: the orchestrator deletes the directory after each coverage run so the merge task sees a clean tree.
- U2 GAP (plan hole, non-blocking): decision #12 promised every wave compiles, but the type change breaks four test files U2 does not own — cli/test/tool/core-templates.test.ts and new-templates.test.ts (U3), cli/test/setup-config/** (U4), cli/test/tool/render.test.ts (U5); 16 tests and 11 type errors red after wave 2. U2 returned it as UNRESOLVED; the orchestrator ruled it a gap, not a block: each file has an owner in a later wave, the fix is inside that owner's Owns, and nothing lands before the final gate. Assumption: U3, U4 and U5 each adapt the test files they own to the new catalog types; the wave gate is recorded red-by-expectation for those files after waves 2 and 3 and must be green from wave 4 on.
- U2 GAP (blueprint, non-blocking): tool invariant 10 cannot be checked from catalog data (a tool's mise entries live in its templates); the catalog test checks only that `base` is a boolean. Assumption: U3's template tests carry it, if anywhere.
- U3 GAP (plan hole, non-blocking, U3/gap/1): the plan did not say how a `_default` dispatcher passes flags and files on (assumed: each dispatcher forwards them and each subtask declares the old single-tool task's flags); did not move node and pnpm out of `_base/mise.toml` (assumed: moved to `_base/mise.dev.toml` as the blueprint lists, exact pins only in mise.toml); did not say where jq and yq go once claude's conf.d/ai file is deleted (assumed: `_base/mise.dev.toml`, since setup:ai:claude runs jq).
- U4 GAP (blueprint, non-blocking, U4/gap/1): the shipped catalog has no max-one category with two tools, so the max-one and row-16 rules are tested only against a mocked catalog; the setup-config entity doc gives no what/why/fix text for MaxOneExceeded, UnremovableReplaced, RequiredToolMissing or VersionTooOld (assumed: U4 wrote it); conventions name the error key `next_command` while the code keeps `fix` (assumed: plan 2 maps one to the other at the process boundary).
- U5 GAP (plan hole, non-blocking, U5/gap/1): the plan did not require templates to stay formatter-valid (assumed: Liquid tags open and close inside TOML/YAML strings so `dprint check` passes); gave no render signature U3's tests could share (assumed: `render(tools, values, …)` kept beside `renderFiles(input)`); gave no GitLab link format (assumed: `/-/commit`, `/-/compare`, `/-/issues` on gitlab.com); the template's commitScopes line and `vscode/settings.json` already failed `dprint check` before this run and are left as they were.
- Coverage residual (non-blocking): after U5 the suite is 98.5% lines vs target 100 — cli/src/bin.ts:13-16 and cli/src/cli.ts:42 (branch), pre-existing and owned by no unit of this plan; every file the plan touched is at 100%.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-10-08-1124-tool-catalog

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
