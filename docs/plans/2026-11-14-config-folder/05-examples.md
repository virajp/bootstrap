# U5 — Examples in the new folder form

- **Read first:** every values file under `repo/examples/`,
  `cli/internal/schema` (read only) and the commits of U3 and U4 on this branch
  (the key names).
- **Lazy-load:** `docs/sync/index.md` (the example shapes),
  `.config/mise/tasks/test/repo` (`example_consumer`, how a temp consumer gets
  its members).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`. D15, the key map, is quoted in
`01-folder-schema.md`.

- **D11, D17:** `merge_model = "pr"` → `landing = "pr"`; `merge_model =
  "direct"` (or absent) with `direct_commit = true` → `direct-commit`; any
  other → `direct-merge`. `merge_from` and `merge_never` move unchanged.
- **D16:** `core_branches` lists `main` and `develop`, in that order, that
  exist in git.
- **D7, D56, D57:** members are `[members.<name>]` with `type = "submodule" |
  "folder" | "sibling"` and `url`; a submodule or sibling member has `[home]`
  with `folder`, `type` and `url`.
- **D34:** the layout stays flat: `repo/examples/<x>/repo.toml` and
  `repo/examples/<x>/tools.toml`.
- **D35, D66, D82:** to migrate or check an example, a temp copy is laid out as
  `.config/bootstrap/` from the flat files and made a git repo. It starts with
  `git init -b <the first branch the example names>` and one commit, then gets
  any other branch it names. Each example gets only the branches its tables or
  comments name: `repo-submodules` and its `iac` member get `develop`;
  `repo-siblings` gets `main`; an example that names none gets `main` and
  `develop`.
- **D67:** before `check`, the temp copy gets its member folders and
  submodules, as `test/repo`'s `example_consumer` does.
- **D45, D81:** U5 runs in wave 4, after U3 and U4, and uses the key names they
  committed.

## Edits

1. Each example's flat `bootstrap.toml` (11 files: `polyrepo`,
   `polyrepo-main`, `polyrepo-develop`, `monorepo`, `repo-submodules` and its
   submodules, `repo-siblings` and its siblings, and the rest) becomes a flat
   `repo.toml` and `tools.toml` in the same folder (D34), each with
   `format = 3`, keys per D15. Produce them with the built engine's
   `bootstrap init --yes` migration in a temp copy laid out per D35, then copy
   the result back flat and review it by hand; keep each example's comments.
2. `repo-siblings`: the base lists its siblings as `[members.<name>]` with
   `type = "sibling"`, `folder = "../<name>"` and a `url`; each sibling has
   `[home]`. The example no longer relies on `.config/vwf.yaml`.
3. Any README or comment inside `repo/examples/` that names the old path is
   updated.

## Verification

- `mise x -- mise run code:format` passes.
- `find repo/examples -name bootstrap.toml` finds nothing.
- With the built engine, `bootstrap check --clone <worktree>` passes in a temp
  copy of each example, laid out per D35 with its members per D67.

## Guardrails

- Touch nothing outside `repo/examples/`.
- Delete with `rm`, never `git rm`.

## Commit

`docs: examples use .config/bootstrap/`
