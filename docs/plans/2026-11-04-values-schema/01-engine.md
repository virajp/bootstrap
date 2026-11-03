# U1 — Engine: the values library and the repo context

- **Read first:** every owned file, top to bottom, before editing; B 128-160,
  455, 587-716 and 744-857 (the path is in the facts; read only).
- **Lazy-load:** `sync-engine/common` (the jq and gomplate wrappers, the log
  function), `sync-engine/select` (the selection and the depends closure).

## Ruling

From the design rulings of 2026-11-04:

> - Values row, item 2: one values file on each side. The values of a tool
>   are keys of its `[tools.<name>]` table, which also selects it; a table
>   for a base tool or a dependency only gives values. No `<tool>.toml`
>   files, no Mac `[overrides.<tool>]`, no "a `<tool>.toml` does not select"
>   rule. Repo: `.config/bootstrap.toml`, `[tools.<name>]` keeps `env` as its
>   only key until item 3. Mac file: `macos.toml` at the root of the private
>   `macos-setup` repo (the user wants the type in the file name; `macos`
>   matches the side name). Repo file keeps `bootstrap.toml` (a rename needs
>   a consumer migration for no function).

> - Values row, item 3: the shared `tools.toml` schema accepts `[defaults]`,
>   `[questions]` (`question`, `detect`, `example`; the guide asks, the
>   engine only validates) and `[contributes.<tool>]` on both sides. Merge
>   order: defaults, then contributions, then `[tools.<name>]`. The repo side
>   moves none of its existing top-level keys (`commit_scopes`, `extensions`,
>   `[claude.plugins.*]`, `[branch.*]`); a new repo tool value goes into
>   `[tools.<name>]`. Repo questions stay in `docs/sync/tools/<tool>.md`; a
>   new repo tool may use `[questions]`. `[[precommit]]` is a special-name
>   contribution; a later move to `[contributes]` is possible.

D2: "The 16 plans form one linear chain through requires. Most of them edit
sync-engine/repo-sync or sync-engine/setup-sync, and parallel plans conflict
there."

D5: "Port the rules of setup-sync as they are. A base or dependency tool is a
tool that pack base adds, also through a nested default = true pack
("pack:security": gitleaks, grype, osv-scanner, and _security_base through
depends), or a tool in the depends closure of a selected tool, a tool that a
mise depends adds included; a <tool> = {} member that the pack does not add
does not count. It is a base or dependency tool also when a [tools.<name>]
table names it. For such a tool, a list in [tools.<name>] appends its new
items, and a value whose type differs from the merged value stops the sync
with code 2. For any other tool, a value of [tools.<name>] replaces the
merged one, and tables merge key by key."

D6: "No. env stays the selection key of the repo side (a list, as today) and
is not in .config. Every other key of [tools.<name>] is a value. A
[packs.<name>] table keeps env as its only key."

D7: "No. The engine validates only the shape of [questions]. A template that
needs an answer uses required and fails that file with its message. Exit 4
stays Mac-only."

D8: "_rs_context adds .all (each selected tool name to its values, an empty
table for a tool without values) to the shared context. For each tool that
has a .tmpl file, the engine writes one context file: the shared context plus
config, the values of that tool. config and all are engine keys, so they win
over a top-level key of the same name in the values file, as members and
tools do today."

D9: "No. [defaults], [questions] and [contributes] are in the tools.toml files
of the clone, so they are clone rules, and only the sync examines them.
--check keeps its local rule for [tools.<name>]: env, when present, is a
list; any other key is now valid."

D10: "A contribution to a tool that bootstrap does not have gives the message
<tools root>/<tool>/tools.toml: [contributes.<x>] names a tool that bootstrap
does not have, with the path from the tools root variable (repo/tools). The
library logs every message and returns the highest code, as the sync does."

## Edits

