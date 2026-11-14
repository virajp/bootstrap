# U3 — The code:lsp:check task

- **Read first:** `repo/tools/mise/tools.toml` (the `[tasks.*]` entries),
  `repo/tools/mise/tasks/code-format.tmpl` (the style of a task),
  `repo/tools/mise/tasks/helpers`, and the `bootstrap lsp --json` output of U1.
- **Lazy-load:** `repo/tools/git/tasks/merge` (the
  `${BOOTSTRAP_BIN:-mise x -- bootstrap}` call pattern and the D85 exit-2 rule
  of R1a U4).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D5:** the mise tool ships `code/lsp/check` (task `code:lsp:check`). It
  reads `${BOOTSTRAP_BIN:-mise x -- bootstrap} lsp --json` (with the D85
  exit-2 rule of R1a U4 when no engine is installed) and, for each server,
  tests that its command resolves.
- **D24:** read the command and the source of each server from `data.checks`
  (not from the Claude entries in `data.servers`, whose command is `mise`).
  A `source` of `"host"` uses `xcrun --find`; any other uses `mise which`.
- **D16:** for a pinned server, use `MISE_ENV=dev mise which <command>`,
  which resolves only a mise-managed binary; for a host server
  (`sourcekit-lsp`), use `xcrun --find sourcekit-lsp`. Never `command -v`.
- **D28:** a server is in `bootstrap lsp --json` (and so checked) only when
  its tool's envs include `dev` and its source pin's `requires_found` is met.
- **D29:** if `bootstrap lsp --json` exits 2, the task exits 2 with that fix
  line; any other failure of it exits 1.
- **D5, D22, D23:** the task starts no server. A missing server prints one
  line with the fix: `mise install` for a pinned server; for `sourcekit-lsp`
  (`source = "host"`), "install the Xcode Command Line Tools
  (`xcode-select --install`) or Xcode". It exits 1 if any server is missing,
  and 0 when all resolve or the repo has no server.

## Edits

1. **`repo/tools/mise/tasks/code-lsp-check`** (new) — D5, in the style of the
   other tasks of the mise tool.
2. **`repo/tools/mise/tools.toml`** — `[tasks."code/lsp/check"]`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.

## Guardrails

- Touch nothing outside `repo/tools/mise/`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: code:lsp:check tests that each LSP server resolves`
