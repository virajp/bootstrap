# U2 — Write docs/sync/vwf.md, the entry point for vwf and stackgen

- **Wave:** 1
- **Depends on:** —
- **Owns:** `docs/sync/vwf.md` (new)
- **Model:** opus
- **Kind:** edit
- **Read first:** `docs/sync/index.md` (whole file: the adoption steps, the
  `tools.toml` keys, the values file keys) and `docs/sync/flutter.md` (the
  structure of a tool file, especially "What the tool ships" with its Kind
  column `Synced`, `Seed`, `Generated`).
- **Lazy-load:** `README.md` (the sync contract), `CLAUDE.md` ("How the sync
  operates"), `tools/base/.config/mise/tasks/_scripts/repo-sync:8-110` (exit
  codes and the environment variables).

## Ruling

From index.md, assumed decisions:

> D7 — `docs/sync/vwf.md` accepts `bootstrap: add tools [<tool>, …]` and
> `bootstrap: remove tools [<tool>, …]`, each with `reason: <why>`. For each
> tool it follows `docs/sync/<tool>.md`. A plugin never edits
> `.config/bootstrap.toml` itself; the agent that reads `vwf.md` does.

> D8 — The agent removes the tool from `tools`, deletes each file that
> `docs/sync/<tool>.md` lists as Synced (they carry the bootstrap header),
> keeps each Seed (the repository owns it), then runs
> `mise run setup:configs`, which removes the stale generated mise files.

> D9 — The instruction is the consent: the plugin asked the user before it
> sent it. The agent applies the change, runs `mise run setup:configs`, shows
> the changed files and the reason, and does not commit. The caller commits.

> D10 — The agent stops and changes nothing when: the repository has no
> `.config/bootstrap.toml` (it says the repository is not adopted, and that
> `docs/sync/index.md` is the adoption); a tool name is not a folder in
> `tools/` (it names the tool); `remove` names a tool that another selected
> tool needs through `depends` or a mise `depends` (it names that tool). `add`
> of a tool that is already in `tools` changes nothing for that tool. `remove`
> of a tool that is not in `tools` changes nothing for that tool.

> D11 — Not in this plan. `vwf.md` documents only `add tools`,
> `remove tools` and `reason:`.

> D12 — `docs/sync/<tool>.md` and `docs/sync/vwf.md` are shipped instructions
> that agents run, so the edit units own them. The docs unit owns
> `README.md`, `CLAUDE.md`, `docs/sync/index.md` and `docs/sync/migrate.md`.

The contract that `virajp/claude-plugins` expects (index.md, Facts): a plugin
shallow-clones bootstrap (`git clone --depth 1
git@github.com:virajp/bootstrap.git` into a `mktemp -d` folder), reads
`docs/sync/vwf.md` in the clone, and gives it instructions. The plugins name
this path and no other bootstrap path. The materializer's reason string is
`stackgen pack <type>/<slug> pinned`.

## Edits

1. **`docs/sync/vwf.md`** (new) — a Markdown file in STE, in the style of the
   other files of `docs/sync/`, with these sections:
   1. Title and purpose: the one entry point of `bootstrap` for the vwf and
      stackgen plugins. A plugin reads only this file; this file sends the
      agent to the other files of `docs/sync/`.
   2. **How a plugin uses this file** — the shallow clone into a temporary
      folder, read `docs/sync/vwf.md` in the clone, give the instructions,
      remove the clone after. The plugin never edits
      `.config/bootstrap.toml` itself.
   3. **Instructions** — the exact format, in a `text` code block:

      ```text
      bootstrap: add tools [<tool>, …]
      reason: <why>
      ```

      and the same with `remove tools`. One message can hold several
      instructions; each has its own `reason:` line. Give the example reason
      `stackgen pack app-framework/flutter pinned`.
   4. **Before each instruction** — the D10 stop conditions, each with the
      message the agent gives. The tool names are the folders of `tools/` in
      the clone.
   5. **Add tools** — steps: add each name to `tools` in
      `.config/bootstrap.toml` (sorted, no duplicates); for each added tool and
      each tool its `depends` adds, read `docs/sync/<tool>.md` in the clone
      and do its "Adapt to the repository" and "Migration" items; then run
      `mise run setup:configs`.
   6. **Remove tools** — the D8 steps, in order, and the D10 check first.
      Say that a tool another selected tool needs cannot be removed alone.
   7. **After the change** — the D9 rule: show the changed files
      (`git status --short`) and the reason to the caller; do not commit; the
      caller commits.
2. Do not document `bootstrap: adopt` or any other instruction (D11).
3. Do not add the bootstrap sync header: `docs/` is for this repository only
   and is never synced.

## Verification

- Wave-gate lines 1 and 2 (index.md, Wave gate) pass.
- `grep -c 'bootstrap: add tools' docs/sync/vwf.md` and
  `grep -c 'bootstrap: remove tools' docs/sync/vwf.md` each print 1 or more.
- `grep -c 'reason:' docs/sync/vwf.md` prints 1 or more.
- `grep -n 'bootstrap: adopt' docs/sync/vwf.md` prints nothing.

## Guardrails

- Do not touch `docs/sync/index.md`, `README.md` or `CLAUDE.md` — U3 owns
  them (U3 adds `vwf.md` to the files table of `index.md` and to the layout
  rule of `CLAUDE.md`). Report what they need as `DOCS FALSIFIED:` lines.
- Do not name a specific tool's files in `vwf.md`; refer to them through
  `docs/sync/<tool>.md`, because U1 writes the new tool files in the same
  wave.
- Delete with `rm`, never `git rm`.

## Commit

`docs(sync): add vwf.md, the entry point for vwf and stackgen`
