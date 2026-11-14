# U1 — The test dispatchers and the NN- filter

- **Read first:** `repo/tools/mise/tools.toml` (the
  `[tasks."code/format/_default"]`
  entry and its neighbours, about lines 84-122),
  `repo/tools/mise/tasks/code-format.tmpl` (the dispatcher pattern, the
  `own_tasks` loop at about line 34), `repo/tools/mise/tasks/helpers`
  (`own_tasks` at about lines 256-264, `has_own_task` at about 247-250).
- **Lazy-load:** `repo/tools/mise/tasks/code-lint.tmpl`,
  `repo/tools/mise/tasks/setup.tmpl`, `docs/design/mise.md` (the task-script
  rule at about line 638).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D1:** R1c handles only the ports that are not R2's. `code:format`,
  `code:lint`, `code:sec`, `code:precommit` and `setup` do not change.
- **D2:** the mise tool ships `test/_default`, `test/e2e/_default` and
  `test/golden/_default` (tasks `test`, `test:e2e`, `test:golden`) as group
  dispatchers that loop `own_tasks`, as `code:format` does. With no part, a
  dispatcher prints one line and exits 0.
- **D3:** `test` runs only its own `NN-` parts; it does not run `test:e2e` or
  `test:golden`.
- **D4 (the filter):** `own_tasks` takes only executable files whose name
  matches `NN-<name>` (two digits, a hyphen, a name), and skips folders. Every
  part that a tool ships today already has such a name.
- **D13, D19:** each test dispatcher declares in its `#USAGE` the member flags
  of `code-format.tmpl` (`--all`, `--<member>`) and no other flag, plus one
  variadic argument. It passes the member flags as `code-format.tmpl` does,
  and the variadic arguments unchanged to each part. A part with no `#USAGE`
  gets them as plain positional arguments.
- **D24:** a part's own flag (for example `--update`) reaches the part with no
  `--`: on mise 2026.11.7 an undeclared flag goes into the variadic argument.
  Pass the raw arguments (`"$@"`) to each part, after you remove the member
  flags; do not use the joined `usage_args` string, which loses the quoting of
  an argument with spaces.
- **D16:** a dispatcher with no part prints one line and exits 0, as
  `code:format` already prints a success line with no parts.

## Edits

1. **`repo/tools/mise/tasks/test.tmpl`, `test-e2e.tmpl`, `test-golden.tmpl`**
   (or plain files, if no template value is needed) — each a dispatcher in the
   style of `code-format.tmpl`: a `#MISE description`, `source` of the helpers,
   the member flags of `code-format.tmpl` (D13), a loop over `own_tasks test`
   (or `test:e2e`, `test:golden`) that runs each part with the arguments and
   flags, then a success line. With no part, print one line that says no part
   exists for the port, and exit 0.
2. **`repo/tools/mise/tools.toml`** — three `[tasks."test/_default"]`,
   `[tasks."test/e2e/_default"]` and `[tasks."test/golden/_default"]` entries,
   in the order and the style of the `code/*/_default` entries.
3. **`repo/tools/mise/tasks/helpers`** — `own_tasks` lists only files whose
   name matches `^[0-9][0-9]-` and that are not folders. Keep its sort order.
   `has_own_task` does not change.

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run code:lint` passes (shellcheck covers the new scripts).
- `grep -n 'test/_default\|test/e2e/_default\|test/golden/_default'
  repo/tools/mise/tools.toml`
  finds the three entries.

## Guardrails

- Touch nothing outside `repo/tools/mise/`.
- Do not change `code-sec`, `setup.tmpl` or any other dispatcher.
- This repo's own `.config/mise/tasks/` is not a synced copy you edit; do not
  touch it.
- Delete with `rm`, never `git rm`.

## Commit

`feat: the test, test:e2e and test:golden ports as group dispatchers`
