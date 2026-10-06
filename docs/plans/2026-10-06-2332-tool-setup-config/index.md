---
type: vwf-plan
title: Tool catalog and setup config
requires: []
backlog: []
backlog_pieces: []
covers:
  - docs/blueprint/entities/tool/index.md
  - docs/blueprint/entities/setup-config/index.md
---

# Plan — Tool catalog and setup config (2026-10-06)

## Status

**APPROVED**

APPROVED 2026-10-06 by the user

## Consent

| Action                                            | Granted |
| ------------------------------------------------- | ------- |
| Merge to the integration branch and push on green | yes     |
| After landing: archive the plan folder            | run     |
| After landing: `mise run code:graph`              | run     |
| After landing: `/vwf:handoff` for slice 2         | ask     |
| Release cli publicly                              | none    |

**The mode recorded here is the consent.** A `run` step runs on a green landing
without a prompt; an `ask` step stops the run once before it, reports what it
would do, and waits. The release row records the user's answer "No release yet":
no version bump, no publish.

## Goal

The `cli` package exists in the pnpm workspace and holds the tool catalog of
[Tool](../../blueprint/entities/tool/index.md) (13 core tools, `github`,
`gitlab`, and every target path), one template per target path, a renderer that
turns templates and setup values into file contents, and the code that reads,
validates and writes `.config/bootstrap.yaml` per
[Setup config](../../blueprint/entities/setup-config/index.md). After landing,
both entity docs read `implementation: complete`. No command uses them yet;
`bootstrap init` is plan 2.

## Slice

Entities [Tool](../../blueprint/entities/tool/index.md) and
[Setup config](../../blueprint/entities/setup-config/index.md), planned together
because each links the other (one chain element).

Plan 1 of 2 — requires nothing; required by
`docs/plans/2026-10-06-2333-setup-repository` (flow `cli/110-setup-repository`).

## Facts the survey established

- `cli/` does not exist. There is no `pnpm-workspace.yaml`. The root
  `package.json` is a placeholder (name `bootstrap`, `test` script fails).
  `pnpm-lock.yaml` is empty.
- Stack: `project/typescript-effect-cli`
  (`.claude/stackgen/templates/typescript-effect-cli.md`): commands as values,
  `effect` platform services behind layers, `NodeRuntime.runMain` at the process
  boundary, no direct `node:fs` or `process.exit`, Vitest with `@effect/vitest`,
  strict TypeScript, `@/` alias, one shared `tsconfig.base.json`. Repo stack
  `repo/pnpm-workspace`: members listed by name, `.npmrc` with
  `ignore-scripts=true` and `fund=false`, `minimumReleaseAge` cooldown. Deploy
  `npm-package`: one `bin` entry with a `#!/usr/bin/env node` shebang, `files`
  allowlist.
- npm registry (2026-10-06): `effect` 4.0.1 and `@effect/platform-node` 4.0.1
  are `latest`; `@effect/cli` 0.77.2 is the v3 line. Effect v4 holds the CLI in
  `effect/cli` (`Command`, `Flag`, `Argument`, `Prompt` with `Text`,
  `MultiSelect`, `Confirm`). `liquidjs` 10.30.0, `yaml` 2.9.1.
- LiquidJS takes `tagDelimiterLeft/Right`, `outputDelimiterLeft/Right`,
  `strictVariables`, and an in-memory `templates` map.
- Template sources: every core target path except `.config/fnox.toml` exists in
  this repo (tool name, then source): `mise` — `.config/mise.toml`,
  `.config/miserc.toml`, `.config/mise/conf.d/_base/*`, the listed tasks; `git`
  — `.gitignore`, `tasks/code/git-config`; `pre-commit` —
  `.config/pre-commit-config.yaml`, `.config/git-conventional-commits.yaml`,
  `tasks/code/precommit`, `tasks/setup/precommit`; `vscode` —
  `.vscode/settings.json`, `.vscode/extensions.json`; `dprint` — `dprint.json`,
  `.config/dprint.json`, `tasks/code/format`; `taplo` — `.config/taplo.toml`;
  `gitleaks` — `.config/gitleaks.toml`, `tasks/code/sec`; `grype` —
  `.config/grype.yaml`; `virajp-linter` — `.config/linter.yaml`,
  `eslint.config.mjs`, `tasks/code/lint`; `claude` —
  `.config/mise/conf.d/ai/mise.dev.toml`, `.config/claude-status.json`,
  `tasks/setup/ai`; `graphify` — `tasks/code/graph`, `.graphifyignore`;
  `mempalace` — `mempalace.yaml`. `tasks/` means `.config/mise/tasks/`.
