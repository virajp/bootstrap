# U3 — Tests for the [files] tables, the old keys and merge mode

- **Read first:** `.config/mise/tasks/test/repo-sync`, top to bottom, before
  editing.
- **Lazy-load:** `sync-engine/merge`, `sync-engine/repo-sync`,
  `tools/python/tools.toml`, `tools/node/tools.toml`,
  `MS/.config/mise/tasks/test/setup-sync` (read only: the merge cases
  :923-1347, their fixtures and their expected results).

## Ruling

From the design rulings of 2026-11-04:

> - Seeds on both sides use the setup-sync form: `[files."<target>"]` (target
>   path, no `.tmpl`), `mode = "seed"`, `requires = [...]` (exact paths
>   relative to the target root, each must exist). `seed = [...]` and
>   `[seed_requires]` go (exit 2 naming the new key). No `[files]` table =
>   `mode = "generated"`. The repo mise-entry `requires` (a file name searched
>   4 folders deep) is a second meaning of the word: solve it in the "mise
>   config" row.

> - Merge mode (`mode = "merge"`, JSON/YAML/TOML, `remove`, `remove_items`)
>   moves into the shared engine and is available to repo tools. No repo
>   tool uses it until a real need (candidates: `.vscode/settings.json`,
>   `.claude/settings.json`). Repo side order: merge, dprint format, compare.
>   `[overrides.gitignore]` stays as it is. Merge mode = tool author declares
>   enforced keys; overrides = consumer changes defaults; two mechanisms. The
>   three override shapes (`[overrides.gitignore]`, `[overrides.mise*]`, Mac
>   `[overrides.<tool>]`) get aligned in the "Values" row.

> - `requires` has one meaning: exact paths, in `[files]` tables. New key
>   `requires_found` = a list of file names searched down to 4 folders below
>   the root, valid in a mise entry and in a `[files]` table, repo side only
>   (Mac → exit 2). The `files` list of a mise entry goes: each file states
>   its condition in its own `[files]` table (python, node).

From this plan:

D3: "This plan extends test:repo-sync with cases for its new rules. A current
case changes only where a ruling changes its expected result: the cases that
write seed, [seed_requires] or a mise files list now write the new keys."

D7: "It does not write the target. A generated target keeps the behavior of
the mise files list of today: a copy with the bootstrap header is removed (dry
run: would remove <path> (no <file>)), else the log says skip file <path> (no
<file>). A seed logs skip seed <path> (no <file>). A merge target logs skip
merge <path> (no <file>) and stays as it is. <file> names the first entry that
is not there, in the order of the lists: requires first, then requires_found.
Plan 5 replaces the removal with the manifest."

D8: "Every tool of the clone, selected or not, in the same pass that reads all
manifests today (_rs_mise_index). The rules: a known key in each [files]
table, a known mode and format, a target that a file of the tool produces
(<target> or <target>.tmpl), no tool folder with both <p> and <p>.tmpl, merge
only with json, yaml or toml, remove and remove_items only with merge and of
the right shape, the old keys. Thus the test of bootstrap finds an error
before a consumer selects the tool."

D9: "requires: a list of exact paths relative to the repository root
(MISE_PROJECT_ROOT); each must exist (-e); an absolute path or a .. part stops
with code 2. requires_found: a list of file names; each must be found by the
rule of `_rs_has_file` (git tracks it or does not ignore it, at the root or down
to 4 folders below it, no submodule); a name with a / stops with code 2. Both
lists must be non-empty when present. A target with both lists needs every
entry of both."

D10: "sync-engine/merge holds a port of `_SS_MERGE_JQ` and `_ss_merge` with the
rules of the setup-sync header (Merge files). The repo order is render, parse
check, merge with the live target, dprint, compare. The repo side follows no
symlink: the target path is the path in the repository. A merge target has no
header, is never removed by the sync, and keeps the mode of an existing target
(a new one takes the mode of its source). The format of a merge target is its
format key, else its extension; the JSONC rule of plan 2 does not apply to
it."

