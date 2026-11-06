# U3 — Tests for the merged packs model

- **Read first:** `.config/mise/tasks/test/repo-sync`, top to bottom, before
  editing.
- **Lazy-load:** `sync-engine/select`, `tools/packs.toml`, the three scanner
  `tools.toml` files (the output of U1 and U2).

## Ruling

From the design rulings of 2026-11-04:

> - One merged packs model for both sides (user's proposal, 2026-11-04):
>   - A tool can be in two packs; the engine dedupes. The guide shows it once.
>   - `packs.toml` shape (both sides): each key of `[pack.<name>]` is a member,
>     `<tool> = { default = true }` or `<tool> = {}` (implicit `false`). A
>     nested pack is a quoted `"pack:<name>"` key (not `p:`). The only entry
>     key is `default`; any other stops with exit 2. No pack-level keys (no
>     `depends`, `tools`, `default = [...]`, `required`, `description`).
>   - `[packs.<name>]` adds each `default = true` member (option A). A new
>     `default = true` member reaches every consumer of the pack; a
>     `default = false` member arrives only with `[tools.<name>]`. Still open,
>     as in `docs/design/packs.md`: no way to drop one default member.
>   - `required` goes: a common tool (`_security_base`) becomes a `depends` of
>     each member tool (gitleaks, grype, osv-scanner), so a single-tool
>     selection is correct. Repo consumers get a `migrate.md` item for it.

From this plan:

D3: "This plan extends test:repo-sync with cases for its new rules. A current
case changes only where a ruling changes its expected result: the cases that
write a [pack.*] table in the old shape now write the new shape."

D5: "Every [pack.*] table of packs.toml, expanded or not: the shape of each
entry, the old keys, a pack named like a tool folder, a loop over every
\"pack:<name>\" key (default true or not), and an unknown member (code 3).
Thus the test of bootstrap finds an error in its own packs.toml before a
consumer selects that pack."

D6: "A key depends, required, tools or default whose value is a list:
'tools/packs.toml: [pack.<name>] <key> is the old format - write each member
as <tool> = { default = true } or <tool> = {}' (code 2). A key that starts
with p:: '... \"p:<x>\" is the old form - write \"pack:<x>\" = { default =
true }' (code 2). Any other value that is not a table: '... <key> must be a
table, { default = true } or {}' (code 2). An entry key other than default, or
a default that is not a boolean: code 2 with the key in the message. The
message of an unknown member names the pack (code 3)."

D7: "\"pack:<x>\" = {} does not expand when [packs.<outer>] adds the outer
pack; the nested pack arrives only through its own [packs.<x>] entry. The loop
check still reads it (D5)."

D9: "In no pack. gitleaks, grype and osv-scanner each get depends =
[\"_security_base\"], so a pack selection and a single [tools.<scanner>] entry
both bring it, with the environments of the scanners that need it."

D11: "In file order, so the tool order and the pre-commit hook order of today
stay. An awk scan of each [pack.<name>] section reads the keys in file order
(one <key> = {...} line per member; a member that spans lines stops with code
2). gomplate still validates the shape. tools/packs.toml lists the base
members in the order of today: mise, repo-sync, git, pre-commit, dprint,
linter, vscode, then \"pack:security\"; security lists gitleaks, grype,
osv-scanner. A test proves that the tool order and the pre-commit hook order
do not change."

## Edits

1. **`.config/mise/tasks/test/repo-sync` — the current cases.**
   - `pack_loop` and `pack_named_like_tool` (:187-194) append packs in the
     new shape (`"pack:b" = { default = true }`, `node = { default = true }`).
     The cases keep their codes and messages.
   - The comment above `base_tools` (:157) says that `_security_base`
     arrives through the `depends` of the scanners. The list stays.
2. **New cases** (a change function appends to `$clone/tools/packs.toml`,
   or writes a test pack):
   - each old key (`depends = [...]`, `required = [...]`, `tools = [...]`,
     `default = [...]`) in a pack that no values file selects: 2 with the
     old-format message (D5, D6);
   - a `"p:x" = { default = true }` key: 2, the message names `"pack:x"`;
   - an entry key other than `default`; `default = "yes"`; a member value
     that is a string: 2 each;
   - a loop through `"pack:x" = {}` members only: 2 (D5, D7);
   - an unknown member in an unselected pack: 3, the message names the pack;
   - a pack `extra` with `node = {}` and `python = { default = true }`:
     `[packs.extra]` selects `python` and not `node`; `[packs.extra]` with
     `[tools.node]` selects both;
   - a nested `"pack:security" = {}` in a pack `outer`: `[packs.outer]`
     does not select `gitleaks`;
   - a tool in two selected packs: the dry run writes its files once and the
     run exits 0;
   - `[tools.gitleaks]` with a test pack `base` that lacks `"pack:security"`
     (a change function that rewrites the base pack of the scratch clone):
     the output syncs a file of `_security_base`.
   - "the tool order and the hook order do not change" (D11): a dry run of
     a values file with only `environments = [ "dev" ]` lists the tools of
     the pack base in the order of today (mise, repo-sync, git, pre-commit,
     dprint, linter, vscode, then the security tools), and the rendered
     `.config/pre-commit-config.yaml` has its `repos:` entries in the order
     of today. Take the expected order from the current tree before U1 and
     U2 land, for example from `git show <branch base>:` files or a run of
     the test at the branch base.
   - a member on two lines: 2 (D11).
3. Every other current case passes unchanged.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes: every case.
- `grep -n "depends = \[ \"p:" .config/mise/tasks/test/repo-sync` finds
  only the old-key cases.

## Guardrails

- Do not touch `sync-engine/**` or `tools/**`; U1 and U2 own them. Report a
  defect in them as a `GAP:` or an `UNRESOLVED:`.
- A test pack name must not be a tool folder, except in the case that tests
  that rule.
- The test needs no network beyond what `mise x <tool>@latest` needs, and
  writes nothing outside its temp folder.
- Write shell for bash 3.2 and bash 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`test: cover the merged packs model` — written by the orchestrator after the
unit gate and the wave gate, not by the unit.
