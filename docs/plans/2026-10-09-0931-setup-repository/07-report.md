# U7 — Result output: human, JSON and progress

- **Wave:** 4
- **Depends on:** U2
- **Owns:** `cli/src/report/human.ts` (new), `cli/src/report/json.ts` (new),
  `cli/src/report/progress.ts` (new), `cli/test/report/output.test.ts` (new)
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/report/output.test.ts` (pure over a result value):
  - Human: one group per non-empty key in the JSON key order (`created`,
    `changed`, `deleted`, `unchanged`, `kept`, `orphaned`, `replaced`,
    `warnings`), each headed by the status word, its glyph and a count, then one
    path per indented line; `replaced` shows `<old> → <new>`; a warning shows
    `! <text>`; an empty key is left out.
  - Closing line: "commit the changes, then run
    `MISE_ENV=dev mise run setup:all`" when a tool was added (first run, new
    tool, replacement, repair) or a `mise` config file was created, changed or
    deleted; else "commit the changes" when a file was created, changed, deleted
    or replaced; else no closing line. A dry run shows the closing line of the
    real run.
  - JSON: one document with `exit`, `created`, `changed`, `deleted`,
    `unchanged`, `kept`, `orphaned`, `replaced` (`{from, to}`), `warnings`,
    `next_command` only when the next command is printed, and `"dry_run": true`
    under `--dry-run`; every list sorted.
  - Color: off with `NO_COLOR`, `--no-color`, a non-terminal stdout and always
    under `--json`; on a terminal each glyph uses its color role.
  - `--quiet` prints no result and no progress; `--json` ignores `--quiet`.
    `--verbose` adds each per-file decision.
  - Progress: a spinner per step when stderr is a terminal; one stable line per
    finished step when it is not, in CI or under `--json`.
  - It fails because the report modules do not exist.
- **Read first:** `docs/blueprint/design-system.md` `#terminal-ux`,
  `docs/blueprint/conventions.md` `#errors`, flow 110 step 9,
  `cli/src/report/errors.ts`.
- **Lazy-load:** `docs/blueprint/design-system.md` color tokens.

## Ruling

| # | Decision          | Ruling                                                          | Rejected            | Unit |
| - | ----------------- | --------------------------------------------------------------- | ------------------- | ---- |
| 8 | Color and spinner | No new dependency: ANSI codes by color role and a small spinner | `picocolors`, `ora` | U7   |

## Edits

1. **`cli/src/report/human.ts`** — the human result per the test.
2. **`cli/src/report/json.ts`** — the JSON success document.
3. **`cli/src/report/progress.ts`** — the progress lines and spinner on stderr.
4. **`cli/test/report/output.test.ts`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.
- `mise x -- mise run p:cli:e2e` exits 0.

## Guardrails

- Do not touch `cli/src/report/errors.ts` (U2).
- Results go to stdout; progress and warnings go to stderr.
- Roles map to the basic ANSI palette by role name, never a fixed hex value.
- Copy the closing lines byte for byte from `conventions.md` `#errors`.

## Commit

`feat: print init results as text or JSON with progress`
