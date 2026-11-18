DECISION 2026-11-15 ★4 | repo:bootstrap, repo/tools/mise,
repo/tools/dprint, repo/tools/linter, the code:format and code:lint gates |
the gates examine the uncommitted files by default and every file with --all;
one Python helper, gate_files, makes the list of each gate from a global
code_files list and a #FILES header, and is the only place that removes a file
| alt rejected: the staged files only, per-gate git ls-files, tool config
excludes, the orchestrator computes the lists, all in bash

# The file sets of the code:format and code:lint gates

The user settled this in the artifact "Gate File Sets" (Q1 to Q8), then in the
plan interview of `docs/plans/2026-11-15-gate-file-sets` (its decisions D1 to
D59). The user asked: "We need mechanism where by default only uncommitted (or
staged) files are formatted and lint'ed."

## Reversal

The CLAUDE.md gate rule of "Task scripts" is reversed (D20). It said: a gate
without files lists them with
`git ls-files --cached --others --exclude-standard`, and a tool that scans
its project itself (ruff, pyrefly, detekt) uses that list only to find whether
a file exists. Now `gate_files` makes the list in every mode, and each tool
gets only that list. The name `--all` of `code:format` and `code:lint` comes
back with a new meaning, every file of the repository, after D100 of
`2026-11-14-config-folder` removed the member flag.

## The rulings

- **The default set (D1).** Without files and without `--all`, a gate examines
  the uncommitted files: the staged files
  (`git diff --cached --name-only --diff-filter=d`), the changed files that
  are not staged (`git ls-files --modified`) and the new files
  (`git ls-files --others --exclude-standard`). Rejected: the staged files
  only; the uncommitted files plus the commits of the branch.
- **The global list (D2, D9).** The key `code_files` of the `mise` tool, with
  `includes` and `excludes`; an empty `includes` means every file. Its
  `[defaults]` has the old dprint excludes (`**/*-lock.yaml`, `**/*.svg`,
  `**/package-lock.json`). Another tool adds patterns in
  `[contributes.mise]`, a repository in `[tools.mise]`. The sync renders the
  lists to `.config/bootstrap/config-files.toml`. Rejected: a hand-written
  file; the dprint tool owns the list; the path `.config/code-files.toml`.
- **The list of a gate (D3, D44, D53).** A
  `#FILES include="..." exclude="..."` header in the top block of the gate
  file, also for a gate of a repository. No header, no `include` key or
  `include=""`: every file. Two `#FILES` lines, an unknown key, a key twice or
  a value without double quotes stop with exit 2. Rejected: keys of the task
  entry in `tools.toml`.
- **The helper (D4, D5, D13, D34).** Each gate calls `gate_files "$0"` of
  `helpers` with its flags and files; the orchestrator computes no list. The
  file logic is Python, `_scripts/gate_files.py`, run with
  `mise x python@3.14 -- python3`; the gates and the orchestrators stay bash.
  The helper writes the list to a temporary file, one path for each line, and
  prints its path; the gate deletes it with a `trap` on `EXIT` (D24).
  Rejected: the orchestrator computes each list; all in bash;
  `mise x -- python3`, which gives the system Python outside dev and ci.
- **The gate tasks stay (D6).** A tool keeps its gate in its own folder, and a
  repository can add a gate. Rejected: the tool commands move into the
  orchestrator.
- **One place removes a file (D7, D8).** In a gate run, `gate_files` is the
  only place that removes a file: the gates drop `--force-exclude` and every
  own filter, and `--debug` prints each removed file as
  `drop <path>: <layer> <kind> ["<pattern>"]` on stderr. The sync also renders
  the `code_files` excludes into `dprint.jsonc`, for an editor and for the
  formatter of the sync; a seed such as `.config/swiftlint.yml` keeps its own
  text. Rejected: the helper filters first and each tool also applies its
  config excludes; a flag for each tool.
- **The repository filter (D16, D48, D59).** The helper never recurses into a
  submodule. It drops, in this order: a path with a line break or a leading
  `-` (`repo unsafe-name`, D59, the U5 security fix), a path that git does not
  list (`repo unlisted`), a gitlink (`repo gitlink`), a nested repository
  (`repo nested-repo`), a symlink (`repo symlink`) and a path that is not a
  regular file (`repo deleted`). Then the global filter (`global include`,
  `global exclude "<pattern>"`) and the gate filter (`gate include`,
  `gate exclude "<pattern>"`, `gate shebang`) (D33). D59 amends D33 and D48
  for `repo unsafe-name`. Rejected: `git ls-files --recurse-submodules`;
  keep a symlink to a regular file.
- **The patterns (D12, D25, D40, D46, D52).** The gitignore glob form without
  negation. The dprint merge spec renders each pattern in the dprint form; a
  pattern `p` without `/` gives `**/p` and `**/p/**`.
- **ShellCheck (D11, D27, D47).** ShellCheck moves to its own gate
  `code:lint:06-shellcheck` of the `linter` tool, with
  `#FILES include="*.sh *.bash" shebang="sh bash dash ksh" exclude="*.tmpl"`.
  A zsh file is not checked. `shell_files_in_scope` of `helpers` goes.
  Rejected: the linter gate keeps both tools.
- **The linter gate (D10).** Its `#FILES` header includes the file types of
  `@virajp.dev/linter` 2.1.0 and excludes its built-in ignore list. The linter
  package does not change; an option to turn off its list is parked.
- **The flags (D17, D28, D30, D37, D42).** `code`, `code:format`,
  `code:lint` and each gate declare `--all`; `--all` with files stops with
  exit 2. The orchestrators pass `--all` only to a gate that declares it.
  `gate_files.py` accepts `--fix` and ignores it. The pre-commit hooks do not
  change: they pass the staged files.
- **The argument limit (D14, D23, D26).** dprint reads the list with
  `--stdin-files`, detekt through its argument file, and each other tool
  through `tr '\n' '\0' | xargs -0`. A path with a line break is not
  supported. Rejected: arguments as before.
- **Module gates (D12, D31, D54).** android, go, gomod and uv map the files of
  the list to their modules. A change that only deletes files gives an empty
  list; `--all` covers it.
- **Task output (D45).** With `--debug`, the drop lines go to stderr and can
  come before the header of a gate.

## Non-goals (D21) and the hand steps (D32, D35, D36, D38)

- `code:sec` and its scanners keep their own file lists. The shipped PR
  templates and CI do not change. The SwiftLint seed keeps its text.
- M45 of `docs/sync/migrate.md` moves the excludes of `.config/linter.yaml`,
  `dprint.jsonc`, `[tools.dprint] excludes` and the tool configs into
  `code_files.excludes`, adds `"!**/*.py"` to the shuck associations, and adds
  `--all` and `gate_files` to the own gates. This repository does it at its
  next self-sync, by step 4 of "Change a shared file" of CLAUDE.md.
- The vwf gate lines of a consumer must use `--all`: parked for claude-plugins
  (D55).
