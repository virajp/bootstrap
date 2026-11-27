# U4 — Gates

- **Read first:** `index.json` `gates`.

## Ruling

> The fixed gates unit: it runs the one generator of this plan (the golden
> file, D11 g), then the full wave gate and the plan gate lines, and reports
> each result. It bumps no version.

> D11 (edge rules): (a) The D3 skip line uses the existing skip form of
> file.go:61-66: `skip file  <show> (no backup in macos.toml)`, with the same
> outcome as a requires skip. (b) U1 adds its new case rows to the case table
> of e2e/mac/mac_test.go. (c) For a backup entry, the check of the first
> segment dotfiles is case-insensitive, and a .git segment at any depth, in
> any case, is refused. (d) The unknown-key message lists root on the Mac side
> only; the repo-side list does not change
> (cli/internal/selection/form_test.go:86
> stays green). (e) The seed list of a removed Mac tool
> (cli/internal/core/plan.go:1358-1360, cli/internal/tui/tools.go:889-893)
> resolves a backup seed with the same resolve function. (f) The
> handled-target set and the entry lookup (file.go:53, :66) key a backup item
> by its row path @backup/<target>. (g) U4 regenerates
> cli/internal/source/testdata/toml-yq.golden for the new fixture. (h) U3
> updates docs/setup/index.md (:74-75, :360-365) and
> docs/setup/migrate-dotfiles.md (:190-191). (i) For a backup entry, the
> requires message says a path relative to the backup clone.

> D13 (last edge rules): (a) U1 owns all of cli/internal/core/, so the
> accessor of D12 b can sit in core.go beside Home() and be loaded in
> state.go. (b) U4 keeps the golden lines of the new fixture tools.toml and of
> each new or changed fixture macos.toml of U1; any other changed line is a
> QUESTION. (c) The path that the dry run and mac check print for a backup
> item has the form of each other output line: ~/... when the clone is below
> $HOME, else the absolute path.

## Edits

1. **`cli/internal/source/testdata/toml-yq.golden`** — in `cli/`, run
   `UPDATE_GOLDEN=1 mise x -- go test ./internal/source`, and keep only the
   lines of the new fixture `tools.toml` and of each new or changed fixture
   `macos.toml` (D13 b); report any other changed line as a
   QUESTION.
2. Run each wave gate and plan gate line of `index.json`, and report each
   result.

## Verification

- `mise x -- mise run code:lint --all` is green.
- `env CLAUDE_PLUGINS_CLONE=$HOME/Projects/github.com/virajp/claude-plugins mise
  x -- mise run test`
  is green.

## Guardrails

- Bump no version. Edit no file but the golden. If a gate fails, report it as a
  QUESTION.

## Commit

`test: run the gates of the backup target` — the orchestrator makes no commit
when the unit changed nothing.
