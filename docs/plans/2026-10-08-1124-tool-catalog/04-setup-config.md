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
    `max: one` category fail; a tool whose `requires` is not in `values.tools`
    fails (`taplo` without `dprint`). Each failure is a typed error with what,
    why and fix, and exit code 3.
  - Repair on read: a `values.tools` that misses an unremovable tool, with no
    other tool of its category recorded, returns the repaired config and the
    warning `added back unremovable tool \`<name>\``.
  - Write format: keys in schema order; `values.tools`, `values.commit_scopes`
    and `files` sorted; `format` and `version` set to the running cli's; no
    comments.
  - Version guard: an older recorded `version` fails with exit code 1 and the
    next command `bootstrap init`; an absent file fails with exit code 1 and
    `not set up — run \`bootstrap init\``; an equal version passes.
  - Refresh of `files` (a pure function over the recorded config, the set of
    paths the running cli renders, and the set of recorded paths present on
    disk): rendered paths are listed; a recorded path not rendered and present
    on disk stays and is returned as `orphaned`; a recorded path absent on disk
    leaves silently; `.config/bootstrap.yaml` never stays.
  - Rewrite rule: the new text differs from the recorded text when only
    comments, key order or list order differ, or when `version` or `format`
    differ.
- **Read first:** `docs/blueprint/entities/setup-config/index.md`,
  `docs/blueprint/entities/setup-config/schema.yaml`, `cli/src/tool/catalog.ts`,
  `cli/src/tool-category/catalog.ts`, `cli/src/cli.ts` (the `NoCommand`
  exit-code pattern), every owned file.
- **Lazy-load:** `docs/blueprint/conventions.md` `#errors`, `#config`; Effect
  Schema and `yaml` docs via Context7.

## Ruling

| # | Decision                                                                   | Ruling                                                                                                            | Rejected                      | Unit       |
| - | -------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- | ----------------------------- | ---------- |
| 2 | Drift: code refuses unremovable tools in `values.tools` (`CoreToolListed`) | `values.tools` lists every selected tool, base and unremovable tools included (setup-config invariant 3)          | Keep core tools implicit      | U4         |
| 3 | Drift: config has `kept` / `deleted` keys and `PathKeptAndDeleted`         | Remove them; the schema is closed over `format`, `version`, `values`, `files`                                     | Keep them as extensions       | U4         |
| 5 | Version guard in plan 1                                                    | Build it now (user, 2026-10-08)                                                                                   | Defer to slice 2              | U4         |
| 7 | Exit codes of entity errors                                                | Typed errors carry their exit code, as `NoCommand` does (`cli/src/cli.ts:14-17`): Validity → 3, version guard → 1 | Map the codes only in the CLI | U4         |
| 9 | Setup-config invariant 1 (written last; a safety target)                   | Lands with plan 2's atomic write; plan 1 has no disk writer                                                       | A disk writer in plan 1       | — (plan 2) |

## Edits

1. **`cli/src/setup-config/schema.ts`** — remove `kept` and `deleted`; add
   `minItems: 1` to `values.tools`; set the `commit_scopes` item pattern to
   `^[a-z][a-z0-9]*(-[a-z0-9]+)*$`. Keep the stricter `values.repo` pattern
   (G13).
2. **`cli/src/setup-config/validity.ts`** — remove `PathKeptAndDeleted`; add the
   max-one rule (through `categoriesByName`) and the requires rule. Every
   Validity error carries exit code 3.
3. **Repair on read** — a function that takes a valid config and returns the
   repaired config plus a list of warnings. A missing unremovable tool whose
   category holds another recorded tool is a Validity failure, not a repair.
   `read` returns the warnings beside the config.
4. **Version guard** — a function over the recorded config (or its absence) and
   the running version, with the two typed errors at exit code 1.
5. **Refresh and write** — a pure function for setup-config invariants 2, 4, 5
   and 6, returning the next config and the `orphaned` list. `write` returns the
   text in schema key order with the three lists sorted.
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

## Commit

`feat: validate, repair and guard the setup config per the rescoped contract`
