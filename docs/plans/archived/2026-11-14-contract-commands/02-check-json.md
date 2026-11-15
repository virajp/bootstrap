# U2 — check --json

- **Read first:** `cli/internal/repo/repo.go` (`check()` and the code that
  calls it with `--check`), `cli/internal/fail/fail.go` (`Collector`), the
  `cli/internal/contract` package that U1 adds, and the `check` branch of
  `cli/internal/cli/cli.go`.
- **Lazy-load:** `cli/internal/report/report.go` (how a stage writes its lines
  today), `docs/design/cli.md` ("The exit codes").

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. D1, D11 and D12 are quoted in
`01-envelope-contract.md`.

- **D8:** `check --json` prints `data` = `{"valid": bool, "checks": [...]}`.
  `checks` has one entry per stage of `check()` in
  `cli/internal/repo/repo.go`, in order, as the code is after R1a:
  `{"name", "status": "pass"|"fail"|"skipped", "findings": [{"code",
  "severity": "error"|"warning", "message", "path"}]}`. A `fail.Collector` per
  stage collects the findings. A stage that `check()` skips because its input
  failed has status `skipped`. The stage name is the kebab-case name of what
  the stage examines; the stages of today are `envs`, `values`, `old-forms`,
  `env-rules`, `overrides`, `members` and `scripts`, and a stage R1a adds takes
  a name the same way. `path` is `""` when no file is known. The envelope
  `code` is the highest stage code, as `check()` returns today.
- **D32:** if `check` fails before any stage runs (in prepare or openClone),
  `data` = `{"valid": false, "checks": []}`, and the cause goes in `errors`.
- **D31:** `mac check --json` is refused with code 2 (U1 refuses it in the
  flag table).
- **D35:** add `check` to the `commands` list of `contract`.

## Edits

1. **`cli/internal/repo/repo.go`** — give each stage of `check()` a
   `fail.Collector` and a name. Return the stage list with the code, so the
   caller can print it. The human output without `--json` does not change: the
   same lines, in the same order, on the same stream.
2. **The `check` branch of `cli/internal/cli/cli.go`** — allow `--json`; print
   the envelope with `data` of D8; `valid` is true only when the code is 0. An
   early failure prints the D32 form. Add `check` to the `commands` list of
   `contract` (D35).
3. **Tests** — a Go test per status (`pass`, `fail`, `skipped`) and one for an
   unknown key in `.config/bootstrap/repo.toml` (R1a D8): the `values` stage
   fails with one finding whose `path` is that file, and the stages that need
   it are `skipped`.

## Verification

- `mise x -- mise run test:go` passes.
- `bootstrap check` without `--json` prints the same lines as before this unit
  for a valid repo and for a repo with an unknown key (compare with a test
  golden or the test of today).

## Guardrails

- Touch nothing outside `cli/`.
- Do not change which stage runs or skips; only report it.
- Delete with `rm`, never `git rm`.

## Commit

`feat: bootstrap check --json reports each stage and its findings`
