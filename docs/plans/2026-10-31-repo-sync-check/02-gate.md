# U2 — Gate: code:lint:repo-sync

- **Read first:** `tools/python/.config/mise/tasks/code/lint/ruff` (the model
  gate), `tools/repo-sync/.config/mise/tasks/setup/configs/sync` (how a task
  loads repo-sync).
- **Lazy-load:** `tools/linter/.config/mise/tasks/code/lint/_default.tmpl`,
  `tools/mise/.config/mise/tasks/_scripts/helpers`.

## Ruling

D1: "User: \"repo_sync --check will be better way of doing this without
needing the bootstrap clone.\" A --check mode of repo_sync, with no bootstrap
clone, no REPO_SYNC_CLONE and no network."

D4: "A gate task code:lint:repo-sync in the repo-sync tool. code:lint runs each
code:lint:<name> task with the staged files, and the existing lint hook runs
code:lint, so no new hook is added. It follows the gate rules of CLAUDE.md:
without gomplate (mise x -- which gomplate) it prints nothing and exits 0;
with a file list that does not include .config/bootstrap.toml it prints
nothing and exits 0; with no file list it runs when .config/bootstrap.toml
exists."

D12: "The gate has no #MISE tools directive. The repo-sync tool already
declares [mise.tools.gomplate], so mise run setup installs it in each
consumer. The gate finds it with mise x -- which gomplate; without it, the
gate prints nothing and exits 0, so a commit never needs the network."

## Edits

1. **`tools/repo-sync/.config/mise/tasks/code/lint/repo-sync`** (new,
   executable, `755`) — a bash task in the shape of the model gate:
   - the shebang, then the bootstrap header for
     `tools/repo-sync/.config/mise/tasks/code/lint/repo-sync`;
   - `#MISE description="Validate .config/bootstrap.toml with repo_sync
     --check"`,
     `#MISE hide=true`, and no `#MISE tools` directive (D12);
   - `#USAGE flag "--debug"`, `#USAGE flag "--fix"` (accepted and ignored:
     `code:lint` passes it to each gate), and
     `#USAGE arg "[files]" var=#true` with help that says only
     `.config/bootstrap.toml` is used;
   - source `helpers`; the applies-checks of D4, in that order, each with no
     output; then `print_header "Validating .config/bootstrap.toml ..."`;
   - unset `REPO_SYNC_CLONE`, source
     `${MISE_PROJECT_ROOT}/.config/mise/tasks/_scripts/repo-sync`, and run
     `repo_sync --check`; exit with its code.
   - A file list is relative to `MISE_PROJECT_ROOT`; match the exact path
     `.config/bootstrap.toml`.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n tools/repo-sync/.config/mise/tasks/code/lint/repo-sync` passes, and
  `test -x` on the file passes.
- `head -3` of the file shows the shebang and the two header lines.

## Guardrails

- Touch only the one new file. Do not add a `[[precommit]]` table or change
  `tools/repo-sync/tools.toml` (D4).
- Do not edit a doc; report each doc that the new task makes wrong as
  `DOCS FALSIFIED:`.
- Write shell for bash 3.2 and bash 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`feat: add the code:lint:repo-sync gate for bootstrap.toml` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
