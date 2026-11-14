# U1 — The engine: the [lsp] block, bootstrap lsp and describe's lsp field

- **Read first:** `cli/internal/source/manifest.go` (the `Manifest` type, about
  lines 48-83; `repoLines`, about 180+, and the `[formatter]` parse, about
  210-219, the model for a small typed table), `cli/internal/source/check.go`
  (the repo-side loop, about lines 100-117, the duplicate mise owner check,
  about 106-111, and `repoRules`, about 121-161), the `describe` and
  `contract` code that R1b adds (`cli/internal/cli/`, `cli/internal/contract/`),
  and the tool selection reader of the repo.
- **Lazy-load:** `repo/tools/kotlin/tools.toml` (the `http:kotlin-lsp` pin,
  about lines 23-30),
  `docs/plans/2026-11-14-contract-commands/04-catalog-describe.md`
  (R1b D10), `docs/design/cli.md`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D4:** a tool declares `[lsp.<server>]` with `command` (string, needed),
  `args` (list of strings, default empty), `extensions` (table of
  `".<ext>" = "<language id>"`, needed, not empty), `startup_timeout`
  (integer milliseconds, optional) and `source` (needed: the name of a mise
  tool that this tool pins, or `"host"`). `bootstrap lsp --json` prints, in
  the R1b envelope, `data` = `{"servers": {<server>: <entry>}}` for the
  servers of the repo's selected tools, each entry in Claude's `lspServers`
  shape: `{"command": "mise", "args": ["x", "--", <command>, <args>...],
  "extensionToLanguage": <extensions>, "startupTimeout": <startup_timeout>}`
  (`startupTimeout` left out when not set). A `source = "host"` server uses
  the same `mise x --` launch.
- **D6:** `bootstrap lsp` without `--json` prints one line per server: the
  name, the tool, the command and the extensions. `describe --json` gets an
  `lsp` field with the same `servers` object; this is additive (R1b D7), so
  the contract stays 1. `contract` lists `lsp` in `commands`. Outside a repo,
  or with no `.config/bootstrap/`, `bootstrap lsp` exits 2 like `describe`
  (R1b D28).
- **D12:** a server follows the `requires_found` of its source pin: when that
  pin is not active in the repo (for example `pipx:pyrefly` with no
  `pyproject.toml`), `bootstrap lsp` and `describe` leave the server out.
  **D19:** the same holds for a dev-only pin when the tool's envs do not
  include `dev`.
- **D28 (the full rule, replacing D12 and D19):** a server is listed only when
  the tool's envs include `dev` and its source pin's `requires_found` is met;
  a `"host"` source has no `requires_found`.
- **D30:** `command` is one binary with no whitespace; the rest goes in
  `args`. `check` refuses a `command` with whitespace (code 2, beside D7).
- **D32:** if a Go test asserts that the port `code:lsp:check` is not
  present, update it.
- **D13:** `source` is one string: the pin that provides the command.
- **D24:** `bootstrap lsp --json` `data` also has `checks`, beside `servers`:
  `{<server>: {"command": <the first word of the [lsp] command>, "source":
  <the source>}}`. The `servers` entries stay verbatim Claude entries (D4).
  `describe`'s `lsp` field carries both maps.
- **D7:** `check` refuses, with code 2: an `[lsp.<server>]` with no `command`
  or no `extensions`, or empty `extensions`; one server name in two tools; a
  `source` that is neither `"host"` nor the name of a mise tool that the same
  tool pins (in `[mise.tools]` or `[mise_env.<env>.tools]`).

## Edits

1. **`cli/internal/source/manifest.go`** — a typed `LSP` field on `Manifest`,
   parsed in `repoLines` in the style of `[formatter]`, with a "bad" line for a
   wrong shape.
2. **`cli/internal/source/check.go`** — the D7 rules, the duplicate-name check
   beside the duplicate mise owner check.
3. **`cli/internal/cli/`** — the `lsp` command (D4, D6), its usage line, its
   `--json`; the `lsp` field of `describe`; `lsp` in `contract`'s `commands`.
4. **Tests** — Go tests for the parse, each D7 rule, the D4 entry shape (with
   and without `startup_timeout`), `describe`'s `lsp` field, and exit 2 outside
   a repo.

## Verification

- `mise x -- mise run test:go` passes.

## Guardrails

- Touch nothing outside `cli/`. The tool files are U2's.
- Delete with `rm`, never `git rm`.

## Commit

`feat: the [lsp] block, bootstrap lsp, and describe's lsp field`
