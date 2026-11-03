# U3 — The last Mac packs and the full sample selection

- **Read first:** `macos/packs.toml` and `macos/examples/macos.toml`, top to
  bottom, then
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/packs.toml`
  and
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tests/setup-sync/sample-config/macos-setup.toml`.
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

"Values row, item 2: one values file on each side. The values of a tool
are keys of its `[tools.<name>]` table, which also selects it; a table
for a base tool or a dependency only gives values. [...]"

From this plan:

D5: "In the order of MS/tools/packs.toml: the packs in the old file order,
and the members of each pack in the order of its old tools list. This plan
inserts each member at its old position."

D6: "Every member. The old rule was "absent: every tool of the pack is a
default", so iwork, ms-office, chat and cli-extras get default = true on
each member."

D7: "One empty [tools.<name>] table for each of the 52 tools, so the file
selects every tool outside pack base, as
MS/tests/setup-sync/sample-config/macos-setup.toml does. test:macos-tools
then renders every Mac tool."

D8: "No new test case. test:macos-tools renders every selected tool and
parses every JSON, YAML and TOML output, so the selection of D7 is the test
of the 52 tools."

D9: "Only the tool names outside pack base. macos/examples/macos.toml also
holds [tools.<name>] tables of base tools for their answers (P10 to P12); the
check ignores them."

D2: "No. The 16 plans form one linear chain through requires, because most of
them edit sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans
conflict there. This plan requires 2026-11-04-macos-tools-extras."

## Edits

A member is `default = true` when the old pack has no `default` list (D6),
or when its old `default` list names the member. Every other member is `{}`.
`[x]` marks a member that an earlier plan wrote; keep it as it is.

1. **`macos/packs.toml`** — after this edit the file holds the 14 packs in
   the old file order (D5): `base`, `ai`, `essentials`, `security`, `dev`,
   `terminal`, `apple-dev`, `iwork`, `ms-office`, `disk`, `media`, `chat`,
   `notes`, `cli-extras`. Add the members and the new tables:
   - `ai`: [claude-desktop], [claude-cli], [claude-status], [codeburn],
     `opencode = {}`, `orca = {}`, `zed-delta = {}`.
   - `essentials`: `alfred`, `vorssaint`, `easyfind`, `the-unarchiver`,
     `brave-browser`, `github-desktop` each `{ default = true }`;
     `google-chrome = {}`, [tailscale], `magnet = {}`, `sill = {}`,
     `taskexplorer = {}`, `battery-health = {}`, `setapp = {}`,
     [diff-so-fancy], [git-tools], [glab].
   - `security`: no change.
   - `dev`: [devtools], [dprint], `visual-studio-code = { default = true }`,
     `lcov = {}`, [orbstack], `gcloud = {}`, [gem], [superfile],
     `zed = {}`, `sublime-text = {}`.
   - `terminal`: [ghostty], `warp = {}`.
   - new `[pack.apple-dev]`: `xcode = { default = true }`,
     `developer = {}`, `testflight = {}`, `transporter = {}`,
     `android-studio = {}`.
   - new `[pack.iwork]`: `keynote`, `numbers`, `pages`, each
     `{ default = true }`.
   - new `[pack.ms-office]`: `microsoft-excel`, `microsoft-powerpoint`,
     `microsoft-word`, each `{ default = true }`.
   - new `[pack.disk]`: `daisydisk = { default = true }`,
     `disk-doctor = {}`, `disk-map = {}`, `duplicate-detective = {}`.
   - new `[pack.media]`: `vlc = { default = true }`, `final-cut-pro = {}`,
     `handbrake = {}`.
   - `chat`: [telegram], `whatsapp = { default = true }`.
   - new `[pack.notes]`: `notion = { default = true }`,
     `obsidian = { default = true }`, `kindle = {}`, `cardhop = {}`,
     `pdf-protector = {}`, `pdf-squeezer = {}`.
   - `cli-extras`: `btop = { default = true }`, [dasel],
     `imagemagick = { default = true }`, [pstree], `tree`, `watch`, `wget`,
     `yq`, each `{ default = true }`.
     Move a table when its position differs from the old file order. Keep
     the file comments in the style of the file.
2. **`macos/examples/macos.toml`** — add one empty `[tools.<name>]` table
   for each of the 52 tools of U1 and U2, in the place and order that the
   file uses for the selected tools (D7). After this edit the file selects
   the same 73 tools outside pack `base` as
   `MS/tests/setup-sync/sample-config/macos-setup.toml`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- Both files parse: `mise x -- yq -p toml -o json . <file> >/dev/null`.
- For each of the 14 old packs, the member list of `macos/packs.toml`
  equals the old `tools` list, in the same order; the members with
  `default = true` equal the old `default` list, or all members when the old
  pack has no `default`. Compare with `yq` over both files (for example
  `.pack.<name> | keys` against `.pack.<name>.tools`).
- The tool names of `[tools.*]` in `macos/examples/macos.toml` outside pack
  `base` equal the `tools` list of
  `MS/tests/setup-sync/sample-config/macos-setup.toml` (D9). A `[tools.*]`
  table of a base tool (an answer) is not part of the compare.

## Guardrails

- Touch only `macos/packs.toml` and `macos/examples/macos.toml`. U1 and U2
  own the tool folders.
- Do not change a member, a value or an answer that an earlier plan wrote,
  except the position of a pack table.
- Write the comments in ASD-STE100 Simplified Technical English.
- Delete with `rm`, never `git rm`.

## Commit

`feat: complete the Mac packs and the sample selection` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
