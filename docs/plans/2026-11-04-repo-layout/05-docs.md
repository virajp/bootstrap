# U5 — Docs

- **Read first:** every owned file that the list below names, at the named
  passages, before editing.
- **Lazy-load:** the diffs of U1, U2 and U3 (`git log -p` of the branch), the
  `DOCS FALSIFIED:` lines that the earlier units returned.

## Ruling

D4: "The user, by hand, after the landing. It is out of scope, as in the
earlier plans."

D6: "Both hold. With a P05 manifest, the stale rule removes the consumer copy
of .config/mise/tasks/_scripts/repo-sync, because the sync no longer ships it
and nothing in a consumer sources it after P01. Without a manifest, the copy
stays until the parked cleanup, whose migrate.md item covers it.
docs/sync/tools/repo-sync.md says so; no migrate.md item in P06."

D9: "git log <commit_id>..HEAD -- tools repo docs/sync CLAUDE.md. It keeps
tools until commit_id is set again, because the contract changes before the
move are under tools/. P15 later adds docs/design."

## Edits

1. **The explicit list.** Find each passage by its text.
   - **`CLAUDE.md`**: "Layout rule": shared repo files are in
     `repo/tools/<tool>/`, the packs in `repo/packs.toml`, the examples in
     `repo/examples/` (with `last_sync.toml`); `tools/` holds only the
     forwarding file, which the sync does not ship; `sync-engine/` holds the
     engine. The example-review command becomes
     `git log <commit_id>..HEAD -- tools repo docs/sync CLAUDE.md` (D9). The
     header rule
     shows `# Synced from bootstrap: repo/tools/<tool>/<path>. Do not edit this
     file.`
     "Task scripts", "Change a shared file", "mise config" and every other
     `tools/<tool>` path become `repo/tools/<tool>`. `docs/sync/tools/<tool>.md`
     is "one file for each folder in `repo/tools/`".
   - **`README.md`**: each `tools/` and `examples/` path of bootstrap.
   - **`docs/sync/index.md`**, **`docs/sync/vwf.md`**,
     **`docs/sync/kotlin-dsl.md`**,
     **`docs/sync/packs/base.md`**, **`docs/sync/packs/security.md`**: each
     bootstrap path `tools/...`, `tools/packs.toml` and `examples/...`.
     `docs/sync/tools/docker.md` stays unchanged: its `.tools/` is a
     consumer path.
   - **`docs/sync/migrate.md`**: only a passage that tells the reader to open
     a bootstrap path today; an old item that describes history stays.
   - **`docs/design/bootstrap-config.md`**, **`docs/design/packs.md`**,
     **`docs/design/tools.md`**, **`docs/design/mise.md`**: each bootstrap
     path.
   - **`docs/sync/tools/repo-sync.md`**: the forwarding file is no longer
     shipped; with a P05 manifest the stale rule removes the synced copy
     `.config/mise/tasks/_scripts/repo-sync`; without one the copy stays
     until the parked cleanup (D6).
2. **`vwf:docs-sync`.** Run it over the full branch delta
   (`/Users/virajpatel/Projects/github.com/virajp/claude-plugins/.dev-marketplace/plugins/vwf/skills/docs-sync/SKILL.md`),
   and apply each finding and each `DOCS FALSIFIED:` line of the earlier
   units that the list does not cover, inside the owned files.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -rnE '(^|[^/])tools/<tool>' CLAUDE.md README.md docs/sync docs/design`
  finds nothing.
- `grep -n 'tools repo docs/sync CLAUDE.md' CLAUDE.md` finds the command.
- `mise x -- dprint check README.md CLAUDE.md docs` passes.

## Guardrails

- Touch only the owned files. Report a doc outside the list as a `GAP:`.
- A consumer path such as `.config/mise/tasks/...` stays.
- Write identifiers in code spans.
- Delete with `rm`, never `git rm`.

## Commit

`docs: describe the repo/ side folder` — written by the orchestrator after
the unit gate and the wave gate, not by the unit.
