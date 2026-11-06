# U1 — Engine output

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `repo/tools/mise/.config/mise/tasks/_scripts/helpers`, for
  `style_enabled` and the colors. Read it only. Do not source it.

## Ruling

D1: "Both front-ends: repo_sync and setup_sync get the list of changes and the
table. The user chose this."

D2: "A verbose flag. --verbose of setup:configs and updateall:configs sets
SYNC_ENGINE_VERBOSE=1. The variable also works when a user sets it, and only the
value 1 turns it on. With it, the engine prints each current per-file line in
its
current format, then the table. The test tasks set the variable. Each entry task
consumes the flag and does not pass it to the engine. Both tasks add it to

`#USAGE`, and the Mac task also adds it to usage(). The user chose this."

D16: "Failed counts files. Each failed target gets exactly one failed outcome
and
one list line, from its first failure message; this covers each failure message
of repo-sync and setup-sync, also where no existing file stays. A target that
fails with no message today (repo-sync:1467, :1473, :1589, write:90) also gets a
failed outcome, with the text cannot write the target. A later kept existing
line
for the same target prints only with verbose. A failed list line is ✗ failed
<path>: <text> with style on, and failed <path>: <text> with style off. <path>
is
the target path, and the text does not repeat it. The line FAILED: <list> stays
by default in both front-ends, before the table, and the line <ok>/<total> files
in sync goes. The last line uses 1 file changed and 1 file failed for one file,
and it omits the count with no failed file: The sync failed with code <rc>. With
an empty commit, the Mac info line reads bootstrap an unknown commit · <n>
tools.
With verbose, setup-sync keeps its lines using bootstrap and tools:. By default
the bootstrap_commit line is Set bootstrap_commit to <short commit>, or Would
set
bootstrap_commit to <short commit> in a dry run, as an info line before the
table."

D3: "The list, the table and the color rule are in sync-engine/common. The color
rule is the same as style_enabled of helpers: FORCE_COLOR turns it on, NO_COLOR
turns it off, otherwise it needs a terminal and TERM not dumb. The engine does
not call helpers."

D4: "One function, \_se_outcome <kind> <path> [text], replaces each per-file log
line. It increments the count of the kind and prints the correct line. The kinds
are changed, new, seeded, removed, unchanged, kept, skipped and failed. In Bash
3.2, each count is a plain variable."

D8: "The list shows only changed, new, seeded, removed and failed, each with its
glyph and color: changed ✎ yellow, new ✚ green, seeded ❋ blue, removed ✖ red,
failed ✗ red. The table has the rows Changed, New, Seeded, Removed, Unchanged
(=), Kept seeds (◦), Skipped (⊘), Failed and Total, in that order, and always
shows all rows. A row with a zero count is dim. Total is the sum of the rows.
Each run prints one table, so --all prints one table per member. The line 'using
bootstrap', the line 'tools:' and the line 'environments:' become one info line
by default. The other messages of \_se_log stay. Manifest lines 'kept <path>
(link)' and 'kept <path> (another repository)' have no count and print only with
verbose. A manifest write counts as changed or new. 'set bootstrap_commit'
becomes an info line. D12 gives the last line and when the table prints. D13
gives the forms of each line."

D11: "A line on a failure path, which includes the symlink lines, kept existing
and cannot remove, is failed, with its current message as the text. The kind
kept
holds only kept seed, and the row Kept seeds counts only that kind. A failure of
the mise config of a tool or of the manifest file has no count, and its message
stays."

D12: "\_se_summary <rc> prints the table and the last line once, as the last
output of the run: after the stale rule, the manifest write and
bootstrap_commit. Each return after the file loop calls it, also on a failure. A
run that stops before the file loop has no table. On the Mac side, it prints
after mise bootstrap, with one table across the init and main stages, and return
4 after the init stage has no table. n is the sum of Changed, New, Seeded and
Removed. With rc 0: ✅ Synced: <n> files changed, or ✅ Synced: no file changed.
In a dry run with rc 0: ✅ Dry run: <n> files would change, or ✅ Dry run: no
file would change. With another rc: a red ❌ The sync failed with code <rc>:
<failed count> files failed."

