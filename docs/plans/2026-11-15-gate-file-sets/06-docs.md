# U6 — Docs

- **Read first:** the DOCS FALSIFIED lines that U1 to U4 returned, then the
  passages that the survey found (section "Facts the survey established" of
  `index.md`).
- **Lazy-load:** `docs/sync/migrate.md` (the form of an item),
  `docs/memory/decisions/` (the form of a record).

## Ruling

- **D20 (reversal):** the CLAUDE.md gate rule of "Task scripts" changes: the
  helper `gate_files` makes the list in every mode, and each tool gets only
  that list. The name `--all` of `code:format` and `code:lint` means every file
  of the repository. Write `docs/memory/decisions/2026-11-15-gate-file-sets.md`
  with the rulings D1 to D17, the reversal and the rejected alternatives.
- **D9, D10:** `docs/sync/migrate.md` gets one item: move the excludes of
  `.config/linter.yaml` (`ignores`), of `[tools.dprint] excludes`, and the path
  excludes that a consumer relied on in a tool config (ruff, pyrefly,
  SwiftLint) into `[tools.mise] code_files.excludes`; the gates now examine
  only the uncommitted files without `--all`.
- **D21:** the docs say that `code:sec` keeps its own lists.
- **D32:** this repository's own excludes (`.config/linter.yaml`, the root
  `dprint.jsonc`) move at its next sync, by the hand step of CLAUDE.md
  "Change a shared file" step 4; this unit edits neither file.
- **D28:** the docs say that `mise run code` takes `--all`.
- **D45:** CLAUDE.md "Task output" gets one line: with `--debug`, the drop
  lines of `gate_files` go to stderr and can come before the header.
- **D35, D36, D37, D38:** the migrate.md item also says: add `"!**/*.py"` to
  the shuck associations of `dprint.jsonc`; move the own excludes of
  `dprint.jsonc` and `[tools.dprint] excludes` into
  `[tools.mise] code_files.excludes`; add `--all` and a `gate_files` call to
  each own `code:format` and `code:lint` gate of the repository.
- `docs/sync/tools/linter.md` gets the new gate `code:lint:06-shellcheck`;
  `docs/sync/tools/mise.md` gets `code_files`, `config-files.toml`,
  `gate_files` and the `#FILES` header; `docs/sync/tools/dprint.md` says that
  its excludes come from `code_files`.
- `docs/design/tools.md`: the `#FILES` header of a gate, beside "The `[tasks]`
  table".

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings, every
   DOCS FALSIFIED line, and the passages of the survey:
   `CLAUDE.md` (Task scripts, the gate rules), `README.md`,
   `docs/design/mise.md`, `docs/design/repo.md`, `docs/design/tools.md`,
   `docs/design/cli.md`, `docs/sync/index.md`, `docs/sync/migrate.md`,
   `docs/sync/tools/{mise,linter,dprint,flutter,kotlin,swift,python,go,android}.md`.
2. Write the decisions record and the migrate item of the Ruling.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn "whole tree" CLAUDE.md docs/sync/tools` finds no passage about a
  gate.

## Guardrails

- Touch nothing outside the owned paths.
- Do not edit a plan folder in `docs/plans/`.

## Commit

`docs: the gate file sets, the code_files key and gate_files`
