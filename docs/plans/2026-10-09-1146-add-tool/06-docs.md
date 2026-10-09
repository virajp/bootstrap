# U6 — Docs

- **Wave:** 4
- **Depends on:** U1, U2, U3, U4, U5
- **Owns:** the repo's human-facing docs — `README.md`, `CLAUDE.md`, `docs/**`
  outside `docs/blueprint/` and `docs/plans/`
- **Model:** opus
- **Kind:** edit
- **Read first:** every `DOCS FALSIFIED:` line the earlier units returned.
- **Lazy-load:** the docs that `vwf:docs-sync` names.

## Ruling

No assumed decision binds this unit. No reversal was confirmed in the interview,
so no `docs/memory/decisions/` file is written.

## Edits

1. Run `vwf:docs-sync` over the run's branch delta.
2. Apply its findings and every `DOCS FALSIFIED:` line from U1–U4.
3. The survey found no `cli/README.md` and no `cli/CHANGELOG.md`; do not create
   them.

## Verification

- `mise x -- mise run code:format -- --check` reports nothing on the edited
  docs, if the task accepts `--check`; else `dprint check` on them.

## Guardrails

- Do not edit `docs/blueprint/**`; the executor writes the stamps.
- Do not edit `docs/plans/**`.

## Commit

`docs: reconcile docs with the add command`
