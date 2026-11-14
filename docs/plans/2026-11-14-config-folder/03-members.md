# U3 — Members: siblings, url and the home back-link

- **Read first:** `docs/design/repo.md` (members), `docs/sync/index.md`
  (topologies), `cli/internal/schema`, the migration U2 added in
  `cli/internal/oldform`, the code that reads `[member.*]` (grep `member` in
  `cli/internal`).
- **Lazy-load:** `repo/examples/repo-submodules/`,
  `repo/examples/repo-siblings/`, `repo/examples/monorepo/`.

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`.

- **D7:** `[member.*]` becomes `[members.*]` with `url` and
  `type = "sibling"`; each submodule and sibling member gets a `[home]`
  back-link; members are read from `repo.toml`, never from a vwf file.
- **D15:** `[members.<name>]` has `folder`, `type`
  (`submodule | folder | sibling`), `url`, `label` and `worktree_install`.
  `topology` (`repo | monorepo | multi-repo`) and `topology_reason` sit at the
  root of `repo.toml`.
- **D62:** `--all` and `--<member>` are flags of the `setup:configs` task (U4),
  not of the engine. There is no `init --all`: the user runs `bootstrap init` in
  each member.
- **D63, D89:** `setup:configs` passes `BOOTSTRAP_HOME=<path to the home repo>`
  and `BOOTSTRAP_MEMBER=<name>` when it runs in a member. The member's own
  `bootstrap sync` then writes `[home]`.
- **D56, D57:** `[home]` holds `folder` (the path from the member back to the
  home), `type` (from the home's `[members.<name>].type`) and `url` (the home's
  `repo_url`, left out when the home has none).
- **D30:** `[home]` is written only into a member that has
  `.config/bootstrap/`. The task skips a member with no folder or the old form
  (D62, U4).
- **D55:** `folder` members (part of a monorepo tree) get no `[home]` and no
  warning.
- **D57, D79:** `root_only` tools are skipped only in a member whose `[home]`
  says `type = "submodule"`. A sibling member is its own root, so they run
  there.
- **D81:** U3 and U4 run in wave 3 in `teams` mode and agree the key names; U5
  (wave 4) uses the names they commit.

## Edits

1. **Schema.** Add `[members.<name>]`, `[home]`, `topology` and
   `topology_reason` to the `repo.toml` table (D15, D56). `url` is required for
   `type = "sibling"`. `folder` of a sibling is a path relative to the home repo
   root and may start with `../`.
2. **Migration.** U2's hook: `[member.<name>]` becomes `[members.<name>]` with
   the same keys. A repo's `.config/vwf.yaml` is not read: the user or a later
   vwf migration (R4) adds sibling members with `bootstrap config set`.
3. **The engine side of members:** `MEMBERS` and `worktree_install` read
   `[members.*]`. A `sibling` member is handled like a `submodule` member, in
   its own folder, when the folder exists; a missing folder is skipped with an
   info line that names the member and its `url`.
4. **The back-link** (D30, D56, D57, D63, D89): when `sync` runs with
   `BOOTSTRAP_HOME` and `BOOTSTRAP_MEMBER` set, it writes `[home]` into the
   member's own `repo.toml`, only when `[home]` is absent or differs. The write
   appears in the change list like any other write.
5. **`root_only`** (D57): skipped only when `[home].type` is `submodule`.
6. **Commit scopes.** The default scopes (today from the members) come from
   `[members.*]` when `[commits].scopes` is absent.
7. **Tests.** Go tests for the schema rules, the migration, a sibling member
   present and missing, the back-link write with and without `url`, and
   `root_only` in a submodule and a sibling member.

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run test:go` passes.
- `grep -rn 'vwf' cli/ --include=*.go` finds no read of a vwf file.

## Guardrails

- Touch nothing outside `cli/`. U4 owns `repo/tools/` (including
  `setup:configs` and `code:worktrees`); U5 owns `repo/examples/`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: members in repo.toml with siblings and a home back-link`
