# U3 — Docs

- **Read first:** the owned docs at the sections named below.
- **Lazy-load:** `${CLAUDE_PLUGIN_ROOT}/skills/docs-sync/SKILL.md`.

## Ruling

> D1: A key root on a Mac [files] entry: "home" (the default) or "backup".
> Any other value stops the sync. On the repo side, a root key stops the sync.

> D2: A real file in the clone, at <backup>/<path> (the root of the clone, not
> below dotfiles/). The link rule ignores it: no link from $HOME.

> D3: The sync skips the entry, as a skipped target, with a log line that
> names the reason (no backup in macos.toml).

> D4: The Mac manifest row path of a backup target is @backup/<path>. The
> stale rule resolves @backup/ to the clone. When backup is unset, the run
> keeps each old @backup/ row and deletes nothing. When backup moves from
> clone A to clone B, a row resolves to B, and a file left in A stays:
> accepted, and the docs say so.

> D5: requires of a backup entry is relative to the clone. The dry run and mac
> check print the full path in the clone.

> D6: bootstrap never commits and never runs git in the clone. The user
> commits the new files, as with the backup links today.

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

> D12 (more edge rules): (a) A root = "backup" entry accepts mode copy,
> template and seed; mode merge with root backup is an error of the entry, so
> no contribution item can target it. (b) core and the TUI get the clone from
> a new accessor beside Home() (for example Backup(), the absolute clone path,
> or "" when backup is unset); a backup seed is left out of the seed list when
> backup is unset or the clone folder is missing. (c) For a backup item, the
> walk top of file.go:75-80 is the real path of the clone, so the walk never
> reaches a parent folder of the clone.

## Edits

Write all prose in ASD-STE100.

1. Run `/vwf:docs-sync` over the branch delta of the run, and apply its
   findings and every `DOCS FALSIFIED:` line of U1.
2. **`docs/design/tools.md`** ("The `[files]` table" :206) — add the `root`
   row to the key table (side: Mac), with D1 and D5.
3. **`docs/design/mac.md`** — "The backup clone" (:360): a `root = "backup"`
   entry and D2, D3, D6; "The manifest" (:432): the `@backup/` row (D4);
   "Where the Mac side writes" (:446): the clone root gets the files of
   backup entries.
4. **`docs/setup/index.md`** and **`docs/setup/migrate-dotfiles.md`** — per D11
   h.
5. **`docs/design/cli.md`** ("The manifest and the stale rule" :1599) — the
   `@backup/` row and its stale rule (D4).

## Verification

- `mise x -- mise run code:format --all` is green.
- `grep -n 'root' docs/design/tools.md` shows the new row.

## Guardrails

- Do not touch any file outside `owns`. Do not edit `docs/plans/archived/`,
  `docs/scratchpad/` or `docs/memory/`.

## Commit

`docs: the root key of a Mac [files] entry`
