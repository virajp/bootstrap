# U2 — Teach repo-sync the tables, the packs and the environment rules

- **Wave:** 2
- **Depends on:** U1
- **Owns:** `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`,
  `tools/repo-sync/.config/mise/tasks/setup/configs/_default`,
  `tools/repo-sync/.config/mise/tasks/setup/configs/sync`
- **Model:** opus
- **Kind:** edit
- **Read first:** the three owned files, top to bottom; `tools/packs.toml`;
  `tools/mise/tools.toml`; `tools/git/tools.toml`.
- **Lazy-load:** `tools/mise/.config/mise/tasks/_scripts/helpers` (for
  `members`).

## Ruling

Quoted from `index.md`:

- D1: "`envs = ["dev", "ci"]` means that this tool can be installed in either
  or both these environments." "A dependency inherits the environments of its
  parent; a tool that the consumer selects gets the environments that the
  consumer chooses."
- D2: "`[tools.<folder>]` and `[packs.<name>]` tables. 'keys are tool folders
  and only env'. `[packs.*]`: 'Same applies to packs as well'."
- D3: "An entry without `env` gets `default_envs`."
- D4: "'base uses default_envs'. Each tool of `base` gets its `default_envs`."
- D5: "`environments` limits every result. An `env` name that is not in
  `environments`, or that a mise tool of the tool does not allow in its
  `envs`, stops the sync with code 2. A dependency whose `envs` lacks an
  inherited environment stops the sync with code 2."
- D6: "`tools = [...]` (a list) or `packs = [...]` stops the sync with code 2;
  the message names the table format. No upgrade path."
- D7: "The `base` pack name is fixed in `repo-sync` and is always added. Empty
  or absent `[tools]` is valid."
- D8: "A pack loop stops the sync with code 2 and shows the loop
  (`a -> b -> a`). A pack name that is also a tool folder stops it with code
  2. An unknown pack stops it with code 3. The `[packs.*]` keys of the values
  file have no `p:` prefix."
- D15: "`_rs_merge_model` uses the `git` tool path; `setup/configs/sync` loads
  `tools/repo-sync/.../repo-sync`; the `setup-all` and `setup-<submodule>`
  aliases go to the `mise` tool. `_base` goes away."
- D16: "`repo-sync` skips `pre-commit autoupdate` when
  `REPO_SYNC_NO_AUTOUPDATE=1`, so the tests need no network."

## Edits

1. **Read the values file** (`_rs_resolve`, and a small reader next to
   `_rs_list`): read the keys of `[tools]` and `[packs]` and the `env` list of
   each entry. If `tools` or `packs` is a list, stop with code 2 and a message
   that shows the table format, for example `[tools.node] env = ["dev"]`.
2. **Expand the packs** (a new function, used by `_rs_resolve` and by
   `_rs_tool_extensions`): read `<clone>/tools/packs.toml` with gomplate
   (`-c p=...`). Start from `base` plus each `[packs.*]` key. For each pack add
   its `depends` and its `required`; a name with `p:` is a pack and is
   expanded in turn. Stop with code 2 and the chain (`a -> b -> a`) for a loop.
   Stop with code 2 if a pack name is also a folder of `tools/`. Stop with code
   3 for an unknown pack. Each tool of a pack gets the `env` of the pack entry,
   or else its own `default_envs`; each tool of `base` gets its own
   `default_envs`.
3. **Environments of each tool.** A tool of `[tools.*]` gets its `env`, or
   else its `default_envs`. A dependency of a tool (the `depends` key, and the
   tool of each mise `depends`) inherits the environments of each tool that
   needs it, until nothing changes. Then keep only the names of
   `environments`. Each mise tool of a tool gets the environments of its tool.
   Remove the `_RS_DIRECT` rule (dev only for a dependency).
4. **The errors of D5,** each code 2 with a message that names the tool, the
   mise tool and the environment:
   - an `env` name of the values file that is not in `environments`;
   - an environment of a tool that one of its mise tools does not allow in its
     `envs` (a mise tool without `envs` allows `["dev"]`);
   - an inherited environment that a dependency does not allow.
5. **Read `default_envs`** in `_rs_manifest`, as a necessary key of each
   `tools.toml`.
6. **Fixed paths.** `_rs_mise_render`: remove `folder=_base`; give the
   `setup-all` and `setup-<submodule>` aliases to the `mise` tool.
   `_rs_merge_model`: use `git/.config/mise/tasks/code/merge`.
   `setup/configs/sync`: load
   `${REPO_SYNC_CLONE}/tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`,
   and change its comment "The tools come from `tools`".
7. **Template context.** `.tools` is the resolved list of tools, packs
   expanded. If gomplate rejects an alias with a hyphen, for example
   `tool_repo-sync`, replace each `-` with `_` in the alias, and use the same
   name where the template reads it.
8. **`REPO_SYNC_NO_AUTOUPDATE=1`** skips the `pre-commit autoupdate` step and
   prints one line that says so. Add it to the environment list in the header.
9. **Header comments** (lines 15-50 and the error text at the old line 230):
   describe the tables, the packs, `default_envs` and the rules above. Remove
   each mention of `base` as a tool and of `_base`.

## Verification

- `bash -n` passes for the three files.
- `mise x -- shellcheck -x -s bash` shows no new warning on the changed lines.
- `rg -n '\bbase\b' tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
  shows only the `base` pack.
- With a scratch values file `[tools]` empty and `environments = ["dev"]`, a
  dry run with `REPO_SYNC_CLONE` set to a scratch clone of the run branch and
  `REPO_SYNC_NO_AUTOUPDATE=1` exits 0 and lists the files of the ten `base`
  tools.
- `mise x -- mise run code:precommit` passes (the wave gate).

## Guardrails

- Touch only the three owned files.
- Write shell for bash 3.2 and 5.x: no namerefs, no `${var,,}`, no associative
  arrays.
- Under `set -e`, write a command substitution that can fail as
  `VAR=$(cmd) || VAR=''`.
- A failure must still leave every existing file unchanged.

## Commit

`feat: resolve the tool and pack tables in repo-sync`
