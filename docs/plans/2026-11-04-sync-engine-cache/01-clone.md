# U1 — The clone library and the repo-sync tasks

- **Read first:** every owned file, top to bottom, before editing. Also
  `MS/tools/macos-setup/.config/mise/tasks/updateall/configs/_default`
  (lines 91-216, read only), the source of `normalize_url`, the explicit
  branch and the non-git folder rule.
- **Lazy-load:** `tools/mise/.config/mise/tasks/_scripts/helpers`
  (`print_header`, `print_info`, `print_error`), `_rs_context` of
  `sync-engine/repo-sync` (U2 adds the key `.sync_engine` there).

## Ruling

From the design rulings of 2026-11-04:

> - The shared library folder is `sync-engine/` (not `engine/`).

> - One shared cache for repos and the Mac: the clone in a `repo/` subfolder,
>   the lock and `macos-generated.tsv` next to it. Origin compare is
>   normalized (repos use SSH, the Mac uses HTTPS). Branch `main` explicit. A
>   non-git, non-empty clone folder stops with exit 2 (no delete). A dry run
>   releases the lock early.

> - The new engine treats a clone that is not in `repo/` (the old layout, the
>   clone at the cache root) as an old clone and removes it before it clones
>   into `repo/`.

> - The shared cache folder is `~/.cache/sync-engine/` (`repo/`, `lock/`,
>   `macos-generated.tsv`), not `~/.cache/bootstrap/` and not `bootstrap2`.
>   **Why:** an old consumer stub runs once more after the update; on a new
>   layout in `~/.cache/bootstrap` it does `rm -rf` and deletes the manifest,
>   and its lock `~/.cache/bootstrap.lock` does not exclude the new lock. A
>   version number in the name has no meaning after the transition. Each run
>   removes the legacy `~/.cache/bootstrap/`, `~/.cache/bootstrap.lock` and
>   `~/.cache/macos-setup/` (not while an old stub holds the old lock). The
>   first run moves `~/.cache/macos-setup/generated.txt` to
>   `macos-generated.tsv` before that removal.

Note: D12 of this plan narrows the last two sentences. This plan removes only
`~/.cache/bootstrap/` and `~/.cache/bootstrap.lock`; plan 10 handles
`~/.cache/macos-setup/`.

> - Engine variables: `SYNC_ENGINE_CLONE` (temp copy, stub → child task),
>   `SYNC_ENGINE_URL`, `SYNC_ENGINE_CACHE`. Internal only, not user config:
>   test:repo-sync and test:setup-sync use them (today `REPO_SYNC_URL` 8x,
>   `MACOS_SETUP_REPO_URL` 5x, `MACOS_SETUP_REPO_PATH` 3x), and
>   `SYNC_ENGINE_URL`
>   tests an unpushed change. The user saw no value in them and let me keep
>   them on this reason. The forwarding file at the old repo-sync path accepts
>   `REPO_SYNC_CLONE` during the transition.

> - No local engine copy in a consumer. `setup:configs:sync` without
>   `SYNC_ENGINE_CLONE` stops with exit 2. The gate `code:lint:repo-sync`
>   sources the engine from `~/.cache/sync-engine/repo/`; no cache → silent
>   exit 0, never a clone.

> - Clone URL: `https://github.com/virajp/bootstrap.git` on both sides
>   (public, fetch only, no SSH key needed on a new Mac). The normalized
>   origin compare stays as a guard for `SYNC_ENGINE_URL` in tests. The
>   engine never clones the private `macos-setup`: `setup` clones it once
>   after `gh auth login`; the engine reads the working clone at
>   `MACOS_SETUP_CONFIG_PATH`.

From this plan:

D6: "The stub template embeds sync-engine/clone at render time with the
gomplate function file.Read. The repo front-end adds the absolute path of the
sync-engine/ folder of its clone copy to the template context as
.sync_engine. One source, no copy in tools/."

D9: "Each function of sync-engine/clone starts with *se* and each global
variable with *SE*. The later shared libraries of sync-engine/ use the same
prefix. The library defines only functions; _se_clone_init sets the paths
from the environment when the caller calls it."

D10: "Through _se_say <info|error> <text>. When the caller defined print_info
and print_error (the helpers of the repo side), _se_say uses them. Else it
prints 'sync-engine: <text>' on stderr. The Mac entry task of a later plan has
no helpers."

D11: "Both SYNC_ENGINE_CLONE and REPO_SYNC_CLONE, set to the same copy, for
the root and for each member, until the parked cleanup plan. A member that has
not synced yet has the old setup/configs/sync, which reads only
REPO_SYNC_CLONE and then sources the forwarding file."

D12: "Each run, a dry run too, calls _se_legacy_cleanup while it holds the new
lock. For ${XDG_CACHE_HOME:-$HOME/.cache}/bootstrap it first takes the old
lock bootstrap.lock (mkdir, with the stale-lock rule of today for a dead
holder); when a live pid holds it, the run skips the removal and says nothing.
With the old lock taken, it removes bootstrap/ and then bootstrap.lock. Once
the run holds the old lock, it releases it on success, on failure and on EXIT:
_se_clone_cleanup, which the trap of the caller calls, releases it too. This
plan does not touch ~/.cache/macos-setup/: the old Mac task can still run
until plan 10, and the Mac entry task of plan 10 moves generated.txt to
macos-generated.tsv and removes that folder."

