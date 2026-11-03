# U2 — Short script headers in sync-engine

- **Read first:** each file of `sync-engine/`, top to bottom, and the
  design doc that D3 names for it.
- **Lazy-load:** the other files of `docs/design/`.

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"Contract docs: one place for each rule. [...] Script headers keep usage,
flags, a pointer (the 300-line setup-sync header contract goes). [...]"

From this plan:

D5: "The shebang when the file has one, one line on the job of the file, the
usage (the functions that other files call, or the entry function with its
flags), the environment variables that the file reads, and one pointer line
"Contract: docs/design/<file>.md". A file outside the mapping points to
docs/design/sync-engine.md. A file that is not bash skips the bash -n check
and the pointer line. The return codes and the long contract text go to the
design doc. Where the design doc lacks a rule, the long header lines stay in
this plan, and U2 returns a GAP for each. Comments inside a function body
stay."

D3: "sync-engine.md: the rules that both sides share (cache, clone, lock,
engine variables, tool calls, templates, the write order, the parse check,
merge mode, the header and its detection, the manifest and the stale rule,
the values merge, the shared exit codes 2, 3 and 5, bootstrap_commit).
repo-sync.md: the repo front-end; it links to sync-engine.md for the exit
codes. setup-sync.md: the Mac front-end; it describes exit 4 and links to
sync-engine.md for the other codes. packs.md: the packs model. tools.md: the
tools.toml schema, the [files] keys and the rule for the header of a source
file. mise.md: the mise generator, environments, [overrides.mise*], zz-repo
and zz-macos. bootstrap-config.md: the keys of .config/bootstrap.toml.
fnox.md: no change. Each other doc links to the one place."

D1: "No. The plan changes docs and comments only; no runnable line changes.
U2's Verification proves that the diff of sync-engine/ touches comment lines
only."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-macos-tools-apps."

## Edits

1. **Each file of `sync-engine/`** (`clone`, `common`, `write`, `select`,
   `merge`, `manifest`, `values`, `mise`, `repo-sync`, `setup-sync`, and
   any other file there) — cut the header comment to the parts of D5. The
   pointer line names: `sync-engine.md` for `clone`, `common`, `write`,
   `merge`, `manifest` and `values`; `packs.md` for `select`; `mise.md`
   for `mise`; `repo-sync.md` for `repo-sync`; `setup-sync.md` for
   `setup-sync`; `sync-engine.md` for any other file (D5). A file that is
   not bash gets no pointer line.
2. **Before each cut**, find each rule of the removed text in the named
   design doc (or in the doc that D3 names for that rule). When a rule is
   missing there, keep its comment lines, and return one
   `GAP: <file> keeps the header rule <rule>; <design doc> lacks it` for
   each (D5).
3. Do not change a comment inside a function body, and do not change a code
   line.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n` passes on each bash file of `sync-engine/` (a file that is not
  bash skips it, D5).
- The diff touches comment lines only:
  `git diff -U0 -- sync-engine | grep -E '^[-+][^-+]' | grep -vE
  '^[-+][[:space:]]*#'`
  prints nothing.
- Each bash file of `sync-engine/` has one line that matches
  `Contract: docs/design/`.
- `mise x -- mise run test` passes.

## Guardrails

- Touch only `sync-engine/**`. Do not touch `docs/design/`; U1 owns it.
- Do not touch the forwarding file
  `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`.
- Delete with `rm`, never `git rm`.

## Commit

`docs: cut the sync-engine script headers to usage and a pointer` —
written by the orchestrator after the unit gate and the wave gate, not by
the unit.
