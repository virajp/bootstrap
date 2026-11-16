DECISION 2026-11-16 ★4 | repo:bootstrap, cli/internal/repo,
cli/internal/misegen,
repo/tools/node | requires_found is a list met when any one name is found, and
it is valid in a [mise_env.<env>.tools] entry too | alt rejected: all-of
requires_found, a separate requires_any_found key, a per-[lsp] requires_found
key, [mise.tools] pins with envs ["dev"]

# requires_found means any one name, and gates a mise_env entry too

The user ruled this during the run of `docs/plans/2026-11-14-lsp` (its
decisions D35 to D40).

- `requires_found` stays a list of file names, searched at the root and down to
  4 folders below it. It is met when **any one** of the names is found. This
  holds for a `[files]` entry, a `[tasks]` entry, a `[mise.tools]` entry and a
  `[mise_env.<env>.tools]` entry (D38).
- A `[mise_env.<env>.tools]` entry may carry `requires_found`. The sync writes
  the entry into `mise.<env>.toml` only when the gate is met, and drops the key
  from what it writes. `requires` and `files` stay valid only in
  `[mise.tools]` (D40).
- Why: a dev-only pin in `[mise.tools]` with `envs = ["dev"]` stops every sync
  that puts its tool in ci, and the base `linter` tool brings `node` into ci.
  So a dev-only pin with a file gate lives in `[mise_env.dev.tools]`.
- A skipped entry's skip line names every name: "a", "a or b", "a, b or c"
  (D39).
- The node tool: node alone gets no LSP server and no LSP pins. The dev pins
  `npm:typescript` and `npm:typescript-language-server`, and the one server
  `typescript-lsp`, need a `package.json` or a `tsconfig.json` (D35, D37). An
  LSP server follows its source pin's gate (D28); an `[lsp]` block has no gate
  of its own.
