# U5 — Docs

- **Read first:** every passage that the list below names, at the named
  lines, before editing.
- **Lazy-load:** the diffs of U1 to U3 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned, the vwf
  documentation-standards skill.

## Ruling

From the design rulings of 2026-11-04:

> - The shared cache folder is `~/.cache/sync-engine/` (`repo/`, `lock/`,
>   `macos-generated.tsv`), not `~/.cache/bootstrap/` and not `bootstrap2`.
>   **Why:** an old consumer stub runs once more after the update; on a new
>   layout in `~/.cache/bootstrap` it does `rm -rf` and deletes the manifest,
>   and its lock `~/.cache/bootstrap.lock` does not exclude the new lock. A
>   version number in the name has no meaning after the transition. Each run
>   removes the legacy `~/.cache/bootstrap/`, `~/.cache/bootstrap.lock` and
>   `~/.cache/macos-setup/` (not while an old stub holds the old lock). The
>   first run moves `~/.cache/macos-setup/generated.txt` to
>   `macos-generated.tsv` before that removal.

Note: D12 of this plan narrows the last two sentences. This plan removes only
`~/.cache/bootstrap/` and `~/.cache/bootstrap.lock`; plan 10 handles
`~/.cache/macos-setup/`.

> - Clone URL: `https://github.com/virajp/bootstrap.git` on both sides
>   (public, fetch only, no SSH key needed on a new Mac). The normalized
>   origin compare stays as a guard for `SYNC_ENGINE_URL` in tests. The
>   engine never clones the private `macos-setup`: `setup` clones it once
>   after `gh auth login`; the engine reads the working clone at
>   `MACOS_SETUP_CONFIG_PATH`.

From this plan:

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan is the first of the chain."

D4: "No. mise run setup:configs in bootstrap and in each consumer stays a hand
step (Out of scope), as in earlier plans."

D5: "It becomes the forwarding file and stays until the parked cleanup plan,
because the old consumer stubs and the old setup/configs/sync tasks source
it."

D16: "The method of the repo-sync-check plan: the explicit list of passages in
05-docs.md, then vwf:docs-sync over the full branch delta, then grep checks.
The new docs/sync/migrate.md item gets the next free number after the highest
item (M27 when no other plan added one first)."

## Edits

1. **The explicit list.** Line numbers are from the survey; find each passage
   by its text if the lines moved.
   - **`CLAUDE.md`**:
     - "Layout rule": add `sync-engine/` to the files for this repository
       only: the shared library folder, never copied to a consumer, with
       `clone` and `repo-sync` today.
     - "How the sync operates" (:130-205): the cache is
       `${XDG_CACHE_HOME:-~/.cache}/sync-engine/` (`SYNC_ENGINE_CACHE`
       changes it) with the clone in `repo/` and the lock folder `lock/`;
       the URL is `https://github.com/virajp/bootstrap.git`
       (`SYNC_ENGINE_URL` changes it) on the branch `main`; the origin
       compare is normalized; a non-git, non-empty `repo/` stops with code
       2; each run removes `~/.cache/bootstrap/` and `~/.cache/bootstrap.lock`
       under the old lock (D12 of this plan); the
       stub embeds `sync-engine/clone`; `setup:configs:sync` sources
       `${SYNC_ENGINE_CLONE}/sync-engine/repo-sync` and stops with code 2
       without it; `repo_sync` never clones; the gate sources the engine of
       the cache and exits 0 without it. Replace `REPO_SYNC_CLONE` and
       `REPO_SYNC_URL` with the new names, also in the test bullet (:204).
     - "Task scripts", `_scripts/` list: `repo-sync` is now the forwarding
       file, which the cleanup plan removes; the sync code is
       `sync-engine/repo-sync`.
   - **`README.md`** (:133): the cache folder, the HTTPS URL.
   - **`docs/sync/tools/repo-sync.md`**: the table of files (:19-26: the
     library row becomes the forwarding file); the cache passage (:42-55);
     the gate passage (:67-80); the variable table (:121-130: add
     `SYNC_ENGINE_CACHE`, rename the two variables, `XDG_CACHE_HOME` is the
     base of the cache).
   - **`docs/sync/index.md`** (:669): code 5 is a clone or fetch that
     failed, or a `SYNC_ENGINE_CLONE` that is not usable.
   - **`docs/sync/migrate.md`**: the new item (D16) and its row in "Find the
     items that apply": "A sync that adds `sync-engine/`". Text: the cache
     moved to `~/.cache/sync-engine/`; the first sync after the update runs
     the old stub one time, and the second run removes
     `~/.cache/bootstrap/` and `~/.cache/bootstrap.lock` by itself; the
     consumer deletes nothing by hand. Update the first row of the table,
     which names "the old synced `setup/configs/sync`", when it is no longer
     true.
   - **`docs/design/*.md`**: each passage on the cache, the clone or the two
     variables, as `vwf:docs-sync` finds it.
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`).
   Apply each finding and each `DOCS FALSIFIED:` line of the earlier units
   that the list does not cover, inside the owned files.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rn 'REPO_SYNC_CLONE\|REPO_SYNC_URL\|cache}/bootstrap\|cache/bootstrap'
  CLAUDE.md README.md docs/sync docs/design`
  finds only the migrate item and the passages on the transition.
- `grep -ln 'sync-engine' CLAUDE.md README.md docs/sync/tools/repo-sync.md
  docs/sync/migrate.md`
  lists all four files.
- `mise x -- dprint check CLAUDE.md README.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- Do not edit `docs/plans/**`.
- Write identifiers in code spans.
- Each new paragraph obeys the style of the file it is in (ASD-STE100, short
  sentences).
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the sync-engine cache and the moved repo-sync library` —
written by the orchestrator after the unit gate and the wave gate, not by the
unit.
