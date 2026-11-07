# U2 — Tests: the manifest and the stale rule

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `sync-engine/manifest`, `sync-engine/repo-sync` (U1's
  output).

## Ruling

From the design rulings of 2026-11-04:

> - Stale rule on both sides: remove each file with the header that this run
>   does not write, then its empty parent folders up to the target root; a
>   file without the header is never removed. One manifest format:
>   `<tool><TAB><path>`, sorted, first line `# Written by bootstrap. Do not
>   edit this file.` (reader skips `#` lines), written only after a clean
>   run. Mac: `~/.cache/sync-engine/macos-generated.tsv`. Repo (user's
>   proposal): `.config/repo-generated.tsv`, committed (moves with the branch;
>   a cache copy breaks with clones, worktrees, other machines). Each
>   submodule has its own. The first repo sync has no manifest and removes
>   nothing; `migrate.md` lists old leftovers. No `git grep` header scan.

D3: "This plan extends test:repo-sync with cases for its new rules. An
existing case changes only where a ruling changes its expected result."

D5: "Nothing. The first sync has no manifest and removes nothing.
docs/sync/migrate.md lists the old leftovers, and the owner deletes them by
hand as before."

D6: "No. The stale removal and the manifest write run only after a run
without a failure, in this order: stale removal, manifest write,
bootstrap_commit. A failed removal is a failure, so the manifest is not
written and the next clean run tries again."

D7: "The manifest is a committed file of the consumer, so the engine does not
trust it. A row without a TAB, with an empty tool or path, with an absolute
path, or with a .. segment is skipped with a log line and never removed. A
listed path that is a symlink, or that has a symlinked folder between the
target root and the file, is not removed (repo-sync never follows a link)."

D8: "A dry run prints would remove <path> for each stale file and would write
.config/repo-generated.tsv when the new content differs from the file. It
writes and removes nothing. A real run writes the manifest only when its
content changes."

D10: "One row <tool><TAB><path> for each file of mode generated that the run
writes or finds unchanged, and for each generated conf.d mise file. A seed, a
merge file, a skipped file (an environment that the repo does not use, a
requires or requires_found file that is not present) and a removed merge task
have no row. The path is relative to MISE_PROJECT_ROOT, without .tmpl."

D11: "Every path that this run handled: each generated, seed and merge file
that the run wrote or kept. Only a listed path that the run did not handle is
a candidate for removal."

D12: "The stale rule skips the paths in _RS_SKIP, because _rs_remove already
removes them. A dry run logs one would remove line for each file."

D13: "LF line endings, no trailing blanks, and exactly one LF at the end, so
the pre-commit hooks never rewrite it."

D14: "sync-engine/manifest takes a link mode. repo: a path that is a symlink,
or under a symlinked folder, is never removed (D7). macos: the stale rule
resolves links and removes the real file, then prunes empty parents up to the
target root ($HOME), never past a symlinked folder (the rule of setup-sync),
because on the Mac ~/.config/mise is a link into the dotfiles/ of the owner,
and the mise config of a deselected tool must go."

## Edits

1. **`.config/mise/tasks/test/repo-sync`** — add a helper that writes a
   manifest into the consumer of a case (`$consumer/.config/repo-generated.tsv`,
   the first line, then the given rows), and a helper that writes a file with
   the bootstrap header. Add these cases (dry run unless the case says real
   run, `case_args=' '`):
   1. No manifest: a header file `.config/mise/conf.d/oldtool/mise.dev.toml`
      stays; the log has no `would remove` (D5).
   2. A manifest lists that header file under `oldtool`: the log has
      `would remove <consumer>/.config/mise/conf.d/oldtool/mise.dev.toml`.
   3. A manifest lists a file without the header: no `would remove` for it.
   4. A manifest lists a file that the run writes: no `would remove` for it.
   5. A manifest with a `..` row, an absolute row and a row without a TAB:
      code 0, a `skipped` line for each, no `would remove` for them (D7).
   6. A listed header file under a symlinked folder of the consumer: no
      removal, the log has `(link)` (D7).
   7. Real run: the manifest is written; its first line is the constant; its
      rows are sorted (`LC_ALL=C sort -c` on the rows passes); it lists a
      generated conf.d file; it lists no seed (for example no
      `.config/mise/conf.d/zz-repo/` row) (D10).
   8. Real run with a listed stale header file in `a/b/stale.txt`: the file
      and the empty folders `a/b` and `a` are gone; a folder that holds
      another file stays.
   9. A run with a failure (for example a template that fails): code not 0,
      no `would remove`, and in a real run the old manifest is unchanged (D6).
   10. A manifest lists a seed that the consumer has (for example
       `.config/mise/conf.d/zz-repo/mise.dev.toml` with the header added by
       the case): no `would remove` for it (D11).
   11. With only `main` in the consumer, a manifest lists
       `.config/mise/tasks/code/merge/develop` with the header: exactly one
       `would remove` line for it (D12).
   12. Real run: the manifest ends with exactly one LF and has no CR and no
       trailing blank (`tail -c 1` is a newline, `grep -c $'\r'` is 0,
       `grep -c ' $'` is 0) (D13).
   13. The macos link mode, a direct call: in a subshell, source
       `sync-engine/common` and `sync-engine/manifest` of the scratch clone;
       make `$tmp/home/.config/mise` a link to `$tmp/dotfiles/mise`, and a
       header file `$tmp/dotfiles/mise/conf.d/old/mise.toml`; call the stale
       function with mode `macos`, a row for
       `.config/mise/conf.d/old/mise.toml`,
       no handled path and root `$tmp/home`. The real file and the empty
       `old` folder are gone; the link `$tmp/home/.config/mise` stays (D14).
   14. Dry run with a manifest that differs: the log has
       `would write <consumer>/.config/repo-generated.tsv`; the file is
       unchanged (D8).
2. Change the existing requires case that expects the removal of a synced
   gate (`files, no pyproject.toml: a synced gate is removed`, renamed by
   P03): its change function also writes a manifest that lists the gate, and
   the case expects `would remove <path>`. Add one case without a manifest:
   the gate stays and the log has `skip file`. No other existing case
   changes its expected result (D3).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes, every old and new case.
- `grep -c 'repo-generated.tsv' .config/mise/tasks/test/repo-sync` is at
  least 8.

## Guardrails

- Do not touch `sync-engine/**`; U1 owns it. Report an engine bug as
  `UNRESOLVED:` with the case.
- A real-run case writes only inside its case folder.
- Write shell for bash 3.2 and bash 5.x. Under `set -e`, write a command
  substitution that can fail as `VAR=$(cmd) || VAR=''`.
- Delete with `rm`, never `git rm`.

## Commit

`test: cover the repo manifest and the stale rule` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
