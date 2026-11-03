# U2 — Mac packs file and sample values file

- **Read first:** `dprint.jsonc`, `repo/packs.toml` (the format after P04) and
  `repo/examples/polyrepo/bootstrap.toml` (the model of a sample values file).
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

> - Values row, item 6: `bootstrap_commit` on both sides (Mac: in
>   `macos.toml`, replaces `macos_setup_commit`), same write rules. Each Mac
>   sync with a new commit changes `macos.toml` in the macos-setup clone; the
>   user commits it (`ops: sync bootstrap`). The Values row is closed.

Plan decision D3: "[questions.config_path] (question and example
~/Projects/github.com/janedoe/macos-setup) and [mise.env]
MACOS_SETUP_CONFIG_PATH = "<% index .config `config_path` | required
`macos-setup: set config_path in [tools.macos-setup] of macos.toml` |
regexp.Replace `^~` .env.HOME %>". A ~ at the start becomes $HOME, because
mise does not expand ~ in [env]. The first run needs the variable from the
shell; each later shell gets it from mise."

D15: "No. U2 adds macos/packs.toml to the excludes of the root dprint.jsonc.
The global dprint config sorts TOML keys, and the engine reads the members of
a pack in file order, so the member order must stay as written (as P04 does
for repo/packs.toml)."

## Edits

1. **`macos/packs.toml`** (new). A comment in ASD-STE100 at the top: the packs
   of the Mac tools in `macos/tools/`; the shape of a pack (each key is a
   member, `{ default = true }` or `{}`, a nested pack is `"pack:<name>"`);
   pack `base` is always added and must hold `macos-setup`; the import plans
   add the other members. Then:

   ```toml
   [pack.base]
   macos-setup = { default = true }
   ```

2. **`macos/examples/macos.toml`** (new). The sample values file of the Mac
   side; it is not shipped, so it has no sync header. A comment in
   ASD-STE100 at the top: the file lives at the root of the private
   `macos-setup` clone; `MACOS_SETUP_CONFIG_PATH` names its folder; each
   `[tools.<name>]` table selects a tool and gives its values; each
   `[packs.<name>]` table adds the `default = true` members of the pack; the
   sync writes `bootstrap_commit`. Then:

   ```toml
   bootstrap_commit = ""

   [tools.macos-setup]
   config_path = "~/Projects/github.com/janedoe/macos-setup"
   ```

   with a one-line comment above `config_path` that names its question.

3. **`dprint.jsonc`** (D15): add `"macos/packs.toml"` to the `excludes` list
   of the root `dprint.jsonc`. Keep each other entry and the comment. If an
   earlier plan already lists `repo/packs.toml` there, add the new entry next
   to it.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x yq@latest -- yq -p toml -o json '.' macos/packs.toml` and the same
  for `macos/examples/macos.toml` parse.
- `grep -c '^\[pack\.' macos/packs.toml` is 1.
- `grep -n '"macos/packs.toml"' dprint.jsonc` finds the exclude line.

## Guardrails

- Touch only `macos/packs.toml`, `macos/examples/macos.toml` and
  `dprint.jsonc`.
- Write no pack-level key and no member other than `macos-setup`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the Mac packs file and the sample macos.toml` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
