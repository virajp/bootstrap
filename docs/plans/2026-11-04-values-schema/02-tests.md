# U2 — Tests: values, questions and contributions

- **Read first:** every owned file, top to bottom, before editing.
- **Lazy-load:** `sync-engine/values`, `sync-engine/repo-sync` (U1's output).

## Ruling

From the design rulings of 2026-11-04:

> - Values row, item 3: the shared `tools.toml` schema accepts `[defaults]`,
>   `[questions]` (`question`, `detect`, `example`; the guide asks, the
>   engine only validates) and `[contributes.<tool>]` on both sides. Merge
>   order: defaults, then contributions, then `[tools.<name>]`. The repo side
>   moves none of its existing top-level keys (`commit_scopes`, `extensions`,
>   `[claude.plugins.*]`, `[branch.*]`); a new repo tool value goes into
>   `[tools.<name>]`. Repo questions stay in `docs/sync/tools/<tool>.md`; a
>   new repo tool may use `[questions]`. `[[precommit]]` is a special-name
>   contribution; a later move to `[contributes]` is possible.

D3: "This plan extends test:repo-sync with cases for its new rules. An
existing case changes only where a ruling changes its expected result."

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

1. **`.config/mise/tasks/test/repo-sync`** — add change functions that give a
   fixture tool of the scratch clone (for example a new folder
   `$clone/repo/tools/vtest/` with a `tools.toml` and a template
   `vtest.txt.tmpl` that prints `.config` and `.all` values; commit the
   change in the scratch clone if the engine reads committed files only) and
   that make the consumer file, so the dry run shows a diff. Add cases:
   1. `[defaults]` only: the output has the default.
   2. `[tools.vtest] key = "x"` wins over the default; `env` is not in
      `.config` (D6).
   3. A second fixture tool contributes a list item and a scalar: the list
      has the new item; the scalar wins over the default.
   4. Two tools contribute different scalars to one key: code 2 with the
      message that names both `tools.toml` files.
   5. A contribution to a tool that is not selected: code 0, no effect.
   6. A contribution to a tool that bootstrap does not have: code 3.
   7. Base or dependency tool (D5): a list in `[tools.<name>]` appends; a
      scalar over a list: code 2.
   8. A tool that is not base or dependency: a list in `[tools.<name>]`
      replaces.
   9. `[questions.q]` without `question`: code 2; with a key `foo`: code 2;
      with `detect = 1`: code 2; a valid question: code 0.
   10. `.all` holds the values of another selected tool.
   11. `[packs.<name>]` with a key other than `env`: code 2 (D6).
   12. A sync case: `[tools.<name>]` with a key other than `env`: code 0
       (D6).
   13. A base tool through the nested pack `"pack:security"` (for example
       `gitleaks`) gets the list-append rule; `_security_base` through
       `depends` too (D5).
   14. Two errors in different tools (for example a bad question and a
       contribution to an unknown tool): both messages, code 3 (D10).
2. Change the one existing `--check` case (test lines 696-698 today) that
   expects code 2 for a key other than `env` in a `[tools.*]` entry: it now
   expects code 0 (D6, D9). No other existing case changes its expected result
   (D3).

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes, every old and new case.

## Guardrails

- Do not touch `sync-engine/**` or `repo/**` of the working tree; a fixture
  tool lives only in the scratch clone of a case. Report an engine bug as
  `UNRESOLVED:` with the case.
- Write shell for bash 3.2 and bash 5.x. Under `set -e`, write a command
  substitution that can fail as `VAR=$(cmd) || VAR=''`.
- Delete with `rm`, never `git rm`.

## Commit

`test: cover tool values, questions and contributions` — written by the
orchestrator after the unit gate and the wave gate, not by the unit.
