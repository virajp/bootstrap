# U2 — The mode of the ssh folder

- **Read first:** `docs/design/setup-sync.md`, `docs/design/sync-engine.md`,
  then
  every owned file. In `sync-engine/setup-sync`, read `_ss_file` from :518. In
  `.config/mise/tasks/test/setup-sync`, read the helpers from :86 and
  `case_init_chmod`.
- **Lazy-load:** `sync-engine/common` (`_se_mode`, `_se_outcome`) and
  `sync-engine/write` (`_se_finish`). Read them only.

## Ruling

D6: "When setup_sync handles a target below .ssh/, it makes sure once in each
run
that the folder `$_SS_HOME/.ssh` has mode 700, on a new folder and on an
existing
folder. The check uses the path in the home folder, not the real path of a
linked
file. If ~/.ssh is itself a link, the check reads and sets the mode of the
target
of the link. A real run sets the mode, reports it, and sets `_SS_CHANGED`. A dry
run changes nothing, and it reports the change only for an existing folder with
another mode. It reports nothing for a missing folder. When the mode is already
700, the sync prints nothing about the folder."

## Edits

1. **`sync-engine/setup-sync`** — Add a function `_ss_ssh_folder <dry_run>`
   above
   `_ss_file`, with a comment that gives the reason: a new `~/.ssh` otherwise
   takes the mode of the umask. It does this, for `dir=$_SS_HOME/.ssh` and
   `show=$_SS_TILDE/.ssh`:
   - If `_SS_SSH_DONE` is set, it returns 0. Each path below that returns 0 sets
     `_SS_SSH_DONE=1` first. Thus the function acts once in each run, and after
     a
     failed `chmod` each later target below `.ssh/` also fails and is not
     written. `_se_outcome` counts one failure for that path.
   - If `dir` is not a folder (`[[ -d $dir ]]` is false) and dry run is 1, it
     returns 0 and prints nothing.
     It does not call `_se_mode` here, because `stat` of a missing path prints
     an error.
   - If `dir` is a folder and the mode from `_se_mode "$dir/."` is `700`, it
     returns
     0 and prints nothing. The `/.` makes `stat` read the target of a link.
   - With dry run 1: it calls
     `_se_outcome changed "$show" '' "would chmod 700 $show"` and returns 0.
   - With dry run 0: it runs `mkdir -p -- "$dir"`, then `chmod 700 -- "$dir"`.
     If
     a command fails, it calls
     `_se_outcome failed "$show" 'cannot chmod' "cannot chmod $show"` and
     returns
     1. On success it calls `_se_outcome changed "$show" '' "chmod 700  $show"`
        and sets `_SS_CHANGED=1`.
2. **`sync-engine/setup-sync`, `_ss_file`** — After the
   `if ((dry_run)); then destdir=... fi` block that ends after the `mkdir -p` at
   :572, add this line:

   ```text
   [[ $target != .ssh/* ]] || _ss_ssh_folder "$dry_run" || return 1
   ```

   The function must run before the file is written. The new folder must not
   have another mode when the file goes in.

   The `mkdir -p -- "$destdir"` at :572 runs first in a real run. Thus a new
   folder exists when `_ss_ssh_folder` runs, and it reports `chmod 700`. Do not
   change that order.
3. **`macos/fixtures/repo/tools/sshdir/`** — Add `tools.toml` with a description
   and `[files]` tables for `.ssh/config` and `.ssh/config.identity` with
   `format = "text"`, and a small file for each. Copy the header and the shape
   of the nearest fixture
   tool.
4. **`macos/fixtures/config/sshdir/macos.toml`** — Select the `sshdir` tool.
   Copy
   the shape of `macos/fixtures/config/ok/macos.toml`.
5. **`.config/mise/tasks/test/setup-sync`** — Add these cases and their
   `run_case` lines, next to `case_init_chmod`. Each case uses
   `use_config sshdir`.
   - `case_ssh_new`: with no `$home/.ssh`, `run_real` makes `$home/.ssh` with
     mode
     700 and writes `config` in it. A second `run_real` has no `chmod 700` line.
   - `case_ssh_existing`: with `$home/.ssh` at mode 755, `run_real` has the
     output
     `chmod 700  ~/.ssh`, and the mode is 700 after the run.
   - `case_ssh_link`: with `$home/.ssh` at mode 755 and `$home/.ssh/config` as a
     link to a file outside `$home/.ssh`, `run_real` sets the mode of
     `$home/.ssh`
     to 700.
   - `case_ssh_dry`: with `$home/.ssh` at mode 755, `run_dry` has the output
     `would chmod 700 ~/.ssh`, and the mode stays 755. With no `$home/.ssh`,
     `run_dry` makes no folder, and the output has no `stat:` line. The output
     has exactly one `would chmod 700 ~/.ssh` line, for the two targets of the
     fixture.
   - `case_ssh_folder_link`: with `$home/.ssh` as a link to a folder at mode
     755, `run_real` sets the mode of that folder to 700. A second `run_real`
     has no `chmod 700` line.

## Verification

- `mise x -- mise run code:format` passes.
- `mise x -- mise run test:setup-sync` passes, also the five new cases.
- `mise x -- mise run test:macos-tools` passes.
- `mise x -- mise run test:repo-sync` passes.

## Guardrails

- Write for bash 3.2 and bash 5.x.
- Under `set -e`, write a command substitution that can fail as
  `VAR=$(cmd) || VAR=''`.
- Do not change `macos/tools/ssh/`.
- Do not change `sync-engine/write` or `sync-engine/common`. `repo_sync` has no
  `~/.ssh` target.
- Do not change a mode other than the mode of `~/.ssh`.

## Commit

`fix: keep the ssh folder at mode 700 in the Mac sync`
