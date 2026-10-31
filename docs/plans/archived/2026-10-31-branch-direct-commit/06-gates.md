# U6 — Gates

- **Read first:** `examples/last_sync.toml`.

## Ruling

D8: "[...] The gates unit sets commit_id last, so no contract commit of the
plan comes after it."

## Edits

1. **Generators:** the plan names none.
2. **`examples/last_sync.toml`** — set `commit_id` to the output of
   `git rev-parse HEAD`, run after the docs unit's commit. Change no other
   line.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `git log <new commit_id>..HEAD -- tools docs/sync CLAUDE.md` is empty
  before this unit's commit.

## Guardrails

- Touch only `examples/last_sync.toml`.
- Do not bump a version.
- Delete with `rm`, never `git rm`.

## Commit

`docs: set the examples commit_id after the direct_commit review` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
