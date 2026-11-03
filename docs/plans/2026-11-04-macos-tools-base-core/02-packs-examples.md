# U2 — Packs and sample values for the base core

- **Read first:** `macos/packs.toml` and `macos/examples/macos.toml`, top to
  bottom.
- **Lazy-load:**
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/packs.toml`.

## Ruling

From the design rulings of 2026-11-04, verbatim:

> - `packs.toml` shape (both sides): each key of `[pack.<name>]` is a member,
>   `<tool> = { default = true }` or `<tool> = {}` (implicit `false`). A
>   nested pack is a quoted `"pack:<name>"` key (not `p:`). The only entry
>   key is `default`; any other stops with exit 2. No pack-level keys (no
>   `depends`, `tools`, `default = [...]`, `required`, `description`).

> - `[packs.<name>]` adds each `default = true` member (option A). A new
>   `default = true` member reaches every consumer of the pack; a
>   `default = false` member arrives only with `[tools.<name>]`. Still open,
>   as in `docs/design/packs.md`: no way to drop one default member.

> - Values row, item 2: one values file on each side. The values of a tool
>   are keys of its `[tools.<name>]` table, which also selects it; a table
>   for a base tool or a dependency only gives values. No `<tool>.toml`
>   files, no Mac `[overrides.<tool>]`, no "a `<tool>.toml` does not select"
>   rule. Repo: `.config/bootstrap.toml`, `[tools.<name>]` keeps `env` as its
>   only key until item 3. Mac file: `macos.toml` at the root of the private
>   `macos-setup` repo (the user wants the type in the file name; `macos`
>   matches the side name). Repo file keeps `bootstrap.toml` (a rename needs
>   a consumer migration for no function).

Plan decisions:

D7: "No. test:macos-tools (P10) renders each new tool through
macos/examples/macos.toml, and it fails when a tool folder is not selected. U2
selects diff-so-fancy in the example, because no base tool depends on it
before P12."

D11: "[pack.essentials], as diff-so-fancy = {}, because the old default list
of essentials omits it. git brings it into each Mac through depends (P12)."

P10 D15 (`docs/plans/2026-11-04-macos-entry-tool`), the reason for the member
order: "No. U2 adds macos/packs.toml to the excludes of the root dprint.jsonc.
The global dprint config sorts TOML keys, and the engine reads the members of
a pack in file order, so the member order must stay as written (as P04 does
for repo/packs.toml)."

## Edits

1. **`macos/packs.toml`**:
   - In `[pack.base]`, add these members, each `= { default = true }`: `cli`,
     `fnox`, `fonts`, `homebrew`, `mas`, `mise`, `node`, `python`,
     `starship`, `touchid-sudo`. Keep `macos-setup`. Order: the members of
     `[pack.base]` (and of any
     other pack this unit edits) keep the order of the old `tools` list in
     `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/packs.toml`:
     insert each new member at its old position relative to the members
     already there (P10 D15).
   - Add `[pack.essentials]` with `diff-so-fancy = {}` (D11).
2. **`macos/examples/macos.toml`**: add `[tools.diff-so-fancy]` (an empty
   table) with a one-line comment: the table selects the tool, which is not
   a base tool (D7). No tool of this group has a question, so add no answer.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- Both files parse: `mise x yq@latest -- yq -p toml -o json '.' <file>`.
- `mise x yq@latest -- yq -p toml '.pack.base | keys | length' macos/packs.toml`
  prints 11.
- `mise x yq@latest -- yq -p toml '.pack.base | keys' macos/packs.toml`
  lists the members in the order of the old `tools` list of pack base.

## Guardrails

- Touch only `macos/packs.toml` and `macos/examples/macos.toml`.
- Write no pack-level key; each entry has only `default`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the base-core Mac tools to the packs and the sample values` —
written by the orchestrator after the unit gate and the wave gate, not by the
unit.
