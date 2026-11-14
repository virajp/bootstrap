# U4 — The fetch rule, catalog and describe

- **Read first:** the clone and fetch code of `cli/internal/core/` (the
  fetch-time file of R1a D52 and the 7-day check of R1a D23), the `config set`
  and `unset` code of R1a, `cli/internal/source/manifest.go` (`Manifest`),
  `cli/internal/source/check.go` (`Index`, `Source.Open()`),
  `cli/internal/source/packs.go` (`Source.ReadPacks()`),
  `cli/internal/source/catalog.go` (`Source.ReadCategories()`), the
  `.config/bootstrap/` readers of R1a, and the `cli/internal/contract` package
  that U1 adds.
- **Lazy-load:** `cli/internal/cli/cli.go`, `docs/design/cli.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. D1, D3, D11, D12 and D13 are quoted in
`01-envelope-contract.md`.

- **D5:** (1) if the last successful fetch is 30 minutes old or less, use the
  cached clone with no fetch; (2) if it is older, or no fetch time is
  recorded, fetch; (3) if that fetch fails, use the cached clone when its last
  successful fetch is less than 7 days old, with a warning in the envelope;
  (4) else stop with code 5 — also when no fetch time is recorded (D44). The
  fetch time is the file R1a D52 writes in the
  cache root. Under `--clone` or `BOOTSTRAP_CLONE` there is no fetch and no age
  rule; the clone is used as it is. User's words: "Fetch when the last fetch is
  > 30mins and if the fetch fails then use the existing clone cache if it's
  > <7days old; otherwise return with an error (exit 5)".
- **D6:** one shared function applies D5 to `config set`, `config unset`,
  `catalog` and `describe`. It replaces R1a's 7-day skip for `config set` and
  `unset`. The other write commands still fetch every time.
- **D14:** `catalog --json` prints `data` = `{"fetched_at", "categories",
  "tools", "packs"}`: `fetched_at` is the time of the last successful fetch as
  ISO-8601 UTC, or `""` under `--clone` or `BOOTSTRAP_CLONE`; `categories` is
  what `Source.ReadCategories()` returns; `tools` is one entry per tool of the
  `Index` that `Source.Open()` returns, from its `Manifest`, with its name,
  category, description, questions and defaults; `packs` is what
  `Source.ReadPacks()` returns. Each field name is the Go field in snake case.
  Without `--json`, `catalog` prints the tools by category as human lines.
- **D29, D20:** `catalog` reads the repo side (`source.Repo`) only.
- **D28:** `catalog` needs only the clone and works outside a repo.
  `describe` outside a repo, or in a repo with no `.config/bootstrap/`, stops
  with code 2 and the fix line `→ Run bootstrap init`, as other repo commands
  do after R1a.
- **D10:** `describe --json` prints `data` = `{"repo", "tools", "members",
  "ports"}`: `repo` is the `repo.toml` values; `tools` is the selected tools
  with their values from `tools.toml`; `members` is each `[members.<name>]`
  with its type, folder, url and `present` (the folder exists); `ports` is D3
  with `present`. No `lsp` field.
- **D40:** without `--json`, `describe` prints human lines on stdout, one
  section per part (repo, tools, members, ports); each port shows present or
  missing. The same style as `list`.
- **D19, D27:** the engine runs `mise tasks ls --json --no-global --hidden` in
  the repo root once and matches the task names; a port is a task of this
  repo. A plain port is `{"name", "present"}`. A patterned port is
  `{"name": <the pattern>, "present": <true when at least one task matches
  (D39)>, "tasks": <every matching task name>}`, where `<branch>` and `<id>`
  match one
  colon-free segment. If mise fails (for example an untrusted config),
  `describe` stops with code 1 and an error that names `mise trust`.

## Edits

1. **The clone code of `cli/internal/core/`** — one function that applies D5
   and returns the clone path, the fetch time and an optional warning. Write
   its Go tests first, with a fake fetch and a fake clock: 10 minutes old (no
   fetch), 2 hours old (fetch), fetch fails at 2 days (cache plus warning),
   fetch fails at 8 days (code 5), no fetch-time file (fetch), no fetch-time
   file and the fetch fails (code 5, D44), `--clone` (no fetch).
2. **`config set` and `config unset`** — use that function (D6).
3. **A `catalog` command** in `cli/internal/cli/` — D14, with and without
   `--json`.
4. **A `describe` command** in `cli/internal/cli/` — D10 and D19, with and
   without `--json`.
5. **`cli/internal/cli/cli.go`** — usage lines for both; add both to the
   `commands` list of `contract` (D13).
6. **Tests** — `catalog --json` against a test clone, also outside a repo;
   `describe --json` on a scratch repo with a member that exists, one that
   does not, and a `git:merge:main` task; `describe` with no
   `.config/bootstrap/` (code 2).

## Verification

- `mise x -- mise run test:go` passes.
- `grep -rn '7 \* 24\|168 \* time.Hour' cli/internal` finds only the one
  shared function.

## Guardrails

- Touch nothing outside `cli/`.
- Do not change the fetch of the other write commands.
- Delete with `rm`, never `git rm`.

## Commit

`feat: bootstrap catalog, bootstrap describe, and one fetch rule for the clone`
