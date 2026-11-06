# U2 — Mac packs and the sample values file for the extras

- **Read first:** `macos/packs.toml` and `macos/examples/macos.toml`, top to
  bottom, then
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/packs.toml`
  and
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tests/setup-sync/sample-config/1password.toml`.
- **Lazy-load:** `sync-engine/select` (the pack rules),
  `.config/mise/tasks/test/macos-tools` (how the test reads the example).

## Ruling

From the design rulings of 2026-11-04 (verbatim):

"`packs.toml` shape (both sides): each key of `[pack.<name>]` is a member,
`<tool> = { default = true }` or `<tool> = {}` (implicit `false`). A
nested pack is a quoted `"pack:<name>"` key (not `p:`). The only entry
key is `default`; any other stops with exit 2. No pack-level keys (no
`depends`, `tools`, `default = [...]`, `required`, `description`)."

"`[packs.<name>]` adds each `default = true` member (option A). A new
`default = true` member reaches every consumer of the pack; a
`default = false` member arrives only with `[tools.<name>]`. [...]"

"A tool can be in two packs; the engine dedupes. The guide shows it once."

"Values row, item 2: one values file on each side. The values of a tool
are keys of its `[tools.<name>]` table, which also selects it; a table
for a base tool or a dependency only gives values. [...]"

From this plan:

D5: "In the order of the old tools list of that pack in MS/tools/packs.toml.
A plan that adds a member inserts it at its old position, so after P14 each
pack has the old order."

D6: "No new test case. test:macos-tools renders every tool that
macos/examples/macos.toml selects and parses every JSON, YAML and TOML
output, so U2's selection of the 20 tools is their test."

D7: "One [tools.<name>] table for each of the 20 tools, empty except
[tools.1password] with ssh_vault = "Sample". This follows
MS/tests/setup-sync/sample-config/macos-setup.toml, which selects every tool
outside pack base."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-macos-tools-base-shells."

## Edits

A member is `default = true` when the old pack has no `default` list, or
when its old `default` list names the member. Every other member is `{}`.

1. **`macos/packs.toml`** — add these tables after the existing tables, in
   the order of `MS/tools/packs.toml` (`ai`, `essentials`, `security`,
   `dev`, `terminal`, `chat`, `cli-extras`), each member at its old
   position (D5):
   - `[pack.ai]`: `claude-desktop = { default = true }`,
     `claude-cli = {}`, `claude-status = {}`, `codeburn = {}`.
   - `[pack.essentials]` (it exists with `diff-so-fancy = {}`): add
     `tailscale = {}` before `diff-so-fancy`, and `git-tools = {}` and
     `glab = {}` after it.
   - `[pack.security]`: `1password = { default = true }`,
     `1password-cli = {}`, `1password-safari = {}`, `bitwarden = {}`.
   - `[pack.dev]`: `devtools = { default = true }`,
     `dprint = { default = true }`, `orbstack = {}`, `gem = {}`,
     `superfile = {}`.
   - `[pack.terminal]`: `ghostty = { default = true }`.
   - `[pack.chat]`: `telegram = { default = true }` (the old pack has no
     `default`).
   - `[pack.cli-extras]`: `dasel = { default = true }`,
     `pstree = { default = true }` (the old pack has no `default`).
     Keep the file comments in the style of the file. Do not add a pack-level
     key.
2. **`macos/examples/macos.toml`** — add one `[tools.<name>]` table for each
   of the 20 tools, in the place and order that the file uses for the
   selected tools (D7). `[tools.1password]` holds `ssh_vault = "Sample"` with
   an STE comment, for example `# Sample answer of the ssh_vault question.`
   The other 19 tables are empty. Add no `env` key.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- Both files parse: `mise x -- yq -p toml -o json . <file> >/dev/null`.
- `grep -nE '^(tools|default|depends|required) *=' macos/packs.toml` finds
  nothing, and `grep -n '"p:' macos/packs.toml` finds nothing.
- Each of the 20 tools is a member of exactly one pack in
  `macos/packs.toml`, and has one `[tools.<name>]` table in
  `macos/examples/macos.toml`.

## Guardrails

- Touch only `macos/packs.toml` and `macos/examples/macos.toml`. Do not
  touch `macos/tools/`; U1 owns the 20 tool folders.
- Do not remove or change a member or a value that an earlier plan wrote.
- Write the comments in ASD-STE100 Simplified Technical English.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the Mac packs and the sample values of the extra tools` —
written by the orchestrator after the unit gate and the wave gate, not by
the unit.
