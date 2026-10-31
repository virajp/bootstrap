# U3 — Examples: add direct_commit and review them

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `docs/design/bootstrap-config.md`, `docs/sync/index.md`
  (step 3, the keys of the values file), `examples/last_sync.toml` (read only).

## Ruling

D2: "One core branch, either main or develop, depending on the repo's need:
direct_commit on that branch drops it from the hook, and with no branch left
the sync removes the hook. With both main and develop: develop only (the hook
keeps --branch main), or both (the sync removes the hook). User: \"There can
never be a situation where main is allowed for direct-commit but develop is
NOT.\" That case stops the sync with code 2."

D4: "A boolean direct_commit in [branch.main] and [branch.develop]. [...]"

D8: "Add direct_commit to the examples where it fits, review all the examples
against the contract commits since commit_id (git log <commit_id>..HEAD --
tools docs/sync CLAUDE.md), and set commit_id. The gates unit sets commit_id
last, so no contract commit of the plan comes after it."

D12: "Assumed: as a commented line under [branch.develop], with one comment
line that gives the rule, so that a consumer who copies an example does not
allow direct commits by accident. Only polyrepo-develop and monorepo get the
line. No other example gets it, and no example shows it in [branch.main]."

## Edits

1. **Review.** Read `commit_id` in `examples/last_sync.toml`, then run
   `git log --oneline <commit_id>..HEAD -- tools docs/sync CLAUDE.md`. For
   each commit that changes the contract of the values file (a key, a table,
   a default, a rule that stops the sync), change each owned example that the
   commit makes wrong. Examples must stay valid values files for their shape.
   Each example keeps its empty `bootstrap_commit`.
2. **`direct_commit`.** Per D12, add a commented `direct_commit` line with one
   comment line of the rule to `examples/polyrepo-develop/bootstrap.toml`
   (`[branch.develop]`, `:19`) and `examples/monorepo/bootstrap.toml`
   (`[branch.develop]`, next to `:57`/`:62`). No other example gets the line.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- pre-commit run check-toml --config .config/pre-commit-config.yaml
  --files <each owned file>`
  passes.
- `grep -rn 'direct_commit' examples` finds the commented lines of step 2 and
  no uncommented `direct_commit = true`.
- Each example still has `bootstrap_commit = ""`.

## Guardrails

- Do not touch `examples/last_sync.toml`; U6 sets `commit_id`.
- Do not touch any doc; report a doc that an example change makes wrong as
  `DOCS FALSIFIED:`.
- Delete with `rm`, never `git rm`.

## Commit

`docs: add direct_commit to the examples and review them` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
