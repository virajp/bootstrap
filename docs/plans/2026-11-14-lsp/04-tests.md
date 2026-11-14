# U4 — Tests

- **Read first:** `.config/mise/tasks/test/repo` (its case helpers, its
  scratch repos and how a case runs the engine), and the work of U1 to U3.

## Ruling

D4, D6 and D7 are quoted in `01-engine-lsp.md`; D1 to D3 and D8 in
`02-tool-blocks.md`; D5 in `03-lsp-check.md`. This unit tests the built engine
and the synced task end to end. Every case that runs a task exports
`BOOTSTRAP_BIN=<the engine that test/repo builds>`. No case uses the network.

## Edits

1. **`.config/mise/tasks/test/repo`** — add cases:
   1. A scratch repo that selects go and node: `bootstrap lsp --json` lists
      `gopls` and `typescript-lsp` in the D4 shape; `describe --json` has the
      same `lsp.servers`.
   2. A scratch clone whose tool has an `[lsp.x]` with no `extensions`:
      `check` exits 2. Two tools with one server name: `check` exits 2. A
      `source` that the tool does not pin: `check` exits 2.
   3. `code:lsp:check` with stubs (D20): run the synced task file directly with
      `bash`, not through `mise run`, with a `BOOTSTRAP_BIN` stub that prints a
      `lsp --json` envelope with one server, and a `PATH` shim for `mise`
      whose `which` prints a path for that command; the task exits 0. With the
      shim's `which` failing: exit 1 and the fix line (D16).
   4. A scratch repo with no language tool: `code:lsp:check` exits 0.
   5. A scratch repo that selects python with no `pyproject.toml`:
      `bootstrap lsp --json` has no `pyrefly`; with a `pyproject.toml`, it has
      `pyrefly` (D12). A tool whose envs do not include `dev` has no server in
      the output (D28).
   6. A clone whose tool has `command = "pyrefly lsp"`: `check` exits 2
      (D30). A `BOOTSTRAP_BIN` stub that exits 2: `code:lsp:check` exits 2
      (D29).
   7. Search this file for a case that asserts the port `code:lsp:check` is
      not present (from R1b or R1c), and update it (D32).

## Verification

- `mise x -- mise run test:repo` passes.

## Guardrails

- Touch nothing outside `.config/mise/tasks/test/repo`.
- Delete with `rm`, never `git rm`.

## Commit

`test: the LSP blocks, bootstrap lsp and code:lsp:check`
