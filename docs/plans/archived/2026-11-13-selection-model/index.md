# Plan — The selection model: tool and pack commands, a values file of tools, and the new packs (2026-11-13)

## Goal

After this plan lands, the CLI uses one grammar for the selection: `add tool`,
`add pack [--exclude]`, `remove tool` and `replace tool`, on both sides. The
values file holds a table for each selected tool and the key `format = 2`, and
never a `[packs.*]` table. A pack is a shortcut of `add`. `packs.toml` has the
new form and the new packs of the branch `packs-shortcut`. Each command writes
one time and syncs one time through `Plan` and `Apply`, and `Apply` puts the
values file back when the sync fails. `init` migrates a values file of today.

This is plan 2a of the TUI work (artifact "Bootstrap TUI Screens", draft 43,
https://claude.ai/artifact/QGqhhDrKesC3N2Kz7x1nr5; artifact "Bootstrap
Selection Model", https://claude.ai/artifact/JM1D6D53Lqk3SBbSGrRUGu). It
requires plan 1b. Plan 2c (the TUI) requires this plan.

**Reversals of standing behaviour** (each one a ruling of the user): the
grammar `add <tool>` stops with code 2 (D8); `[packs.*]` in a values file is
an old form (D5); the pack `base` is no longer implicit, except its required
members (D3); `detect --no` is gone (D7); the key `default` of `packs.toml` is
gone (D4).

**After the landing, by hand** (not in `after_landing`, because a plan never
carries a release): the user tags `bootstrap@3.0.0` at once (D13), then runs
`bootstrap init` in each consumer and `bootstrap mac init` on the Mac.

## Facts the survey established

The line numbers are from `main` before plans 1a and 1b landed (D19 a). Find
the code by the name of its function or key.

- Dispatch: `Usage` at `cli/internal/cli/cli.go:31`, `Run` at `:150`, the
  `allowed` table at `:181-193`, the extra flags at `:211-219`, `takesNames`
  at `:220`, the dispatch switch at `:242-263`, `parse` at `:268` (value flags
  at `:294-295`). `prepare` at `:384`, `openClone` at `:456`.
- The change flow: `app.open` at `cli/internal/cli/select.go:48` (it needs an
  existing values file, `:58-62`); `known` `:158`; `planAdd` `:183`;
  `planRemove` `:210` (the refusals at `:216-235` and `:275-296`, the orphan
  fixpoint at `:254-274`); `edited` `:330`; `add` `:344`; `addNames` `:357`;
  `remove` `:370`; `replace` `:393` (clean-tree rule `:425-434`); `detect`
  `:529` (classification `:558-600`, tree `:601-650`, the `--no`/`--yes` branch
  `:659-692`); `detectScripts` `:741`; `addPackages` `:713` (Mac). Mac package
  scan: `installedPackages` and `scanMac` in `cli/internal/cli/packages.go:153`
  and `:212`.
- `init` at `cli/internal/cli/init.go:132` (new file headers `repoHeader` and
  `macHeader` at `:28-29`, `:150-155`); `finish` at `:231-368`: temp copy
  (`:236-249`), `selectionOf` (`:252`), `ask` (`:256`), `editTOML` with the
  answers (`:260-273`), the plan with `frontEndOf` and `--dry-run` into
  `io.Discard` (`:279-310`, it sets BOOTSTRAP_VALUES or MACOS_SETUP_REPO_PATH
  and `oldform.Migrating`), the print of the changes (`:316-336`), dry-run exit
  (`:338-343`), `stopDirty` (`:344`), `consent` (`:347`, `git.go:142`), the
  before script (`:350`), `oldform.Apply` (`:355`), the sync (`:365`). There is
  no restore when the sync fails; `ch.cur` and `ch.existed` (`:210-211`) hold
  the old bytes.
- The packs: `Source.ReadPacks` `cli/internal/source/packs.go:101`
  (`IsDefaultTrue` `:46`, `HasPack` `:49`, `HasMember` `:59`, `scanOrder`
  `:180`); `selection.Select` `cli/internal/selection/selection.go:34` (base,
  then `[packs.*]`, then `[tools.*]`), `Expand` `:67` (the `default = true`
  members), `PackLoop` `:91`, `CheckPacks` `:117` (`required` only for a
  default member of base, `:141`), `Walk` `:189`. `selection.BasePack` is the
  constant `"base"` (`cli/internal/selection/repo.go:11`).
- `list`: `cli/internal/cli/list.go:24`, `listRoots` `:70` (pack roots today).
- The old forms: `cli/internal/oldform/oldform.go`: `Repo` `:62` and `Mac`
  `:91` give the code-2 messages, `MigrateRepo` `:148` and `MigrateMac` `:235`
  give the changes, `Apply` `:356` writes them. `Hint` and `MacHint` end each
  message. Contract: `docs/design/bootstrap-config.md`.
- `bootstrap_commit` is written by `target.Env.RecordCommit`
  (`cli/internal/repo/repo.go:448`, `cli/internal/macos/macos.go:575`).
- The branch `packs-shortcut` (`420e3e6`, local, not merged) changes
  `repo/packs.toml`, `macos/packs.toml`, `macos/tools/brave-browser/tools.toml`
  and deletes `macos/tools/devtools/tools.toml`. The engine of today refuses its
  packs files (`[pack.base] dprint: required = true is only for a member of pack
  base with default = true`).
- The values file of this repository (`.config/bootstrap.toml`) has
  `[tools.claude]` and `[tools.go]` only; `macos-setup/macos.toml` has
  `[packs.ai]`, `[packs.apple-dev]`, `[packs.chat]`, `[packs.cli-extras]`,
  `[packs.dev]`, `[packs.disk]`, `[packs.essentials]`, `[packs.iwork]`,
  `[packs.media]` and more. `macos/examples/macos.toml:46` names
  `[tools.devtools]`; `macos/examples/macos.toml:5` describes `[packs.<name>]`.
- The docs that describe the commands and the packs: `docs/design/cli.md`,
  `docs/design/packs.md`, `docs/design/bootstrap-config.md`,
  `docs/design/tools.md`, `docs/design/mac.md`, `docs/sync/index.md` (it runs
  `bootstrap detect --no` at `:268` and `:327`), `docs/sync/vwf.md`,
  `docs/sync/packs/base.md`, `security.md`, `container-build.md`,
  `docs/setup/index.md`, `docs/setup/tools.md`, `README.md`, `CLAUDE.md`. The
  last item of `docs/sync/migrate.md` is M36; plan 1b adds M37, so this
  plan adds M38.
- Tests: `test:repo` writes many scratch `tools.toml` and values files;
  `.config/mise/tasks/test/mac:1984-2016` run `mac detect --no` and `--yes`;
  `.config/mise/tasks/test/repo:3117` has `detect_no`. Go tests in
  `cli/internal/cli/cli_test.go` call `Run` with buffers.
- Process-global state that `Plan` and `Apply` must keep serial:
  `os.Setenv` of MISE_PROJECT_ROOT, MACOS_SETUP_REPO_PATH, BOOTSTRAP_CLONE
  (`cli.go:406`, `:426`, `:478`), BOOTSTRAP_VALUES around the plan
  (`init.go:280-297`), `oldform.Migrating`, `report.CloneShown`.
- The commit types are in `.config/git-conventional-commits.yaml`; the
  repository has no scopes.

## Unit contract

Every unit prompt carries, in order: its ruling quoted from this plan, its
owned paths plus "touch nothing outside this list", the facts section, the
plan's shared files, and the return block below. A unit never bumps a version,
never runs a generator, never edits a doc, never adds a dependency the plan
does not list, never commits. A unit deletes with plain `rm`, never `git rm` —
it stages nothing. A unit never writes a lock file with Write, Edit, `sed` or a
script.

A unit returns exactly this block and nothing else — no file contents, no diff:

    CHANGED: <path> — <one line>            (one per file)
    DECIDED: <what> — <why>                 (choices made inside scope, or none)
    DOCS FALSIFIED: <path> — <passage>      (reported, never edited; or none)
    QUESTION Q<n>: <the question, readable alone>
    OPTIONS Q<n>: <option 1, the recommendation> — <consequence> | <option 2> — <consequence>
    WHY Q<n>: <path:line or plan section> — <why the plan does not decide it>

A block with no question carries the one line `QUESTION: none`.

## Out of scope

- The TUI (`bootstrap tui`, `bootstrap mac tui`, `init` that opens the TUI):
  plan 2c. Until then `init` with a terminal takes the `--no-tui` path (D6).
- `remove pack`, `--include` and optional pack members: the user rejected them
  (D2).
- An engine that reads the packs form of today: the user rejected it (D4).
- A change of the members of a pack other than the branch `packs-shortcut`.

## Parked

- Plan 2c, the TUI, per the artifact "Bootstrap TUI Screens", draft 43. It
  requires this plan and uses `Plan` and `Apply`.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-11-13-selection-model

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
