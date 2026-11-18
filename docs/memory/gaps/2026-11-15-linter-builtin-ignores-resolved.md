GAP 2026-11-15 ★2 | plan:gate-file-sets, repo:linter (@virajp.dev/linter),
repo/tools/linter | the built-in ignore list of the linter duplicates the gate
header of code-lint-linter | resolved 2026-11-18: code-lint-linter runs
`linter --no-ignore`, so the ignores of the linter config have no effect and
gate_files is the only list

**Source** The interview of the gate file-sets plan (the plan of the
`code:format` and `code:lint` file model). The user agreed to park this item.

What the plan decides:

- One Python helper is the only place that removes a file for a gate. Each gate
  names its includes and excludes in a `#FILES` header.
- The `#FILES` header of `code-lint-linter` copies the built-in ignore list of
  `@virajp.dev/linter` 2.1.0 (`dist/config/ignores.js`: `**/node_modules/`,
  `**/dist/`, `**/*.d.ts`, `**/AGENTS.md`, `.claude/**/*.md`,
  `**/.worktrees/**` and others). Thus ESLint never gets a file that it ignores,
  and prints no "File ignored" warning.
- A consumer moves its `.config/linter.yaml` `ignores` into
  `[tools.mise] code_files.excludes` (a `migrate.md` item).

What is open: the linter package keeps its built-in list, so two lists hold the
same patterns. When a new linter version changes its list, the gate header is
out of date. The fix is in the linter repository: an option that turns off the
built-in list, or `warnIgnored: false` (ESLint `--no-warn-ignored`), so that the
helper holds the only list. Then the gate header can drop the copy.
