# U4 — Setup config: closed schema, validity, repair, write format, version guard

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `cli/src/setup-config/**`, `cli/test/setup-config/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/setup-config/setup-config.test.ts` — one failing
  case per rule of the
  [Setup config](../../blueprint/entities/setup-config/index.md) doc that the
  code does not meet:
  - Schema: a `kept` or `deleted` key fails (closed schema); `values.tools: []`
    fails (`minItems: 1`); `commit_scopes` `a-` and `a--b` fail.
  - Validity: an unremovable tool in `values.tools` passes; two tools of one
    `max: one` category fail; a `max: one` category that misses its unremovable
    tool and holds another tool fails (row 16); a tool whose `requires` is not
    in `values.tools` fails (`taplo` without `dprint`, through `checkRequires`).
    Each failure is a typed error with what, why and fix, and exit code 3.
  - Check order (row 14): a file with an older `version` that also fails
    Validity fails with exit code 3, not 1.
  - Repair on read: a `values.tools` that misses an unremovable tool, with no
    other tool of its category recorded, returns the repaired config, the list
    of tools added back, and the warning
    `added back unremovable tool \`<name>\``.
  - Write format: keys in schema order; `values.tools`, `values.commit_scopes`
    and `files` sorted; `format` and `version` set to the running cli's; no
    comments.
  - `requireSetUp` (row 15): an absent file fails with exit code 1 and
    `not set up — run \`bootstrap init\``; a present file passes.
  - `versionGuard` (row 15): an older recorded `version` fails with exit code 1
    and the next command `bootstrap init`; an equal version passes. It accepts
    only a config that passed Validity (by type).
  - Refresh of `files` (row 17; a pure function over the recorded config, the
    mode, the set of paths the running cli renders, and the set of recorded
    paths present on disk):
    - Full mode: rendered paths are listed; a rendered path absent on disk is
      listed.
    - Remove mode: the paths of the removed tools leave; no rendered path is
      added except the paths of a tool that Repair on read added back.
    - Both modes: a recorded path not rendered and present on disk stays and is
      returned as `orphaned`; a recorded path not rendered and absent on disk
      leaves silently; `.config/bootstrap.yaml` never stays.
  - Rewrite rule (row 18): in the init and tui-apply modes, the file is
    rewritten when the new text differs from the recorded text, also when only
    comments, key order or list order differ, or when `version` or `format`
    differ. In the add and remove modes, it is rewritten only when the run has a
    tool change or a repair; with neither it is not rewritten, also when the
    text differs.
- **Read first:** `docs/blueprint/entities/setup-config/index.md`,
  `docs/blueprint/entities/setup-config/schema.yaml`, `cli/src/tool/catalog.ts`,
  `cli/src/tool-category/catalog.ts`, `cli/src/cli.ts` (the `NoCommand`
  exit-code pattern), every owned file.
- **Lazy-load:** `docs/blueprint/conventions.md` `#errors`, `#config`; Effect
  Schema and `yaml` docs via Context7.

## Ruling

