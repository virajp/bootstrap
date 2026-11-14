GAP 2026-11-14 ★3 | plan:2026-11-14-lsp, plugin:stackgen, skill:doctor,
skill:execute | the plugins still keep their own LSP data and park their LSP
check | parked: R3 or R5 in claude-plugins

**Source** The R2b interview of `docs/plans/2026-11-14-lsp/`.

From R2b on, bootstrap holds the LSP server of each language tool in an
`[lsp]` block, `bootstrap lsp --json` prints the servers of a repo in Claude's
`lspServers` shape, and `code:lsp:check` tests that each one resolves.

What claude-plugins must change, in R3 (plugins use the contract) or R5 (stack
without packs):

- stackgen writes vwf-tools' `lspServers` from `bootstrap lsp --json`, and
  stops using the packs' `lsp_servers` blocks (for example
  `stacks/language/kotlin/pack.yaml`). The server keys stay `kotlin-lsp`,
  `sourcekit-lsp`, `dart-lsp` and `typescript-lsp`; bootstrap adds `gopls`,
  `pyrefly`, `rust-analyzer` and `jdtls`.
- vwf doctor and `/vwf:execute` can run `code:lsp:check` for their parked LSP
  check (claude-plugins
  `docs/memory/gaps/2026-11-13-execute-lsp-check-parked.md`,
  doctor's `checks.mjs`).
