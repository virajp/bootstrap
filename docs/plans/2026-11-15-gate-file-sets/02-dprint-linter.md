# U2 — The dprint, linter and ShellCheck gates

- **Read first:** every owned file that an edit below names, top to bottom:
  `repo/tools/dprint/tools.toml`,
  `repo/tools/dprint/files/dprint.jsonc.merge.toml`,
  `repo/tools/dprint/tasks/code-format-dprint`, `repo/tools/linter/tools.toml`,
  `repo/tools/linter/tasks/code-lint-linter`. Then the U1 result:
  `repo/tools/mise/tasks/helpers` (`gate_files`) and
  `repo/tools/mise/tasks/gate_files.py` (read only).
- **Lazy-load:** the installed linter, for its file types and its built-in
  ignores: `$(dirname $(dirname $(readlink -f $(mise x -- which
  linter))))/@virajp.dev/linter/dist/config/ignores.js`
  and `dist/config/presets/*.js` (the `files:` globs).

## Ruling

- **D13:** each gate calls `gate_files "$0"` with its flags and files, and
  gives the list file to its tool. An empty list: exit 0 with no output.
- **D14:** dprint reads the list file with `--stdin-files`; the linter and
  ShellCheck get it through `xargs`.
- **D7:** the gates drop every own filter. User: "I am more keen on
  standardizing the approach so it becomes easier to debug".
- **D8, D9:** the dprint merge spec renders the `excludes` of dprint.jsonc
  from `.all.mise.code_files.excludes` (the global list of the mise tool). The
  `excludes` key of the dprint `[defaults]` goes; its three values are the
  mise defaults now.
- **D12:** `05-dprint` has no `#FILES` header.
- **D10:** the `#FILES` header of `code-lint-linter` includes the file types
  of `@virajp.dev/linter` 2.1.0 (`*.astro *.css *.html *.json *.jsonc *.md
  *.toml *.yaml *.yml *.js *.mjs *.cjs *.jsx *.ts *.mts *.cts *.tsx`, and the
  JSONC file names that its json preset names) and excludes its built-in
  ignore list, written as the same patterns. User: yes.
- **D11:** ShellCheck moves to a new gate `code:lint:06-shellcheck` of the
  linter tool, header `#FILES include="*.sh *.bash" shebang="sh bash dash ksh"
  exclude="*.tmpl"`. It runs `shellcheck --severity=warning` as today. User:
  yes.
- **D17:** each gate declares `#USAGE flag "--all"`, `--fix`, `--debug` and the
  variadic `[files]`.
- **D3, D4:** the gate names its lists in its `#FILES` header and gets its
  list from `gate_files`; the orchestrator computes no list.
- **D19:** U3 edits the other gates in the same wave; the two units share no
  file.
- **D22:** add `"!**/*.py"` to `[defaults.blocks.shuck] associations` of
  `repo/tools/dprint/tools.toml`, so shuck never formats `gate_files.py`.
- **D23:** each `xargs` call is `tr '\n' '\0' < "<list file>" | xargs -0
  <tool>`.
- **D24:** right after the `gate_files` call: `trap 'rm -f "<list file>"'
  EXIT`.
- **D25:** the merge spec renders each pattern in the dprint form: a pattern
  without `/` gets the prefix `**/`, and a pattern with a trailing `/` gets
  `**` after it.
- **D27:** the helper drops a `*.sh` file whose shebang is zsh, so the
  ShellCheck gate needs no zsh check.
- **D38 (replaces the removal of the dprint `excludes` key in D9):** the
  dprint tool keeps `excludes = []` in `[defaults]`, so a consumer value
  `[tools.dprint] excludes` stays a known key. The merge spec renders
  `.all.mise.code_files.excludes`, then `.config.excludes`.
- **D40 (the dprint form, with D25):** a leading `/` is removed and the rest
  stays as it is. A pattern without `/` gets the prefix `**/`. A pattern
  whose only `/` is trailing (`build/`) becomes `**/build/**`. A pattern with
  an inner `/` and a trailing `/` becomes `<pattern>**`.
- **D46, D52:** each pattern without any `/` (such as `node_modules` or
  `*.svg`) renders as two dprint patterns: `**/<p>` and `**/<p>/**`.
- **D44:** put the `#FILES` line right after the `#USAGE` lines of the gate.
- **D47:** the helper decides by the interpreter of the `#!` line, with the
  known shells `sh bash dash ksh mksh zsh fish csh tcsh`.
- **D36:** dprint also applies the own excludes that a repository keeps in its
  `dprint.jsonc`; that is accepted, and the migrate item of U6 moves them.

## Edits

1. **`repo/tools/dprint/tasks/code-format-dprint`** — declare `--all`; get the
   list with `gate_files`; with an empty list, exit 0 with no output; run
   `mise x -- dprint <check|fmt> --allow-no-files --log-level <level>
   --stdin-files < <list file>`. Remove the two branches of today (files, or
   the whole config). Change the `[files]` help: "the default is the
   uncommitted files".
2. **`repo/tools/dprint/tools.toml`** — set `excludes = []` in `[defaults]`
   (D38) and change its comment: the excludes come from `code_files` of the
   mise tool. Add `"!**/*.py"` to `[defaults.blocks.shuck] associations`
   (D22).
3. **`repo/tools/dprint/files/dprint.jsonc.merge.toml`** — the `excludes`
   items range over `.all.mise.code_files.excludes`, then
   `.config.excludes` (D38), each written in the dprint form of D40. Change
   its comment.
4. **`repo/tools/linter/tasks/code-lint-linter`** — declare `--all`; add the
   `#FILES` header (D10); get the list with `gate_files`; run the linter
   through `xargs` with `--fix` and `--debug` as today. Remove the ShellCheck
   part.
5. **`repo/tools/linter/tasks/code-lint-shellcheck`** (new) — the
   `Synced from bootstrap` header, `#MISE description="Check the shell
   scripts with ShellCheck"`, `#MISE hide=true`, the four `#USAGE` lines, the
   `#FILES` header (D11). Without `shellcheck` (`mise x -- which shellcheck`),
   exit 0 with no output. With a list: `print_header "Checking the shell
   scripts with ShellCheck ..."`, then `xargs` runs
   `mise x -- shellcheck --severity=warning`.
6. **`repo/tools/linter/tools.toml`** — add `[tasks."code/lint/06-shellcheck"]`
   with `source = "code-lint-shellcheck"`, beside the entry of `05-linter`.

## Verification

- `mise x -- mise run code:format` passes.
- `grep -rn -- '--force-exclude\|shell_files_in_scope\|ls-files'
  repo/tools/dprint/tasks repo/tools/linter/tasks`
  finds nothing.
- `bash -n` passes on each changed task file.

## Guardrails

- Touch nothing outside `repo/tools/dprint/` and `repo/tools/linter/`. Do not
  edit `repo/tools/mise/` (U1's); report a needed change there as a QUESTION.
- Do not change the pre-commit hook entries.
- Write shell for bash 3.2 and bash 5.x. Quote the list file path.
- Delete with `rm`, never `git rm`.

## Commit

`feat: the dprint, linter and ShellCheck gates take their files from gate_files`
