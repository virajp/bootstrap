# U3 — --json on the write commands

- **Read first:** `cli/internal/core/plan.go` (`PlanResult`: Changes, Files,
  Skipped, Pending, Warnings, Notes; about lines 101-126 before R1a),
  `cli/internal/core/apply.go`, `cli/internal/report/report.go` (`Outcome`,
  about line 158 before R1a, and where outcomes are counted), the `RunWith`
  function, the `add`, `remove`, `replace`, `init`, `detect` and `sync`
  branches of `cli/internal/cli/`, and the `cli/internal/contract` package that
  U1 adds.
- **Lazy-load:** the consent and prompt code of `cli/internal/cli/`,
  `docs/design/cli.md` ("Consent", "The output of a run").

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. D1, D11 and D12 are quoted in
`01-envelope-contract.md`.

- **D4:** with `--json` a command never prompts. It must also get `--yes` or
  `--dry-run`, else it stops with code 2. `data` holds the plan (changes,
  files, skipped, pending, as `core.PlanResult` has them) and, after an apply,
  one outcome per target. This needs an observer on `report.Outcome`, and
  `RunWith` must return the outcomes.
- **D21:** `sync` is exempt from the `--yes`-or-`--dry-run` rule. `sync --json`
  applies with no `--yes`, as `sync` does today, and `--dry-run` still works.
  The rule applies to `add`, `remove`, `replace`, `init` and `detect`.
- **D22:** under `--json --yes`, a plan with questions that have no answer
  (`p.Pending`) stops with code 2 and writes nothing. `data.plan.pending` lists
  the questions, and one error names them. The plugin gives the answers with
  `config set` and runs the command again.
- **D23:** `data.plan` = `{changes, files, skipped, pending, notes}`. The plan
  warnings go to the envelope `warnings` (D11). Side, Cur/Next, Sel, AskOnly,
  Before, Seeds and Reasons stay out. Each nested field (`oldform.Change`,
  `target.Planned`) is the Go field in snake case; each enum such as `Kind` is
  a lower-case string. One outcome item per `report.Outcome` call:
  `{kind, path, message}`.
- **D45:** under `--json --dry-run`, pending questions do not call `ask()`.
  The command exits 0; `data.plan.pending` lists the questions, and one
  warning names them.
- **D20:** `--json` on the `mac` write commands is refused with code 2 (U1).
- **D30:** a plain `detect --json` (no `--add`) is a read: it needs no `--yes`
  or `--dry-run`, and `data` = `{"detected": [...]}`, one entry per tool that
  `detect` reports today, its fields in snake case. `detect --add --json`
  follows D4: it needs `--yes` or `--dry-run`, and its `data` is the plan form
  of D23.

## Edits

1. **`cli/internal/report/report.go`** — add an observer that receives each
   `Outcome` call (`kind`, `path`, `message`, D23). Without an observer the
   behaviour does not change.
2. **`RunWith` and its callers** — return the outcomes that the observer
   collected.
3. **The JSON form of the plan** — one type for `data.plan` with the D23
   fields; the nested types in snake case, each enum as a lower-case string.
4. **The `add`, `remove`, `replace`, `init`, `detect` and `sync` branches** —
   allow `--json`. Under `--json`: no prompt, no TUI. A plain `detect --json`
   prints the D30 form. For `add`, `remove`, `replace`, `init` and
   `detect --add`, without `--yes` or `--dry-run`, stop with code 2 and one
   error that names both flags (D21). With pending questions under `--yes`,
   stop as D22 says. With `--dry-run`, `data` = `{"plan": ..., "outcomes": []}`;
   after an apply, `data` = `{"plan": ..., "outcomes": [...]}`. All human lines
   (the change list, the trees) go to stderr. Add each command to the
   `commands` list of `contract` (D13).
5. **Tests** — for each of the five: `--json` with neither flag exits 2 and
   writes no file. For all six: `--json --dry-run` returns the plan and no
   outcome; an apply on a scratch repo returns the outcomes. One test for D22
   (pending questions, code 2, no write), and one for `sync --json` with no
   `--yes`.

## Verification

- `mise x -- mise run test:go` passes.
- Each of the six commands without `--json` prints the same lines as before
  this unit.

## Guardrails

- Touch nothing outside `cli/`.
- Do not change what a command writes; only what it reports.
- Delete with `rm`, never `git rm`.

## Commit

`feat: --json on add, remove, replace, init, detect and sync`
