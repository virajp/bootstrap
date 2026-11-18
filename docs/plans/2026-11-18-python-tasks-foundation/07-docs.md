# U7 — Docs

- **Read first:** the DOCS FALSIFIED lines that U1 to U5 returned, then the
  passages of "Facts the survey established" in `index.md`.
- **Lazy-load:** `docs/memory/decisions/` (the form of a record),
  `docs/plans/archived/2026-11-15-gate-file-sets/index.json` (D4 and D5 that
  D3 reverses).

## Ruling

- **D3 (reversal):** write `docs/memory/decisions/2026-11-18-python-tasks.md`
  with every ruling of this plan, the reversals of D3 and the
  rejected alternatives. Quote the user's words that the decisions quote.
- **D4, D5, D7, D8, D11:** CLAUDE.md "Task scripts" gets the rules of a
  Python task: the file form, the `.py` name, the import preamble, the
  standard library only, `# USAGE` and `usage_<flag>`, `helpers.tool` and
  `helpers.has_tool` for the `mise x --` rule, `shutil.which` for the tool
  exception, `sys.exit` codes of `bootstrap_get` and `members`. CLAUDE.md
  says `# MISE` with a space where it shows a Python header. The shell rules
  stay, marked as the rules of a shell task, until plan 2 deletes the shell
  libraries.
- **D9, D10, D12, D18, D19:** the docs of the libraries name the Python
  modules beside the shell libraries: CLAUDE.md "`_scripts/`" list,
  `docs/sync/tools/mise.md`, `docs/sync/tools/git.md`,
  `docs/sync/tools/swift.md`.
- **D13:** `docs/sync/tools/python.md`, `docs/sync/tools/dprint.md` and
  `docs/design/tools.md` say that dprint is the one formatter of every `.py`
  file, that `code:format:40-ruff` is gone, and that ruff and pyrefly run
  only with a `pyproject.toml`. `docs/sync/index.md` "Files that bootstrap
  no longer ships" gets a row for `code/format/40-ruff`.
- **D25:** the `migrate.md` item M47, with its row in "Find the items that
  apply": a repository with format settings in `[tool.ruff]` or
  `[tool.ruff.format]` of its `pyproject.toml` moves them to the ruff block of
  `dprint.jsonc` with `[tools.dprint.files."dprint.jsonc"]`
  (`docs/sync/tools/dprint.md` 33-60), under the setting names of the dprint
  ruff plugin. Find those names in the plugin docs (Context7 or its README)
  before you write them, and name only settings that the plugin accepts.
- **D6:** `docs/sync/index.md` (about line 640) and
  `docs/sync/tools/python.md` (about line 59) drop `UV_NO_CACHE`.
- **D12:** `docs/design/tools.md` ("The `#FILES` header of a gate") accepts
  `# FILES`.
- **D24:** the `migrate.md` item M46, with its row in "Find the items that
  apply": remove an `[overrides.mise.env] UV_NO_CACHE` of a consumer, or move
  it to `conf.d/zz-repo/mise.<env>.toml`. The override example of
  `docs/sync/index.md` (about line 640) uses `UV_PYTHON_DOWNLOADS` instead.
  Plan 2 writes the item of the deleted libraries (M48).
- **D24 (Mac):** `docs/setup/index.md`, the part that brings a Mac up to
  date, gets the step: remove an `[overrides.mise.env] UV_NO_CACHE` of the
  Mac values.
- **D13 (row):** the row of `code/format/40-ruff` in "Files that bootstrap
  no longer ships" names M47 in its Do column; the stale rule removes the
  synced file itself.
- **D28, D29:** `docs/sync/tools/python.md` says that ruff and pyrefly skip
  `.config/mise/tasks/`; CLAUDE.md shows the one `# MISE tools={...}` line of
  a task with a tool exception.
- **D21, D22, D23:** CLAUDE.md gives the preamble with
  `sys.dont_write_bytecode = True`, mode 755, `bold=True`, and the interface
  of `write_worktrees_page`.

## Edits

1. Run `vwf:docs-sync` over the branch delta, and apply its findings, every
   DOCS FALSIFIED line, and the passages of the survey.
2. Write the decisions record of D3.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -n UV_NO_CACHE docs CLAUDE.md -r` finds no passage that says it is
  set.

## Guardrails

- Touch nothing outside the owned paths.
- Do not edit a plan folder in `docs/plans/`.
- Write each changed passage in ASD-STE100.

## Commit

`docs: the rules of a Python task and the Python libraries`
