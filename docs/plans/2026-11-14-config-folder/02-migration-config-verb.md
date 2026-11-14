# U2 — The init migration, the config verb and settings

- **Read first:** `cli/internal/oldform/oldform.go`,
  `cli/internal/core/apply.go`, `cli/internal/core/clone.go`,
  `cli/internal/cli/cli.go`, `cli/internal/cli/config.go`,
  `cli/internal/config/`, `cli/internal/cli/init.go`, the package U1 added
  (`cli/internal/schema`) and the paths constants U1 added.
- **Lazy-load:** `docs/design/cli.md` ("The command line", "The config file",
  "The exit codes"), `cli/internal/format/`, `cli/internal/fail/fail.go`,
  `docs/plans/2026-11-13-tui/01-shell.md` (D11(b), the migration as one change).

## Ruling

This file states each ruling in its final form. The ids point at the
`decisions` list of `index.json`, where each ruling and its history are kept.
D15, the key map, is quoted in `01-folder-schema.md`.

**The migration**

- **D4, D10:** `bootstrap init` splits `.config/bootstrap.toml` into
  `.config/bootstrap/repo.toml` and `tools.toml`, each with `format = 3`, moves
  `.config/repo-generated.tsv` into the folder, and removes the old file. It
  does not touch `.config/releases.yaml`; phase R3 moves it.
- **D5, D18:** the content of `.config/bootstrap.overrides.toml` becomes
  `[overrides.tools.*]` in `tools.toml`, except an override of a key that moves
  to `repo.toml`: its value goes directly into `repo.toml` and wins over the old
  file's value. The overrides file is removed. The schema refuses an
  `[overrides.tools.*]` entry for a key that lives in `repo.toml`.
- **D11, D17:** `[tools.git.branch.<b>]` becomes `[branches.<b>]` with
  `landing`: `merge_model = "pr"` → `pr`; `merge_model = "direct"` (or absent)
  with `direct_commit = true` → `direct-commit`; any other → `direct-merge`.
  `merge_from` and `merge_never` move unchanged. An absent `landing` reads as
  `direct-merge`.
- **D16, D76:** `core_branches` lists `main` and `develop`, in that order, but
  only the branches that exist in git (locally or on `origin`). With neither,
  `init` stops with code 2.
- **D47:** a `[tools.git.branch.<b>]` table for a branch that is not `main` or
  `develop`, or that does not exist in git, stops `init` with code 2. The error
  names the table; the fix line is `→ Remove the table or create the branch`.
- **D48, D87:** a fresh `init` (no old file) writes the `main` and `develop`
  that exist; with neither, the current branch — born or unborn, of any name.
- **D32:** `[commits]` gets `scopes` (was `[tools.git].commit_scopes`) and
  `style = "conventional"`, the only enum value.
- **D70, D83:** the migration is atomic. It writes the folder to a temp path,
  renames it into place, then removes the old files. If the sync then fails,
  `core.Apply`'s restore puts back `.config/bootstrap.toml`,
  `.config/bootstrap.overrides.toml` and the old `.config/repo-generated.tsv`,
  and removes `.config/bootstrap/`.
- **D71:** the TUI shows the folder migration as one change, "migrate to
  .config/bootstrap/", applied by the same core migration.

**Old forms and broken folders**

- **D28, D53, D77:** on a repo with only the old file, `version`, `settings`,
  `init` and every `mac` command run. Every other command, `tui` included,
  stops with code 2 and the fix line `→ Run bootstrap init`.
- **D29:** when both the old file and the folder exist, every command except
  `version`, `settings` and the `mac` commands exits 2, `init` included, with
  the fix line `→ Remove one: compare .config/bootstrap.toml with
  .config/bootstrap/`.
- **D88:** when the folder exists but `repo.toml` or `tools.toml` is missing,
  every command except `init` exits 2 with `→ Run bootstrap init`; `init`
  creates the missing file.

**The commands**

- **D3:** the engine-settings command `config` becomes `settings`, with the
  same flags and output. `bootstrap config get|set|unset|show` is the
  repo-config verb.
- **D21, D41, D22, D50, D51:** `config get <key>` of a valid key exits 0 and
  prints an envelope `{"key": <path>, "value": <v>, "source": "file" |
  "default"}` in every format (a JSON object, a TOML document, a YAML map). The
  value can be a table, a list or a scalar. An unset key prints its schema
  default with `source: "default"`; a key with no default prints `null`. An
  unset `tools.*` key, or a key of an unknown tool, prints `null` without the
  catalog. The default of a `releases.<p>` key is `"NONE"`. Only an invalid key
  (code 2) or a read failure (code 1) is an error.
- **D19:** `config` reads the `BOOTSTRAP_VALUES` folder when it is set, and
  then skips the old-form stop. During `init` and `sync`, the engine points
  `BOOTSTRAP_VALUES` at the temp folder of the planned layout.
