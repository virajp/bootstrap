# Plan — B27 — user config file ~/.config/bootstrap/config.toml for the engine (2026-11-11)

## Goal

A user can set the defaults of the engine one time in
`~/.config/bootstrap/config.toml` (or `$XDG_CONFIG_HOME/bootstrap/config.toml`),
and does not export variables in each shell. Each value comes from a flag, then
a variable, then the file, then the default. `bootstrap config` shows each value
and its source. Backlog item B27 holds the agreed rulings 1–8 (decision D0).
The plan reverses no standing decision.

## Facts the survey established

- The engine reads its settings as variables of the process, through
  `a.getenv` (`os.Getenv`):
  - `BOOTSTRAP_CLONE`, `BOOTSTRAP_CACHE`, `BOOTSTRAP_URL`:
    `cli/internal/cli/clone.go:38-55` (`acquireClone`). `defaultURL` is the
    default URL. The default cache is
    `${XDG_CACHE_HOME:-$HOME/.cache}/bootstrap`.
  - `MACOS_SETUP_REPO_PATH`: `cli/internal/cli/cli.go:304-319` (`prepare`), and
    `cli/internal/macos/macos.go:195`, `cli/internal/macos/file.go:219` (it
    passes
    the value to the templates). The mise env of an adopted Mac exports it from
    `config_path` of `macos.toml` (`macos/tools/mise/tools.toml:18`).
  - `BOOTSTRAP_VERBOSE`: `cli/internal/report/report.go:116` (only `1`).
  - `BOOTSTRAP_JOBS`: `cli/internal/scripts/scripts.go:132`. `--jobs` sets it
    with `os.Setenv` in `parse` (`cli/internal/cli/cli.go:258`).
  - `BOOTSTRAP_VALUES`: `cli/internal/cli/cli.go:342`. Not a key (D2).
- `Run` (`cli/internal/cli/cli.go:95`) dispatches the commands. `allowed`
  (`cli.go:128`) lists the flags of each command. `parse` (`cli.go:203`) reads
  the flags; only `answer` and `jobs` take a value now. `Usage` (`cli.go:31`)
  is the help text.
- Each command except `version` opens the clone: `sync`, `check`, `init`,
  `questions` through `a.openClone`, and `add`, `remove`, `replace`, `detect`,
  `list` through `a.open` (`cli/internal/cli/select.go:46`).
- `github.com/BurntSushi/toml` v1.6.0 is in `cli/go.mod`. `toml.DecodeFile`
  returns `toml.MetaData`; `MetaData.Undecoded()` lists the keys that no field
  took, for the unknown-key rule.
- `TestEngineNamesNoTool` fails when a string literal of the engine names a
  tool. The new code names no tool.
- The test tasks: `.config/mise/tasks/test/repo` exports `BOOTSTRAP_VERBOSE=1`
  at line 37 and has the verbose cases near line 2560. `test/mac` and
  `test/macos-tools` set `XDG_CONFIG_HOME` to a temp folder in each case
  (`mac:143`, `macos-tools:114`); that stays as it is. `test/go` runs
  `go vet`, a `gofmt` check and `go test` in `cli/`.
- The docs that describe the variables: `docs/design/cli.md` ("The engine
  variables", "The command line"), `docs/design/mac.md`, `docs/design/repo.md`,
  `README.md`, `CLAUDE.md` ("Tests"), `docs/sync/index.md`,
  `docs/setup/index.md`, `docs/setup/tools.md`.
- The repo has no commit scopes. The commit types are `feat`, `fix`, `test`,
  `docs`, `ops` and the others of `.config/git-conventional-commits.yaml`.

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

- `BOOTSTRAP_VALUES` as a key: it names a file of one repository.
- `BOOTSTRAP_BIN` as a key: the engine does not read it.
- `color`, `[mac] bootstrap`: new behavior with no variable now; `FORCE_COLOR`,
  `NO_COLOR` and `--no-bootstrap` exist.
- `yes`: it removes the consent step of each writing command.
- `timings`: B26 is not built.
- `update_interval`: new behavior; a later backlog item can add it.
- A `verbose` key: the user ruled that verbose is a flag only (D3).
- JSON output of `bootstrap config`.

## Parked

None.

## Launch

This folder is already committed and pushed on the branch it was planned on, so
the run's worktree — cut from the integration branch — can see it.

Run in a fresh context — a fresh session, or a runner that `all` dispatches:

/vwf:execute docs/plans/2026-11-11-user-config-file

or let the queue pick it, by priority:

/vwf:execute next

or run every runnable plan, highest priority first:

/vwf:execute all
