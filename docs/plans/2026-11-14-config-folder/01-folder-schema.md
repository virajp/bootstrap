# U1 — The folder model and the key schema

- **Read first:** `docs/design/bootstrap-config.md` (the whole file),
  `cli/internal/values/values.go`, `cli/internal/repo/repo.go`,
  `cli/internal/core/core.go:123-125` (the values path, D73),
  `cli/internal/repo/file.go` (about line 374), `cli/internal/repo/rules.go`
  (about line 360), `cli/internal/target/commit.go`,
  `cli/internal/oldform/oldform.go`.
- **Lazy-load:** `cli/internal/manifest/`, `cli/internal/misegen/templates.go`,
  `cli/internal/source/`, `cli/internal/cli/select.go`,
  `cli/internal/cli/init.go`.

## Ruling

D15, the key map:

> repo.toml: format, repo_name, product_name, repo_url, core_branches,
> [branches.<b>] (landing, merge_from, merge_never), [commits] (style, scopes;
> was [tools.git].commit_scopes), environments, bootstrap_commit, [worktree]
> files (was [tools.mise].worktree_files), topology, topology_reason,
> [members.<name>] (folder, type submodule|folder|sibling, url, label,
> worktree_install), [home] (folder, url) in a member, release_env,
> [deploy.<env>] (<project> = <url>). tools.toml: format, [tools.<name>] (env
> and tool values, files overrides), [overrides.tools.*] (was the overrides
> file), [overrides.mise.*], [overrides.mise_env.*]. releases.toml: format,
> <project> = NONE|PATCH|MINOR|MAJOR. repo-generated.tsv: unchanged content.
> Keys are snake case.

D8, verbatim:

> Go table + tool catalog (Recommended): repo.toml and releases.toml: one Go
> table in a new package internal/schema (key path, type, enum values,
> required). tools.toml: a key is valid when the tool exists in the clone's
> catalog and the key is one of its questions, defaults or env; overrides keep
> today's rules. `check` uses the same schema, so an unknown key in repo.toml
> is now code 2.

D5, verbatim:

> Merge into tools.toml: Its [tools.*] overrides move into tools.toml, under an
> [overrides.tools.*] table. One file fewer, but tools.toml mixes selection and
> overrides.

D9, verbatim:

> Write only: It writes the value and prints "run bootstrap sync" when the key
> affects generated files. Simpler, but the repo is out of step until the next
> sync.

D10 (the releases schema part): one key per project, `NONE|PATCH|MINOR|MAJOR`.

D11: `[branches.<b>]` holds `landing = "pr" | "direct-merge" | "direct-commit"`,
`merge_from` and `merge_never`; `core_branches` at the root lists the branches.

D12: `BOOTSTRAP_VALUES` names the `.config/bootstrap/` folder.