- Repo-specific content in those sources that must become template values or be
  removed: the repo name (`conf.d/_base/mise.toml:24`, `mempalace.yaml:1`,
  `conf.d/ai/mise.dev.toml:29`, `.config/claude-status.json:3`,
  `.config/git-conventional-commits.yaml:27,28,32`); one `MERGE_MODEL` variable
  (`conf.d/_base/mise.toml:23`, read in `_scripts/merge:150`) where the setup
  config has `merge_model.develop` and `merge_model.main`; `commitScopes: []`
  (`.config/git-conventional-commits.yaml:10`); `MEMBERS`
  (`conf.d/_base/mise.toml:22`); vwf text (`setup/worktree:9`,
  `_scripts/placeholder:62-63`, `code/all:15`); Doppler (`setup/secrets`,
  `.vscode/settings.json:78`, `.vscode/extensions.json`); dart, flutter, astro,
  tailwind and pubspec entries in `.vscode/*`; `site/` and claude-plugins
  excludes (`.config/dprint.json:9-20`, `.config/linter.yaml:16-27,62-72`,
  `.graphifyignore:15`, `.gitignore:58-59`); a vwf memory block in `.gitignore`
  (64-65); `setup/all:40-43` names `@askviraj/linter` and calls tool tasks
  directly; `setup/deps/*` run pnpm directly.
- `.gitignore` has no `*.bak` line (Tool invariant 3).
- No source exists for `.config/fnox.toml`, the `github` paths or the `gitlab`
  paths.
- Harness: no test, build or typecheck task covers a `cli` project; `code:all`
  runs format, lint and security scan only. The `p/i/*` tasks target a missing
  `installer/` (parked).
- Backlog: unreadable — no GitHub project `bootstrap` under `virajp`; no ids.
- `/vwf:doctor cli` (2026-10-06): no blocking finding; no LSP question.

## Assumed decisions — confirm or override at review

| #  | Decision                       | Ruling                                                                                                                                                                                                                                | Rejected                                         | Unit       |
| -- | ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------ | ---------- |
| 1  | Effect major                   | Effect v4: `effect` 4.0.1 and `@effect/platform-node` 4.0.1; the CLI from `effect/cli`                                                                                                                                                | `@effect/cli` 0.77 on Effect v3                  | U1         |
| 2  | Build tool                     | tsdown builds `cli/dist` (earlier product decision)                                                                                                                                                                                   | plain `tsc` emit                                 | U1         |
| 3  | Template engine                | LiquidJS with tag delimiters `<%` `%>` and output delimiters `<%=` `%>`, `strictVariables: true` (earlier product decision)                                                                                                           | default `{% %}` / `{{ }}` delimiters             | U3         |
| 4  | Template storage               | "Embed at build": templates live as files under `cli/templates/<tool>/<target path>`; the build bundles them into the JS as an in-memory map; nothing reads template files at run time                                                | ship templates as package files read at run time | U3, U5, U6 |
| 5  | Setup config parsing           | the `yaml` package reads and writes `.config/bootstrap.yaml`; Effect Schema validates it against `schema.yaml`                                                                                                                        | a hand-written parser                            | U4         |
| 6  | Review placement               | "One per plan": one review row after the last code unit                                                                                                                                                                               | an extra review row before the templates         | U7         |
| 7  | Drift: repo sources vs catalog | the templates conform to the Tool catalog and the setup values; the repo's own tracked files stay unchanged (they are sources, not targets)                                                                                           | edit this repo's files to match                  | U5, U6     |
| 8  | Merge model in templates       | templates set `MERGE_MODEL_DEVELOP` from `values.merge_model.develop` and `MERGE_MODEL_MAIN` from `values.merge_model.main`; `MEMBERS` renders empty (the setup config has no members value)                                          | one `MERGE_MODEL`                                | U5         |
| 9  | `setup/deps` dispatchers       | authored new: each of `setup/deps/{all,install,outdated,audit,upgrade,cleanup}` runs the `setup/deps/<verb>/<tool>` task of every tool that has one and does nothing when there is none; no 1.0 tool has one                          | copy the pnpm tasks                              | U5         |
| 10 | `setup/all`                    | orchestrates by task name only; it runs a tool's setup task (`setup:ai`, `setup:secrets`, `setup:precommit`) only when that task exists                                                                                               | call tool tasks unconditionally                  | U5         |
| 11 | Template values                | every template reads only the setup values: `repo`, `commit_scopes`, `merge_model.develop`, `merge_model.main`, `tools`; derived names (e.g. the repo name, the last path segment of `repo`) are computed in the renderer, not stored | per-tool values                                  | U3, U5     |

