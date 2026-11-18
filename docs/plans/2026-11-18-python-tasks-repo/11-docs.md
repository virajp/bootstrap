# U11 — Docs

- **Read first:** the DOCS FALSIFIED lines that U1 to U9 returned, then the
  passages of "Facts the survey established" in `index.md`.
- **Lazy-load:** `docs/sync/migrate.md` (the form of an item and of the "Find
  the items that apply" table), `docs/sync/index.md` ("Files that bootstrap
  no longer ships").

## Ruling

- **D8:** the rows and the item for the deleted libraries. Add one row for
  each of `_scripts/helpers`, `_scripts/checks`, `_scripts/placeholder`,
  `_scripts/worktrees-page`, `_scripts/merge` and `_scripts/xcode`, with the
  `.py` module that replaces it, and one row that says that each task file
  without an extension now has a `.py` name, and that the stale rule removes
  the old copy.
- **M48** in `docs/sync/migrate.md` (M46 and M47 are the items of the foundation
  plan), with its row in "Find the items that
  apply": a repository with an own task that sources a shell library of
  `_scripts/`, or an own orchestrator part that calls a shell function of
  `helpers`, moves that task to a Python uv script with the import preamble
  (the rules of CLAUDE.md "Task scripts"). Without such a task, nothing to do.
- **D1:** CLAUDE.md "Task scripts" loses the rules of a shell task that only
  the shell libraries needed: the bash 3.2 rule, `VAR=$(cmd) || VAR=''`,
  `|| exit $?`, `source .../helpers`, and the shell gate pattern. The
  `_scripts/` list names only the `.py` modules and `gate_files.py`. The
  "tool exception" paragraph names only `github:virajp/bootstrap`,
  `pre-commit` and `dprint` (D4).
- `docs/sync/tools/<tool>.md` of each tool of this plan: the task file names,
  and no yq for a task.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings, every
   DOCS FALSIFIED line, and the passages of the survey.
2. Write the rows and M48 of the Ruling.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn 'source "\${MISE_PROJECT_ROOT}' CLAUDE.md docs/` finds no rule
  that a shipped task follows.

## Guardrails

- Touch nothing outside the owned paths.
- Do not edit a plan folder in `docs/plans/`.
- Write each changed passage in ASD-STE100.

## Commit

`docs: the Python tasks of repo/tools and migrate item M48`
