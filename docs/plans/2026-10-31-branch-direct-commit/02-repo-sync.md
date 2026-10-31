# U2 — repo-sync: validate direct_commit and render the hook args

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `tools/git/tools.toml` (the hook table, `:21-29`),
  `tools/pre-commit/.config/pre-commit-config.yaml.tmpl` (the loop, `:28-47`),
  `docs/design/bootstrap-config.md` (the `[branch.*]` keys).

## Ruling

D2: "One core branch, either main or develop, depending on the repo's need:
direct_commit on that branch drops it from the hook, and with no branch left
the sync removes the hook. With both main and develop: develop only (the hook
keeps --branch main), or both (the sync removes the hook). User: \"There can
never be a situation where main is allowed for direct-commit but develop is
NOT.\" That case stops the sync with code 2."

D4: "A boolean direct_commit in [branch.main] and [branch.develop]. repo-sync
validates it before it writes any file, in the same step as `_rs_mise_env` and
`_rs_overrides`: a non-boolean value, the key in any other [branch.*] table,
and main true without develop true when both branches exist each stop the
sync with code 2. The merge tasks do not read the key."

D5: "It stops the sync with code 2. The message names the table and says that
the branch does not exist."

D7: "In repo-sync, where _rs_context builds .precommit: find the hook with the
id no-commit-to-branch, remove each --branch <b> pair whose branch has
direct_commit = true, and remove the hook when no branch is left. The
template does not change. Accepted: a no-commit-to-branch hook in the
consumer file .config/repo-sync/pre-commit.yaml changes the same way."

D13: "Yes. The sync validates the key when it is present, whatever its value:
false in a table other than [branch.main] or [branch.develop], and false for
a core branch that does not exist, each stop the sync with code 2. false has
no effect on the render."

D14: "Only --branch <b> pairs, the form of the git tool. A hook with no args,
or with --branch=<b> or -b <b>, stays as it is. The sync removes a hook that
had --branch pairs and has none left, also when it has other args: with no
--branch arg, no-commit-to-branch of pre-commit-hooks v6.0.0 protects main
and master by default."

D15: "Yes. The render drops the --branch <b> pair of each core branch (main,
develop) that _rs_branch_exists does not find, and of each branch with
direct_commit = true. So the hook protects only the core branches that exist
and do not allow direct commits, and a one-branch repo with direct_commit on
that branch has no hook (D2). Side effect, accepted: each one-branch consumer
gets --branch <its branch> only at its next sync, with no new key; a new
develop branch comes back at the sync after it exists."

## Edits

1. **`tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`** — add the
   validation and the render:
   - A new function, for example `_rs_direct_commit "$values"`, called in the
     validation step at `:1671`, next to `_rs_mise_env` and `_rs_overrides`,
     so a failure returns before any file is written. It reads every
     `[branch.<name>]` table of the values file that has `direct_commit`, in
     the way that the script already reads TOML (gomplate). It returns 2,
     after one `_rs_log` message that names the table, when:
     - the value is not a boolean;
     - `<name>` is not `main` or `develop`;
     - `<name>` is a branch that `_rs_branch_exists` does not find (D5);
     - both `main` and `develop` exist, `[branch.main].direct_commit` is
       `true`, and `[branch.develop].direct_commit` is not `true` (D2).
       The first three checks apply to a present key whatever its value
       (D13). `direct_commit = false` has no effect on the render. The
       function records the branches with `direct_commit = true` for the
       render, for example in a variable that `_rs_context` reads.
   - In `_rs_context` (`:1273-1299`), where the line at `:1298` builds the
     joined `precommit` list (tool entries, then the consumer entries), drop
     each `--branch <b>` pair of a hook with `id` `no-commit-to-branch` whose
     `<b>` has `direct_commit = true`, or whose `<b>` is `main` or `develop`
     and `_rs_branch_exists` does not find it (D15). Match only the
     `"--branch", "<b>"` pair form (D14). When a hook that had `--branch`
     pairs has none left, drop the hook from its entry, also when it has other
     args. A hook with no args, or with another form, stays as it is. When an
     entry has no hook left, drop the entry. In a repository
     with both core branches and no `direct_commit = true`, the output is the
     same as today.
   - Update the header comment of the context (`:125-136`) and the list of
     values-file errors (`:163-174`) for the new key.
2. **`.config/mise/tasks/test/repo-sync`** — add cases with `run_case`
   (`:76-95`). Use the pattern of `gitignore_file` (`:483`, `:492-499`) and
   `check_log_lacks` (`:367`) to examine the rendered
   `.config/pre-commit-config.yaml` of the dry run. Make the branches that a
   case needs in its scratch repository. Cases:
   - both branches, no key: the args keep `--branch main` and
     `--branch develop` (code 0);
   - `main` only, no key: only `--branch main` (code 0, D15);
   - a consumer hook in `.config/repo-sync/pre-commit.yaml` with
     `--branch develop` and `--pattern x`, in a `main`-only repository: the
     sync removes that hook (code 0, D14);
   - `main` only, `[branch.main].direct_commit = true`: no
     `no-commit-to-branch` hook (code 0);
   - `develop` only, `[branch.develop].direct_commit = true`: no hook
     (code 0);
   - both branches, `[branch.develop].direct_commit = true`: only
     `--branch main` (code 0);
   - both branches, both `true`: no hook (code 0);
   - both branches, `[branch.main].direct_commit = true` only: code 2 and the
     message;
   - a non-boolean value: code 2;
   - the key in `[branch.feature]`: code 2;
   - `[branch.develop].direct_commit = true` in a `main`-only repository:
     code 2 (D5).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes, with the new cases and every
  existing case.
- `bash -n tools/repo-sync/.config/mise/tasks/_scripts/repo-sync` passes.

## Guardrails

- Do not change `tools/pre-commit/.config/pre-commit-config.yaml.tmpl` (D7).
- Do not touch `tools/git/`; U1 owns the hook text and the merge library.
- Do not touch the synced copy `.config/mise/tasks/_scripts/repo-sync` of
  this repository.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- The template delimiters are `<% %>`; gomplate in the script uses its own
  calls as the existing line `:1298` does.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add direct_commit to the branch tables of bootstrap.toml` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