D4 (this unit's part): the new files carry `format = 3`.

### Rulings review, round 1

These rulings override any earlier text of this file that they contradict.

- **D16** — The migration takes main and develop, in that order, but only the
  branches that exist in git (locally or on origin), the rule of docs/sync
  Step 2. core_branches is a required key of repo.toml. With neither branch,
  init stops with code 2.
- **D17** — direct-merge, today's default (direct with no direct commit).
- **D18** — User: "No overrides for keys in `repo.toml`, changes must be made
  directly in that file (ideally using bootstrap cli)". The migration writes
  such an override's value directly into repo.toml (the override wins over the
  old file's value). [overrides.tools.*] keeps only keys that stay in [tools.*];
  the schema refuses an [overrides.tools.*] entry for a key that lives in
  repo.toml.
- **D19** — config reads the BOOTSTRAP_VALUES folder when it is set. During init
  and sync the engine points BOOTSTRAP_VALUES at the temp folder of the planned
  layout, so scripts read the new keys even mid-migration. With BOOTSTRAP_VALUES
  set, config skips the old-form stop.
- **D21** — User: "A config get is not a failure so it must emit either the
  default (with status showing default) or the actual value but never an error
  code. Other verbs/commands will continue raising error of missing config".
  config get of a valid key exits 0 and prints {"key": <path>, "value": <v>,
  "source": "file" | "default"} (the same three fields in toml and yaml); a key
  with no schema default prints value null with source default. Only an invalid
  key (code 2) or a read failure (code 1) is an error. Scripts read .value.
- **D27** — unset of a required key exits 2. A whole table (members.<n>,
  branches.<b>) may be unset. Any set or unset whose result breaks a schema
  rule, the cross-key rules of D31 included, exits 2 and writes nothing.
- **D31** — Each core branch must exist (locally or on origin). Each
  [branches.<b>] must name a core branch. If main is direct-commit, develop
  (when it is a core branch) must be direct-commit too, today's rule in the new
  terms. The schema enforces the second and third as cross-key rules (D27).
- **D32** — An enum with one value, conventional, the default. Nothing reads it
  in R1a; it is the place for the commit model the user named in the design
  review.
- **D33** — U1 cites the selection model's code it builds on:
  cli/internal/core/core.go:123-125 (the values path) and
  cli/internal/core/plan.go:314 and :383 (the temp values copy), beside the
  main-checkout lines.

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D48** — The unborn branch of HEAD counts as existing, so a fresh init writes
  core_branches = ["<that branch>"].
- **D51** — "NONE".
- **D60** — format, repo_name and core_branches.

### Rulings review, round 3

These rulings override any earlier text of this file that they contradict.

- **D71** — The TUI shows the folder migration as one change, "migrate to
  .config/bootstrap/", applied by the same core migration; every write goes to
  the new files, the TUI's Apply included (U1 Edit 3).
- **D73** — cli/internal/core/core.go:123-125 replaces "cli/internal/cli/cli.go
  (prepare, about line 438)".
- **D74** — None. repo-generated.tsv keeps its content; only the three TOML
  files carry format = 3.

### Rulings review, round 4

These rulings override any earlier text of this file that they contradict. D20
no longer has a version check (D84).

- **D86** — env and files.* are always valid. U4 declares plugins.user and
  plugins.project as defaults (empty tables) in repo/tools/claude/tools.toml, so
  the catalog rule covers them.
- **D87** — A fresh init writes the main and develop that exist; with neither,
  the current branch, born or unborn, of any name. D48 allows any name.
- **D88** — Every command except init exits 2 with "Run bootstrap init"; init
  creates the missing file.
- **D90** — Every repo.toml key except deploy.*, release_env, topology,
  topology_reason, bootstrap_commit and format. No releases.* key.

## Edits

1. **One place for the paths.** Add the folder and file names as constants in
   one package (for example `cli/internal/values`): `.config/bootstrap/`,
   `repo.toml`, `tools.toml`, `releases.toml`, `repo-generated.tsv`. Replace
   every literal `.config/bootstrap.toml`, the derived overrides path and
   `.config/repo-generated.tsv` in the engine with them. `BOOTSTRAP_VALUES`
   now names the folder (D12).
2. **Read the folder.** The engine reads `repo.toml` and `tools.toml` and builds
   the same in-memory model that the one file gives it today, so `sync`,
   `check`, `add`, `remove`, `replace`, `detect`, `list` and `questions` work
   unchanged on the new layout. `[overrides.tools.*]` in `tools.toml` replaces
   `ReadOverrides` on the old overrides file. Tool code that read a moved key
   (`[tools.git].commit_scopes`, `[tools.git.branch.*]`,
   `[tools.mise].worktree_files`, `[member.*]`) reads its new place in
   `repo.toml`. The gomplate context that templates see keeps its names, or the
   unit reports each name it must change under `DECIDED`.
3. **Write the folder.** Every write the engine makes goes to the new files:
   `bootstrap_commit` (`target.Env.RecordCommit`) into `repo.toml`; tool tables
   and answers (`init`, `add`, `remove`, `replace`, `detect --add`) into
   `tools.toml`; the manifest into `.config/bootstrap/repo-generated.tsv`. A new
   repo gets both files with `format = 3`, created by `init` as today's single
   file is created.
4. **The schema package.** Add `cli/internal/schema`: one Go table per file kind
   for `repo.toml` and `releases.toml` (key path, type, enum values, required,
   and a flag "a sync reads it", which D9 uses), and a function that checks a
   `tools.toml` key against the clone's tool catalog. The required keys of
   `repo.toml` are `format`, `repo_name` and `core_branches` (D60). A key path
   with a name segment
   (`branches.<b>`, `members.<name>`, `deploy.<env>`) matches any name.
5. **`check` uses the schema.** An unknown key, a wrong type or a value outside
   its enum in any of the files is code 2, with the file, the key and the
   reason.
6. **Tests.** Update every Go test and every `testdata` fixture in `cli/` to the
   folder layout, and add table tests of the schema (a valid file, an unknown
   key, a bad enum, a missing `repo_name`, a `tools.toml` key that the catalog
   does not know).
7. The migration from the old file, the `config` verb, `settings` and the
   members behaviour are not this unit's: U2 and U3 add them. Until U2 lands,
   a repo with only the old file is not supported by the engine in this
   worktree, and no test of this unit uses one.

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run test:go` passes, `TestEngineNamesNoTool` included.
- `grep -rn
  'bootstrap.toml\|bootstrap.overrides.toml\|config/repo-generated.tsv' cli/
  --include=*.go` finds only the constants of edit 1 and the old-form detection
  that `oldform` keeps.

## Guardrails

- Touch nothing outside `cli/`.
- Do not change `repo/`, `macos/`, `docs/` or `.config/`.
- The Mac side (`macos.toml`) is out of scope: keep it as it is.
- Keep comments in a TOML file through the existing yq path when the engine
  edits it; never rewrite a whole file through BurntSushi.
- Delete with `rm`, never `git rm`.

## Commit

`feat: values live in .config/bootstrap/ with a key schema`