D13: "One rule turns glyphs and colors on and off together: the rule of D3, so
NO_COLOR on a terminal also removes the glyphs. Style on: an info line is ℹ️
<text> with no color, and the table has the box characters of the sample in
index.md, with a left label column and a right count column. Style off: a list
line is <word> <path>, an info line is [INFO] <text>, a last line is [OK] <text>
or [ERROR] <text>, and the table has ASCII borders (+, -, |). The new lines have
no sync-engine: or setup-sync: prefix."

D15: "sync-engine/setup-sync:822-823 (using bootstrap, tools:) become one info
line: bootstrap <short commit> · <n> tools. Each call of log() in the Mac entry
task is an info line, and a call that reports a failure is an error line. The
printf lines at configs/sync:19 and :24 follow D7 too."

D9: "The list has 'would change', 'would add', 'would seed' and 'would remove'
lines. The diff of a change stays. The table counts what the run would do."

D17: "In a dry run, Would update the local sync task prints only if a file
differs. Otherwise the line is The local sync task is up to date. A file skipped
as a symlink does not count as a difference. A non-zero table row takes the
color
of its kind in the list. Unchanged, Kept seeds and Skipped have no color, and
Total has no color. A row with a zero count is dim. log() and the two printf
lines of the Mac entry use the same style rule as the engine, not a plain
terminal test. The user accepted this."

## Edits

1. **`sync-engine/common`** — Add `_se_outcome`, the color rule, the table
   printer (`_se_summary`) and the info printer. All of this output goes to
   stderr, as `_se_log` does. Read `SYNC_ENGINE_VERBOSE` here. Make the
   `bootstrap_commit` lines (:93, :140) info lines by default, with the text of
   D16, and keep their current text with verbose.
2. **`sync-engine/write`** — In `_se_finish` and the folder-path removal, call
   `_se_outcome`. The kind is `new` when the target did not exist, `seeded` for
   a new seed, `changed` for an update or a chmod, and `unchanged` for no
   difference. Keep the diff of a dry run. The parse failure is `failed`, with
   the current message as the text.
3. **`sync-engine/merge`** — Use `failed` for each failure message, and let
   `_se_finish` report the result of a merge.
4. **`sync-engine/manifest`** — A stale removal is `removed`, and `cannot
   remove`
   is `failed`. The manifest write is `changed` or `new`. The two `kept` lines
   have no count and print only with verbose.
5. **`sync-engine/repo-sync`** — Use `kept` for `kept seed`, and `failed` for
   `kept existing` and the symlink lines (D11). Use `skipped` for each `skip`
   line. Replace the lines at 1739, 1754 and 1755 with one info line, for
   example
   `bootstrap 5961aaf · 14 tools · environments: dev ci`. With verbose, keep the
   three lines. Call `_se_summary` as D12 tells. With verbose, keep the end
   lines
   at 1892-1896 and then print the table. By default, the table and its last
   line
   replace these lines, except the `FAILED:` line at :1895, which stays (D16).
6. **`sync-engine/setup-sync`** — Make the same changes at :555, :562-565, :631,
   and at the end lines :745, :835 and :885. The `FAILED:` lines at :744 and
   :831
   stay (D16). Change :822-823 as D15 tells.
   Call `_se_summary` as D12 tells.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `_SE_OUTCOME`-style counts are plain variables. The code has no associative
  array, no nameref and no `${var,,}`.
- With `SYNC_ENGINE_VERBOSE=1`, each current per-file string is still printed.
  Examine this with `git diff`: the old strings stay in the verbose path.
- The tests are the work of U3. They can fail at this stage.

## Guardrails

- Do not touch `sync-engine/clone` or an entry task: U2 owns them.
- Do not change an exit code, or the order of the stale rule, the manifest and
  `bootstrap_commit`.
- Do not change any message that is not a per-file line, an end line or one of
  the lines this file names.
- Write for Bash 3.2 and 5.x. Under `set -e`, a count increment must not fail:
  do not use `((x++))`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: list only changed files and print a table of counts in the sync`
