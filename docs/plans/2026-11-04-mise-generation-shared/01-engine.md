# U1 — Engine: the shared mise library with a side parameter

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `sync-engine/common` (the gomplate wrapper, the log
  function, the header constant), `sync-engine/values` (P07),
  `sync-engine/write`.

## Ruling

From the design rulings of 2026-11-04:

> - mise config: both sides generate `conf.d/<tool>/mise.<env>.toml` from
>   `[mise.*]` and `[mise_env.<env>.*]` of `tools.toml`; no tool ships a mise
>   fragment. Context7 (release notes 2026.9.14): global `conf.d` folder
>   fragments load `mise.toml` and `mise.<env>.toml`. Valid sections get
>   `[bootstrap.*]` and `[dotfiles]` on both sides. The 83 plain Mac
>   fragments move into `tools.toml`; the 6 templated ones (claude-cli, cli,
>   devtools, fish, macos-defaults, network: packages, aliases, PROJECTS_DIR,
>   dock position, login shell) stay open until the "Values" row.

> - The Mac side uses no environments (user, 2026-11-04): every generated Mac
>   mise config is `~/.config/mise/conf.d/<tool>/mise.toml`. **Why:** processes
>   without the shell startup files (first `mise bootstrap` of `setup`, Dock
>   apps such as VS Code, launchd, `ssh host cmd`) do not have `MISE_ENV=dev`,
>   and they lose `mise.dev.toml`. The engine keeps the shared code path for
>   `mise.<env>.toml`, so the Mac can use it later; until then a Mac tool with
>   `[mise_env.*]`, or an `env` key in the Mac selection, stops with exit 2.
>   The repo rule "no `conf.d/*/mise.toml`" stays for the repo side.

> - `requires` has one meaning: exact paths, in `[files]` tables. New key
>   `requires_found` = a list of file names searched down to 4 folders below
>   the root, valid in a mise entry and in a `[files]` table, repo side only
>   (Mac → exit 2). The `files` list of a mise entry goes: each file states
>   its condition in its own `[files]` table (python, node).

> - Values row, item 4: overrides use the repo shape on both sides.
>   `[overrides.mise.<path>]` valid on both (same rules, `# override:`
>   comment); `[overrides.mise_env.*]` → exit 2 on the Mac;
>   `[overrides.gitignore]` repo only (Mac `~/.config/git/ignore` is a git
>   tool file, changed with `[tools.git]` values). `[tools.<name>]` = values
>   for templates; `[overrides.mise.*]` = changes to generated mise config.
>   The repo rule "an override cannot add a key" is the input of item 5.

> - Values row, item 5 (values into generated mise config):
>   1. Map values (packages, aliases, `[env]`, defaults domains) move directly
>      into `[mise.*]` of `tools.toml` (`[defaults.packages]` →
>      `[mise.bootstrap.packages]`); changed with `[overrides.mise.*]`.
>   2. User additions go into `[mise.*]` of `macos.toml`; the engine generates
>      `~/.config/mise/conf.d/zz-macos/mise.toml` (loads last, wins; the Mac
>      twin of `zz-repo`). "An override cannot add a key" stays (else a new
>      alias lands in all 12 tools with `[shell_alias]`).
>   3. The engine renders the generated mise file with gomplate and the
>      tool's values, so a `[mise.*]` string can hold `<% %>` (e.g.
>      `PROJECTS_DIR = "<% .config.projects_dir %>"`); `required` and
>      `test.Fail` checks keep working.
>   4. A conditional section (fish `login_shell` bool) becomes a value:
>      `[mise.bootstrap.user] login_shell = "<% .config.login_shell %>"`.
>      All 6 templated Mac fragments become `tools.toml` tables.

Item 5.3 above says "renders the generated mise file". D6 supersedes that
mechanism: the engine renders the `tools.toml` before it generates, so the
generated file holds the rendered value.

> - User principle: every mise section (`[env]`, `[shell_alias]`,
>   `[settings]`, `[tasks]`, ...) exists at user level and at repo level, so
>   the tool model must not limit one side. ("env" here meant the mise
>   `[env]` section, not environments; I first misread it.)

D2: "The 16 plans form one linear chain through requires. Most of them edit
sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans conflict
there."

D5: "Only in dev, like [tasks], when the tool is in dev. CI is not a machine
to provision."

D6: "When the tools.toml of a tool contains <%, the engine renders that file
as a gomplate template into a temp copy, and the generator reads the copy, so
the generated file holds the rendered value. The render context has .config
(the values of the tool), .all (the values of every selected tool) and .env
(HOME, USER), on both sides. The overrides of the values file are not
rendered. A <% %> expression sits inside a TOML basic string ("..."), and
every string literal inside the expression is a Go raw string in backticks. A
failed render (a required or test.Fail check) stops the mise config of that
tool with code 2 on the current failure path: the existing conf.d files of the
tool stay. A rendered copy that is not valid TOML gives the message <tools
root>/<tool>/tools.toml: not valid TOML after render: <message>, code 2, for
that tool."

D7: "One file conf.d/<tool>/mise.toml for each tool with content. It writes
every valid section, the dev-only sections included, because the Mac uses no
environments. A [mise_env.*] table, the envs key of a mise tool and the
requires_found key of a mise tool are tool errors: the files of that tool
stay, the other tools continue, and the run returns 2. An
[overrides.mise_env.*] table is a values-file error: the whole run stops with
code 2 before any write. [overrides.mise.*] works as on the repo side, an
override cannot add a key, and the comment is # override: macos.toml (the file
name only)."

