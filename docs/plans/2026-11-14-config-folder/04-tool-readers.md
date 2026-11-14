# U4 — Tool scripts read through bootstrap config get

- **Read first:** every file below before editing it; `cli/internal/schema`
  (read only) for the key names.
- **Lazy-load:** `docs/design/repo.md` (the branch model and the merge tasks),
  `repo/tools/git/tools.toml`, `repo/tools/mise/tools.toml`.

## Ruling

D6, verbatim:

> Through bootstrap config get (Recommended): Each script calls `bootstrap
> config get <file>.<key> --format json` instead of a yq read of a file path.
> One reader, schema-checked; the path can move again without touching tools.
> The scripts already run where the engine is installed.

D7, verbatim:

> All of it in R1a (Recommended): [member.*] becomes [members.*] with `url`
> and `type = "sibling"`, each member gets a [home] back-link, and
> code:worktrees and the --<member>/--all flags read repo.toml, never
> vwf.yaml. bootstrap then names no vwf file for topology, which R4 needs.

D11, the meaning of `landing`: `pr` lands only through a PR; `direct-merge`
merges without a PR and allows no direct commit; `direct-commit` allows direct
commits and direct merges. `core_branches` lists the branches.

D13, D81: in `teams` mode, agree the key names with U3 (wave 3); U5 (wave 4)
uses the names you commit.

### Rulings review, round 1

These rulings override any earlier text of this file that they contradict.

- **D16** — The migration takes main and develop, in that order, but only the
  branches that exist in git (locally or on origin), the rule of docs/sync
  Step 2. core_branches is a required key of repo.toml. With neither branch,
  init stops with code 2.
- **D17** — direct-merge, today's default (direct with no direct commit).
- **D19** — config reads the BOOTSTRAP_VALUES folder when it is set. During init
  and sync the engine points BOOTSTRAP_VALUES at the temp folder of the planned
  layout, so scripts read the new keys even mid-migration. With BOOTSTRAP_VALUES
  set, config skips the old-form stop.
- **D20** — The engine exports BOOTSTRAP_BIN (its own executable path) to every
  script and task it runs. A task mise runs uses ${BOOTSTRAP_BIN:-mise x --
  bootstrap}, today's pattern. With no engine, a script that needs a value exits
  2 with an install hint (D85); an old engine answers `config get` with a usage
  error, and the task stops with the hint (D84). There is no version check.
- **D21** — User: "A config get is not a failure so it must emit either the
  default (with status showing default) or the actual value but never an error
  code. Other verbs/commands will continue raising error of missing config".
  config get of a valid key exits 0 and prints {"key": <path>, "value": <v>,
  "source": "file" | "default"} (the same three fields in toml and yaml); a key
  with no schema default prints value null with source default. Only an invalid
  key (code 2) or a read failure (code 1) is an error. Scripts read .value.
- **D31** — Each core branch must exist (locally or on origin). Each
  [branches.<b>] must name a core branch. If main is direct-commit, develop
  (when it is a core branch) must be direct-commit too, today's rule in the new
  terms. The schema enforces the second and third as cross-key rules (D27).
- **D37** — U4 changes their sources under repo/tools/ (for example
  git/files/gitignore.merge.toml, the dprint and vscode merge specs). This
  repo's synced copies (.gitignore, dprint.jsonc, .vscode/extensions.json) stay
  out of scope until the user's hand migration.

### Rulings review, round 2

These rulings override any earlier text of this file that they contradict.

- **D42** — D20 wins: ${BOOTSTRAP_BIN:-mise x -- bootstrap}.
- **D43** — dev counts as 4 or newer, the same rule as CheckVersion
  (cli/internal/core/clone.go).
- **D44** — U4's Verification sets BOOTSTRAP_BIN to the worktree build.
- **D49** — A segment with a dot is double-quoted, as in yq:
  repo.branches."release.1".landing. Scripts always quote the branch segment.

### Rulings review, round 3

These rulings override any earlier text of this file that they contradict.

- **D62** — They are flags of the setup:configs task, not of the engine. There
  is no init --all: the user runs bootstrap init in each member. D54 reads:
  setup:configs --all skips an old-form member with the warning "member <name>:
  run bootstrap init there" and continues.
- **D63** — The member's own bootstrap sync. setup:configs passes
  BOOTSTRAP_HOME=<path to the home repo> when it runs in a member; the engine
  then writes [home] folder, url and type from it (type from the home's
  [members.<name>].type).
- **D64** — U4: repo/tools/mise/tasks/setup-configs.tmpl, setup.tmpl and the
  helpers library. members() reads `config get repo.members` (submodule and
  sibling members); the task skips an old-form member and continues; it passes
  BOOTSTRAP_HOME.
