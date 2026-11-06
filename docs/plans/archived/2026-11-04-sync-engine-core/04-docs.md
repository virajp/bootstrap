# U4 — Docs

- **Read first:** every passage that the list below names, before editing.
- **Lazy-load:** the diffs of U1 and U2 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned, the vwf
  documentation-standards skill.

## Ruling

From the design rulings of 2026-11-04:

> - The engine starts gomplate, jq and yq as `mise x <tool>@latest --` from
>   `/`: no pins (drop setup-sync's gomplate@5.3.1, jq@1.8.2, yq@4.54.1), no
>   `#MISE tools.gomplate` directive for the engine. Works before a first sync,
>   on a new Mac, in test scratch folders. The template data stays per
>   front-end; only the delimiters and the gomplate call are shared.

From this plan:

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires plan 1 only."

D4: "No. mise run setup:configs in bootstrap and in each consumer stays a hand
step (Out of scope), as in earlier plans."

D7: "repo_sync no longer checks command -v gomplate; it checks command -v mise
and returns 2 without it. The gate code:lint:repo-sync keeps its mise x --
which gomplate check, so a commit without gomplate stays silent, and the
[mise.tools.gomplate] table of tools/repo-sync/tools.toml stays. The #MISE
tools.pre-commit line of setup/configs/sync stays."

D13: "No. The parse check reads only the render of the files of bootstrap, and
a consumer changes nothing. The docs unit still describes the new code 2."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`CLAUDE.md`**:
     - "Layout rule": `sync-engine/` holds `clone`, `common`, `write`,
       `select` and `repo-sync`.
     - "How the sync operates": the parse check (jq for JSON, yq for YAML
       and TOML, the JSONC rule of D6, code 2, the old file stays); the
       exit-codes bullet adds it to code 2; the sync needs `mise` and starts
       gomplate, jq and yq with `mise x <tool>@latest --` from `/`.
     - "Task scripts": the exception bullet on `#MISE tools.<tool>="latest"`
       no longer names `gomplate` in `setup:configs:sync`; it keeps
       `pre-commit` there. Add one sentence: the engine in `sync-engine/`
       calls gomplate, jq and yq only through its three wrappers.
   - **`docs/sync/tools/repo-sync.md`**: the mise tools table (gomplate is for
     the gate `code:lint:repo-sync`); the list of the steps of `repo_sync`
     (the parse check).
   - **`docs/sync/index.md`**: the table of exit codes (code 2 adds a target
     that does not parse).
   - **`docs/design/*.md`** and **`README.md`**: each passage that
     `vwf:docs-sync` finds on the gomplate directive or on the write steps.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply each finding and each `DOCS FALSIFIED:` line of the earlier units
   that the list does not cover, inside the owned files. Add no
   `docs/sync/migrate.md` item (D13).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'parse' CLAUDE.md docs/sync/tools/repo-sync.md docs/sync/index.md`
  finds the new passages.
- `grep -n 'gomplate' CLAUDE.md` has no passage that says
  `setup:configs:sync` gets gomplate from a `#MISE` line.
- `mise x -- dprint check CLAUDE.md README.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Do not edit `docs/plans/**`.
- Write identifiers in code spans.
- Each new paragraph obeys the style of the file it is in (ASD-STE100, short
  sentences).
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the shared engine libraries and the parse check` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
