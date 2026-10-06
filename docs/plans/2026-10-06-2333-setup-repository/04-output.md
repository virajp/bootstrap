# U4 — Output: human report and --json documents

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `cli/src/report/**`, `cli/test/report/**`
- **Model:** opus
- **Kind:** code
- **Test first:** `cli/test/report/report.test.ts` — (a) the init success
  document has exactly `exit`, `created`, `backed_up` (`{path, backup}`),
  `unchanged`, `next_command`, every list sorted by path; the dry-run document
  adds `"dry_run": true`; the error document has `exit` and
  `error = {what, why, next_command}`, plus `unrestored` (`{path, backup}`) on a
  failed rollback; an invalid document fails validation instead of printing; (b)
  human output: results on stdout, progress and errors on stderr; no colour when
  `NO_COLOR` is set, when the stream is not a TTY, with `--no-color`, or under
  `--json`; `--quiet` prints errors only; `--json` ignores `--quiet`. Fails
  because the module does not exist.
- **Read first:** flow 110 step 8 and Modes, `docs/blueprint/conventions.md`
  `#errors`, `docs/blueprint/design-system.md` Terminal UX.
- **Lazy-load:** Effect Schema and terminal service docs via Context7.

## Ruling

| # | Decision            | Ruling                                                                                    | Rejected           | Unit |
| - | ------------------- | ----------------------------------------------------------------------------------------- | ------------------ | ---- |
| 7 | `--json` validation | every `--json` document is an Effect Schema value, encoded and validated before it prints | build JSON by hand | U4   |

## Edits

1. **`cli/src/report/json.ts`** — Schemas for the init success, dry-run and
   error documents; an encoder that validates before returning text.
2. **`cli/src/report/human.ts`** — the human result for init (created, backed up
   as `original → backup`, unchanged, the one next command), colour roles
   `success` / `warning` / `error` / `emphasis` mapped to the terminal, never
   fixed shades.
3. **`cli/src/report/output.ts`** — the output policy: stdout vs stderr,
   `--quiet`, `--json`, colour on or off, spinner when interactive and stable
   lines otherwise.
4. **`cli/test/report/**`** — the test-first cases.

## Verification

- `mise x -- mise run p:cli:test` exits 0.
- `mise x -- mise run p:cli:check` exits 0.

## Guardrails

- The site's no-em-dash rule does not apply to CLI text: CLI messages keep the
  exact strings the conventions quote ("not a git repository — run `git init`
  first", "interrupted — nothing written", "interrupted — repository restored").
- Do not touch `cli/src/repo/**`, `cli/src/apply/**`, `cli/src/commands/**`.

## Commit

`feat: print the human and json reports`
