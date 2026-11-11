# U2 — The test tasks: BOOTSTRAP_CONFIG and the config cases

- **Read first:** the top of each owned task (to the first case), and the
  verbose cases of `.config/mise/tasks/test/repo` near line 2560.
- **Lazy-load:** the rest of `.config/mise/tasks/test/repo`, `CLAUDE.md`
  ("Task scripts", "Tests").

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

- **D13** — The comment on --verbose in updateall:configs (gap 3)

  Fix the comment only at macos/tools/mise/tasks/updateall-configs:56-58: it
  must not say that the engine refuses --verbose. The task does not pass
  --verbose to the engine.

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

1. **`.config/mise/tasks/test/repo`, `test/mac`, `test/macos-tools`,
   `test/go`** — near the `export BOOTSTRAP_VERBOSE=1` line (or the top of
   `test/go`), export `BOOTSTRAP_CONFIG` to a path in the temp folder of the
   task that does not exist (`test/go`: `/nonexistent/bootstrap/config.toml`,
   D15), with a comment: the tests never read the real
   config file of the user. Do not change `XDG_CONFIG_HOME`.
2. **`.config/mise/tasks/test/repo`** — new cases, in the style of the
   existing cases, that run the built binary:
   - `bootstrap config` with a test file (`BOOTSTRAP_CONFIG` points at it):
     each source shows (`file`, `default`, `variable <NAME>`, `flag --<name>`),
     and the first line shows the path and its source.
   - A `--config <path>` flag takes the place of `BOOTSTRAP_CONFIG`.
   - A file with `verbose = true`, a file with `jobs = "4"`, and a file that is
     not TOML: exit 2, and the message names the key (where there is one) and
     the file.
   - A missing file at `BOOTSTRAP_CONFIG`: `bootstrap config` exits 0 and shows
     each default, with no message on stderr.
   - `clone` in the file: a `bootstrap sync --dry-run` uses that clone with no
     `BOOTSTRAP_CLONE` (the case unsets it).
   - `--verbose`: in the verbose cases, one more run with `--verbose` and
     `BOOTSTRAP_VERBOSE` unset prints the per-file log lines.

3. **`macos/tools/mise/tasks/updateall-configs`** — at lines 56-58, change
   only the comment so that it does not say that the engine refuses
   `--verbose` (D13). Change no code of the task.
4. **`.config/mise/tasks/test/mac`** — at lines 1619-1620, change only the
   comment so that it does not say that the engine stops on `--verbose`
   (D19).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test` passes.
- `grep -n 'BOOTSTRAP_CONFIG' .config/mise/tasks/test/*` finds each of the
  four tasks.

## Guardrails

- Do not touch `cli/` (U1) or a doc (U4).
- Write bash for bash 3.2 and bash 5.x.
- Do not change `XDG_CONFIG_HOME` for the config (ruling 8).

## Commit

`test: point the test tasks at a missing config file, add the config cases`
