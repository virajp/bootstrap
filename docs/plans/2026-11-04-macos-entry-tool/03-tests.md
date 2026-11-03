# U3 — Tests: test:macos-tools and the entry-task cases

- **Read first:** the owned files `.config/mise/tasks/test/setup-sync` and
  `.config/mise/tasks/test/_default`, the old test
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/.config/mise/tasks/test/tools`,
  and the U1 and U2 output, top to bottom.
- **Lazy-load:** `.config/mise/tasks/test/repo-sync` (how it renders the repo
  stub `_default.tmpl` with gomplate, and its cache and lock cases),
  `sync-engine/setup-sync`, `sync-engine/clone`, the old lock cases of
  `/Users/virajpatel/Projects/github.com/virajp/macos-setup/.config/mise/tasks/test/setup-sync`
  (`case_lock_wait`, `case_lock_dead`, `case_leftovers`).

## Ruling

From the design rulings of 2026-11-04, verbatim:

> - Tests: `test:repo-sync`, `test:setup-sync`, `test:macos-tools` (renamed
>   from macos-setup `test:tools`), and an orchestrator `test`
>   (`test/_default`) that runs all three in parallel. CLAUDE.md rule:
>   `sync-engine/` → `mise run test`; `repo/` → `test:repo-sync`; `macos/` →
>   `test:setup-sync` + `test:macos-tools`. No test in hooks or in
>   `code:merge` (a shared task; consumers have no such tests).

> - One shared cache for repos and the Mac: the clone in a `repo/` subfolder,
>   the lock and `macos-generated.tsv` next to it. Origin compare is
>   normalized (repos use SSH, the Mac uses HTTPS). Branch `main` explicit. A
>   non-git, non-empty clone folder stops with exit 2 (no delete). A dry run
>   releases the lock early.

Plan decisions:

D2: "updateall/configs/_default.tmpl embeds sync-engine/clone at render time
with gomplate file.Read and the .sync_engine path of the template context, as
the repo stub does since P01. It parses the flags, refuses a nested run, and
checks MACOS_SETUP_CONFIG_PATH (set, an existing folder) before any network
use. Then, in the order of P01 D12, it takes the new lock, runs the legacy
cleanup (D14), clones or refreshes the cache clone, and copies it to a temp
folder. It runs
macos/tools/macos-setup/.config/mise/tasks/updateall/configs/sync of the copy
with SYNC_ENGINE_CLONE set to the copy. A real run holds the lock to the end
of that child; a dry run releases it after the copy."

D3: "[questions.config_path] (question and example
~/Projects/github.com/janedoe/macos-setup) and [mise.env]
MACOS_SETUP_CONFIG_PATH = "<% index .config `config_path` | required
`macos-setup: set config_path in [tools.macos-setup] of macos.toml` |
regexp.Replace `^~` .env.HOME %>". A ~ at the start becomes $HOME, because
mise does not expand ~ in [env]. The first run needs the variable from the
shell; each later shell gets it from mise."

D5: "U3 adds them to test:setup-sync. Each case renders
updateall/configs/_default.tmpl with gomplate, as test:repo-sync renders the
repo stub, and runs it with --no-bootstrap, SYNC_ENGINE_URL set to a local
bare origin and SYNC_ENGINE_CACHE in the case folder. The origin is a git copy
of the fixture tree with the real macos/tools/macos-setup/ in place of the
fixture stub. Cases: the old clone cases on the new cache (a first clone, a
refresh after a new origin commit, a clone with another origin, the SSH and
HTTPS forms of one URL, a non-empty non-git folder gives 2, an origin that
cannot be cloned gives 5, the running line); the old lock cases (a run waits
for a live holder, a dead holder); the old flag cases (an unknown flag under
bash gives 2, --dry-run under bash and under mise run);
MACOS_SETUP_CONFIG_PATH unset gives 2 and no clone; a nested run gives 2; the
legacy cleanup of D14; and the leftovers case: a real run gives 0, a dry run
0, a run with a bad macos.toml 2, and after each, no lock and no temp copy."

D6: "Every tool of macos/tools/. It copies macos/examples/macos.toml into a
temp config folder and calls setup_sync of the working tree directly (as
test:setup-sync does) with --no-bootstrap in a temp HOME. It fails when a tool
folder of macos/tools/ is not in the selected tools that the run logs, so an
import plan that does not select a new tool in the example fails."

D7: "Case 0: no [files] table gives mode = "seed" to a json, yaml or toml
target. Case 1: the first real run exits 4 when an init-stage tool is
selected, else 0; a second run exits 0; every .json, .yaml, .yml and .toml
file in the temp HOME parses with jq or yq. Case 2: the example without the
answer of one question (config_path of macos-setup) exits 2, and the message
names the key."

D10: "In this plan: U3 adds test:macos-tools and the entry-task cases of
test:setup-sync, and the test orchestrator runs test:macos-tools too. The
cases of P09 do not change."

D14: "The cleanup of sync-engine/clone (the old ~/.cache/bootstrap/ and
~/.cache/bootstrap.lock), and the Mac part that P01 leaves alone: when
<cache>/macos-generated.tsv is missing and ~/.cache/macos-setup/generated.txt
exists, the task moves that file to <cache>/macos-generated.tsv; then it
removes ~/.cache/macos-setup/, but not while a live pid holds
~/.cache/macos-setup/lock. It runs after the task takes the new lock."

## Edits

1. **`.config/mise/tasks/test/macos-tools`** (new, mode 755). The port of the
   old `test/tools`:
   - `#MISE description="Render every Mac tool with the sample values file, with
     no network"`;
     a comment that says what the three cases prove.
   - The old preamble: `set -euo pipefail`, the unset of the `GIT_*`
     variables of a git hook, the helpers source, `MISE_DATA_DIR` and
     `MISE_CACHE_DIR` exported before `HOME` changes, a temp folder with its
     real path and a trap that removes it.
   - The tree: copy `sync-engine/` and `macos/` of the working tree into a
     temp folder, `git init`, add and commit it with a `_git` that has no
     user config, as `test:setup-sync` does.
   - Each case: a temp `HOME`, `XDG_*` folders, `SYNC_ENGINE_CACHE`,
     `TMPDIR` and `GIT_CONFIG_GLOBAL`; `MACOS_SETUP_CONFIG_PATH` is a temp
     folder with a copy of `macos/examples/macos.toml`. The run is
     `bash -c 'source "$SYNC_ENGINE_CLONE/sync-engine/setup-sync" && setup_sync
     "$@"' setup-sync --no-bootstrap`
     with `SYNC_ENGINE_CLONE` set to the tree.
   - The selection check (D6): after the first run, each folder name of
     `macos/tools/` must be in the `tools:` line that `setup_sync` logs;
     else the case fails and names each missing tool.
   - Case 0, case 1 and case 2 of D7. Case 1 finds whether a tool of
     `macos/tools/` has `stage = "init"` and expects 4 then 0, or 0 then 0.
     Parse each file with the jq and yq of `mise x <tool>@latest --`, no pin.
     Case 2 removes `config_path` from the copy and expects exit 2 and the
     `required` message of D3.
