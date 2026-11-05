# U3 — Tests

- **Read first:** the helpers of each owned file: `case_body` (:111),
  `case_rerun` (:1075) and `CASE_CALL` in `test/repo-sync`, `sync_run` (:164)
  and
  the entry runs (:1265, :1276, :1389) in `test/setup-sync`, and :111 in
  `test/macos-tools`.
- **Lazy-load:** the rest of each file, only where a case needs it.

## Ruling

D2: "A verbose flag. --verbose of setup:configs and updateall:configs sets
SYNC_ENGINE_VERBOSE=1. The variable also works when a user sets it, and only the
value 1 turns it on. With it, the engine prints each current per-file line in
its
current format, then the table. The test tasks set the variable. Each entry task
consumes the flag and does not pass it to the engine. Both tasks add it to

`#USAGE`, and the Mac task also adds it to usage(). The user chose this."

D5, D6, D8, D9, D11, D12, D13, D14 and D16 are the behaviour that the new cases
test. Read them
in `index.json`, with `plan.mjs get`.

D13: "One rule turns glyphs and colors on and off together: the rule of D3, so
NO_COLOR on a terminal also removes the glyphs. Style on: an info line is ℹ️
<text> with no color, and the table has the box characters of the sample in
index.md, with a left label column and a right count column. Style off: a list
line is <word> <path>, an info line is [INFO] <text>, a last line is [OK] <text>
or [ERROR] <text>, and the table has ASCII borders (+, -, |). The new lines have
no sync-engine: or setup-sync: prefix."

## Edits

1. **Each owned file** — Export `SYNC_ENGINE_VERBOSE=1` near the top of the
   task,
   so each existing case keeps its output. Make sure that no case clears the
   environment and loses the variable.
2. **`.config/mise/tasks/test/repo-sync`** — Add these cases. Each new case
   unsets
   the variable when it tests the default output:
   - The default output lists a changed file and a new file, does not list an
     unchanged file, and has the table with correct counts.
   - A new seed is `seeded`, and an existing seed counts in `Kept seeds`.
   - A failed file shows in the list, and the exit code is 2.
   - A dry run writes nothing, and it shows a `would change` line.
   - With output that is not a terminal, the output has no glyph and no escape
     code. With `FORCE_COLOR=1`, it has both.
   - The rendered `_default`: a differing local `sync` is replaced and the line
     `Updated the local sync task` prints. Equal files print
     `The local sync task is up to date`. A dry run copies nothing.
   - The old-cache line prints when the old folder exists, and does not print
     when only the old lock exists.
3. **`.config/mise/tasks/test/setup-sync`** — Add a case for the default output
   of `setup_sync`, with the list and the table. Add a case for
   `updateall:configs --verbose`, which shows the per-file lines.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test` passes.
- `grep -c 'SYNC_ENGINE_VERBOSE' ` of each owned file is at least 1.

## Guardrails

- Do not change an engine file or an entry task. A failing test that shows a bug
  in U1 or U2 is a `GAP:` line with the failing case.
- Keep each new case in its own folder, as the other cases do.
- Delete with `rm`, never `git rm`.

## Commit

`test: test the changed-only output, the table and the local sync task`
