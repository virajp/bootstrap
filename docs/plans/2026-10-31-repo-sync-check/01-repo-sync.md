# U1 — repo-sync: the --check mode and its tests

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `tools/repo-sync/.config/mise/tasks/setup/configs/sync`,
  `tools/mise/.config/mise/tasks/_scripts/helpers`.

## Ruling

D1: "User: \"repo_sync --check will be better way of doing this without
needing the bootstrap clone.\" A --check mode of repo_sync, with no bootstrap
clone, no REPO_SYNC_CLONE and no network."

D2: "User: \"It should validate all rules which are outside the scope of
shell-lint\", confirmed as: every content rule of repo_sync that needs only
the values file and the local git repository. TOML syntax stays with the
check-toml hook."

D3: "Finish B14. The seven rules that need the clone stay with setup:configs:
a tool or a pack that is not in bootstrap (code 3); a pack loop or a pack
named like a tool; an env that a mise tool does not allow; an [overrides.mise*]
path that no selected tool has; [overrides.graphifyignore] without the claude
tool; a remove line that the shared .gitignore or .graphifyignore does not
have; a root_only tool in a submodule."

D5: "--check calls the same validation functions as the sync. Where a function
mixes a local rule and a clone rule (for example the env rules in `_rs_resolve`
with the tool lookup), the local part becomes its own function, and both the
sync and --check call it. A rule never differs between the two paths."

D6: "--check runs each local check, prints each message, and returns the
highest code. The sync does not change: it keeps its && chain and its order."

D7: "No. --check examines exactly the existing local rules. [...]"

D8: "--check with --dry-run is a usage error, code 2, with the usage text.
--check makes no clone, writes no file and does not set bootstrap_commit. The
[member.*] rules (`_rs_members`) and the neither-main-nor-develop rule
(`_rs_merge_model`) are local, so --check runs them."

D11: "A check whose input failed is skipped, with no message, because the
cause is already reported: when `_rs_envs` fails, the local env rules are
skipped, and `_rs_overrides` in its local-only mode skips only its env and
envname problems (both need the environments list) and still reports its
other problems; when `_rs_values_sel` fails, the local env rules are skipped.
The env-rule function takes the result of
`_rs_values_sel` and does not call it a second time, so no message shows
twice."

D13: "`_rs_overrides` stays one function with a local-only mode (for example
an argument), which skips the problems that need the tool list or the shipped
file (none, the path of no selected tool, graphifyignore without claude, a
remove line that is not shipped). In the local-only mode it loads no
`tool_<name>` templates and reads no clone folder; it reads `_RS_TOOLS` as
empty, and `_RS_ENVS` as empty when `_rs_envs` failed, with `${var:-}`, so it
runs under `set -u`. The sync calls it as today, so its messages keep their
order and interleaving; --check calls it in the local-only mode."

D14: "Accepted as exempt: the old name of the values file (a warning that
needs REPO_SYNC_VALUES unset) and the missing values file. Both run before the
--check path, in the code that the sync already uses, and case_body always
exports REPO_SYNC_VALUES. Every other local rule gets a case."

## Edits

1. **`tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`**
   - **Arguments** (`S:1665-1676`): accept `--check`. `--check` together with
     `--dry-run` prints the usage and returns 2. Add `--check` to the usage
     text (`S:13`, `S:1659`, `S:1672`) and to the header comment (the return
     codes `S:166-181`, the env vars `S:183-192`). Also rewrite the gomplate
     comment (`S:139-141`, "the calling task provides it with
     `#MISE tools.gomplate="latest"`"): the `code:lint:repo-sync` gate has no
     such directive and finds gomplate with `mise x -- which gomplate` (D12 of
     this plan).
   - **Split.** Move each local rule that today sits in a function with a
     clone step into its own function, and call that function from the
     original place, so the sync behaves exactly as before (same order, same
     messages, same codes). At least: the env rules of `_rs_resolve`
     (`S:517-542`). Do not split `_rs_overrides`: give it a local-only mode
     (D13) that keeps the old `.config/repo-sync/gitignore`, the problems
     `table`, `key`, `gikey`, `gilist`, `giempty`, `envname`, `env`, `lock`,
     and a line in both `add` and `remove`, and skips `none`, the path of no
     selected tool, graphifyignore without `claude`, and a `remove` line that
     is not shipped. Do not split a function whose rules are all local
     (`_rs_values_sel`, `_rs_envs`, `_rs_direct_commit`, `_rs_merge_model`,
     `_rs_members`).
   - **The `--check` path.** After the values-file and gomplate checks
     (`S:1678-1690`) and before the clone (`S:1695`), when `--check` is set
     (the old-name warning at `S:1682` already ran, and stays its only copy):
     run each local check once, in the order of the sync — `_rs_envs`
     (`S:1729`), `_rs_values_sel` (called inside
     `_rs_resolve` today, `S:513`), the local env rules, `_rs_merge_model`
     (`S:1736`), the `core_tools` warning (`S:1744`), `_rs_overrides` in its
     local-only mode,
     `_rs_direct_commit`, `_rs_members` — with `|| rc=$?` on each, so a failed
     check does not stop the next (D6). Each check prints its own messages,
     as in the sync. `_rs_members` takes a name that starts each of its
     messages; the sync passes "template context" (`S:1324`), and `--check`
     passes the relative path of the values file. On success, `_rs_members`
     prints its member rows on stdout (`S:256-271`); `--check` sends that
     stdout to `/dev/null` and keeps its stderr messages. Keep the highest
     code. On no
     error, print one line with
     `_rs_log`, for example `<values file>: valid (local rules)`. Return the
     highest code. Make no clone, write no file, and do not touch
     `bootstrap_commit`. If a local check needs a value that only the clone
     gives (for example the tool list for an env rule), give it what the
     values file alone shows, and skip only the clone part.
2. **`.config/mise/tasks/test/repo-sync`**
   - Let `run_case` and `case_body` run `repo_sync --check` instead of
     `--dry-run` for a case, with `REPO_SYNC_CLONE` unset, for example with a
     new optional argument or an env variable. The existing cases do not
     change.
   - Add `--check` cases: a valid values file (code 0 and the valid line);
     one case for each local rule of D2 (the code and the message of the
     sync), except the two that D14 exempts; one case with two errors in
     different checks (both messages and
     the highest code, D6); `--check --dry-run` (code 2); and a case that
     proves no clone: `REPO_SYNC_URL` set to a path that does not exist, code
     0 for a valid file. `case_body` runs `unset REPO_SYNC_URL` today
     (`:107`), so add a case variable, for example `CASE_REPO_SYNC_URL`, that
     `case_body` exports as `REPO_SYNC_URL` after that line when it is set.
   - Run each existing case again. All pass unchanged, which proves that the
     split did not change the sync.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes, with the new cases and every
  existing case.
- `bash -n tools/repo-sync/.config/mise/tasks/_scripts/repo-sync` passes.

## Guardrails

- Do not change the order, the messages or the codes of the sync.
- Add no new rule (D7).
- Do not touch the synced copy `.config/mise/tasks/_scripts/repo-sync` of this
  repository.
- Do not touch `tools/repo-sync/.config/mise/tasks/code/`; U2 owns the gate.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add repo_sync --check for the local rules of bootstrap.toml` — written
by the orchestrator after the unit gate and the wave gate, not by the unit.