- **D20:** the engine exports `BOOTSTRAP_BIN` (its own executable path) to every
  script and task it runs.
- **D24:** with no `.config/bootstrap/` (and no `BOOTSTRAP_VALUES`), every
  `config` subcommand exits 2 with `→ Run bootstrap init`. `set` creates
  `releases.toml` with `format = 3` when it is absent. `show` of an absent
  `releases.toml` prints an empty table.
- **D9, D26, D90:** `set` and `unset` only write. They print
  `· run bootstrap sync to apply it` for every `tools.*` key and for every
  `repo.toml` key except `deploy.*`, `release_env`, `topology`,
  `topology_reason`, `bootstrap_commit` and `format`. They never print it for
  `releases.*`.
- **D10:** `set releases.<p> <level>` keeps the higher of the old and new level,
  NONE < PATCH < MINOR < MAJOR, and prints what it kept. `unset releases.<p>`
  removes the key.
- **D25:** `set tools.<name>.<key>` for a tool that is not selected exits 2 with
  `→ Run bootstrap add tool <name>`.
- **D27:** `unset` of a required key exits 2. A whole table
  (`members.<n>`, `branches.<b>`) may be unset. Any `set` or `unset` whose
  result breaks a schema rule, the cross-key rules included, exits 2 and writes
  nothing.
- **D23, D52, D86:** `get` never needs the clone. `set` and `unset` of a
  `tools.*` key validate against the clone's catalog: `env` and `files.*` are
  always valid; any other key must be a question or a default of the tool.
  After each successful fetch the engine writes the fetch time to a file in the
  cache root. `set` and `unset` skip the fetch while that time is less than 7
  days old, and fetch otherwise; stale or absent and offline exits 5. Under
  `--clone` or `BOOTSTRAP_CLONE` there is no TTL. Other write commands still
  fetch every time.
- **D49:** a key segment that contains a dot is double-quoted, as in yq:
  `repo.branches."release.1".landing`.

## Edits

1. **The old form and broken folders** in `oldform` and the command preflight,
   per D28, D29 and D88.
2. **The migration in `init`** (`MigrateRepo` or a new function beside it),
   per D4, D5, D10, D11, D16, D17, D18, D32, D47 and D48. It plans one change
   per file, so `init --dry-run` lists them; the TUI shows the whole migration
   as one change (D71). Comments on a moved key move with it where yq can keep
   them; a comment that cannot move is dropped, and the unit says so under
   `DECIDED`. `[member.*]` is U3's: leave a clear hook for it. The migration
   runs after the selection model's `format = 2` migration, so a file of any
   older form reaches `format = 3` in one `init`.
3. **Atomic write and restore** (D70, D83), in the migration and in
   `cli/internal/core/apply.go`.
4. **`BOOTSTRAP_VALUES` and `BOOTSTRAP_BIN`** for scripts and tasks the engine
   runs (D19, D20).
5. **`bootstrap settings`** (D3). `bootstrap config` with no subcommand is a
   usage error (code 2) that names `get`, `set`, `unset`, `show` and the new
   `settings` command.
6. **`bootstrap config`**: `get <key>`, `set <key> <value>`, `unset <key>` and
   `show [<file>]`, per D9, D10, D21-D27, D41, D49-D52, D86 and D90. The first
   key segment names the file: `repo.`, `tools.`, `releases.`; `show` with no
   file prints all three. `--format json|yaml|toml` on `get` and `show`; the
   default is `json` when stdout is not a terminal, else `toml`; output goes to
   stdout. A `set` value is typed by the schema: a string, a boolean, an
   integer,
   or a list or table given as JSON (`'["main","develop"]'`). Writes use the yq
   path so comments stay.
7. **The cache fetch time** (D52) in `cli/internal/core/clone.go`.
8. **Usage and help.** Update `Usage`, `argsOf`, `flagUsage` and the flag-allow
   table in `cli.go` for `settings` and `config`.
9. **Tests.** Go tests for each migration rule above (every `landing` mapping,
   the overrides fold and merge, the manifest move, an already-migrated repo is
   a no-op, the code-2 cases of D16, D29, D47 and D88, the restore of D83), for
   every `config` subcommand and format, the envelope and defaults, the releases
   rule, the TTL, and `settings`.

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run test:go` passes.
- A built engine, in a temp copy of this repo with its `.git`, passes the plan's
  K3 (K1 runs in U4, D68).

## Guardrails

- Touch nothing outside `cli/`.
- The Mac side is out of scope.
- Do not add a dependency: TOML, YAML and JSON output use the libraries the
  module already has.
- Delete with `rm`, never `git rm`.

## Commit

`feat: init migrates to .config/bootstrap/; add config verb; config becomes
settings`