D13: "No. The gate only reads <cache>/repo/sync-engine/repo-sync. A fetch at
the same time is rare, and the next commit runs the gate again."

D14: "No. The sync never copies a file of sync-engine/ into a consumer, so the
header rule of tools/ does not apply. Each file starts with the bash shebang,
so shell_files_in_scope lints it, and a short comment that names its use."

D17: "The caller. sync-engine/clone defines _se_clone_cleanup (it releases a
held new lock and a held old lock, and removes the copy) and sets no trap. The
stub sets trap
_se_clone_cleanup EXIT, trap 'exit 130' INT and trap 'exit 143' TERM, as the
stub does today."

D18: "Mode 755 for every file of sync-engine/ that has a shebang, clone
included, because the check-shebang-scripts-are-executable hook refuses a file
with a shebang and no execute bit."

## Edits

1. **`sync-engine/clone`** (new, mode 755 (D18), first line `#!/usr/bin/env
   bash`,
   then a comment of at most 10 lines: what the library does, that a caller
   sources it, the three variables). Port the code of A_TMPL:31-112 and
   B_DEF:91-216 into these functions, in this order. Write no top-level
   statement other than function definitions and comments.
   - `_se_clone_init`: sets `_SE_OLD_HELD=''` and
     `_SE_CACHE=${SYNC_ENGINE_CACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/sync-engine}`,
     `_SE_REPO=$_SE_CACHE/repo`, `_SE_LOCK=$_SE_CACHE/lock`,
     `_SE_URL=${SYNC_ENGINE_URL:-https://github.com/virajp/bootstrap.git}`,
     `_SE_BRANCH=main`, `_SE_HELD=''`, `_SE_COPY=''`.
   - `_se_say <info|error> <text>`: D10. Find the helpers with
     `declare -F print_info > /dev/null`.
   - `_se_normalize_url <url>`: a byte copy of `normalize_url` of B_DEF.
   - `_se_git_remote <args>`: `GIT_TERMINAL_PROMPT=0 GIT_SSH_COMMAND="ssh -o
     BatchMode=yes" git "$@"`.
   - `_se_stale_lock <lock>`: `remove_stale_lock` of A_TMPL:54-71 with the
     lock path as its parameter (`<lock>.takeover` as its own lock). Returns
     0 when it removed the lock of a dead holder.
   - `_se_lock`: `mkdir -p` of `$_SE_CACHE`, then the wait loop of
     A_TMPL:76-87 on `$_SE_LOCK` with `_se_stale_lock`. The wait message is
     `Waiting for another sync (pid <pid>) ...` through `_se_say info`. Then
     `_SE_HELD=1` and the pid file `$_SE_LOCK/pid`.
   - `_se_unlock`: when `_SE_HELD` is set, `rm -rf -- "$_SE_LOCK"`, then
     `_SE_HELD=''`.
   - `_se_legacy_cleanup`: D12, with `base=${XDG_CACHE_HOME:-$HOME/.cache}`
     (not `$_SE_CACHE`). It acts only when `$base/bootstrap` or
     `$base/bootstrap.lock` exists. Take the old lock `$base/bootstrap.lock`
     with `mkdir`; when `mkdir` fails, call `_se_stale_lock` and try `mkdir`
     one more time; when that fails too, skip the removal. With the old lock
     taken, set `_SE_OLD_HELD` to its path, remove `$base/bootstrap`, then
     remove the old lock and clear `_SE_OLD_HELD`. Say
     `Removed the old cache <path>` through `_se_say info`. Return 0 always;
     a failed `rm` is not an error of the sync. Do not touch
     `$base/macos-setup` (D12).
   - `_se_fetch`: first `unset GIT_DIR GIT_WORK_TREE GIT_INDEX_FILE`.
     Then:
     1. Old layout: when `$_SE_CACHE/.git` exists, remove each entry of
        `$_SE_CACHE` other than `repo`, `lock`, `lock.takeover` and
        `macos-generated.tsv`, and say `Removed an old clone at <cache>`.
     2. `$_SE_REPO` absent, or an empty folder that is not a link: clone.
     3. A folder with `.git`: read the raw origin with
        `git -C "$_SE_REPO" config --get remote.origin.url`. When its
        normalized form equals the normalized `$_SE_URL`, say
        `Updating the clone in <repo>`, run
        `_se_git_remote -C "$_SE_REPO" fetch --quiet --depth 1 origin
        "$_SE_BRANCH"`
        (fail: say `Cannot fetch <url> - nothing synced` as error, return
        5), then `reset --quiet --hard FETCH_HEAD` and `clean --quiet -ffdx`.
        Else say `The clone in <repo> has the origin '<origin>', not <url> -
        cloning again`
        and clone.
     4. Else: say `<repo> is not a git clone and not empty - move it, or set
        SYNC_ENGINE_CACHE to another folder`
        as error, return 2. Delete nothing.
        The clone step: say `Cloning into <repo>`, `rm -rf` the folder,
        `mkdir -p` its parent, then
        `_se_git_remote clone --quiet --depth 1 --single-branch --branch
        "$_SE_BRANCH" -- "$_SE_URL" "$_SE_REPO"`.
        On a failure remove the folder, say `Cannot clone <url> - nothing
        synced`
        as error, and return 5.
   - `_se_copy <prefix>`: `_SE_COPY=$(mktemp -d
     "${TMPDIR:-/tmp}/<prefix>.XXXXXX")`,
     then `cp -R -- "$_SE_REPO/." "$_SE_COPY"`. Returns 1 on a failure.
   - `_se_clone_cleanup`: `_se_unlock`; when `_SE_OLD_HELD` is set, remove
     that old lock folder; then `rm -rf` of `$_SE_COPY` when it is set.
     Returns 0.