| #  | Decision                                                                                                                                   | Ruling                                                                                                                                                                                                                                                                                                                          | Rejected                           | Unit       |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- | ---------- |
| 2  | Drift: code refuses unremovable tools in `values.tools` (`CoreToolListed`)                                                                 | `values.tools` lists every selected tool, base and unremovable tools included (setup-config invariant 3)                                                                                                                                                                                                                        | Keep core tools implicit           | U4         |
| 3  | Drift: config has `kept` / `deleted` keys and `PathKeptAndDeleted`                                                                         | Remove them; the schema is closed over `format`, `version`, `values`, `files`                                                                                                                                                                                                                                                   | Keep them as extensions            | U4         |
| 5  | Version guard in plan 1                                                                                                                    | Build it now (user, 2026-10-08)                                                                                                                                                                                                                                                                                                 | Defer to slice 2                   | U4         |
| 7  | Exit codes of entity errors                                                                                                                | Typed errors carry their exit code, as `NoCommand` does (`cli/src/cli.ts:14-17`): Validity → 3, version guard → 1                                                                                                                                                                                                               | Map the codes only in the CLI      | U4         |
| 9  | Setup-config invariant 1 (written last; a safety target; a dirty setup file is refused only when its render differs from the working copy) | Lands with plan 2's atomic write; plan 1 has no disk writer                                                                                                                                                                                                                                                                     | A disk writer in plan 1            | — (plan 2) |
| 13 | Plan constraint: catalog-driven code (user, 2026-10-08)                                                                                    | No tool name or category name appears as a literal in `cli/src/{tool,tool-category,setup-config}` outside `cli/src/tool/catalog.ts` and `cli/src/tool-category/catalog.ts`. Every rule reads catalog fields. A test scans these folders and fails on a literal                                                                  | Enforce it in the review only      | U2, U4, U5 |
| 14 | Check order on read                                                                                                                        | `read` applies Validity (exit 3) first. The version guard (exit 1) accepts only a valid config. An older `version` plus a Validity failure exits 3                                                                                                                                                                              | Guard first                        | U4         |
| 15 | The absent file and the older version                                                                                                      | Two functions. `requireSetUp`: an absent file exits 1 "not set up — run `bootstrap init`"; `add`, `remove` and `show` call it. `versionGuard`: an older `version` exits 1, next command `bootstrap init`; `add`, `remove` and `tui` call it. The tui first run calls neither                                                    | One guard with a command parameter | U4         |
| 16 | Validity: a `max: one` category that misses its unremovable tool and holds another tool                                                    | Fails Validity, exit 3. The check runs before Repair on read                                                                                                                                                                                                                                                                    | Repair it                          | U4         |
| 17 | Refresh of `files` (setup-config invariants 2, 4, 5, 6)                                                                                    | Full mode (init, add, tui apply): the rendered set. Remove mode: the recorded paths minus the paths of the removed tools, plus the paths of a repaired tool. In both modes a path that is not rendered stays as `orphaned` when it is on disk and drops when it is absent; a rendered path stays also when it is absent on disk | One mode for all commands          | U4         |
| 18 | Rewrite rule (setup-config invariant 7)                                                                                                    | A pure function. Init and the tui apply rewrite when the text differs from its render. Add and remove rewrite only when the run has a tool change or a repair                                                                                                                                                                   | Leave it to each flow              | U4         |

## Edits

1. **`cli/src/setup-config/schema.ts`** — remove `kept` and `deleted`; add
   `minItems: 1` to `values.tools`; set the `commit_scopes` item pattern to
   `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`. Keep the stricter `values.repo` pattern
   (G13).
2. **`cli/src/setup-config/validity.ts`** — remove `PathKeptAndDeleted`; add the
   max-one rule (through `categoriesByName`), the missing-unremovable rule (row
   16) and the requires rule (through `checkRequires`). Every Validity error
   carries exit code 3.
3. **Repair on read** — a function that takes a valid config and returns the
   repaired config plus a list of warnings. A missing unremovable tool whose
   category holds another recorded tool is a Validity failure, not a repair.
   `read` applies Validity first (row 14), then the repair, and returns the
   warnings beside the config.
4. **`requireSetUp` and `versionGuard`** (row 15) — two functions, each with its
   typed error at exit code 1. `versionGuard` takes the validated config type,
   so a caller cannot run it before Validity.
5. **Refresh and write** — a pure function for setup-config invariants 2, 4, 5
   and 6 with the full and remove modes (row 17), returning the next config and
   the `orphaned` list. A pure rewrite rule for invariant 7 (row 18). `write`
   returns the text in schema key order with the three lists sorted.
6. **`cli/test/setup-config/**`** — the test-first cases; remove the
   `CoreToolListed` and `kept`/`deleted` cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `grep -rn "kept\|deleted\|CoreToolListed" cli/src/setup-config` finds no
  schema key or error of the old model.

## Guardrails

- Do not touch `cli/src/tool/**`, `cli/src/tool-category/**` or
  `cli/templates/**`.
- No `node:fs` import and no disk write. `write` returns text; plan 2 writes it.
- Copy the message texts from the entity doc byte for byte.
- Row 13: no tool name or category name as a literal. The rules read
  `removable`, `max`, `category` and `requires` from the catalogs.
  `cli/test/catalog-driven.test.ts` (U2) must pass.

## Commit

`feat: validate, repair and guard the setup config per the rescoped contract`