## New dependencies

All added by U1 to `cli/package.json`:

- `effect` 4.x — the runtime, `effect/cli`, Schema; preferred over `@effect/cli`
  0.77 (v3).
- `@effect/platform-node` 4.x — Node file system, process and terminal layers,
  `NodeRuntime.runMain`.
- `liquidjs` 10.x — template rendering with custom delimiters; product decision.
- `yaml` 2.x — read and write `.config/bootstrap.yaml`; no YAML parser exists in
  the tree.
- dev: `typescript` — type checking; `vitest` and `@effect/vitest` — tests
  (stack convention); `tsdown` — build (product decision); `@types/node` — Node
  types.

## Units

| Id | Wave | Unit file                                    | Kind   | Owns                                                                                                                                                                                                                                                                                                                                              | Depends on             | Status  | Commit |
| -- | ---- | -------------------------------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- | ------- | ------ |
| U1 | 1    | [01-scaffold.md](01-scaffold.md)             | code   | `pnpm-workspace.yaml`, `package.json`, `.npmrc`, `tsconfig.base.json`, `cli/package.json`, `cli/tsconfig.json`, `cli/vitest.config.ts`, `cli/tsdown.config.ts`, `cli/src/bin.ts`, `cli/src/cli.ts`, `cli/test/cli.test.ts`, `.config/mise/tasks/p/cli/build`, `.config/mise/tasks/p/cli/test`, `.config/mise/tasks/p/cli/check`, `pnpm-lock.yaml` | —                      | pending |        |
| U2 | 2    | [02-tool-catalog.md](02-tool-catalog.md)     | code   | `cli/src/tool/catalog.ts`, `cli/test/tool/catalog.test.ts`                                                                                                                                                                                                                                                                                        | U1                     | pending |        |
| U3 | 3    | [03-renderer.md](03-renderer.md)             | code   | `cli/src/tool/render.ts`, `cli/src/tool/templates.ts`, `cli/test/tool/render.test.ts`, `cli/test/fixtures/templates/**`                                                                                                                                                                                                                           | U2                     | pending |        |
| U4 | 3    | [04-setup-config.md](04-setup-config.md)     | code   | `cli/src/setup-config/**`, `cli/test/setup-config/**`                                                                                                                                                                                                                                                                                             | U2                     | pending |        |
| U5 | 4    | [05-core-templates.md](05-core-templates.md) | code   | `cli/templates/{mise,git,pre-commit,vscode,dprint,taplo,gitleaks,grype,virajp-linter,claude,graphify,mempalace}/**`, `cli/test/tool/core-templates.test.ts`                                                                                                                                                                                       | U3                     | pending |        |
| U6 | 4    | [06-new-templates.md](06-new-templates.md)   | code   | `cli/templates/{fnox,github,gitlab}/**`, `cli/test/tool/new-templates.test.ts`                                                                                                                                                                                                                                                                    | U3                     | pending |        |
| U7 | 5    | [07-review.md](07-review.md)                 | review | —                                                                                                                                                                                                                                                                                                                                                 | U1, U2, U3, U4, U5, U6 | pending |        |
| U8 | 6    | [08-docs.md](08-docs.md)                     | edit   | the repo's docs (README, CLAUDE.md, `docs/**` outside `docs/blueprint/` and `docs/plans/`)                                                                                                                                                                                                                                                        | all                    | pending |        |
| U9 | 7    | [09-gates-and-bump.md](09-gates-and-bump.md) | edit   | — (no release: no version file, no generator)                                                                                                                                                                                                                                                                                                     | U8                     | pending |        |

