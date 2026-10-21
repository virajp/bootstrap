# U1 — Split the base tool into ten tools, and add the packs

- **Wave:** 1
- **Depends on:** —
- **Owns:** `tools/**`. For the three engine files of U2 (`_scripts/repo-sync`,
  `setup/configs/_default`, `setup/configs/sync`) this unit only moves the file
  and changes its header line; U2 edits the content.
- **Model:** opus
- **Kind:** edit
- **Read first:** `tools/base/tools.toml`, then every file under `tools/base/`,
  `tools/claude/tools.toml`, `tools/claude/.config/mise/tasks/setup/ai`,
  `tools/vscode/tools.toml`, `docs/design/packs.md`.
- **Lazy-load:** the other `tools/*/tools.toml` files, when you add
  `default_envs` to them.

## Ruling

Quoted from `index.md`:

- D3: "Each `tools.toml` has a fifth necessary key `default_envs`, a list ...
  An entry without `env` gets `default_envs`."
- D7: "`tools/packs.toml` has `[pack.<name>]` with `depends` (tool folders, or
  `p:<name>` for a pack) and optional `required` (tools always installed, not in
  `depends`)."
- D9: "The orchestrators stay in their tools: 'That's not possible since they
  are part of `base` pack and that is not removable'."
- D10: "`security` is a pack: `depends = ["gitleaks", "grype", "osv-scanner"]`,
  `required = ["_security_base"]`. `_security_base` holds `code:sec` and the
  `sec` hook; each scanner ships `code:sec:<tool>`."
- D11: "`code/sec/_default` runs `own_tasks code:sec` and passes `--staged` and
  `--debug` to each. `code:sec:gitleaks` keeps both modes. `code:sec:grype` and
  `code:sec:osv-scanner` print nothing and exit 0 with `--staged`."
- D12: "`dprint.dprint` → `dprint`; `hverlin.mise-vscode`,
  `redhat.vscode-yaml`, `tamasfe.even-better-toml` → `mise`;
  `dbaeumer.vscode-eslint`, `foxundermoon.shell-format` → `linter`;
  `gruntfuggly.todo-tree`, `oderwat.indent-rainbow`,
  `pkief.material-icon-theme`, `usernamehw.errorlens` → `vscode`."
- D13: "keep `claude` as is ... move `code:graph` inside `claude` ... add logic
  in `setup:ai` where it silently skips when it's being run inside
  git-submodule ... `claude` will NOT be `root_only`".
- D14: "In the `mise` tool with `envs = ["dev", "ci"]`."

## Edits

Move each file with its git mode (use `mv`, never `git mv`). In each moved
file, change the header path `tools/base/<path>` to `tools/<new tool>/<path>`.
Keep all other content byte for byte, except where an edit below says so.

1. **The ten new tool folders.** Each gets a `tools.toml` with the five keys
   `depends`, `default_envs`, `extensions`, `root_only = false` and `seed`, a
   one-line comment that says what the tool is, and its `[mise.*]`,
   `[mise_env.*]` and `[[precommit]]` tables, moved from `tools/base/tools.toml`
   with their comments:

   | Tool             | Files from `tools/base/`                                                                                                                                                                                                                                                     | `tools.toml` content                                                                                                                                                                                                                                                                   | `default_envs`  |
   | ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------- |
   | `mise`           | `.config/mise.toml.tmpl`, `.config/mise/conf.d/_repo/mise.ci.toml`, `.config/mise/conf.d/_repo/mise.dev.toml`, `.config/mise/tasks/_scripts/helpers`, `checks`, `placeholder`, `setup/_default.tmpl`, `setup/mise`, `setup/worktree.tmpl`, `code/_default`, `code/worktrees` | seeds: the two `_repo` files; mise tools `jq` and `yq` with `envs = ["dev", "ci"]`; `[mise.shell_alias] setup`; `[mise.tasks.init]`; `[mise_env.dev.settings]` and `[mise_env.ci.settings]` (`os`); extensions `hverlin.mise-vscode`, `redhat.vscode-yaml`, `tamasfe.even-better-toml` | `["dev", "ci"]` |
   | `repo-sync`      | `.config/mise/tasks/_scripts/repo-sync`, `setup/configs/_default`, `setup/configs/sync`                                                                                                                                                                                      | mise tool `gomplate`                                                                                                                                                                                                                                                                   | `["dev"]`       |
   | `git`            | `.gitignore`, `.config/git-conventional-commits.yaml.tmpl`, `.config/mise/tasks/_scripts/merge`, `code/git-config`, `code/merge/_default`, `code/merge/develop`, `code/merge/main`                                                                                           | hooks: the local `git-config` hook; a `[[precommit]]` for `https://github.com/pre-commit/pre-commit-hooks` (same `rev`) with only `no-commit-to-branch`; the `conventional-commits` repo entry                                                                                         | `["dev", "ci"]` |
   | `pre-commit`     | `.config/pre-commit-config.yaml.tmpl`, `setup/precommit`, `code/precommit`                                                                                                                                                                                                   | mise tool `pre-commit` (depends `python`, `uv`); `[mise.env] PRE_COMMIT_HOME`; `[mise.shell_alias] precommit`; the `pre-commit-hooks` entry with every hook except `no-commit-to-branch`                                                                                               | `["dev"]`       |
   | `dprint`         | `code/format/_default.tmpl`, `setup/dprint`                                                                                                                                                                                                                                  | mise tool `dprint`; the local `format` hook; extension `dprint.dprint`                                                                                                                                                                                                                 | `["dev", "ci"]` |
   | `linter`         | `code/lint/_default.tmpl`, `setup/linter`                                                                                                                                                                                                                                    | mise tools `npm:@askviraj/linter` and `shellcheck`; the local `lint` hook; extensions `dbaeumer.vscode-eslint`, `foxundermoon.shell-format`                                                                                                                                            | `["dev", "ci"]` |
   | `_security_base` | `code/sec` becomes `code/sec/_default` (edit 3)                                                                                                                                                                                                                              | the local `sec` hook                                                                                                                                                                                                                                                                   | `["dev", "ci"]` |
   | `gitleaks`       | `.config/gitleaks.toml.tmpl` (a seed); new `code/sec/gitleaks`                                                                                                                                                                                                               | seed `.config/gitleaks.toml.tmpl`; mise tool `gitleaks`                                                                                                                                                                                                                                | `["dev", "ci"]` |
   | `grype`          | `.config/grype.yaml` (a seed); new `code/sec/grype`                                                                                                                                                                                                                          | seed `.config/grype.yaml`; mise tool `grype`                                                                                                                                                                                                                                           | `["dev", "ci"]` |
   | `osv-scanner`    | `.config/osv-scanner.toml` (a seed); new `code/sec/osv-scanner`                                                                                                                                                                                                              | seed `.config/osv-scanner.toml`; mise tool `osv-scanner`                                                                                                                                                                                                                               | `["dev", "ci"]` |

   The task paths above are under `.config/mise/tasks/`. Keep the order of the
   hooks inside each tool as it is in `tools/base/tools.toml`, and keep a real
   tag in each `rev`.