- **D65** — Not for repo.toml keys (D18). derive reads the branch keys from
  repo.toml only.
- **D68** — K1 needs U4's scripts in the clone, so U4's Verification runs K1,
  and U2's runs K3 only. The plan keeps K1 at the end.

### Rulings review, round 4

These rulings override any earlier text of this file that they contradict. D20
no longer has a version check (D84).

- **D81** — D13 now reads: wave 3 has two edit units (U3, U4) that align key
  names in teams mode; U5 uses the names they committed. D14 now reads: the
  review unit is in wave 6.
- **D84** — D20's version check is removed. A script calls config get; when an
  old engine answers with a usage error (exit 2), the task stops with the
  install hint. No version comparison.
- **D85** — The exit-2 rule applies only to scripts that need a value: git
  merge, derive, check, setup-ai, code-worktrees and setup-configs.
  code-lint-bootstrap keeps today's silent exit 0 when no engine exists.
- **D86** — env and files.* are always valid. U4 declares plugins.user and
  plugins.project as defaults (empty tables) in repo/tools/claude/tools.toml, so
  the catalog rule covers them.
- **D89** — setup:configs also passes BOOTSTRAP_MEMBER=<name>; the engine reads
  the type from the home's [members.<name>].

## Edits

Each read of the values file becomes a `bootstrap config get … --format json`
call
that reads `.value` (D21: an unset key prints its default with
`source: "default"`; exit 1 is a read failure and stops the task, as today).

1. **`repo/tools/git/tasks/merge`** (lines about 184-201): read
   `repo.branches.<b>.landing`, `merge_from`, `merge_never`. `pr` opens a PR;
   `direct-merge` and `direct-commit` merge directly.
2. **`repo/tools/git/scripts/derive`** (line about 18): derive the branch pair
   from `repo.core_branches`, and the protected branches (the
   `no-commit-to-branch` arguments) as every core branch whose `landing` is not
   `direct-commit`. It reads no overrides for these keys (D65).
3. **`repo/tools/git/scripts/check`** (line about 10): the same reads.
4. **`repo/tools/claude/tasks/setup-ai`** (lines about 85-89): read
   `tools.tools.claude.plugins.<scope>`.
5. **`repo/tools/mise/tasks/code-worktrees`**: read the members from
   `repo.members` and the home from `repo.home`; remove every read of
   `.config/vwf.yaml` and `.config/vwf-membership.yaml`.
6. **`repo/tools/mise/tasks/code-lint-bootstrap`** (lines about 5-59): the gate
   runs on changes under `.config/bootstrap/`.
7. **Comments and messages** that name the old path:
   `node/tasks/setup-deps-node`,
   `vscode/tasks/setup-vscode`, `mise/tasks/setup-configs.tmpl`,
   `mise/files/.config/mise.toml.tmpl`, `git/files/gitignore.merge.toml`,
   `git/files/.config/git-conventional-commits.yaml.tmpl` (the scopes come
   from `[commits].scopes`), `git/tasks/code-merge-main`,
   `pre-commit/files/.config/pre-commit-config.yaml.merge.toml`.
8. **The member loop (D64):** `repo/tools/mise/tasks/setup-configs.tmpl`,
   `setup.tmpl` and the `helpers` library read the members from
   `config get repo.members` (submodule and sibling), skip an old-form member
   with the warning "member <name>: run bootstrap init there" and continue,
   and pass `BOOTSTRAP_HOME` to the run in each member (D63).
9. Any other read or mention of `.config/bootstrap.toml`,
   `bootstrap.overrides.toml` or `.config/repo-generated.tsv` under
   `repo/tools/` gets the same change. Report each under `CHANGED`.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn
  'bootstrap.toml\|bootstrap.overrides.toml\|config/repo-generated.tsv\|vwf.yaml\|vwf-membership'
  repo/tools/` finds nothing.
- With `BOOTSTRAP_BIN` set to a build of this worktree's engine (D44), `bash -n`
  passes on each changed script, and the git `derive` script prints the expected
  protected branches in a temp repo whose `repo.toml` has `core_branches =
  ["main","develop"]` and `[branches.develop] landing = "direct-commit"`.

## Guardrails

- Touch nothing outside `repo/tools/`.
- Scripts stay bash 3.2-compatible and call tools through `mise x --`, as
  today; the engine is called as `${BOOTSTRAP_BIN:-mise x -- bootstrap}` (D20,
  D42).
- Delete with `rm`, never `git rm`.

## Commit

`feat: tool scripts read .config/bootstrap through bootstrap config get`