2. **`tools/repo-sync/.config/mise/tasks/setup/configs/_default.tmpl`** —
   keep the header, the `#MISE`/`#USAGE` lines, the member flags and the
   member loop. Replace lines 22-113:
   - A comment of at most 8 lines: the code below is `sync-engine/clone` of
     the bootstrap clone, embedded when the sync renders this template
     (D6); the cache is `~/.cache/sync-engine/`.
   - The embed: `<% file.Read (printf "%s/clone" .sync_engine) |
     strings.TrimPrefix "#!/usr/bin/env bash\n" %>`
     or the equivalent that drops the shebang line. Verify the exact
     signature of `file.Read` and `strings.TrimPrefix` with Context7
     (gomplate) before you write it.
   - Then: `_se_clone_init`, the three traps of D17,
     `print_header "Fetching bootstrap ..."`, `_se_lock`,
     `_se_legacy_cleanup`, `rc=0; _se_fetch || rc=$?; ((rc == 0)) || exit
     "${rc}"`,
     `_se_copy bootstrap || exit 1`, `_se_unlock`.
   - The child call (line 117) becomes
     `SYNC_ENGINE_CLONE="${_SE_COPY}" REPO_SYNC_CLONE="${_SE_COPY}" mise run
     setup:configs:sync ${DRY_RUN}`
     (D11). In `sync_member` (line 130) set the same two variables.
3. **`tools/repo-sync/.config/mise/tasks/setup/configs/sync`** — keep both
   `#MISE tools.*` lines. Replace lines 16-24: when `SYNC_ENGINE_CLONE` is
   empty, `print_error "SYNC_ENGINE_CLONE is not set - run mise run
   setup:configs, which clones bootstrap"`
   and `exit 2`. Else source `${SYNC_ENGINE_CLONE}/sync-engine/repo-sync`. No
   local fallback. Rewrite the comment above to match.
4. **`tools/repo-sync/.config/mise/tasks/code/lint/repo-sync`** — keep lines
   1-35 (header, flags, the gomplate check, the file filter). Before the
   `print_header` (line 37), add the cache check: set
   `lib="${SYNC_ENGINE_CACHE:-${XDG_CACHE_HOME:-${HOME}/.cache}/sync-engine}/repo/sync-engine/repo-sync"`;
   when `lib` is not a file, `exit 0` with no output. Replace lines 39-43:
   a comment that the gate reads the engine of the shared cache and never
   clones (D13), `unset SYNC_ENGINE_CLONE REPO_SYNC_CLONE`, and
   `source "${lib}"`. Keep `repo_sync --check`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n sync-engine/clone` passes. `test -x sync-engine/clone` passes. `head
  -n 1 sync-engine/clone` is
  `#!/usr/bin/env bash`.
- `grep -c 'Synced from' sync-engine/clone` is 0.
- `grep -nE 'git@github|REPO_SYNC_URL|origin HEAD' sync-engine/clone` finds
  nothing.
- `grep -n 'REPO_SYNC_URL\|bootstrap.lock\|remove_stale_lock'
  tools/repo-sync/.config/mise/tasks/setup/configs/_default.tmpl`
  finds nothing.
- Render the stub with
  `gomplate --left-delim '<%' --right-delim '%>' -c ".=<ctx.json>" -f
  tools/repo-sync/.config/mise/tasks/setup/configs/_default.tmpl`
  and a `ctx.json` with `members`, `member_types`, `member_folders` and
  `sync_engine` (the absolute path of `sync-engine/`). The output passes
  `bash -n` and has `_se_fetch` once.

## Guardrails

- Do not touch `sync-engine/repo-sync`, the forwarding file or the test task;
  U2 and U3 own them.
- Do not touch the synced copies of this repository (`.config/mise/tasks/**`
  outside `test/`).
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- `git` and `mise` take no `mise x --` prefix. Do not add a `#MISE tools`
  line to the stub or the gate.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the shared clone library and the sync-engine cache` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
