# U8 — Docs

- **Wave:** 6
- **Depends on:** U1, U2, U3, U4, U5, U6, U7
- **Owns:** the repo's human-facing docs (README, CLAUDE.md, `docs/**` outside
  `docs/blueprint/` and `docs/plans/`, `cli/README.md`)
- **Model:** opus
- **Kind:** edit

## Ruling

Docs ship with the change. Run `vwf:docs-sync` over this run's branch delta and
apply its findings, plus every `DOCS FALSIFIED:` line U1–U6 returned. The
package README (`cli/README.md`) is what npm shows; it describes
`pnpx @virajp.dev/bootstrap@latest init` (the install command the site uses) and
only the commands that exist.

## Edits

1. Run `vwf:docs-sync` over the branch delta and apply what it finds.
2. Apply each `DOCS FALSIFIED:` line from the unit reports.

## Verification

- `MISE_ENV=dev mise run code:all` exits 0.

## Guardrails

- Do not edit `docs/blueprint/**` or `docs/plans/**`.
- Never document `BOOTSTRAP_TEST_PAUSE_AFTER_WRITE`.

## Commit

`docs: document bootstrap init`
