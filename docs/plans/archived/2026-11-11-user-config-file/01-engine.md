# U1 — The config file, the new flags and bootstrap config in the engine

- **Read first:** `cli/internal/cli/cli.go`, `cli/internal/cli/clone.go`,
  `cli/internal/cli/cli_test.go`, top to bottom, before editing.
- **Lazy-load:** `docs/design/cli.md` ("The command line", "The engine
  variables", "The exit codes"), `cli/internal/report/report.go`,
  `cli/internal/scripts/scripts.go`, `cli/internal/macos/macos.go`.

## Ruling

- **D0** — The rulings of B27 (2026-11-11), agreed before the plan

  1. Precedence: flag, then variable, then config file, then default; each
     existing variable keeps its effect. 2. Path:
     ~/.config/bootstrap/config.toml, or $XDG_CONFIG_HOME/bootstrap/config.toml
     when XDG_CONFIG_HOME is set. 3. A missing file: the defaults, with no
     message. 4. An unknown key or a wrong type: an error; the engine stops with
     code 2 and names the key and the file. 5. One file for the repo side and
     the Mac side, with a [mac] table for the keys of the Mac side only. 6.
     `bootstrap config` prints each key, its value and the source: flag,
     variable (with its name), file (with its path) or default. 7. No sync, seed
     or tool writes the file; only the user, or bootstrap when the user
     explicitly tells it to. The file holds only the settings of the CLI itself,
     never the config of a tool. 8. The path of the file: the flag --config
     <path>, then the variable BOOTSTRAP_CONFIG, then the default path of
     rule 2. The test tasks set BOOTSTRAP_CONFIG to a missing path; a test case
     of the config points it at a test file; the test tasks do not change
     XDG_CONFIG_HOME.

- **D1** — How the engine applies the file

  Run loads the file at start, after the flags are parsed. A flag sets its
  variable of the process (as --jobs sets BOOTSTRAP_JOBS now). For each key with
  no flag and no variable, Run sets the variable of the process from the file.
  The code that reads the variables (clone.go, cli.go prepare, report.go,
  scripts.go, macos) does not change. Run records the source of each value for
  `bootstrap config`.

- **D2** — Which keys the file has

  Four keys: jobs (integer above 0; variable BOOTSTRAP_JOBS), cache_dir (string;
  BOOTSTRAP_CACHE), url (string; BOOTSTRAP_URL), clone (string;
  BOOTSTRAP_CLONE). The user said: "6 keys are fine", then removed verbose (D3),
  then said "Drop the key" for [mac] repo_path (D16).

- **D3** — verbose

  The user said: "`verbose` will always be a flag and never a setting inside
  config". A new flag --verbose sets BOOTSTRAP_VERBOSE=1 for the run.
  BOOTSTRAP_VERBOSE=1 keeps its effect, and only the value 1 turns it on. A
  verbose key in the file is an unknown key: exit 2.

- **D4** — Which commands take which new flag

  The user said: "other keys can be file or flag" and "All keys get a flag".
  --config <path>: each command except version. --cache-dir <path>, --url <url>,
  --clone <path>: each command that opens the clone, which is each command
  except version (init, sync, check, questions, add, remove, replace, detect,
  list, on both sides), and config. --repo-path <path>: each mac command except
  mac init (mac init uses the current folder), and config. --verbose: init,
  sync, check, add, remove, replace, detect, on both sides. --jobs keeps its
  commands. `bootstrap config` takes --config, --jobs, --cache-dir, --url,
  --clone, --repo-path and --help. Each value flag takes `--flag value` and
  `--flag=value`, as --jobs does. An empty value of a flag is a usage error
  (exit 2).

- **D5** — A path in the file

  A value of cache_dir or clone in the file is absolute or starts with `~/`. The
  engine expands a leading `~/` with HOME. Any other value stops with code 2 and
  names the key and the file. Flags and variables keep the current behavior: the
  engine uses their value as given.

- **D6** — The errors of the file

  Exit 2 for: a file that is not valid TOML (names the file), an unknown key or
  table (names the key, for example `mac.verbose`, and the file), a wrong type
  (for example jobs = "4"), an empty string, and jobs below 1. The message has
  the form `bootstrap: <file>: <key>: <problem>` through a.errorf. A missing
  file at any of the three paths: the defaults, no message (ruling 3). A path
  that exists but is a folder or cannot be read: exit 2, names the file. The
  engine reads the file for each command except version and help, and before the
  preflight, so an error stops the command before any other work.

- **D7** — The child processes

  A value of the file goes to the scripts and to mise as a variable of the
  process, the same as an exported variable. Accepted.

- **D8** — The command `bootstrap config`

  `bootstrap config` writes to stdout and changes nothing. Line 1: `file` and
  the path of the config file and the source of the path (`flag --config`,
  `variable BOOTSTRAP_CONFIG` or `default`), and `(missing)` after the path when
  the file does not exist. Then one line for each key, in this order: jobs,
  cache_dir, url, clone, mac.repo_path. Each line has three columns: the key,
  the value and the source. The source is `flag --<name>`, `variable <NAME>`,
  `file` or `default`; the source `file` does not repeat the path. `-` is no
  value (for example clone and mac.repo_path by default). The default of jobs is
  2 x the CPU count, of cache_dir ${XDG_CACHE_HOME:-$HOME/.cache}/bootstrap, of
  url the defaultURL of clone.go. The columns are aligned with spaces, as in the
  example of the plan. It needs no git repository, no mise and no clone. An
  error of the file exits 2 as in D6. `bootstrap mac config` is an unknown
  command. No --json. Example:
  file /Users/me/.config/bootstrap/config.toml default
  jobs 4 flag --jobs
  cache_dir /Users/me/.cache/bootstrap default
  url https://github.com/virajp/bootstrap.git default
  clone - default
  mac.repo_path /Users/me/Projects/macos-setup variable MACOS_SETUP_REPO_PATH

- **D14** — An empty or a bad variable (gaps 6, 7)

  An empty variable (for example BOOTSTRAP_CACHE=) is unset, as acquireClone
  treats it now: the value of the file applies. A non-empty variable is set,
  also when its value is bad (for example BOOTSTRAP_JOBS=abc): it blocks the
  file value, and bootstrap config shows its raw value with the source variable
  <NAME>. The script runner falls back to 2 x the CPU count, as now.

- **D15** — Values of the process in the tests (gaps 8, 9)

  Each Go test that calls Run uses t.Setenv for BOOTSTRAP_JOBS, BOOTSTRAP_CACHE,
  BOOTSTRAP_URL, BOOTSTRAP_CLONE, MACOS_SETUP_REPO_PATH, BOOTSTRAP_VERBOSE and
  BOOTSTRAP_CONFIG first, so Go resets them after the test. A child bootstrap
  that a script starts sees the values: accepted. test/go exports
  BOOTSTRAP_CONFIG=/nonexistent/bootstrap/config.toml.

- **D16** — [mac] repo_path (rulings review gap 5)

  The user said: "Drop the key". On an adopted Mac, the mise env exports
  MACOS_SETUP_REPO_PATH from config_path of macos.toml
  (macos/tools/mise/tools.toml:18), so the variable always beats a key of the
  file. The file has no [mac] repo_path key. A [mac] table with any key is an
  unknown key (exit 2); an empty [mac] table is valid. Ruling 5 stays for a
  later Mac key. The flag --repo-path stays (D4), and bootstrap config still
  shows the line mac.repo_path with the source flag --repo-path, variable
  MACOS_SETUP_REPO_PATH or default (value -).
- **D17** — Round 2 of the rulings review (gaps 1-3)

  1. An empty XDG_CONFIG_HOME is unset: the default path is
     $HOME/.config/bootstrap/config.toml, as for an empty XDG_CACHE_HOME now; a
     non-empty value is used as given. 2. --config <path> also sets
     BOOTSTRAP_CONFIG in the process, the same as each other flag sets its
     variable (D1), so a child bootstrap reads the same file. 3. The commit of
     U2 keeps the type test: its change to updateall:configs is one comment,
     with no change of behavior.- **D19** — Round 4 of the rulings review (gaps
     1, 2)

  1. U2 also fixes the comment at .config/mise/tasks/test/mac:1619-1620 so that
     it does not say that the engine stops on --verbose; it changes no code of
     that case. 2. The text that U1 passes to a.errorf is `<file>: <key>:
     <problem>` (or `<file>: <problem>` with no key); a.errorf adds
     `bootstrap: ` when unstyled and the glyph on a styled terminal, as for each
     other error.

## Edits

1. **`cli/internal/config/config.go`** (new package `config`) — the loader.
   - `Path(flag string, getenv func(string) string) (path, source string)`:
     the flag, then `BOOTSTRAP_CONFIG`, then
     `$XDG_CONFIG_HOME/bootstrap/config.toml`, else
     `$HOME/.config/bootstrap/config.toml`. The source is `flag --config`,
     `variable BOOTSTRAP_CONFIG` or `default`.
   - `Load(path, home string) (File, error)`: a missing file gives an empty
     `File` and no error. Decode with `toml.DecodeFile` into a struct with
     pointer fields (`Jobs *int64`, `CacheDir`, `URL`, `Clone *string`, and an
     empty `Mac` table, D16). A key of `MetaData.Undecoded()` is an
     unknown key. Apply the type, empty-string, `jobs` and path rules of D5 and
     D6, and expand a leading `~/` with `home`. Each error names the file and
     the key, as D6 says.
   - A table of the keys: name in the file (`jobs`, `cache_dir`, `url`,
     `clone`), and `mac.repo_path` for `config` only (D16), variable name, flag
     name. The package names no
     tool.
2. **`cli/internal/config/config_test.go`** (new) — table tests: the path order
   of the three sources; a missing file; each key read; an unknown key and an
   unknown table (`verbose`, `mac.verbose`, `[other]`); a wrong type; an empty
   string; `jobs = 0`; a relative path; `~/` expanded; a file that is not TOML.
3. **`cli/internal/cli/cli.go`** —
   - `allowed`: add `config` and `verbose` and the value flags per D4. Add the
     command `config` (not with `mac`).
   - `parse`: `config`, `cache-dir`, `url`, `clone`, `repo-path` take a value
     like `jobs`. An empty value is a usage error. Record each flag value in
     `opts`. `--verbose` takes no value.
   - In `Run`, for each command except `version` and help, after `parse` and
     before the dispatch: find the path, load the file (an error exits 2 through
     `a.errorf`), then for each key set the variable of the process: from the
     flag when given, else keep a set variable, else from the file. `--verbose`
     sets `BOOTSTRAP_VERBOSE=1`. Record the value and source of each key for
     `config`. Move the `os.Setenv("BOOTSTRAP_JOBS")` of `parse` into this
     step, so that the order is the same for each key.
   - Compute `a.styled` after `BOOTSTRAP_VERBOSE` is set only if the code needs
     it; `report` reads the variable when it prints.
   - `Usage`: add `config` to the commands and the new flags, in the style of
     the current text.
4. **`cli/internal/cli/config.go`** (new) — `func (a *app) config(o opts)
   fail.Code`: print the lines of D8 to `a.stdout`. It calls no preflight, no
   git and no clone. The default of `jobs` is the value that
   `cli/internal/scripts/scripts.go` uses (2 x `runtime.NumCPU()`); read it from
   one place, not a copy, if `scripts` exports it, else compute it the same way
   and name the source line in a comment.
5. **`cli/internal/cli/cli_test.go`** — tests through `Run` with
   `BOOTSTRAP_CONFIG` set to a test file (`t.Setenv`): the order flag >
   variable > file > default for `jobs` and `cache_dir`; `--verbose` sets
   `BOOTSTRAP_VERBOSE`; an unknown key exits 2 and names the key and the file;
   the output of `bootstrap config` for each source; `bootstrap mac config` is
   an unknown command; an empty `--url=` is a usage error. Each test that calls
   `Run` first uses `t.Setenv` for the seven variables of D15, with
   `BOOTSTRAP_CONFIG` at a missing path or a test file, so no test reads the
   real file of the user and no value stays for the next test.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:go` passes, `TestEngineNamesNoTool` included.
- `grep -rn 'BOOTSTRAP_CONFIG' cli/internal` finds the loader and the tests.

## Guardrails

- Do not touch the test tasks in `.config/mise/tasks/test/` (U2) or a doc (U4).
- Do not change the code that reads the variables (`clone.go` `acquireClone`,
  `prepare`, `report.go`, `scripts.go`, `macos`), except to move the
  `BOOTSTRAP_JOBS` write of `parse` (D1).
- No new dependency: use `github.com/BurntSushi/toml` of `cli/go.mod`.
- Write no file: the engine never writes the config file (ruling 7).
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the user config file of the engine and bootstrap config`
