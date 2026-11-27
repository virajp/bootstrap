# U1 — Engine and tests

- **Read first:** every owned file that an edit below names, top to bottom.
- **Lazy-load:** `docs/design/mac.md` ("The backup clone" :360, "The manifest"
  :432), `docs/design/cli.md` ("The manifest and the stale rule" :1599),
  `docs/design/tools.md` ("The `[files]` table" :206), `docs/design/cli.md`
  ("The files of `cli/`" for the golden-file rules).

## Ruling

> D1: A key root on a Mac [files] entry: "home" (the default) or "backup".
> Any other value stops the sync. On the repo side, a root key stops the sync.
> The table key stays a plain path; badRelPath also refuses a key whose first
> segment is @backup, on both sides. A backup entry whose first segment is
> dotfiles or .git is an error of the entry. User, 2026-11-27: "root =
> \"backup\" key".

> D2: A real file in the clone, at <backup>/<path> (the root of the clone, not
> below dotfiles/). The link rule (state.backup, linkPlan) ignores it: no link
> from $HOME.

> D3: The sync skips the entry, as a skipped target, with a log line that
> names the reason (no backup in macos.toml).

> D4: The Mac manifest row path of a backup target is @backup/<path>.
> manifest.BadPath accepts @backup only as the first segment of a Mac row and
> refuses it as any other segment (x/@backup/y); the repo side refuses (skips
> with a log line, as each bad row) a row whose first segment is @backup. The
> stale rule resolves @backup/ to the clone. When backup is unset, the run
> keeps each old @backup/ row in the new manifest unchanged and deletes
> nothing. When backup moves from clone A to clone B, a row resolves to B,
> and a file left in A stays: accepted.

> D5: requires of a backup entry is relative to the clone. The dry run and mac
> check print the full path in the clone.

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

> D13 (last edge rules): (a) U1 owns all of cli/internal/core/, so the
> accessor of D12 b can sit in core.go beside Home() and be loaded in
> state.go. (b) U4 keeps the golden lines of the new fixture tools.toml and of
> each new or changed fixture macos.toml of U1; any other changed line is a
> QUESTION. (c) The path that the dry run and mac check print for a backup
> item has the form of each other output line: ~/... when the clone is below
> $HOME, else the absolute path.

## Edits

Write all comments in ASD-STE100. `TestEngineNamesNoTool`
(`cli/internal/source/notools_test.go:107`) must stay green: name no tool.

1. **`cli/internal/source/inventory.go`** (`Source.entries` :182, the key
   checks about :212-225, `fe.Tmpl` :364) — read the optional `root` key of a
   `[files]` entry into a new field of the entry. Accept `"home"` and
   `"backup"` on the Mac side; any other value is an error of the entry, in
   the form of the other key errors. On the repo side, a `root` key is an
   unknown key (the error that each other unknown key gives). Add the key to
   the list of known keys. In `cli/internal/source/check.go` (`badRelPath`
   :179), refuse a path whose first segment is `@backup`. Refuse a `backup`
   entry whose first segment is `dotfiles` or `.git` (D1).
2. **`cli/internal/macos/plan.go`** (`makePlan` :23) — carry the root on each
   planned item. Two entries with the same target path and different roots are
   not a duplicate.
3. **`cli/internal/macos/file.go`** (`state.file` :39, `homeReal` :44,
   `requires` :55, the walk top :78, `DirPerm` :100) — add one function that
   returns the absolute target of an item: `s.home + "/" + target` for
   `home`, `<clone>/<target>` for `backup`, with the clone path that
   `state.backup` makes absolute (`backup.go:202-205`). Use it in each place
   that builds `s.home + "/" + it.target`, `requires` included (D5). When
   `backup` is unset, a `backup` item is skipped with the log line
   `skip file  <show> (no backup in macos.toml)` (D11 a).
4. **`cli/internal/macos/backup.go`** (`state.backup` :197, the targets of the
   plan :213, `linkPlan` :40) — leave each `backup` item out of the link
   targets (D2).
5. **`cli/internal/manifest/manifest.go`** (`BadPath` :63, used at :126) —
   on the Mac side, accept `@backup` only as the first segment of a row path,
   and refuse it as any other segment; on the repo side (`manifest.Repo`,
   read at `cli/internal/repo/repo.go:370`), refuse a row whose first segment
   is `@backup` (D4). Every other check stays.
6. **`cli/internal/macos/macos.go`** (the stale rule :631) — write the row of a
   `backup` item as `@backup/<target>`, and resolve such a row to the clone.
   With `backup` unset, copy each old `@backup/` row into the new manifest
   unchanged and delete nothing (D4).
7. The dry run and `mac check` output name the full clone path of a backup
   item (D5).
8. **Unit tests** — add cases beside the existing ones:
   `cli/internal/source/source_test.go` (the root key: home, backup, a bad
   value, a root key on the repo side), `cli/internal/macos/backup_test.go`
   (`TestLinkPlan`: a backup item is not linked),
   `cli/internal/manifest/` (a `@backup/` row is accepted, `x/@backup/y` is
   refused).
9. **`e2e/mac/sync_test.go`** (the backup cases :1344-1410, `useBackup`) — add
   cases with a fixture tool in `macos/fixtures/` that has one
   `root = "backup"` entry:
   - with `backup`: the file is a real file at `<clone>/<path>`, not a link,
     and the manifest has the `@backup/` row;
   - without `backup`: the entry is skipped with the D3 line, the run exits
     0, and an old `@backup/` row stays in the manifest;
   - a `backup` entry with the first segment `dotfiles` stops with the entry
     error;
   - the tool leaves the selection: the stale rule deletes the file in the
     clone;
   - the dry run prints the clone path.
     Add the case rows to the case table of `e2e/mac/mac_test.go` (D11 b).

## Verification

- `mise x -- go test ./...` in `cli/` is green.
- `mise x -- go test ./e2e/mac/ ./e2e/macostools/` is green from the root.
- The unit gate line `mise x -- mise run code:format --all` is green.

## Guardrails

- Do not touch any doc. Report each falsified passage as `DOCS FALSIFIED:`.
- Do not regenerate `cli/internal/source/testdata/toml-yq.golden`: U4 does
  (D11 g); `TestTOMLMatchesYq` may skip until then.
- Do not change a tool of `macos/tools/`: plan 3 adds the first real
  `root = "backup"` entries.
- Delete with `rm`, never `git rm`.

## Commit

`feat: a Mac [files] entry can write into the backup clone`
