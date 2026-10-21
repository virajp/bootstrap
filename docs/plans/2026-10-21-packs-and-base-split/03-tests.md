# U3 — Test repo-sync at each commit that changes tools/

- **Wave:** 3
- **Depends on:** U2
- **Owns:** `.config/mise/tasks/test/repo-sync`,
  `.config/repo-sync/pre-commit.yaml`
- **Model:** opus
- **Kind:** edit
- **Read first:** `tools/repo-sync/.config/mise/tasks/_scripts/repo-sync`
  (the header: return codes and environment variables), `tools/packs.toml`,
  `tools/mise/.config/mise/tasks/_scripts/helpers` (the print functions).

## Ruling

Quoted from `index.md`:

- D17: "'Tests must be in `bootstrap` and must run locally before merge to
  main'. `test:repo-sync` is owned by this repository; a local hook in
  `.config/repo-sync/pre-commit.yaml` runs it: 'run the hook on every commit
  that changes tools/'."
- D16: "`repo-sync` skips `pre-commit autoupdate` when
  `REPO_SYNC_NO_AUTOUPDATE=1`, so the tests need no network."
- D5, D6, D8 (the error codes), as `index.md` states them.

## Edits

1. **`.config/mise/tasks/test/repo-sync`** (new, mode 755, bash, no synced
   header, because this repository owns it):
   - `#MISE description="Test repo-sync against scratch values files"`.
   - It copies the working tree's `tools/` into a temporary git repository and
     commits it, so the test sees uncommitted work. It also makes a temporary
     consumer repository with a `main` branch.
   - For each case it writes the values file (and, for the pack cases, a
     changed `tools/packs.toml` or `tools.toml` in the copy), sources
     `repo-sync` from the copy, and runs `repo_sync --dry-run` with
     `REPO_SYNC_CLONE`, `REPO_SYNC_VALUES`, `MISE_PROJECT_ROOT` and
     `REPO_SYNC_NO_AUTOUPDATE=1`.
   - Cases and expected exit codes:

     | Case                                                                      | Code |
     | ------------------------------------------------------------------------- | ---- |
     | no `[tools]`, `environments = ["dev"]`                                    | 0    |
     | `[tools.node]` with `env = ["dev", "ci"]`, `environments = ["dev", "ci"]` | 0    |
     | `tools = ["node"]` (the old list)                                         | 2    |
     | a pack loop `a -> b -> a`                                                 | 2    |
     | a pack named like a tool folder                                           | 2    |
     | `[packs.nope]` (unknown pack)                                             | 3    |
     | `env = ["ci"]` on a tool whose mise tool allows `["dev"]` only            | 2    |
     | `env = ["ci"]` not in `environments = ["dev"]`                            | 2    |
     | a dependency whose `envs` lacks the environment of its parent             | 2    |

   - The first case also checks that the output lists a file of
     `_security_base` (the `required` tool) and of each tool of `base`, and no
     `conf.d/_base/`.
   - It prints one `print_header`, one status line for each case, and exits 1
     if a case fails. It removes the temporary folders on exit.
2. **`.config/repo-sync/pre-commit.yaml`** (new): a YAML list with one
   `repo: local` entry and one hook: `id: test-repo-sync`,
   `entry: mise x -- mise run test:repo-sync`, `language: system`,
   `files: '^tools/'`, `pass_filenames: false`.

## Verification

- `mise x -- mise run test:repo-sync` passes, and each case shows its status.
- With one case changed to expect a wrong code, the task exits 1 (then undo
  the change).
- `bash -n` and `mise x -- shellcheck -x -s bash` pass for the task.
- `mise x -- mise run code:precommit` passes (the wave gate).

## Guardrails

- Touch only the two owned files.
- The task needs no network and writes nothing outside the temporary folders.
- Write shell for bash 3.2 and 5.x.

## Commit

`test: add test:repo-sync and its hook`
