# U2 — Clone and entry tasks

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `repo/tools/repo-sync/.config/mise/tasks/setup/configs/sync`.
  Read it only.

## Ruling

D2: "A verbose flag. --verbose of setup:configs and updateall:configs sets
SYNC_ENGINE_VERBOSE=1. The variable also works when a user sets it, and only the
value 1 turns it on. With it, the engine prints each current per-file line in
its
current format, then the table. The test tasks set the variable. Each entry task
consumes the flag and does not pass it to the engine. Both tasks add it to

`#USAGE`, and the Mac task also adds it to usage(). The user chose this."

D14: "A file that does not exist locally is skipped. A difference in mode counts
as a difference. If a file was copied, the line is Updated. If no file was
copied, the line is up to date. If neither file exists locally, no line prints.
A copied file counts as unchanged in the file loop, so the last line can read no
file changed. The Updated line in the same output reports the copy: accepted.
The copy runs before the sync validates anything, so a run that then stops
leaves
the two fresh files: accepted."

D16, for the copy step: "The copy step skips a file that the clone copy does not
have, and skips a local symlink. A failed copy prints a warning. It prints
Updated
if another file was copied, and otherwise no second line. Up to date prints if
no
file was copied and no copy failed. A dry run copies nothing, so a stale sync or
helpers shows as would change: accepted."

D15: "sync-engine/setup-sync:822-823 (using bootstrap, tools:) become one info
line: bootstrap <short commit> · <n> tools. Each call of log() in the Mac entry
task is an info line, and a call that reports a failure is an error line. The
printf lines at configs/sync:19 and :24 follow D7 too."

D5: "Before each run of setup:configs:sync, for the root and for each member,
setup:configs compares the local setup/configs/sync and \_scripts/helpers with
the same files in the clone copy, and copies a file that differs, with its mode.
It copies only over a file that exists. Then it prints the info line 'Updated
the local sync task to <short commit>' or 'The local sync task is up to date',
below the line 'Updating the clone'. A dry run copies nothing and prints 'Would
update the local sync task to <short commit>'. If the copy fails, the task
prints a warning and continues with the local files."

D6: "'Removed the old cache <path>' prints only if the old folder existed before
the removal and does not exist after it. A run with only the old lock prints
nothing."

D7: "It gets no local-task line, because updateall:configs already runs sync
from the clone copy. On a terminal, its own log lines and the fallback of
\_se_say show the info glyph or the error glyph and no prefix. Otherwise they
keep the prefix. 'setup-sync: running ...' prints only with verbose."

D17: "In a dry run, Would update the local sync task prints only if a file
differs. Otherwise the line is The local sync task is up to date. A file skipped
as a symlink does not count as a difference. A non-zero table row takes the
color
of its kind in the list. Unchanged, Kept seeds and Skipped have no color, and
Total has no color. A row with a zero count is dim. log() and the two printf
lines of the Mac entry use the same style rule as the engine, not a plain
terminal test. The user accepted this."

## Edits

1. **`sync-engine/clone`** — In `_se_legacy_cleanup`, record whether the old
   folder exists before the `rm`. Print the old-cache line only if it existed
   and is gone now (D6). In the fallback of `_se_say`, use the glyphs on a
   terminal, with the same color rule as `style_enabled` (D7). `clone` cannot
   use `sync-engine/common`, because the task embeds `clone` alone.
2. **`repo/tools/repo-sync/.config/mise/tasks/setup/configs/_default.tmpl`** —
   - Add `#USAGE flag "--verbose"`. When it is set, export
     `SYNC_ENGINE_VERBOSE=1` before the sync runs.
   - Add a function that does the copy of D5 for one repository root and prints
     its info line. Call it before the root run (:46) and before each member run
     in `sync_member` (:59), after the guard at :55.
   - The clone copy paths are in the facts of `index.md`.
3. **`macos/tools/macos-setup/.config/mise/tasks/updateall/configs/_default.tmpl`**
   — Add `--verbose` to `#USAGE` (:6-7), to `usage()` (:19) and to the flags at
   :33-50. It exports `SYNC_ENGINE_VERBOSE=1` and removes the flag from the
   arguments that it passes at :125, because `setup_sync` returns 2 on an
   unknown
   flag. Change `log()` at :15 as D7 and D15 tell.
4. **`macos/tools/macos-setup/.config/mise/tasks/updateall/configs/sync`** —
   Print the line at :12 only when `SYNC_ENGINE_VERBOSE` is `1`. Change the
   `printf` lines at :19 and :24 as D15 tells.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- The copy has no effect in a dry run. Read the code path for `--dry-run` to
  make sure.
- The tests are the work of U3. They can fail at this stage.

## Guardrails

- Do not touch the synced copies in `.config/mise/tasks/setup/configs/`: the
  sync
  writes them.
- Do not touch `repo/tools/repo-sync/.config/mise/tasks/setup/configs/sync`. The
  environment variable reaches it through `mise run`.
- Do not touch the engine files that U1 owns.
- Keep the header comment of each file.
- Write for Bash 3.2 and 5.x. The templates use the delimiters `<% %>`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: refresh the local sync task and report the old cache only when removed`