## Shared-file rule

| File                                    | Why it collides                                      | Owner                                                     |
| --------------------------------------- | ---------------------------------------------------- | --------------------------------------------------------- |
| `cli/package.json`, root `package.json` | one manifest; dependency edits race                  | U1 only                                                   |
| `pnpm-lock.yaml`                        | written by the install that adds the dependencies    | U1 only; later units run `pnpm install --frozen-lockfile` |
| `cli/src/tool/catalog.ts`               | the catalog every later unit reads                   | U2 only                                                   |
| `cli/src/tool/templates.ts`             | the embed map; U5 and U6 add templates as files only | U3 only                                                   |
| README, CLAUDE.md, `docs/**`            | n units editing one doc                              | U8 only                                                   |

## Waves

- Wave 1: U1 — the workspace and package everything else needs.
- Wave 2: U2 — the catalog; the renderer and the setup config read it.
- Wave 3: U3, U4 — disjoint trees (`cli/src/tool/` render vs
  `cli/src/setup-config/`).
- Wave 4: U5, U6 — disjoint template folders.
- Wave 5: U7 review; wave 6: U8 docs; wave 7: U9 gates.

Waves are ordering only; `execute` runs the units serially.

## Wave gate

`MISE_ENV=dev mise run code:all` plus the wave review, plus every report read
for `UNRESOLVED:`. The `p:cli:*` tasks exist only after U1, so each unit names
them in its own Verification.

## After landing

| Step                                                       | Mode | Notes                                                                  |
| ---------------------------------------------------------- | ---- | ---------------------------------------------------------------------- |
| archive the plan folder (`plan-management archive`)        | run  | moves this folder to `docs/plans/archived/` and updates the plan index |
| `mise run code:graph`                                      | run  | refreshes the code graph so it includes `cli/`                         |
| `/vwf:handoff` pointing at `/vwf:plan` for slice 2 (check) | ask  | stops once and offers the handoff                                      |

## Gates the orchestrator keeps

- `pnpm --filter cli build` succeeds, and `node cli/dist/bin.mjs --version`
  prints the `cli/package.json` version and exits 0 (pass: exact match).
- `node cli/dist/bin.mjs` with no command prints help and exits 2.

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

## Out of scope

- `bootstrap init` and every other command — plan 2 and slices 2–5.
- Publishing to npm — the release pipeline is parked.
- Editing this repo's own tooling files to match the templates (decision 7).

## Parked

- Release pipeline: npm trusted publishing, the tag shape, the npm version on
  the runner (`docs/memory/gaps/npm-release-open-decisions.md`), and the stale
  `.config/mise/tasks/p/i/*` tasks that target a missing `installer/` — rewrite
  or delete them in the release plan.
- The `site` project (flows `site/100-home`, `site/110-documentation`) —
  slice 6.
- Flows `cli/120-check-drift`, `cli/130-update-repository`, `cli/140-add-tool`,
  `cli/150-remove-tool` — slices 2–5.
- Dogfooding: re-render this repo's own tooling from the templates with
  `bootstrap init`/`update` once those commands exist.

## Run log

| Wave | Unit | Model | Round | Outcome | Detail | Commit |
| ---- | ---- | ----- | ----- | ------- | ------ | ------ |

## Acceptance criteria (from blueprint)

none — no flow touched. The entity invariants are covered by the unit tests of
U2, U4, U5 and U6.

## Gaps surfaced during execution

- …

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the fresh session's worktree — cut from the integration branch — can see it.

Run in a fresh session:

/vwf:execute docs/plans/2026-10-06-2332-tool-setup-config

or let the queue pick it, by priority:

/vwf:execute next