D11: "One message, code 2, per old key per tools.toml (an entry with an old
requires and an old files gives two): '<tools root>/<tool>/tools.toml: seed is
not supported; use [files.\"<target>\"] with mode = \"seed\"'; '<tools
root>/<tool>/tools.toml: [seed_requires] is not supported; use requires in
[files.\"<target>\"]'; '<tools root>/<tool>/tools.toml: files in
[mise.tools.<x>] is not supported; use requires_found in
[files.\"<target>\"]'; and for requires = \"<name>\" in a [mise.tools.<x>]
entry: '<tool>/tools.toml: requires of mise tool <x> is the old key - write
requires_found = [\"<name>\"]'."

D12: "The generated mise.<env>.toml drops requires_found, as it drops
requires today. In a [mise_env.<env>.tools] entry, requires_found, requires and
files stop the sync with code 2 and the message says that the key is valid
only in [mise.tools]."

D13: "No. [overrides.mise.tools.<x>.requires_found] is locked like envs,
depends and requires, and the lock message names requires_found."

D14: "No. A merge source carries no two-line bootstrap header, because the
app owns the target. A YAML or TOML merge source may carry the comment #
Merged by bootstrap: <path>; a JSON merge source carries no comment. The
engine strips nothing. The test fixtures follow this. The docs unit rewrites
the header rule of CLAUDE.md to: each file that is neither a seed nor a merge
source."

## Edits

1. **`.config/mise/tasks/test/repo-sync` — the current cases.**
   - The python cases (about :406-470) that edit `files = [...]` or
     `requires = "..."` of `tools/python/tools.toml` now edit the
     `[files."<path>"]` tables and `requires_found`. Their expected codes and
     messages change only where D7 or D11 changes them.
   - The node seed cases now use the `requires` of
     `[files."pnpm-workspace.yaml"]`.
   - The fixture tool of plan 2 (`tools/fixture/`) loses its `seed = []`
     line.
2. **Old-key cases** (each exit 2 with the exact message of D11): a
   top-level `seed = [...]`; a `[seed_requires]` table; `requires = "x"` in a
   mise entry; `files = [...]` in a mise entry; one entry with both old keys
   gives two messages. Use the fixture tool, not selected, to show that D8
   checks every tool.
   - `requires_found` in a `[mise_env.dev.tools]` entry: 2 (D12); the
     generated `mise.dev.toml` of a tool with `requires_found` has no
     `requires_found` line (D12).
   - `[overrides.mise.tools.<x>.requires_found]`: 2, the lock message names
     `requires_found` (D13).
3. **`[files]` cases:**
   - an unknown key in a `[files]` table: 2;
   - a target that no file of the tool produces: 2;
   - `<p>` and `<p>.tmpl` in one tool folder: 2;
   - an unknown `mode`, an unknown `format`: 2;
   - `requires` with `..` or an absolute path: 2; `requires_found` with a
     `/`: 2;
   - a seed that exists: `kept seed`; a seed whose `requires` path is absent:
     `skip seed`; present: written;
   - a generated target with `requires_found` found (in a folder 3 levels
     down) and not found (5 levels down): written, then `skip file` or
     `would remove` for a copy with the header (D7);
   - a target with `requires` and `requires_found`, one of them absent: not
     written; with both absent, `(no <file>)` names the `requires` entry
     (D7).
4. **Merge cases** with the fixture tool, ported from
   `MS/.config/mise/tasks/test/setup-sync` and adapted to the repo side (no
   symlink, no contribution): `case_merge_json`, `case_merge_yaml` (comments
   stay), `case_merge_toml` (comments stay), `case_merge_absent`,
   `case_merge_empty`, `case_merge_live_bad`, `case_merge_live_kind`,
   `case_merge_live_docs`, `case_merge_src_bad`, `case_merge_src_empty`, and
   the four config errors (merge with `format` text, `remove` on a
   generated file, a `remove` of a bad shape, a `remove_items` without
   `items`). Each merge source follows D14: no bootstrap header; a YAML or
   TOML source may carry `# Merged by bootstrap: <path>`; a JSON source has
   no comment. Run the cases that compare a written result as real runs
   (`case_args=' '`). Add one case that a merge target keeps its mode (an
   existing `0600` target stays `0600`).
5. Every other current case passes unchanged.

## Verification

- `mise x -- mise run code:format` and `mise x -- mise run code:lint` pass.
- `mise x -- mise run test:repo-sync` passes: every case.
- `grep -n 'seed = \[\|seed_requires\|requires = "'
  .config/mise/tasks/test/repo-sync`
  finds only the old-key cases.

## Guardrails

- Do not touch `sync-engine/**` or `tools/**`; U1 and U2 own them. Report a
  defect in them as a `GAP:` or an `UNRESOLVED:`.
- The test needs no network beyond what `mise x <tool>@latest` needs, and
  writes nothing outside its temp folder.
- Write shell for bash 3.2 and bash 5.x.
- Delete with `rm`, never `git rm`.

## Commit

`test: cover the [files] tables, the old keys and merge mode` — written by
the orchestrator after the unit gate and the wave gate, not by the unit.
