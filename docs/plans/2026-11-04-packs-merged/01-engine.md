# U1 — Engine: the merged packs model

- **Read first:** every owned file, top to bottom, before editing
  (`sync-engine/select`, `sync-engine/repo-sync`).
- **Lazy-load:** `tools/packs.toml` (U2 rewrites it),
  `docs/design/packs.md` (read only), `sync-engine/common`.

## Ruling

From the design rulings of 2026-11-04:

> - One merged packs model for both sides (user's proposal, 2026-11-04):
>   - A pack is a group; the user selects the whole pack or single tools of it.
>   - The selection is `[packs.<name>]` and `[tools.<name>]` tables on both
>     sides. setup-sync drops `tools = [...]` in `macos-setup.toml`.
>   - The engine expands `base` and each selected pack, nested `p:<pack>`
>     too (the repo model).
>   - A tool can be in two packs; the engine dedupes. The guide shows it once.
>   - `packs.toml` shape (both sides): each key of `[pack.<name>]` is a member,
>     `<tool> = { default = true }` or `<tool> = {}` (implicit `false`). A
>     nested pack is a quoted `"pack:<name>"` key (not `p:`). The only entry
>     key is `default`; any other stops with exit 2. No pack-level keys (no
>     `depends`, `tools`, `default = [...]`, `required`, `description`).
>   - `[packs.<name>]` adds each `default = true` member (option A). A new
>     `default = true` member reaches every consumer of the pack; a
>     `default = false` member arrives only with `[tools.<name>]`. Still open,
>     as in `docs/design/packs.md`: no way to drop one default member.
>   - No `depends` in a pack entry: a tool's dependencies stay in its
>     `tools.toml`, else `[tools.<name>]` alone misses them.

From this plan:

D5: "Every [pack.*] table of packs.toml, expanded or not: the shape of each
entry, the old keys, a pack named like a tool folder, a loop over every
\"pack:<name>\" key (default true or not), and an unknown member (code 3).
Thus the test of bootstrap finds an error in its own packs.toml before a
consumer selects that pack."

D6: "A key depends, required, tools or default whose value is a list:
'tools/packs.toml: [pack.<name>] <key> is the old format - write each member
as <tool> = { default = true } or <tool> = {}' (code 2). A key that starts
with p:: '... \"p:<x>\" is the old form - write \"pack:<x>\" = { default =
true }' (code 2). Any other value that is not a table: '... <key> must be a
table, { default = true } or {}' (code 2). An entry key other than default, or
a default that is not a boolean: code 2 with the key in the message. The
message of an unknown member names the pack (code 3)."

D7: "\"pack:<x>\" = {} does not expand when [packs.<outer>] adds the outer
pack; the nested pack arrives only through its own [packs.<x>] entry. The loop
check still reads it (D5)."

D8: "As today: the env of the [packs.<name>] entry goes to each default
member, also through a nested pack; without env, each member gets its own
default_envs. The pack base gives each member its own default_envs. A
dependency gets the environments of each tool that needs it."

D11: "In file order, so the tool order and the pre-commit hook order of today
stay. An awk scan of each [pack.<name>] section reads the keys in file order
(one <key> = {...} line per member; a member that spans lines stops with code
2). gomplate still validates the shape. tools/packs.toml lists the base
members in the order of today: mise, repo-sync, git, pre-commit, dprint,
linter, vscode, then \"pack:security\"; security lists gitleaks, grype,
osv-scanner. A test proves that the tool order and the pre-commit hook order
do not change."

## Edits

1. **`sync-engine/select`**
   - `_se_selected`: one gomplate call over the packs file prints, for each
     `[pack.<name>]`, a `pack <name>` line and one line per key:
     `member <pack> <key> <default>` for a table entry (`<default>` is
     `true`, `false`, or `bad:<detail>` for an entry key other than
     `default` or a non-boolean `default`), and `badvalue <pack> <key> <kind>`
     for a value that is not a table. Then validate every pack (D5, D6), in
     this order: old keys and bad values (2), `p:` keys (2), entry keys (2),
     a pack named like a tool folder (2, the message of today), unknown
     members (3: a plain key that is not a tool folder, or a `pack:<x>`
     key whose `<x>` is not a pack), loops over every `pack:` key (2, the
     message of today: `a pack loop: a -> b -> a`).
   - The member order (D11): an awk scan of each `[pack.<name>]` section of
     the packs file reads the member keys in file order, one
     `<key> = { ... }` line per member (a quoted key keeps its quotes off). A
     member whose value spans lines stops with code 2:
     `<shown>: [pack.<name>] <key> must be on one line`. The gomplate call
     validates the shape; the awk scan gives the order. `_se_pack` expands
     the members in that order.
   - `_se_pack`: expand only the members with `default` `true`; a
     `pack:<x>` member with `true` expands `<x>` with the same env and
     source (D7, D8). Print each tool as `<tool> <env,...|-> <source>`, as
     today. The walk and the env rules dedupe a tool that comes more than
     once, as today.
   - The unknown `[packs.<x>]` of a values file keeps its message and code
     3 (`pack '<x>' is not in bootstrap`).
   - Remove every reference to `depends`, `required` and `p:` as the packs
     format.
2. **`sync-engine/repo-sync`** — the header comment: the packs format
   (`tools/packs.toml` with members and `default`, `"pack:<name>"`), the new
   code 2 and code 3 problems. Change the call of `_se_selected` only when
   its signature changes. The `[packs.<name>]` table of a values file keeps
   the pack name without a prefix.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n sync-engine/select sync-engine/repo-sync` passes.
- `grep -n '"depends" "required"\|p:\*\|#p:' sync-engine/select` finds no
  reader of the old format; only the old-key message names `depends`.
- `test:repo-sync` runs green only after U2 and U3; it is U3's gate.

## Guardrails

- Do not touch `tools/**` (U2) or the test task (U3).
- Do not change the selection of `[tools.<name>]`, the depends walk, or the
  env rules beyond D8.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- gomplate runs only through `_se_gomplate`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: read the merged packs model with members and default flags` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