2. **`.config/mise/tasks/test/_default`**: add `"test:macos-tools"` to the
   `#MISE depends` list. Change nothing else.
3. **`.config/mise/tasks/test/setup-sync`**: add the entry-task cases of D5.
   - Once per test, render
     `macos/tools/macos-setup/.config/mise/tasks/updateall/configs/_default.tmpl`
     with gomplate (`mise x gomplate@latest --`, delimiters `<% %>`) and a
     context whose `sync_engine` is the absolute path of the working tree's
     `sync-engine/`, the way `test:repo-sync` renders the repo stub. Make the
     result executable.
   - The origin: a git copy of the fixture tree in which the fixture
     `macos/tools/macos-setup/` is replaced with the real
     `macos/tools/macos-setup/`, then a bare clone of it. Each entry case
     exports `SYNC_ENGINE_URL=file://<bare>` and `SYNC_ENGINE_CACHE` in the
     case folder, and writes a `macos.toml` with
     `[tools.macos-setup] config_path = "<case config folder>"`.
   - The cases of D5, each with the exit code and the message, and each run
     with `--no-bootstrap`. Port the old cases `case_clone_first`,
     `case_clone_refresh`, `case_clone_other_origin`, `case_clone_ssh_https`,
     `case_clone_not_git`, `case_clone_fails`, `case_running_line`,
     `case_lock_wait`, `case_lock_dead`, `case_unknown_flag`, `case_mise_run`
     and `case_leftovers` of
     `/Users/virajpatel/Projects/github.com/virajp/macos-setup/.config/mise/tasks/test/setup-sync`
     to the new cache: the clone at `<SYNC_ENGINE_CACHE>/repo`, the lock at
     `<SYNC_ENGINE_CACHE>/lock`, the new temp-copy name, and
     `SYNC_ENGINE_URL` in place of `MACOS_SETUP_REPO_URL`.
   - `case_leftovers` expects 0 for a real run, 0 for a dry run and 2 for a
     run with a bad `macos.toml` (an unknown top-level key); after each run,
     no lock and no temp copy.
   - The legacy cleanup (D14): a case with
     `<XDG_CACHE_HOME>/macos-setup/generated.txt` and no manifest moves the
     file to `<SYNC_ENGINE_CACHE>/macos-generated.tsv` and removes the old
     folder; a case with an existing manifest keeps the manifest and removes
     the old folder; a case with a live pid in
     `<XDG_CACHE_HOME>/macos-setup/lock/pid` keeps the old folder.
   - Keep each case of P09 as it is.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:macos-tools` passes: the three cases and the
  selection check.
- `mise x -- mise run test:setup-sync` passes: each P09 case and each new
  entry case.
- `mise x -- mise run test` passes and runs the three test tasks.
- `grep -nE 'jq@1|yq@4|gomplate@5' .config/mise/tasks/test/macos-tools`
  finds nothing.

## Guardrails

- Touch only the three owned files. Do not edit `macos/`, `sync-engine/` or
  `macos/fixtures/`; report a fault there as a `GAP:` or `UNRESOLVED:`.
- A case never touches the real `$HOME` or the real cache, and needs no
  network.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- Delete with `rm`, never `git rm`.

## Commit

`test: add test:macos-tools and the updateall:configs cases` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