2. **`tools/packs.toml`** (new, no header, because it is never synced):

   ```toml
   # The packs. A pack is a named list of tools; it has no files of its own.
   # repo-sync always adds the pack `base`. See docs/design/packs.md.
   [pack.base]
   depends = [
     "mise",
     "repo-sync",
     "git",
     "pre-commit",
     "dprint",
     "linter",
     "p:security",
   ]

   [pack.security]
   depends = [ "gitleaks", "grype", "osv-scanner" ]
   required = [ "_security_base" ]
   ```

3. **`code:sec` split.**
   - `tools/_security_base/.config/mise/tasks/code/sec/_default`: keeps the
     `#MISE description`, the `--debug` and `--staged` flags and the header of
     the old `code/sec`. It prints its header, then runs
     `mise run "${task}" <flags>` for each `task` of `own_tasks code:sec`,
     passing `--staged` and `--debug` when set.
   - `tools/gitleaks/.config/mise/tasks/code/sec/gitleaks`: lines 15-20 (the
     `--staged` mode) and 46-67 (the full mode) of the old `code/sec`, with the
     `SCAN_DIR` copy and the EXIT trap.
   - `tools/grype/.config/mise/tasks/code/sec/grype`: lines 24-37 (the
     submodule excludes of grype, and grype). With `--staged` it prints nothing
     and exits 0.
   - `tools/osv-scanner/.config/mise/tasks/code/sec/osv-scanner`: the
     submodule excludes of osv-scanner and lines 39-44. With `--staged` it
     prints nothing and exits 0.
   - Each scanner task is a leaf: it prints `print_subheader` lines under the
     header of the orchestrator, as the old file did, sources
     `_scripts/helpers`, and prefixes each tool call with `mise x --`. Each
     file is executable (mode 755) and has the synced header of its tool.

4. **`claude`.** In `tools/claude/tools.toml`: `root_only = false`; add
   `default_envs = ["dev"]`; replace the comment "Submodules inherit the tools
   from the root" with one that names `setup:ai`, `code:graph`, graphify and
   mempalace; add the `graphify-refresh` hook (moved from
   `tools/base/tools.toml`, with its comment). Move `code/graph` to
   `tools/claude/.config/mise/tasks/code/graph`. In
   `tools/claude/.config/mise/tasks/setup/ai`, add a check before the header:
   if `git -C "${MISE_PROJECT_ROOT}" rev-parse --show-superproject-working-tree`
   prints a path, exit 0 and print nothing.

5. **`vscode`.** Add the four editor preferences of D12 to `extensions` in
   `tools/vscode/tools.toml`, sorted.

6. **`default_envs` in every other tool.** Add `default_envs` to each
   `tools/*/tools.toml` that does not have it. The value is the list of
   environments that every mise tool of the tool allows in its `envs` (a mise
   tool without `envs` allows `["dev"]`). A tool with no mise tool gets
   `["dev"]`. Place the key after `depends`.

7. **Delete `tools/base/`** with `rm -r` after every file is moved. No file of
   `tools/base/` may stay unassigned; if one has no owner in edit 1, report it
   as a `GAP:` and move it to `mise`.

## Verification

- `test ! -e tools/base`
- Each of the ten new folders and `tools/claude`, `tools/vscode` has a
  `tools.toml` with the five keys;
  `rg --files-without-match '^default_envs' tools/*/tools.toml` prints
  nothing.
- `rg -n 'tools/base' tools/` prints nothing.
- Each `[mise.tools.*]` name appears in only one `tools.toml`.
- `bash -n` passes for each moved or new task script; each has mode 755 where
  the old file had it.
- `mise x -- mise run code:precommit` passes (the wave gate).

## Guardrails

- Touch nothing outside `tools/`.
- Do not change the logic of `_scripts/repo-sync` or `setup/configs/*`: U2
  owns it. A wrong path inside them stays for U2.
- Move with `mv`; delete with `rm`, never `git rm` or `git mv`.
- Write shell for bash 3.2 and 5.x.

## Commit

`refactor: split the base tool into ten tools and add the packs`
