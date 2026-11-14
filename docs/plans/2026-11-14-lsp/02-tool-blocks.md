# U2 — The [lsp] blocks and the server pins of the language tools

- **Read first:**
  `repo/tools/{go,node,python,kotlin,swift,flutter,rust,java}/tools.toml`,
  and in `repo/tools/kotlin/tools.toml` the `http:kotlin-lsp` pin (about lines
  23-30: `version_list_url`, `version_regex`, the per-platform URLs) as the
  model for a dev-only http pin.
- **Lazy-load:** `docs/design/tools.md` (the `[mise.tools]` and
  `[mise_env]` keys), `docs/sync/tools/*.md` of these tools.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. D4 (the block shape) is quoted in
`01-engine-lsp.md`.

- **D1:** every language tool gets an `[lsp]` block: kotlin, swift, flutter,
  node, go, python, rust and java.
- **D2:** python uses `pyrefly lsp`; its source is the `pipx:pyrefly` pin
  that the python tool has today.
- **D3:** java uses jdtls from a dev-only `http:jdtls` pin of the Eclipse
  jdtls milestone archive, in the same form as kotlin-lsp. Verify that the
  version list URL and the archive URL resolve; if they do not, stop and ask.
- **D8:** every new server pin is dev-only (`[mise_env.dev.tools.<name>]`), as
  kotlin-lsp is.

The servers, one per tool (D30: `command` is one binary; the rest is `args`):

| Tool    | Server name      | command, args                                        | extensions                                                                                                                  | source                                                                                                     |
| ------- | ---------------- | ---------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| kotlin  | `kotlin-lsp`     | `kotlin-lsp`, args `["--stdio"]`                     | `.kt`, `.kts` → `kotlin`                                                                                                    | the existing `http:kotlin-lsp` pin                                                                         |
| swift   | `sourcekit-lsp`  | `sourcekit-lsp`                                      | `.swift` → `swift`                                                                                                          | `"host"` (D14: `startup_timeout = 1000`)                                                                   |
| flutter | `dart-lsp`       | `dart`, args `["language-server", "--protocol=lsp"]` | `.dart` → `dart`                                                                                                            | `flutter`                                                                                                  |
| node    | `typescript-lsp` | `typescript-language-server`, args `["--stdio"]`     | copy verbatim from claude-plugins `plugins/stackgen/stacks/language/typescript/pack.yaml` (D15: `.tsx` and `.jsx` included) | `npm:typescript-language-server` (D13); new dev pins `npm:typescript-language-server` and `npm:typescript` |
| go      | `gopls`          | `gopls`                                              | `.go` → `go`                                                                                                                | new dev pin `go:golang.org/x/tools/gopls`                                                                  |
| python  | `pyrefly`        | `pyrefly`, args `["lsp"]`                            | `.py`, `.pyi` → `python`                                                                                                    | `pipx:pyrefly`                                                                                             |
| rust    | `rust-analyzer`  | `rust-analyzer`                                      | `.rs` → `rust`                                                                                                              | new dev pin `aqua:rust-lang/rust-analyzer`                                                                 |
| java    | `jdtls`          | `jdtls`                                              | `.java` → `java`                                                                                                            | new dev pin `http:jdtls` (D3)                                                                              |

`startup_timeout`: 60000 for `kotlin-lsp`, `dart-lsp`, `typescript-lsp` (D15)
and `jdtls`; 1000 for `sourcekit-lsp` (D14); not set for the others.

- **D18:** bun and android get no `[lsp]` block, on purpose: android code
  uses the kotlin tool's server, and a bun-only repo has no TypeScript server
  in R2b.
- **D21:** the node tool installs npm packages through `aube`. Verify in a
  scratch repo with node: `MISE_ENV=dev mise install`, then
  `MISE_ENV=dev mise which typescript-language-server`. If it fails, stop and
  ask.
- **D25:** verify in a scratch repo with rust that
  `MISE_ENV=dev mise which rust-analyzer` prints a path under the aqua
  install, not the rustup proxy of `core:rust`; if it prints the proxy, stop
  and ask.
- **D17:** the jdtls launcher `bin/jdtls` is a Python script that uses the host
  `python3` (macOS Xcode CLT and the Linux CI have it); the java tool gets no
  Python dependency. Verify that the `http:jdtls` pin exposes `bin/jdtls` on
  the mise path (for example with the backend's bin-path option); if it
  cannot, stop and ask. The server names and the kotlin, swift and dart entries
  match the
  stackgen packs of claude-plugins, so vwf-tools keeps the same keys.

## Edits

1. **Each of the eight `tools.toml` files** — the `[lsp.<server>]` block of
   the table above, and the new dev pins (D8).
2. **`repo/tools/java/tools.toml`** — the `http:jdtls` pin (D3), with a
   version list URL, a version regex and the archive URL, after you verify
   them.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- Each of the eight files has exactly one `[lsp.` table.

## Guardrails

- Touch nothing outside the eight `tools.toml` files.
- Do not change a pin that exists, except to add the dev pins.
- Delete with `rm`, never `git rm`.

## Commit

`feat: an LSP server block and pin in each language tool`