D8: "[overrides.gitignore] and its checks, the setup-all and setup-<submodule>
aliases of the mise tool (given to the library as a list of extra dev
aliases), the records of the generated files for the file loop and the
manifest, and the call of the library with side repo. Everything else of the
mise generation, the env rules and the [overrides.mise*] checks moves to
sync-engine/mise."

D9: "sync-engine/mise builds the context file itself for each tool whose
tools.toml holds <%, from sync-engine/values, also when the tool has no .tmpl
file. The file is in the work folder of the run; the library removes it after
the render of that tool, and the exit trap removes the work folder."

D10: "Only mise and mise_env. The Mac front-end (P09) stops with code 2 on
[overrides.gitignore], as on any key that the Mac values file does not have."

D11: "sync-engine/mise has a second entry for the own config of a values file.
It generates the [mise.*] tables of the values file into
<conf.d>/zz-macos/mise.toml, with the same section list, the header Synced
from bootstrap: macos.toml [mise]. Do not edit this file. plus the second line
The next sync replaces it, and a manual change is lost., and a render of <% %>
with .all and .env (no .config). [overrides.mise.*] does not apply to it."

## Edits

1. **`sync-engine/mise`** (new, mode 644, a library; the bootstrap header of a
   non-seed file in shell comment syntax; a short usage comment). Use the
   function prefix, the log function, the gomplate wrapper and the header
   constant of `sync-engine/common`. Move into it, from
   `sync-engine/repo-sync`, the defines, the overrides template and the mise
   and `mise_env` checks of `_rs_overrides`, the env rules of `_rs_mise_env`,
   the mise template and the render of `_rs_mise_render`. Rename the
   environment variables that the templates read to the `SYNC_ENGINE_MISE_`
   prefix (for example `SYNC_ENGINE_MISE_ENV`). Each function takes the side
   (`repo` or `macos`) as its first parameter. Any other side value is a usage
   error, code 2.
   - **Sections.** Add `bootstrap` and `dotfiles` to the valid sections. On
     the repo side they are in the dev-only list (D5). The other lists do not
     change.
   - **Render (D6, D9).** Before the generator reads the `tools.toml` of a
     tool, when the file contains `<%`, build its context file in the work
     folder of the run from `sync-engine/values`: `.config` (the values of the
     tool), `.all` (the values of every selected tool) and `.env` (`HOME`,
     `USER`). Render the `tools.toml` with gomplate into a temp copy, remove
     the context file, and give the copy to the generator as `t=`. Without
     `<%`, use the file as it is. A render error logs
     `<tools root>/<tool>/tools.toml: cannot render: <gomplate message>`; a
     copy that does not parse as TOML logs
     `<tools root>/<tool>/tools.toml: not valid TOML after render: <message>`.
     Both return 2 for that tool on the current failure path.
   - **Overrides (D10).** The library reads only the `mise` and `mise_env`
     keys of `[overrides]`.
   - **Side `repo`.** The output is byte-equal to today: the same file names
     `mise.<env>.toml`, the same header, the same `# override:` comments, the
     same messages.
   - **Side `macos` (D7).** One output `mise.toml` for each tool with content,
     every valid section in it, the extra dev aliases ignored. A
     `[mise_env.*]` table, or an `envs` key or a `requires_found` key in a
     `[mise.tools]` entry, is a tool error: a message that names the tool
     file, the files of that tool stay, the other tools continue, and the
     run returns 2. An `[overrides.mise_env.*]` table of the values file
     stops the whole run with code 2 before any write. `[overrides.mise.*]`
     uses the same rules as the repo side, with the comment
     `# override: macos.toml`.
   - **Own config (D11).** A second entry takes a values file, the conf.d
     folder and the selected tools. It generates the `[mise.*]` tables of the
     values file into `<conf.d>/zz-macos/mise.toml`, with the same section
     list and the macos section rules, the header
     `# Synced from bootstrap: macos.toml [mise]. Do not edit this file.` +
     `# The next sync replaces it, and a manual change is lost.`, and a render
     of `<% %>` with `.all` and `.env` (no `.config`). `[overrides.mise.*]`
     does not apply to it. No values-file `[mise]` table: no file.
   - **Header.** The generated header names `<tools root>/<tool>/tools.toml`,
     where the caller gives the tools root (`repo/tools` or `macos/tools`).
2. **`sync-engine/repo-sync`** (D8): source `sync-engine/mise`; remove the
   moved code; call the library with side `repo`, the tools root, the values
   file, the extra dev aliases of the `mise` tool and the context of P07. Keep
   `[overrides.gitignore]` and its checks, `_RS_GENERATED` and the rows of
   the manifest. `repo_sync --check` keeps its local-only override checks:
   the library offers the same local-only mode that `_rs_overrides` has
   today. Update the header comment of the script to point to
   `sync-engine/mise` for the mise rules; keep it short.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n sync-engine/mise sync-engine/repo-sync` passes.
- `grep -n '_RS_MISE_TEMPLATE\|_RS_OVERRIDES_TEMPLATE' sync-engine/repo-sync`
  finds nothing.
- `grep -n '"bootstrap" "dotfiles"\|bootstrap.*dotfiles' sync-engine/mise`
  finds the new sections.
- `mise x -- mise run test:repo-sync` passes with every existing case
  unchanged (the repo output is byte-equal).

## Guardrails

- Do not touch `.config/mise/tasks/test/repo-sync`; U2 owns it.
- Do not change the repo output, the messages or the codes of the existing
  rules.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- Call gomplate only through the wrapper of `sync-engine/common`; keep the
  delimiters `<% %>`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: share the mise config generator between the repo and macos sides` —
written by the orchestrator after the unit gate and the wave gate, not by the
unit.
