# U2 — Packs and sample values for shells and identity

- **Read first:** `macos/packs.toml` and `macos/examples/macos.toml`, top to
  bottom, and the sample answers in
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tests/setup-sync/sample-config/`
  (`git.toml`, `ssh.toml`, `gh.toml`, `macos-defaults.toml`).
- **Lazy-load:**
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/packs.toml`.

## Ruling

From the design rulings of 2026-11-04, verbatim:

> - `packs.toml` shape (both sides): each key of `[pack.<name>]` is a member,
>   `<tool> = { default = true }` or `<tool> = {}` (implicit `false`). A
>   nested pack is a quoted `"pack:<name>"` key (not `p:`). The only entry
>   key is `default`; any other stops with exit 2. No pack-level keys (no
>   `depends`, `tools`, `default = [...]`, `required`, `description`).

> - Values row, item 2: one values file on each side. The values of a tool
>   are keys of its `[tools.<name>]` table, which also selects it; a table
>   for a base tool or a dependency only gives values. No `<tool>.toml`
>   files, no Mac `[overrides.<tool>]`, no "a `<tool>.toml` does not select"
>   rule. Repo: `.config/bootstrap.toml`, `[tools.<name>]` keeps `env` as its
>   only key until item 3. Mac file: `macos.toml` at the root of the private
>   `macos-setup` repo (the user wants the type in the file name; `macos`
>   matches the side name). Repo file keeps `bootstrap.toml` (a rename needs
>   a consumer migration for no function).

Plan decision D7: "No. test:macos-tools (P10) now meets the init stage (bash,
zsh, fish), so its case 1 expects exit 4, then 0. U2 adds the answers of the
questions of git, ssh, gh and macos-defaults to the example, from
MS/tests/setup-sync/sample-config/."

P10 D15 (`docs/plans/2026-11-04-macos-entry-tool`), the reason for the member
order: "No. U2 adds macos/packs.toml to the excludes of the root dprint.jsonc.
The global dprint config sorts TOML keys, and the engine reads the members of
a pack in file order, so the member order must stay as written (as P04 does
for repo/packs.toml)."

## Edits

1. **`macos/packs.toml`**: in `[pack.base]`, add these members, each
   `= { default = true }`: `bash`, `fish`, `gh`, `git`, `macos-defaults`,
   `network`, `ssh`, `zsh`. Order: the members of `[pack.base]` (and of any
   other pack this unit edits) keep the order of the old `tools` list in
   `/Users/virajpatel/Projects/github.com/virajp/macos-setup/tools/packs.toml`:
   insert each new member at its old position relative to the members
   already there (P10 D15). Pack
   base now has the 19 members of the old base. Change no other table.
2. **`macos/examples/macos.toml`**: add one table for each tool with a
   question, with a comment line above each key that names its question, and
   the sample value byte for byte:
   - `[tools.git]`: `email_github`, `email_gitlab`, `name`, `signing_key`
     (from `git.toml`).
   - `[tools.ssh]`: `user = "sampleuser"`.
   - `[tools.gh]`: `editor = "vim"`.
   - `[tools.macos-defaults]`: `dock_orientation = "right"`.
     Keep the tables in byte order of the tool name, after
     `[tools.diff-so-fancy]` and `[tools.macos-setup]` where the order puts
     them.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- Both files parse: `mise x yq@latest -- yq -p toml -o json '.' <file>`.
- `mise x yq@latest -- yq -p toml '.pack.base | keys | length' macos/packs.toml`
  prints 19.
- `mise x yq@latest -- yq -p toml '.pack.base | keys' macos/packs.toml`
  lists the 19 members in the order of the old `tools` list of pack base.
- `mise x yq@latest -- yq -p toml '.tools.git | keys | length'
  macos/examples/macos.toml`
  prints 4.

## Guardrails

- Touch only `macos/packs.toml` and `macos/examples/macos.toml`.
- Write the sample values only: no real name, email or key.
- Delete with `rm`, never `git rm`.

## Commit

`feat: complete pack base and add the sample answers of the identity tools` —
written by the orchestrator after the unit gate and the wave gate, not by the
unit.