1. **`sync-engine/values`** (new, mode 644, a library; the bootstrap header of
   a non-seed file in shell comment syntax; a short usage comment). Use the
   function prefix, the log function and the jq and gomplate wrappers of
   `sync-engine/common`. It defines one entry function with these inputs: the
   selected tool names, the base and dependency tool names (D5), the tools
   root of the clone, and the user values of each tool as one JSON object
   (tool name → table; the front-end builds it). It prints
   `ctx<TAB><tool><TAB><json>` lines, where `<json>` is the merged values of
   that tool, and returns 0; or it logs each message and returns 2 or 3.
   - **Schema checks** of each selected tool's `tools.toml` (code 2, the
     message names `<tools root>/<tool>/tools.toml`): `[defaults]` is a table;
     `[questions]` is a table of tables; a question has a `question` string;
     `detect`, when present, is a string; a question key other than
     `question`, `detect` and `example` is an error; `[contributes]` is a
     table of tables. Port the message texts of B's `_SS_MANIFEST_TPL` (455).
   - **Contributions:** only a selected tool contributes. A contribution to a
     tool that is not in the tools root: code 3, with the message of D10. A
     contribution to a tool of the root that is not selected: ignored.
   - **Merge:** port `_SS_VALUES_JQ` (749-~829) with its functions and its
     messages: `[defaults]`, then the contributions of the selected tools in
     tool-name order (`contributors_merge`: two different scalars → code 2;
     lists append new items), then the user values with `user_merge` (D5).
     The message texts that name `<tool>.toml` or `[overrides.<tool>]` name
     `[tools.<tool>] of <values file>` instead. Build the jq input with one
     gomplate call, as `_SS_INPUTS_TPL` does, with the user values from the
     front-end.
   - **Codes:** the library logs every message and returns the highest code
     (D10).
2. **`sync-engine/repo-sync`**:
   - Source `sync-engine/values` where it sources the other libraries.
   - `_rs_values_sel`: a key other than `env` in a `[tools.*]` entry is now
     valid (D6, D9). `env`, when present, must still be a list. A
     `[packs.*]` entry keeps `env` as its only key. Change the message text of
     the `[packs.*]` error to say "the only key of a pack entry is env, a
     list". `--check` uses this function, so it gets the change too.
   - After the selection and the depends closure, and before `_rs_context`,
     build the user values (each `[tools.<name>]` table of the values file
     without `env`) and call the entry function. The base and dependency set
     is every tool that pack `base` adds (nested `default = true` packs
     included) plus each tool of the depends closure (mise `depends`
     included), also when a `[tools.<name>]` table names it (D5). On code 2 or
     3, stop the sync with that code before any file is
     written, as the other validation steps do.
   - `_rs_context` adds `all` (D8). The engine keys win over the values file.
   - For each selected tool that has a `.tmpl` file, write a context file:
     the shared context plus `config` = the values of that tool. The
     template render of that tool reads it. Remove the context files with
     the other temp files.
   - Update the header comment of the script: rewrite the passage "the only
     key of an entry is `env`" (a `[tools.*]` entry selects with `env` and
     gives values in its other keys; a `[packs.*]` entry has only `env`), and
     the template context names `.config` and `.all`. Keep it short; P15
     rewrites the headers.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `bash -n sync-engine/values sync-engine/repo-sync` passes.
- `grep -n 'contributors_merge\|keep_merge' sync-engine/values` finds the
  ported functions.
- `grep -n 'jq@\|gomplate@' sync-engine/values` finds nothing (the calls go
  through the wrappers of `sync-engine/common`).
- `mise x -- mise run test:repo-sync` passes, except the two "a key other
  than env" `--check` case (test lines 696-698 today) whose expected result
  D6 changes (U2 changes it); name it in a `DECIDED:` line.

## Guardrails

- Do not touch `.config/mise/tasks/test/repo-sync`; U2 owns it.
- Do not move a top-level key of `bootstrap.toml`.
- Write shell for bash 3.2 and bash 5.x: no namerefs, no `${var,,}`, no
  associative arrays. Under `set -e`, write a command substitution that can
  fail as `VAR=$(cmd) || VAR=''`.
- Call jq and gomplate only through the wrappers of `sync-engine/common`.
- Delete with `rm`, never `git rm`.

## Commit

`feat: merge tool values from defaults, contributions and [tools.<name>]` —
written by the orchestrator after the unit gate and the wave gate, not by the
unit.
