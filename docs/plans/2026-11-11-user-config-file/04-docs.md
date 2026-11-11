# U4 — Docs

- **Read first:** `docs/design/cli.md` ("The command line", "The engine
  variables", "The exit codes"), and each other owned file at the passages that
  name an engine variable.
- **Lazy-load:** `CLAUDE.md` ("Where the rules are").

## Ruling

The plan rulings D0, D2, D3, D4, D5, D6, D7, D8, D12, D14 and D16 of
`index.json`, as the
code of U1 implements them.

## Edits

1. Run `vwf:docs-sync` over the branch delta of the run. Apply its findings,
   each `DOCS FALSIFIED:` line of U1 and U2, and the docs of the facts section
   of `index.md`.
2. **`docs/design/cli.md`** — add a section "The config file": the path order,
   the four keys with their variable and flag, the precedence, the rules of D5
   and D6, D7, that no tool, seed or sync writes the file (ruling 7), the empty
   and bad
   variables of D14, and the limit of D12: the gate `code:lint:60-bootstrap`
   and the cleanup of `updateall:configs` find the cache only from
   `BOOTSTRAP_CACHE` or the default. Add
   `BOOTSTRAP_CONFIG` to "The engine variables". Add `config` and the new
   flags to "The command line". Name `--verbose` in "The output of a run".
3. **`docs/design/mac.md`** — `--repo-path` next to `MACOS_SETUP_REPO_PATH`, and
   why the file has no Mac key (D16).
4. **`README.md`, `docs/sync/index.md`, `docs/setup/index.md`,
   `docs/setup/tools.md`** — where a passage tells a user to export a variable
   of the engine, name the config file and `bootstrap config` too.
5. **`CLAUDE.md`** ("Tests") — each test task sets `BOOTSTRAP_CONFIG` to a
   missing path.

Write the docs in ASD-STE100 Simplified Technical English, in the style of the
existing text.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `grep -n 'BOOTSTRAP_CONFIG' docs/design/cli.md CLAUDE.md` finds each file.

## Guardrails

- Do not touch `cli/` or the test tasks.
- Do not edit a synced copy (see `CLAUDE.md`).

## Commit

`docs: describe the user config file of the engine and bootstrap config`
